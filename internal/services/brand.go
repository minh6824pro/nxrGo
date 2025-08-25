package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
)

type BrandService interface {
	Create(ctx context.Context, b *dto.CreateBrandInput) (*models.Brand, error)
	GetByID(ctx context.Context, id uint) (*models.Brand, error)
	Update(ctx context.Context, b *models.Brand) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Brand, error)
	Patch(ctx context.Context, id uint, input *dto.UpdateMerchantInput) (*models.Brand, error)
}
