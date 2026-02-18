package services

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

// AlertService decides whether to create an alert and persists it
type AlertService struct {
	db          *gorm.DB
	aiClient    *AIClient
	emailSvc    *EmailService
	prefSvc     *PreferenceService
}

func NewAlertService(db *gorm.DB) *AlertService {
	return &AlertService{
		db:       db,
		aiClient: NewAIClient(),
		emailSvc: NewEmailService(),
		prefSvc:  NewPreferenceService(db),
	}
}

// severityFromChange computes severity based on change type and magnitude
func severityFromChange(changeType string, changePercent float64) models.AlertSeverity {
	abs := math.Abs(changePercent)

	switch {
	case strings.Contains(changeType, "price_increase") || strings.Contains(changeType, "price_decrease"):
		switch {
		case abs >= 30:
			return models.SeverityCritical
		case abs >= 15:
			return models.SeverityHigh
		case abs >= 5:
			return models.SeverityMedium
		default:
			return models.SeverityLow
		}
	case strings.Contains(changeType, "feature_added") || strings.Contains(changeType, "feature_removed"):
		return models.SeverityHigh
	case strings.Contains(changeType, "messaging_change"):
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

// shouldAlert returns true if the change warrants an alert
func shouldAlert(changeType string, severity models.AlertSeverity, minSeverity string) bool {
	// Always alert on price changes
	if strings.Contains(changeType, "price_increase") || strings.Contains(changeType, "price_decrease") {
		return true
	}
	// Always alert on feature changes
	if strings.Contains(changeType, "feature_added") || strings.Contains(changeType, "feature_removed") {
		return true
	}

	// For other types, check severity threshold
	order := map[models.AlertSeverity]int{
		models.SeverityLow:      1,
		models.SeverityMedium:   2,
		models.SeverityHigh:     3,
		models.SeverityCritical: 4,
	}
	minOrder := order[models.AlertSeverity(minSeverity)]
	if minOrder == 0 {
		minOrder = 2 // default: medium
	}
	return order[severity] >= minOrder
}

// ProcessChange evaluates a detected change, creates an alert if needed, and notifies
func (s *AlertService) ProcessChange(change *models.DetectedChange) error {
	// 1. Compute severity
	severity := severityFromChange(change.ChangeType, change.ChangePercent)

	// 2. Get user preferences
	prefs, err := s.prefSvc.GetPreferencesForPage(change.PageID)
	if err != nil {
		log.Printf("⚠️  AlertService: preference error for page %d: %v", change.PageID, err)
		prefs = &AlertPreference{MinSeverity: "medium"}
	}

	// 3. Decide if alert is needed
	if !shouldAlert(change.ChangeType, severity, prefs.MinSeverity) {
		log.Printf("ℹ️  AlertService: change %d skipped (severity=%s below threshold=%s)", change.ID, severity, prefs.MinSeverity)
		return nil
	}

	// 4. Enrich with AI
	insight, err := s.aiClient.Analyze(
		change.ChangeType,
		change.PageType,
		change.OldPrice,
		change.NewPrice,
		change.ChangePercent,
		change.FeaturesAdded,
		change.FeaturesRemoved,
		change.OldText,
		change.NewText,
	)
	if err != nil || insight == nil {
		insight = &AIInsight{
			Summary:        "Competitor change detected: " + change.ChangeType,
			Recommendation: "Review competitor activity",
			Model:          "rule-based",
		}
	}

	// 5. Persist alert log
	now := time.Now()
	alert := &models.AlertLog{
		ChangeID:       change.ID,
		PageID:         change.PageID,
		AlertType:      change.ChangeType,
		Severity:       severity,
		Summary:        insight.Summary,
		Recommendation: insight.Recommendation,
		Notified:       false,
		NotifyChannel:  "log",
	}

	if err := s.db.Create(alert).Error; err != nil {
		log.Printf("❌ AlertService: failed to save alert for change %d: %v", change.ID, err)
		return err
	}

	log.Printf("🚨 Alert created [%s] change=%d page=%d severity=%s | %s",
		change.ChangeType, change.ID, change.PageID, severity, insight.Summary)

	// 6. Send notifications
	notified := false

	if prefs.NotifyEmail && prefs.UserEmail != "" {
		if err := s.emailSvc.SendAlert(prefs.UserEmail, change.ChangeType, string(severity), insight.Summary, insight.Recommendation, change.PageID); err != nil {
			log.Printf("⚠️  Email notification failed: %v", err)
		} else {
			notified = true
			alert.NotifyChannel = "email"
		}
	}

	if prefs.NotifyWebhook && prefs.WebhookURL != "" {
		if err := s.emailSvc.SendWebhook(prefs.WebhookURL, change.ChangeType, string(severity), insight.Summary, insight.Recommendation, change.PageID, alert.ID); err != nil {
			log.Printf("⚠️  Webhook notification failed: %v", err)
		} else {
			notified = true
			if alert.NotifyChannel == "email" {
				alert.NotifyChannel = "email,webhook"
			} else {
				alert.NotifyChannel = "webhook"
			}
		}
	}

	// 7. Mark as notified
	if notified {
		s.db.Model(alert).Updates(map[string]interface{}{
			"notified":       true,
			"notified_at":    now,
			"notify_channel": alert.NotifyChannel,
		})
	}

	return nil
}

// GetUnprocessedChanges returns detected_changes that have no alert_log yet
func (s *AlertService) GetUnprocessedChanges() ([]models.DetectedChange, error) {
	var changes []models.DetectedChange
	err := s.db.Raw(`
		SELECT dc.*
		FROM detected_changes dc
		LEFT JOIN alert_logs al ON dc.id = al.change_id
		WHERE al.id IS NULL
		ORDER BY dc.detected_at ASC
		LIMIT 50
	`).Scan(&changes).Error
	return changes, err
}
