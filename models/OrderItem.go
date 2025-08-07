package models

type OrderItem struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	OrderID          uint           `gorm:"not null" json:"order_id"`
	Order            Order          `gorm:"foreignKey:OrderID" json:"-"`
	ProductVariantID uint           `gorm:"not null" json:"product_variant_id"`
	Variant          ProductVariant `gorm:"foreignKey:ProductVariantID" json:"variant"`
	Quantity         uint           `json:"quantity"`
	Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
	TotalPrice       float64        `gorm:"type:decimal(10,2)" json:"total_price"`
}
