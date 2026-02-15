package models

import "time"

type Competitor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	Name      string    `gorm:"not null" json:"name"`
	URL       string    `gorm:"type:varchar(512)" json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Project   Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (Competitor) TableName() string {
	return "competitors"
}
