package dto

import "github.com/minh6824pro/nxrGO/models"

type ChangePaymentMethodRequest struct {
	OrderId       uint                 `json:"order_id"`
	PaymentID     int64                `json:"payment_id"`
	PaymentMethod models.PaymentMethod `json:"payment_method"`
}
