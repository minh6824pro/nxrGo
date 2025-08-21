package dto

import (
	"github.com/minh6824pro/nxrGO/models"
)

type ShippingFeeResponse struct {
	Name       string              `json:"name"`
	Mode       models.DeliveryMode `json:"mode"`
	Fee        float64             `json:"fee"`
	DeliveryID uint                `json:"deliveryId"`
	MerchantID uint                `json:"merchant_id"`
	Signature  string              `json:"signature"`
	Timestamp  int64               `json:"timestamp"`
}
