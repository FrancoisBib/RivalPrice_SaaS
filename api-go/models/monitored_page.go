package models

import "time"

type PageType string

const (
	PageTypePricing  PageType = "pricing"
	PageTypeFeatures PageType = "features"
)

type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

type MonitoredPage struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	CompetitorID   uint       `gorm:"not null;index" json:"competitor_id"`
	PageType       string     `gorm:"type:varchar(50);not null" json:"page_type"` // pricing / features
	URL            string     `gorm:"type:varchar(512);not null" json:"url"`
	CSSSelector    string     `gorm:"type:text" json:"css_selector"` // optional for specific targeting
	Frequency      Frequency  `gorm:"type:varchar(20);not null;default:daily" json:"frequency"`
	NextRunAt      time.Time  `gorm:"not null;index" json:"next_run_at"`
	LastCheckedAt  *time.Time `json:"last_checked_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Competitor     Competitor `gorm:"foreignKey:CompetitorID" json:"competitor,omitempty"`
}

func (MonitoredPage) TableName() string {
	return "monitored_pages"
}
