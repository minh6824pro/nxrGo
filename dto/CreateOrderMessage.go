package dto

import "github.com/minh6824pro/nxrGO/models"

type CreateOrderMessage struct {
	OrderID       uint              `json:"order_id"`
	UserID        uint              `json:"user_id"`
	Items         []CreateOrderItem `json:"items"`
	RetryCount    int               `json:"retry_count"`
	CorrelationID string            `json:"correlation_id"`
}

type OrderProcessingResult struct {
	OrderID      uint          `json:"order_id"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
	OrderData    *models.Order `json:"order_data,omitempty"`
}
