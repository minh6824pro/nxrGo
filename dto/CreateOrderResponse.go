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
	PaymentMethod   string              `json:"payment_method"`
	ShippingAddress string              `json:"shipping_address"`
	PhoneNumber     string              `json:"phone_number"`
	DeliveryMode    models.DeliveryMode `json:"delivery_mode"`
	PaymentInfo     PaymentInfoResponse `json:"payment_info"`
	OrderItems      []OrderItemResponse `json:"order_items"`
	CreatedAt       time.Time           `json:"CreatedAt"`
	UpdatedAt       time.Time           `json:"UpdatedAt"`
	OrderType       models.OrderType    `json:"order_type"`
	ParentID        *uint               `json:"-"`
}

type PaymentInfoResponse struct {
	ID                 int64                `json:"id"`
	Total              float64              `json:"amount"`
	ShippingFee        float64              `json:"shipping_fee"`
	Status             models.PaymentStatus `json:"status"`
	PaymentLink        string               `json:"payment_link"`
	CancellationReason string               `json:"cancellation_reason"`
}

type OrderItemResponse struct {
	ID               uint             `json:"id"`
	OrderID          uint             `json:"order_id"`
	OrderType        models.OrderType `json:"order_type"`
	ProductVariantID uint             `json:"product_variant_id"`
	Quantity         uint             `json:"quantity"`
	Price            float64          `json:"price"`
	TotalPrice       float64          `json:"total_price"`
	ProductName      string           `json:"product_name"`
}
