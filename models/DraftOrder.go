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
	Latitude        string        `gorm:"type:varchar(20)" json:"latitude"`
	Longitude       string        `gorm:"type:varchar(20)" json:"longitude"`
	PaymentInfos    []PaymentInfo `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"payment_infos,omitempty"`
	ToOrderID       *uint         `gorm:"column:to_order" json:"to_order"`
	OrderItems      []OrderItem   `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"order_items,omitempty"`
	DeliveryMode    DeliveryMode  `gorm:"type:varchar(20)" json:"delivery_mode"`

	Delivery DeliveryDetail `gorm:"polymorphic:Order;polymorphicValue:draft_order" json:"delivery,omitempty"`

	ParentID  *uint `gorm:"column:parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
