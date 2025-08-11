package dto

import "github.com/minh6824pro/nxrGO/models"

type OrderStatusRequest struct {
	OrderStatus models.OrderStatus `json:"order_status" binding:"required"`
}
