package models

import (
	"time"
)

type Merchant struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(255);unique" json:"name" binding:"required"  `
	Location  string    `gorm:"type:varchar(255)" json:"location" binding:"required"  `
	Latitude  string    `gorm:"type:varchar(255)" json:"-"  `
	Longitude string    `gorm:"type:varchar(255)" json:"-" `
	Products  []Product `gorm:"foreignKey:MerchantID" json:"products,omitempty"`

	// GORM default fields
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
