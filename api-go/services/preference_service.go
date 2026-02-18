package services

import (
	"log"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

// PreferenceService loads user notification settings from DB
type PreferenceService struct {
	db *gorm.DB
}

func NewPreferenceService(db *gorm.DB) *PreferenceService {
	return &PreferenceService{db: db}
}

// GetSettingsForPage resolves the user owning the page and returns their notification settings.
// Chain: monitored_page → competitor → project → user → user_notification_settings
func (s *PreferenceService) GetSettingsForPage(pageID int) (*models.UserNotificationSettings, string, error) {
	// Resolve user email and user_id from page
	type row struct {
		UserID uint
		Email  string
	}
	var r row
	err := s.db.Raw(`
		SELECT u.id AS user_id, u.email
		FROM monitored_pages mp
		JOIN competitors c ON mp.competitor_id = c.id
		JOIN projects p ON c.project_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE mp.id = ?
		LIMIT 1
	`, pageID).Scan(&r).Error

	if err != nil || r.UserID == 0 {
		log.Printf("⚠️  PreferenceService: cannot resolve user for page %d, using defaults", pageID)
		return s.defaultSettings(), "", nil
	}

	// Load notification settings for this user
	var settings models.UserNotificationSettings
	err = s.db.Where("user_id = ?", r.UserID).First(&settings).Error
	if err != nil {
		// No settings row yet → create defaults
		log.Printf("ℹ️  PreferenceService: no settings for user %d, creating defaults", r.UserID)
		settings = models.UserNotificationSettings{
			UserID:               r.UserID,
			NotifyEmail:          true,
			NotifyWebhook:        false,
			MinimumChangePercent: 5.0,
			AlertOnPriceChange:   true,
			AlertOnFeatureChange: true,
			AlertOnMessaging:     false,
		}
		if createErr := s.db.Create(&settings).Error; createErr != nil {
			log.Printf("⚠️  PreferenceService: failed to create default settings: %v", createErr)
		}
	}

	return &settings, r.Email, nil
}

// defaultSettings returns safe defaults when user cannot be resolved
func (s *PreferenceService) defaultSettings() *models.UserNotificationSettings {
	return &models.UserNotificationSettings{
		NotifyEmail:          false,
		NotifyWebhook:        false,
		MinimumChangePercent: 5.0,
		AlertOnPriceChange:   true,
		AlertOnFeatureChange: true,
		AlertOnMessaging:     false,
	}
}
