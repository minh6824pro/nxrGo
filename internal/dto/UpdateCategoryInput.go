package dto

type UpdateCategoryInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
