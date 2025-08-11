package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type VariantOptionValueRepository interface {
	Create(ctx context.Context, variant *models.VariantOptionValue) (*models.VariantOptionValue, error)
	CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.VariantOptionValue) (*models.VariantOptionValue, error)

	//GetByID(ctx context.Context, id uint) (*models.VariantOption, error)
	//Delete(ctx context.Context, id uint) error
	//List(ctx context.Context) ([]models.VariantOption, error)
	//UpdateOrderStatus(ctx context.Context, variantOption *models.VariantOption) error
}
