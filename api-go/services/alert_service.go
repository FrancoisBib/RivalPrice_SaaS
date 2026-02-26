package services

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

// AlertService decides whether to create an alert and persists it.
// Decision is 100% deterministic based on user_notification_settings.
type AlertService struct {
	db       *gorm.DB
	aiClient *AIClient
	emailSvc *EmailService
	prefSvc  *PreferenceService
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

// shouldAlert applies deterministic rules from user_notification_settings
func shouldAlert(change *models.DetectedChange, settings *models.UserNotificationSettings) (bool, string) {
	changeType := change.ChangeType
	abs := math.Abs(change.ChangePercent)

	// Price changes
	if strings.Contains(changeType, "price_increase") || strings.Contains(changeType, "price_decrease") {
		if !settings.AlertOnPriceChange {
			return false, "price alerts disabled by user"
		}
		if abs < settings.MinimumChangePercent {
			return false, "change_percent below minimum_change_percent threshold"
		}
		return true, ""
	}

	// Feature changes
	if strings.Contains(changeType, "feature_added") || strings.Contains(changeType, "feature_removed") {
		if !settings.AlertOnFeatureChange {
			return false, "feature alerts disabled by user"
		}
		return true, ""
	}

	// Messaging changes
	if strings.Contains(changeType, "messaging_change") {
		if !settings.AlertOnMessaging {
			return false, "messaging alerts disabled by user"
		}
		return true, ""
	}

	return false, "unknown change type"
}

// buildAlertMessage assembles factual data + AI summary + recommendation
func buildAlertMessage(change *models.DetectedChange, insight *AIInsight) string {
	var msg strings.Builder

	// Section 1: Factual data (always present)
	msg.WriteString("=== DONNÉES FACTUELLES ===\n")
	if change.OldPrice != "" && change.NewPrice != "" {
		msg.WriteString(fmt.Sprintf("Prix: %s → %s (%.1f%%)\n", change.OldPrice, change.NewPrice, change.ChangePercent))
	}
	msg.WriteString(fmt.Sprintf("Type de changement: %s\n", change.ChangeType))
	msg.WriteString(fmt.Sprintf("Page ID: %d\n", change.PageID))
	msg.WriteString("\n")

	// Section 2: AI Summary
	if insight.Summary != "" {
		msg.WriteString("=== ANALYSE IA ===\n")
		msg.WriteString(insight.Summary + "\n")
		msg.WriteString("\n")
	}

	// Section 3: AI Recommendation
	if insight.Recommendation != "" {
		msg.WriteString("=== RECOMMANDATION ===\n")
		msg.WriteString(insight.Recommendation + "\n")
	}

	return msg.String()
}

// ProcessChange evaluates a detected change using deterministic rules, then notifies
func (s *AlertService) ProcessChange(change *models.DetectedChange) error {
	// 1. Resolve page → user → notification settings
	settings, userEmail, err := s.prefSvc.GetSettingsForPage(change.PageID)
	if err != nil {
		log.Printf("⚠️  AlertService: preference error for page %d: %v", change.PageID, err)
		settings = s.prefSvc.defaultSettings()
	}

	// 2. Deterministic decision
	ok, reason := shouldAlert(change, settings)
	if !ok {
		log.Printf("ℹ️  AlertService: change %d skipped — %s", change.ID, reason)
		return nil
	}

	// 3. Compute severity
	severity := severityFromChange(change.ChangeType, change.ChangePercent)

	// 4. Enrich with AI (summary + recommendation + impact_level)
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
			Model:         "rule-based",
			ImpactLevel:   5,
		}
	}

	// 5. Build full alert message (factual + AI)
	message := buildAlertMessage(change, insight)

	// 6. Persist alert log
	now := time.Now()
	alert := &models.AlertLog{
		ChangeID:        change.ID,
		PageID:          change.PageID,
		AlertType:       change.ChangeType,
		Severity:        severity,
		OldPrice:        change.OldPrice,
		NewPrice:        change.NewPrice,
		ChangePercent:   change.ChangePercent,
		AISummary:       insight.Summary,
		AIRecommendation: insight.Recommendation,
		ImpactLevel:     insight.ImpactLevel,
		AIModel:         insight.Model,
		Message:         message,
		Notified:        false,
		NotifyChannel:   "log",
	}

	if err := s.db.Create(alert).Error; err != nil {
		log.Printf("❌ AlertService: failed to save alert for change %d: %v", change.ID, err)
		return err
	}

	log.Printf("🚨 Alert created [%s] change=%d page=%d severity=%s impact=%d | %s",
		change.ChangeType, change.ID, change.PageID, severity, insight.ImpactLevel, insight.Summary)

	// 7. Send notifications
	notified := false
	channel := "log"

	if settings.NotifyEmail && userEmail != "" {
		if err := s.emailSvc.SendAlert(userEmail, change.ChangeType, string(severity), insight.Summary, insight.Recommendation, change.PageID); err != nil {
			log.Printf("⚠️  Email notification failed: %v", err)
		} else {
			notified = true
			channel = "email"
		}
	}

	if settings.NotifyWebhook && settings.WebhookURL != "" {
		if err := s.emailSvc.SendWebhook(settings.WebhookURL, change.ChangeType, string(severity), message, insight.Recommendation, change.PageID, alert.ID); err != nil {
			log.Printf("⚠️  Webhook notification failed: %v", err)
		} else {
			notified = true
			if channel == "email" {
				channel = "email,webhook"
			} else {
				channel = "webhook"
			}
		}
	}

	// 8. Mark as notified
	if notified {
		s.db.Model(alert).Updates(map[string]interface{}{
			"notified":        true,
			"notified_at":     now,
			"notify_channel":   channel,
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
