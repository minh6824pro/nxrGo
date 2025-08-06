package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error)
	FindByID(ctx context.Context, id uint) (*models.ProductVariant, error)
	Update(ctx context.Context, variant *models.ProductVariant) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context) ([]models.ProductVariant, error)
}
