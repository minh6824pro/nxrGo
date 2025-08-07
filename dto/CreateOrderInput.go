package dto

import "github.com/minh6824pro/nxrGO/models"

type CreateOrderItem struct {
	ProductVariantID uint    `json:"product_variant_id" binding:"required"`
	Quantity         uint    `json:"quantity" binding:"required,min=1"`
	Price            float64 `json:"price" binding:"required"`
	OrderID          uint    `json:"-"`
}

type CreateOrderInput struct {
	UserID          uint                 `json:"-"`
	Total           float64              `json:"total" binding:"required"`
	PaymentMethod   models.PaymentMethod `json:"payment_method" binding:"required,oneof=COD BANK"`
	ShippingAddress string               `json:"shipping_address" binding:"required"`
	ShippingFee     string               `json:"shipping_fee" binding:"required"`
	PhoneNumber     string               `json:"phone_number" binding:"required,len=10"`
	OrderItems      []CreateOrderItem    `json:"order_items" binding:"required,dive,required"`
}
