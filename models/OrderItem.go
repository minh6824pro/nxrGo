package models

//	type OrderItem struct {
//		ID               uint           `gorm:"primaryKey" json:"id"`
//		DraftOrderID          uint           `gorm:"not null" json:"order_id"`
//		Order            Order          `gorm:"foreignKey:DraftOrderID" json:"-"`
//		ProductVariantID uint           `gorm:"not null" json:"product_variant_id"`
//		Variant          ProductVariant `gorm:"foreignKey:ProductVariantID" json:"variant"`
//		Quantity         uint           `json:"quantity"`
//		Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
//		TotalPrice       float64        `gorm:"type:decimal(10,2)" json:"total_price"`
//	}
type OrderType string

const (
	OrderTypeOrder      OrderType = "order"
	OrderTypeDraftOrder OrderType = "draft_order"
)

type OrderItem struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	OrderID          uint           `gorm:"not null" json:"order_id"`
	OrderType        OrderType      `gorm:"type:varchar(20);not null" json:"order_type"`
	ProductVariantID uint           `gorm:"not null" json:"product_variant_id"`
	Variant          ProductVariant `gorm:"foreignKey:ProductVariantID" json:"-"`
	Quantity         uint           `json:"quantity"`
	Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
	TotalPrice       float64        `gorm:"type:decimal(10,2)" json:"total_price"`
	MerchantID       uint           `gorm:"-" json:"-"`
}

/*
How to querry
db.Preload("OrderItems").Find(&order, orderID)
db.Preload("OrderItems").Find(&draftOrder, draftOrderID)
*/
