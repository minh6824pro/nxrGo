package dto

type UpdateVariantOptionInput struct {
	Name string `json:"name" binding:"required"`
}
