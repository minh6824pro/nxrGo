package dto

type VariantCartInfoResponse struct {
	ID          uint    `json:"id"`
	Price       float64 `json:"price"`
	ProductName string  `json:"product_name"`
	ProductID   uint    `json:"product_id"`
	Quantity    uint    `json:"quantity"`
	Option      string  `json:"option"`
	Image       string  `json:"image"`
}
