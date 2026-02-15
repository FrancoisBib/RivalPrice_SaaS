package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	MonitoredPageID uint           `gorm:"not null;index" json:"monitored_page_id"`
	Price          string          `gorm:"type:varchar(100)" json:"price"`
	Availability   string          `gorm:"type:varchar(50)" json:"availability"`
	RawData        json.RawMessage `gorm:"type:jsonb" json:"raw_data"`
	ScrapedAt      time.Time       `json:"scraped_at"`
	MonitoredPage  MonitoredPage   `gorm:"foreignKey:MonitoredPageID" json:"monitored_page,omitempty"`
}

func (Snapshot) TableName() string {
	return "snapshots"
}

type MonitoredPage struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	CompetitorID uint   `gorm:"not null;index" json:"competitor_id"`
	PageType     string `gorm:"type:varchar(50);not null" json:"page_type"`
	URL          string `gorm:"type:varchar(512);not null" json:"url"`
	CSSSelector  string `gorm:"type:text" json:"css_selector"`
}

func (MonitoredPage) TableName() string {
	return "monitored_pages"
}
