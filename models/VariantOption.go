package models

type VariantOption struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(100);not null;unique" json:"name"`

	// Relationships
	Values []VariantOptionValue `gorm:"foreignKey:OptionID" json:"values,omitempty"`
}
