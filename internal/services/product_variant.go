package services

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/models/CacheModel"
)

type ProductVariantService interface {
	Create(ctx context.Context, input dto.CreateProductVariantInput) (*models.ProductVariant, error)
	GetByID(ctx context.Context, id uint) (*models.ProductVariant, error)
	List(ctx context.Context) ([]models.ProductVariant, error)
	Delete(ctx context.Context, id uint) error
	Patch(ctx context.Context, productVariant *models.ProductVariant) error
	IncreaseStock(c *gin.Context, id uint, input dto.UpdateStockRequest) (*models.ProductVariant, error)
	DecreaseStock(c *gin.Context, id uint, input dto.UpdateStockRequest) (*models.ProductVariant, error)
	CheckAndCacheProductVariants(ctx context.Context, ids []uint) ([]CacheModel.VariantLite, error)
	ListByIds(ctx context.Context, list dto.ListProductVariantIds) ([]dto.VariantCartInfoResponse, error)
}
