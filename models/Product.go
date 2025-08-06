package models

import (
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string  `gorm:"type:varchar(255);not null" json:"name"`
	MerchantID    uint    `gorm:"not null" json:"merchant_id"`
	BrandID       uint    `gorm:"not null" json:"brand_id"`
	CategoryID    uint    `gorm:"not null" json:"category_id"`
	AverageRating float64 `json:"average_rating"`
	NumberRating  float32 `gorm:"type:float" json:"number_rating"`
	Image         string  `gorm:"type:varchar(255)" json:"image"`
	Description   string  `gorm:"type:varchar(255)" json:"description"`
	Active        bool    `gorm:"default:true" json:"active"`

	// Relationships
	Merchant Merchant         `gorm:"foreignKey:MerchantID" json:"merchant"`
	Brand    Brand            `gorm:"foreignKey:BrandID" json:"brand"`
	Category Category         `gorm:"foreignKey:CategoryID" json:"category"`
	Variants []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`

	// GORM default fields
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
