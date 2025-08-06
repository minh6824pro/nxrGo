package models

type VariantOptionValue struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	VariantID uint   `gorm:"not null" json:"variant_id"`
	OptionID  uint   `gorm:"not null" json:"option_id"`
	Value     string `gorm:"type:varchar(255);not null" json:"value"` // e.g., Red, L

	// Relationships
	Variant ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	Option  VariantOption  `gorm:"foreignKey:OptionID" json:"option,omitempty"`
}
