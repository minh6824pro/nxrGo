package models

import (
	"time"
)

type Category struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255);unique" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`

	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`

	// GORM default fields
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
