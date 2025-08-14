package models

import "time"

type DraftOrder struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null" json:"user_id"`
	User            User          `gorm:"references:UserID" json:"-"`
	Status          OrderStatus   `gorm:"type:varchar(20)" json:"status"`
	PaymentMethod   PaymentMethod `gorm:"type:varchar(20)" json:"payment_method"`
	ShippingAddress string        `gorm:"type:varchar(255)" json:"shipping_address"`
	PhoneNumber     string        `gorm:"type:varchar(10)" json:"phone_number"`
	PaymentInfos    []PaymentInfo `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"payment_infos,omitempty"`
	ToOrderID       *uint         `gorm:"column:to_order" json:"to_order"`
	OrderItems      []OrderItem   `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"order_items,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
