package dto

import (
	"github.com/minh6824pro/nxrGO/models"
	"time"
)

type CreateOrderResponse struct {
	Data    OrderData `json:"data"`
	Message string    `json:"message"`
}

type OrderData struct {
	ID              uint                `json:"id"`
	UserID          uint                `json:"user_id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	PaymentMethod   string              `json:"payment_method"`
	ShippingAddress string              `json:"shipping_address"`
	ShippingFee     float64             `json:"shipping_fee"`
	PhoneNumber     string              `json:"phone_number"`
	PaymentInfo     PaymentInfoResponse `json:"payment_info"`
	OrderItems      []OrderItemResponse `json:"order_items"`
	CreatedAt       time.Time           `json:"CreatedAt"`
	UpdatedAt       time.Time           `json:"UpdatedAt"`
}

type PaymentInfoResponse struct {
	ID                 uint                 `json:"id"`
	Amount             float64              `json:"amount"`
	Status             models.PaymentStatus `json:"status"`
	PaymentLink        string               `json:"payment_link"`
	CancellationReason string               `json:"cancellation_reason"`
}

type OrderItemResponse struct {
	ID                  uint             `json:"id"`
	OrderID             uint             `json:"order_id"`
	OrderType           models.OrderType `json:"order_type"`
	ProductVariantID    uint             `json:"product_variant_id"`
	Quantity            uint             `json:"quantity"`
	Price               float64          `json:"price"`
	TotalPrice          float64          `json:"total_price"`
	ProductName         string           `json:"product_name"`
	ProductVariantImage string           `json:"product_variant_image"`
}
