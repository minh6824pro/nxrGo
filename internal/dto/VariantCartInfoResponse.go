package dto

type VariantCartInfoResponse struct {
	ID           uint    `json:"id"`
	Price        float64 `json:"price"`
	ProductName  string  `json:"product_name"`
	ProductID    uint    `json:"product_id"`
	Quantity     uint    `json:"quantity"`
	Option       string  `json:"option"`
	MerchantName string  `json:"merchant_name"`
	MerchantID   uint    `json:"merchant_id"`
	Image        string  `json:"image"`
	Timestamp    int64   `json:"timestamp"`
	Signature    string  `json:"signature"`
}
