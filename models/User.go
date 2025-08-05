package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type User struct {
	UserID      uint64 `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Username    string `gorm:"type:varchar(255);unique;not null" json:"username"`
	FullName    string `gorm:"type:varchar(255)" json:"full_name"`
	Email       string `gorm:"type:varchar(255);unique" json:"email"`
	Password    string `gorm:"type:varchar(255);not null" json:"-"`
	PhoneNumber string `gorm:"type:varchar(255)" json:"phone_number"`
	Role        Role   `gorm:"type:enum('ADMIN','USER');default:'USER'" json:"role"`
	Active      bool   `gorm:"type:bit(1);default:1" json:"active"`

	// GORM default fields
	LastLogin *time.Time     `json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
