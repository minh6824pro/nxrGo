package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatePending         OrderStatus = "PENDING"
	OrderStateConfirmed       OrderStatus = "CONFIRMED"
	OrderStateProcessing      OrderStatus = "PROCESSING"
	OrderStateShipped         OrderStatus = "SHIPPED"
	OrderStateDelivered       OrderStatus = "DELIVERED"
	OrderStateDone            OrderStatus = "DONE"
	OrderStateReturnRequested OrderStatus = "RETURN_REQUESTED"
	OrderStateReturnShipping  OrderStatus = "RETURN_SHIPPING"
	OrderStateReturned        OrderStatus = "RETURNED"
	OrderStateCancelled       OrderStatus = "CANCELLED"
)

type PaymentMethod string

const (
	PaymentMethodCOD  PaymentMethod = "COD"
	PaymentMethodBank PaymentMethod = "BANK"
)

type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`    //
	UserID          uint           `gorm:"not null" json:"user_id"` //
	User            User           `gorm:"references:UserID" json:"-"`
	Status          OrderStatus    `gorm:"type:varchar(20)" json:"status"` //
	PaymentMethod   PaymentMethod  `gorm:"type:varchar(20)" json:"payment_method"`
	DeliveryMode    DeliveryMode   `gorm:"type:varchar(20)" json:"delivery_mode"`
	ShippingAddress string         `gorm:"type:varchar(255)" json:"shipping_address"`
	PhoneNumber     string         `gorm:"type:varchar(10)" json:"phone_number"`
	ParentID        *uint          `gorm:"column:parent_id" json:"parent_id,omitempty"`
	PaymentInfos    []PaymentInfo  `gorm:"polymorphic:Order;polymorphicValue:order" json:"payment_infos,omitempty"`
	OrderItems      []OrderItem    `gorm:"polymorphic:Order;polymorphicValue:order" json:"order_items,omitempty"`
	Delivery        DeliveryDetail `gorm:"polymorphic:Order;polymorphicValue:order" json:"delivery,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

var orderStatusOrder = map[OrderStatus]int{
	OrderStatePending:         1,
	OrderStateConfirmed:       2,
	OrderStateProcessing:      3,
	OrderStateShipped:         4,
	OrderStateDelivered:       5,
	OrderStateReturnRequested: 6,
	OrderStateReturnShipping:  7,
	OrderStateReturned:        8,
	OrderStateCancelled:       9,
}

func IsBeforeOrderStatus(s1, s2 OrderStatus) bool {
	order1, ok1 := orderStatusOrder[s1]
	order2, ok2 := orderStatusOrder[s2]

	if !ok1 || !ok2 {
		return false
	}
	return order1 <= order2
}
