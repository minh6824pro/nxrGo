package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type VariantOptionRepository interface {
	Create(ctx context.Context, variant *models.VariantOption) (*models.VariantOption, error)
	GetByID(ctx context.Context, id uint) (*models.VariantOption, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*models.VariantOption, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.VariantOption, error)
	Update(ctx context.Context, variantOption *models.VariantOption) error
}
