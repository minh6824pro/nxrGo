package dto

type CreateProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`

	BrandID   *uint   `json:"brand_id,omitempty"`
	BrandName *string `json:"brand_name,omitempty"`

	CategoryID   *uint   `json:"category_id,omitempty"`
	CategoryName *string `json:"category_name,omitempty"`

	MerchantID   *uint   `json:"merchant_id,omitempty"`
	MerchantName *string `json:"merchant_name,omitempty"`
}
