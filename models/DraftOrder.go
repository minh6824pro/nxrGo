package models

import "time"

type DraftOrder struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null" json:"user_id"`
	User            User          `gorm:"references:UserID" json:"-"`
	Status          OrderStatus   `gorm:"type:varchar(20)" json:"status"`
	Total           float64       `gorm:"type:decimal(10,2)" json:"total"`
	PaymentMethod   PaymentMethod `gorm:"type:varchar(20)" json:"payment_method"`
	ShippingAddress string        `gorm:"type:varchar(255)" json:"shipping_address"`
	ShippingFee     float64       `gorm:"type:decimal(10,2)" json:"shipping_fee"`
	PhoneNumber     string        `gorm:"type:varchar(10)" json:"phone_number"`
	PaymentInfoID   *uint         `json:"payment_info_id,omitempty"`
	PaymentInfo     *PaymentInfo  `gorm:"foreignKey:PaymentInfoID" json:"payment_info,omitempty"`

	OrderItems []OrderItem `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"order_items,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
