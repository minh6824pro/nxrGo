package dto

type CreateVariantOptionInput struct {
	Name string `json:"name" binding:"required"`
}
