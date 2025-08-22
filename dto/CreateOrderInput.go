package dto

import (
	"github.com/minh6824pro/nxrGO/models"
)

type CreateOrderInput struct {
	UserID           uint                  `json:"-"`
	Total            float64               `json:"total" binding:"required"`
	PaymentMethod    models.PaymentMethod  `json:"payment_method" binding:"required,oneof=COD BANK"`
	ShippingAddress  string                `json:"shipping_address" binding:"required"`
	ShippingFee      float64               `json:"shipping_fee" binding:"required"`
	PhoneNumber      string                `json:"phone_number" binding:"required,len=10"`
	OrderItems       []CreateOrderItem     `json:"order_items" binding:"required,dive,required"`
	ShippingFeeInput []ShippingFeeResponse `json:"shipping_fee_input" binding:"required,dive,required"`
	Latitude         string                `json:"lat" binding:"required"`
	Longitude        string                `json:"lon" binding:"required"`
}

type CreateOrderItem struct {
	ProductVariantID uint    `json:"product_variant_id" binding:"required"`
	Quantity         uint    `json:"quantity" binding:"required,min=1"`
	Price            float64 `json:"price" binding:"required"`
	Timestamp        int64   `json:"timestamp" binding:"required"`
	Signature        string  `json:"signature" binding:"required"`
	MerchantID       uint    `json:"merchant_id" binding:"required"`
}

type OrderItemForSplit struct {
	ID         uint `json:"id"`
	MerchantID uint `json:"merchant_id"`
	DeliveryID uint `json:"delivery_id"`
}
