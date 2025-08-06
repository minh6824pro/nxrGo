package dto

type CreateMerchantInput struct {
	Name string `json:"name" binding:"required"`
}
