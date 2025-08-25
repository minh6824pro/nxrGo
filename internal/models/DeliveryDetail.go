package models

type DeliveryDetail struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	OrderID    uint      `gorm:"not null" json:"order_id"`
	DeliveryID uint      `gorm:"not null" json:"delivery_id"`
	OrderType  OrderType `gorm:"type:varchar(20);not null" json:"order_type"`
}
