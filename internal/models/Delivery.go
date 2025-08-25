package models

type DeliveryMode string

const (
	DeliveryModeNormal DeliveryMode = "NORMAL"
	DeliveryModeFast   DeliveryMode = "FAST"
)

type Delivery struct {
	ID           uint         `gorm:"primary_key;auto_increment" json:"id"`
	Name         string       `gorm:"size:255;not null" json:"name"`
	PricePerKm   float64      `gorm:"not null" json:"price_per_km"`
	BasePrice    float64      `gorm:"not null" json:"base_price"`
	DeliveryMode DeliveryMode `gorm:"type:varchar(20)" json:"delivery_mode"`
}
