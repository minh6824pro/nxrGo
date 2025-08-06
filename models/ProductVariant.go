package models

type ProductVariant struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Image     string  `gorm:"type:varchar(255)" json:"image"`
	Version   int     `gorm:"default:0" json:"version"`

	// Relationships
	Product      Product              `gorm:"foreignKey:ProductID" json:"-"`
	OptionValues []VariantOptionValue `gorm:"foreignKey:VariantID" json:"options,omitempty"`
}
