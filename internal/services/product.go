package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/models/CacheModel"
)

type ProductService interface {
	Create(ctx context.Context, input dto.CreateProductInput) (*uint, error)
	GetByID(ctx context.Context, id uint) (*dto.ProductDetailResponse, error)
	List(ctx context.Context) ([]models.Product, error)
	Delete(ctx context.Context, id uint) error
	Patch(ctx context.Context, product *models.Product) error
	GetProductListManagement(ctx context.Context, priceMin, priceMax *float64, priceAsc *bool, totalBuyDesc *bool, page, pageSize int) ([]*CacheModel.ProductMiniCache, int, error)

	GetProductList(ctx context.Context, name string, priceMin, priceMax *float64, priceAsc *bool, totalBuyDesc *bool, page, pageSize int, lat, lon *float64) ([]*CacheModel.ProductMiniCache, int, error)
}
