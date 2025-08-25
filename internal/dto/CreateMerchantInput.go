package dto

type CreateMerchantInput struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location" binding:"required"`
}
