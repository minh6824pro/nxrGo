package models

import "time"

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentSuccess  PaymentStatus = "SUCCESS"
	PaymentFailed   PaymentStatus = "FAILED"
	PaymentCanceled PaymentStatus = "CANCELED"
)

type PaymentInfo struct {
	ID                 uint          `gorm:"primaryKey" json:"id"`
	Amount             float64       `gorm:"type:decimal(10,2)" json:"amount"`
	Status             PaymentStatus `gorm:"type:varchar(20)" json:"status"`
	PaymentLink        string        `gorm:"type:varchar(255)" json:"payment_link"`
	CancellationReason string        `gorm:"type:varchar(255)" json:"cancellation_reason"`
	CancellationAt     *time.Time    `json:"cancellation_at,omitempty"`
}
