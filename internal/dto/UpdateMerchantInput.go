package dto

type UpdateMerchantInput struct {
	Name string `json:"name" binding:"required"`
}
