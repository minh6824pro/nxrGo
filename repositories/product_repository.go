package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/models/CacheModel"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, c *models.Product) (*models.Product, error)
	CreateWithTx(ctx context.Context, tx *gorm.DB, c *models.Product) (*models.Product, error)
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	GetByIdPreloadVariant(ctx context.Context, id uint) (*models.Product, error)
	Update(ctx context.Context, c *models.Product) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Product, error)
	ListWithPagination(ctx context.Context, page int, size int) ([]models.Product, int64, int64, error)
	GetAllProductId(ctx context.Context) ([]uint, error)
	GetProductListFilter(ctx context.Context, priceMin, priceMax *float64, priceAsc *bool, totalBuyDescStr *bool, page, pageSize int) ([]CacheModel.ListProductQueryCache, int, error)
	GetProductListFilterOptimized(ctx context.Context, priceMin, priceMax *float64, priceAsc *bool, totalBuyDescStr *bool, page, pageSize int) ([]CacheModel.ListProductQueryCache, int, error)
}
