package dto

import (
	"github.com/minh6824pro/nxrGO/internal/utils"
)

type OrderEventRequest struct {
	Event utils.OrderEvent `json:"event" binding:"required"`
}
