package dto

type UpdateStockRequest struct {
	Quantity uint `json:"quantity" binding:"required"`
}
