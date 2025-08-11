package dto

import "github.com/minh6824pro/nxrGO/utils"

type UpdateOrderStatusInput struct {
	OrderID uint             `json:"orderID" binding:"required"`
	Event   utils.OrderEvent `json:"event" binding:"required"`
}
