package models

type DeliveryMode string

const (
	DeliveryModeNormal DeliveryMode = "normal"
	DeliveryModeFast   DeliveryMode = "fast"
)

type Delivery struct {
	ID           uint         `gorm:"primary_key;auto_increment" json:"id"`
	Name         string       `gorm:"size:255;not null" json:"name"`
	PricePerKm   float64      `gorm:"not null" json:"price_per_km"`
	BasePrice    float64      `gorm:"not null" json:"base_price"`
	DeliveryMode DeliveryMode `json:"delivery_mode"`
}
