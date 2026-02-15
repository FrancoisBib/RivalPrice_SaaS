package models

import "time"

type PageType string

const (
	PageTypePricing  PageType = "pricing"
	PageTypeFeatures PageType = "features"
)

type MonitoredPage struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CompetitorID  uint      `gorm:"not null;index" json:"competitor_id"`
	PageType      string    `gorm:"type:varchar(50);not null" json:"page_type"` // pricing / features
	URL           string    `gorm:"type:varchar(512);not null" json:"url"`
	CSSSelector   string    `gorm:"type:text" json:"css_selector"` // optional for specific targeting
	CreatedAt     time.Time `json:"created_at"`
	Competitor    Competitor `gorm:"foreignKey:CompetitorID" json:"competitor,omitempty"`
}

func (MonitoredPage) TableName() string {
	return "monitored_pages"
}
