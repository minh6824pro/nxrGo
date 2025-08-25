package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
)

type MerchantService interface {
	Create(ctx context.Context, m *dto.CreateMerchantInput) (*models.Merchant, error)
	GetByID(ctx context.Context, id uint) (*models.Merchant, error)
	//UpdateOrderStatus(ctx context.Context, m *models.Merchant) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Merchant, error)
	Patch(ctx context.Context, id uint, input *dto.UpdateMerchantInput) (*models.Merchant, error)
}
