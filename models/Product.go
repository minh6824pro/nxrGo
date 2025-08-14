package models

import (
	"gorm.io/gorm"
	"time"
)

// @swaggertype object{id=integer,name=string,merchant_id=integer,brand_id=integer,category_id=integer,average_rating=number,number_rating=number,image=string,description=string,active=boolean,merchant=object,brand=object,category=object,variants=array}
type Product struct {
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string  `gorm:"type:varchar(255);not null" json:"name"`
	MerchantID    uint    `gorm:"not null" json:"merchant_id"`
	BrandID       uint    `gorm:"not null" json:"brand_id"`
	CategoryID    uint    `gorm:"not null" json:"category_id"`
	AverageRating float64 `json:"average_rating"`
	TotalBuy      uint    `json:"total_buy"`
	NumberRating  float32 `gorm:"type:float" json:"number_rating"`
	Image         string  `gorm:"type:varchar(255)" json:"image"`
	Description   string  `gorm:"type:varchar(255)" json:"description"`
	Active        bool    `gorm:"default:true" json:"active"`

	// Relationships
	Merchant Merchant         `gorm:"foreignKey:MerchantID;references:ID" json:"merchant"`
	Brand    Brand            `gorm:"foreignKey:BrandID;references:ID" json:"brand"`
	Category Category         `gorm:"foreignKey:CategoryID;references:ID" json:"category"`
	Variants []ProductVariant `gorm:"foreignKey:ProductID;references:ID" json:"variants,omitempty"`

	// GORM default fields
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
