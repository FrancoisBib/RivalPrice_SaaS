package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	MonitoredPageID uint           `gorm:"not null;index" json:"monitored_page_id"`
	Snapshot       json.RawMessage `gorm:"type:jsonb;not null" json:"snapshot"` // Full page version
	Hash           string          `gorm:"type:char(64);not null" json:"hash"`    // Hash for quick comparison
	CreatedAt      time.Time       `json:"created_at"`
	MonitoredPage  MonitoredPage   `gorm:"foreignKey:MonitoredPageID" json:"monitored_page,omitempty"`
}

func (Snapshot) TableName() string {
	return "snapshots"
}
