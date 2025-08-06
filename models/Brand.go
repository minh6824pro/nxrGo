package models

import (
	"time"
)

type Brand struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);unique" json:"name"`

	Products []Product `gorm:"foreignKey:BrandID" json:"products,omitempty"`

	// GORM default fields
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
