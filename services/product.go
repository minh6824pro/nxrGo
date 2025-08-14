package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/models/CacheModel"
)

type ProductService interface {
	Create(ctx context.Context, input dto.CreateProductInput) (*uint, error)
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	List(ctx context.Context) ([]models.Product, error)
	Delete(ctx context.Context, id uint) error
	Patch(ctx context.Context, product *models.Product) error

	GetProductList(ctx context.Context, priceMin, priceMax *float64, priceAsc *bool, totalBuyDesc *bool, page, pageSize int) ([]*CacheModel.ProductMiniCache, int, error)
}
