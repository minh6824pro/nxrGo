package dto

type UpdateProductVariantInput struct {
	Quantity *int     `json:"quantity,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	Image    *string  `json:"image,omitempty"`
	Version  *int     `json:"version,omitempty"`
}
