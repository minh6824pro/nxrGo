package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type BrandService interface {
	Create(ctx context.Context, b *models.Brand) (*models.Brand, error)
	GetByID(ctx context.Context, id uint) (*models.Brand, error)
	Update(ctx context.Context, b *models.Brand) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Brand, error)
	Patch(ctx context.Context, id uint, updates map[string]interface{}) (*models.Brand, error)
}
