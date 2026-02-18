package services

import (
	"log"
	"os"

	"gorm.io/gorm"
)

// AlertPreference holds user notification preferences for a project
type AlertPreference struct {
	ProjectID      uint
	UserEmail      string
	WebhookURL     string
	NotifyEmail    bool
	NotifyWebhook  bool
	MinSeverity    string // low, medium, high, critical
}

// PreferenceService retrieves alert preferences for a project/user
type PreferenceService struct {
	db *gorm.DB
}

func NewPreferenceService(db *gorm.DB) *PreferenceService {
	return &PreferenceService{db: db}
}

// GetPreferencesForPage returns alert preferences for the project owning a page
// For now returns defaults from env vars; can be extended to DB-backed preferences
func (s *PreferenceService) GetPreferencesForPage(pageID int) (*AlertPreference, error) {
	// Try to resolve project from page → competitor → project
	type row struct {
		ProjectID uint
		Email     string
	}
	var r row
	err := s.db.Raw(`
		SELECT p.id AS project_id, u.email
		FROM monitored_pages mp
		JOIN competitors c ON mp.competitor_id = c.id
		JOIN projects p ON c.project_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE mp.id = ?
		LIMIT 1
	`, pageID).Scan(&r).Error

	if err != nil || r.ProjectID == 0 {
		log.Printf("⚠️  PreferenceService: could not resolve project for page %d, using defaults", pageID)
		return s.defaultPreferences(), nil
	}

	return &AlertPreference{
		ProjectID:     r.ProjectID,
		UserEmail:     r.Email,
		WebhookURL:    os.Getenv("ALERT_WEBHOOK_URL"),
		NotifyEmail:   true,
		NotifyWebhook: os.Getenv("ALERT_WEBHOOK_URL") != "",
		MinSeverity:   "medium",
	}, nil
}

func (s *PreferenceService) defaultPreferences() *AlertPreference {
	return &AlertPreference{
		UserEmail:     os.Getenv("ALERT_DEFAULT_EMAIL"),
		WebhookURL:    os.Getenv("ALERT_WEBHOOK_URL"),
		NotifyEmail:   os.Getenv("ALERT_DEFAULT_EMAIL") != "",
		NotifyWebhook: os.Getenv("ALERT_WEBHOOK_URL") != "",
		MinSeverity:   "medium",
	}
}
