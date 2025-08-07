package dto

type CreateProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`

	BrandID *uint `json:"brand_id" binding:"required"`
	//BrandName *string `json:"brand_name,omitempty"`

	CategoryID *uint `json:"category_id" binding:"required"`
	//CategoryName *string `json:"category_name,omitempty"`

	MerchantID *uint `json:"merchant_id" binding:"required"`
	//MerchantName *string `json:"merchant_name,omitempty"`

	// Danh sách các variant của product
	Variants []CreateProductVariantInput `json:"variants" binding:"required"`
}

type CreateProductVariantInput struct {
	Price     float64 `json:"price" binding:"required"`
	Quantity  uint    `json:"quantity" binding:"required"`
	Image     string  `json:"image,omitempty"`
	ProductID uint    `json:"category_id,omitempty"`

	// Danh sách các option-value cho variant (VD: Color=Red, Size=M)
	OptionValues []VariantOptionValueInput `json:"option_values" binding:"required"`
}

type VariantOptionValueInput struct {
	OptionID uint   `json:"option_id" binding:"required"`
	Value    string `json:"value" binding:"required"`
}
