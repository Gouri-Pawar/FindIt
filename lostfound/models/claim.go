package models

import "gorm.io/gorm"

type Claim struct {
	gorm.Model
	ItemID    uint   `gorm:"not null;uniqueIndex:idx_claim_once" json:"item_id"`
	Item      Item   `json:"-"`
	ClaimerID uint   `gorm:"not null;uniqueIndex:idx_claim_once" json:"claimer_id"`
	Claimer   User   `gorm:"foreignKey:ClaimerID" json:"-"`
	Message   string `gorm:"not null" json:"message"`            // description to prove ownership
	Contact   string `json:"contact"`                            // claimer's contact, shared if approved
	Status    string `gorm:"default:pending;index" json:"status"` // pending, approved, rejected

	ClaimerName string `gorm:"-" json:"claimer_name"`
}