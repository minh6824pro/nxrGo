package models

import (
	"time"
)

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentSuccess  PaymentStatus = "SUCCESS"
	PaymentFailed   PaymentStatus = "FAILED"
	PaymentCanceled PaymentStatus = "CANCELED"
)

type PaymentInfo struct {
	ID                 int64         `gorm:"primaryKey" json:"id"`
	OrderID            uint          `gorm:"not null" json:"order_id"`
	OrderType          OrderType     `gorm:"type:varchar(20);not null" json:"order_type"`
	Total              float64       `gorm:"type:decimal(10,2)" json:"amount"`
	Status             PaymentStatus `gorm:"type:varchar(20)" json:"status"`
	ShippingFee        float64       `gorm:"type:decimal(10,2)" json:"shipping_fee"`
	PaymentLink        string        `gorm:"type:varchar(255)" json:"payment_link"`
	CancellationReason string        `gorm:"type:varchar(255)" json:"cancellation_reason"`
	CancellationAt     *time.Time    `json:"cancellation_at,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
