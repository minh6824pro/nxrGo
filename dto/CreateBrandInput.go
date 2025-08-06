package dto

type CreateBrandInput struct {
	Name string `json:"name" binding:"required"`
}
