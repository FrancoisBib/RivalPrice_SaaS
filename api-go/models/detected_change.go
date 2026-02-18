package models

import "time"

// DetectedChange mirrors the detected_changes table (written by Python AI engine)
type DetectedChange struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PageID          int       `gorm:"column:page_id;not null;index" json:"page_id"`
	PageType        string    `gorm:"column:page_type;type:varchar(20)" json:"page_type"`
	ChangeType      string    `gorm:"column:change_type;type:varchar(50);not null" json:"change_type"`
	OldPrice        string    `gorm:"column:old_price;type:varchar(50)" json:"old_price"`
	NewPrice        string    `gorm:"column:new_price;type:varchar(50)" json:"new_price"`
	ChangePercent   float64   `gorm:"column:change_percent" json:"change_percent"`
	OldAvailability string    `gorm:"column:old_availability;type:varchar(50)" json:"old_availability"`
	NewAvailability string    `gorm:"column:new_availability;type:varchar(50)" json:"new_availability"`
	OldFeatures     string    `gorm:"column:old_features;type:text" json:"old_features"`
	NewFeatures     string    `gorm:"column:new_features;type:text" json:"new_features"`
	FeaturesAdded   string    `gorm:"column:features_added;type:text" json:"features_added"`
	FeaturesRemoved string    `gorm:"column:features_removed;type:text" json:"features_removed"`
	OldText         string    `gorm:"column:old_text;type:text" json:"old_text"`
	NewText         string    `gorm:"column:new_text;type:text" json:"new_text"`
	OldHash         string    `gorm:"column:old_hash;type:varchar(64)" json:"old_hash"`
	NewHash         string    `gorm:"column:new_hash;type:varchar(64)" json:"new_hash"`
	DetectedAt      time.Time `gorm:"column:detected_at;not null" json:"detected_at"`
	RawData         string    `gorm:"column:raw_data;type:text" json:"raw_data"`
}

func (DetectedChange) TableName() string {
	return "detected_changes"
}
