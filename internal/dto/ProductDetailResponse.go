package dto

import (
	"github.com/minh6824pro/nxrGO/internal/models"
)

type ProductDetailResponse struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	AverageRating float64 `json:"average_rating"`
	TotalBuy      uint    `json:"total_buy"`
	NumberRating  float32 `json:"number_rating"`
	Image         string  `json:"image"`
	Description   string  `json:"description"`
	Active        bool    `json:"active"`

	// Relationships
	Merchant models.Merchant         `json:"merchant"`
	Brand    models.Brand            `json:"brand"`
	Category models.Category         `json:"category"`
	Variants []VariantDetailResponse `json:"variants"`
}

type VariantDetailResponse struct {
	ID           uint                        `json:"id"`
	Quantity     uint                        `json:"quantity"`
	Price        float64                     `json:"price"`
	ProductID    uint                        `json:"product_id"`
	Image        string                      `json:"image"`
	Timestamp    int64                       `json:"timestamp"`
	Signature    string                      `json:"signature"`
	OptionValues []models.VariantOptionValue `gorm:"foreignKey:VariantID" json:"options"`
}
