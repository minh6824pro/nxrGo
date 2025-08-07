package dto

type CreateOrderMessage struct {
	OrderID    uint              `json:"order_id"`
	UserID     uint              `json:"user_id"`
	Items      []CreateOrderItem `json:"items"`
	RetryCount int               `json:"retry_count"`
}
