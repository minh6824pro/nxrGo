package dto

type CreateProductVariantInput struct {
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	ProductID uint    `json:"product_id" binding:"required"`
	Image     string  `json:"image"`
}
