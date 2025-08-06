package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
)

type ProductVariantService interface {
	Create(ctx context.Context, input dto.CreateProductVariantInput) (*models.ProductVariant, error)
	GetByID(ctx context.Context, id uint) (*models.ProductVariant, error)
	List(ctx context.Context) ([]models.ProductVariant, error)
	Delete(ctx context.Context, id uint) error
	Patch(ctx context.Context, productVariant *models.ProductVariant) error
}
