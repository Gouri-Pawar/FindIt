package models

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Category    string `gorm:"index" json:"category"`
	Location    string `gorm:"index" json:"location"`
	KeptAt      string `json:"kept_at"`
	Contact     string `json:"contact"`
	Type        string `gorm:"not null;index" json:"type"`
	Status      string `gorm:"default:open" json:"status"` // open, claimed, returned
	ImageURL    string `json:"image_url"`
	ImageData   []byte `gorm:"type:bytea" json:"-"` // the photo itself, stored in the database
	ImageType   string `json:"-"`                   // for example image/jpeg
	UserID      uint   `json:"user_id"`
	User        User   `json:"-"`

	PosterName    string `gorm:"-" json:"poster_name"`
	HasContact    bool   `gorm:"-" json:"has_contact"`
	MyClaim       string `gorm:"-" json:"my_claim"`
	PendingClaims int    `gorm:"-" json:"pending_claims"`
}