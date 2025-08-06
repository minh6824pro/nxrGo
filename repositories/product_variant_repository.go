package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error)
	CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.ProductVariant) (*models.ProductVariant, error)
	GetByID(ctx context.Context, id uint) (*models.ProductVariant, error)
	Update(ctx context.Context, variant *models.ProductVariant) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.ProductVariant, error)
}
