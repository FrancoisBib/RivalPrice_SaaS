package models

import "time"

// UserNotificationSettings stores per-user alert preferences
type UserNotificationSettings struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID               uint      `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	NotifyEmail          bool      `gorm:"column:notify_email;default:true" json:"notify_email"`
	NotifyWebhook        bool      `gorm:"column:notify_webhook;default:false" json:"notify_webhook"`
	WebhookURL           string    `gorm:"column:webhook_url;type:varchar(512)" json:"webhook_url"`
	MinimumChangePercent float64   `gorm:"column:minimum_change_percent;default:5" json:"minimum_change_percent"` // alert if |change_percent| >= this
	AlertOnPriceChange   bool      `gorm:"column:alert_on_price_change;default:true" json:"alert_on_price_change"`
	AlertOnFeatureChange bool      `gorm:"column:alert_on_feature_change;default:true" json:"alert_on_feature_change"`
	AlertOnMessaging     bool      `gorm:"column:alert_on_messaging;default:false" json:"alert_on_messaging"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserNotificationSettings) TableName() string {
	return "user_notification_settings"
}
