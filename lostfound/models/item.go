package models

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Category    string `gorm:"index" json:"category"`
	Location    string `gorm:"index" json:"location"`
	Type        string `gorm:"not null;index" json:"type"`   // "lost" or "found"
	Status      string `gorm:"default:open" json:"status"`   // open, claimed, returned
	ImageURL    string `json:"image_url"`
	UserID      uint   `json:"user_id"`
	User        User   `json:"-"`
}