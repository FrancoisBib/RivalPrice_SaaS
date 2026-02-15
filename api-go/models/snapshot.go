package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	MonitoredPageID uint          `gorm:"not null;index" json:"monitored_page_id"`
	Price          string        `gorm:"type:varchar(100)" json:"price"`
	Availability   string        `gorm:"type:varchar(50)" json:"availability"`
	RawData        json.RawMessage `gorm:"type:jsonb" json:"raw_data"`
	ScrapedAt      time.Time     `json:"scraped_at"`
	MonitoredPage  MonitoredPage `gorm:"foreignKey:MonitoredPageID" json:"monitored_page,omitempty"`
}

func (Snapshot) TableName() string {
	return "snapshots"
}
