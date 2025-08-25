package dto

type CreateCategoryInput struct {
	Name        *string `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}
