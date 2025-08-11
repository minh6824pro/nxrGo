package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
)

type OrderService interface {
	Create(ctx context.Context, input dto.CreateOrderInput) (*models.Order, error)
	GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	PayOSPaymentSuccess(ctx context.Context, draftOrderID uint)
	UpdateQuantity(ctx context.Context) error
}
