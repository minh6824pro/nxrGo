package models

import "time"

type OrderStatus string

const (
	OrderStatePending         OrderStatus = "PENDING"
	OrderStateConfirmed       OrderStatus = "CONFIRMED"
	OrderStateProcessing      OrderStatus = "PROCESSING"
	OrderStateShipped         OrderStatus = "SHIPPED"
	OrderStateDelivered       OrderStatus = "DELIVERED"
	OrderStateCancelled       OrderStatus = "CANCELLED"
	OrderStateReturnRequested OrderStatus = "RETURN_REQUESTED"
	OrderStateReturned        OrderStatus = "RETURNED"
	OrderStateRefunded        OrderStatus = "REFUNDED"
)

type PaymentMethod string

const (
	PaymentMethodCOD  PaymentMethod = "COD"
	PaymentMethodBank PaymentMethod = "BANK"
)

type Order struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null" json:"user_id"`
	User            User          `gorm:"references:UserID" json:"-"`
	Status          OrderStatus   `gorm:"type:varchar(20)" json:"status"`
	Total           float64       `gorm:"type:decimal(10,2)" json:"total"`
	PaymentMethod   PaymentMethod `gorm:"type:varchar(20)" json:"payment_method"`
	ShippingAddress string        `gorm:"type:varchar(255)" json:"shipping_address"`
	ShippingFee     float64       `gorm:"type:decimal(10,2)" json:"shipping_fee"`
	PhoneNumber     string        `gorm:"type:varchar(10)" json:"phone_number"`
	PaymentInfoID   *uint         `json:"-"`
	PaymentInfo     *PaymentInfo  `gorm:"foreignKey:PaymentInfoID" json:"payment_info,omitempty"`

	OrderItems []OrderItem `gorm:"polymorphic:Order;polymorphicValue:order" json:"order_items,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
