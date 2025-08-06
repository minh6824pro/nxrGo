package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
)

type VariantOptionService interface {
	Create(ctx context.Context, variantOption *dto.CreateVariantOptionInput) (*models.VariantOption, error)
	GetByID(ctx context.Context, id uint) (*models.VariantOption, error)
	List(ctx context.Context) ([]models.VariantOption, error)
	Delete(ctx context.Context, id uint) error
	Patch(ctx context.Context, id uint, variantOption *dto.UpdateVariantOptionInput) (*models.VariantOption, error)
}
