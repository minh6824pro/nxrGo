package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"

	"gorm.io/gorm"
)

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error)
	CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.ProductVariant) (*models.ProductVariant, error)
	GetByID(ctx context.Context, id uint) (*models.ProductVariant, error)
	GetByIDNoPreload(ctx context.Context, id uint) (*models.ProductVariant, error)
	GetByIDForRedisCache(ctx context.Context, id uint) (*models.ProductVariant, error)
	GetByIDSForRedisCache(ctx context.Context, productVariantIds []uint) ([]models.ProductVariant, error)
	CheckExistsAndQuantity(ctx context.Context, id uint, quantity uint) error
	Update(ctx context.Context, variant *models.ProductVariant) error
	IncreaseQuantity(ctx context.Context, quantityMap map[uint]uint) error
	DecreaseQuantity(ctx context.Context, quantityMap map[uint]uint) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.ProductVariant, error)
	CheckAndDecreaseStock(ctx context.Context, pvID uint, quantity uint) (*models.ProductVariant, error)
	GetByIDSForProductMiniCache(ctx context.Context, productIds []uint) ([]models.ProductVariant, error)
	ListByIds(ctx context.Context, list dto.ListProductVariantIds) ([]models.ProductVariant, error)
}
