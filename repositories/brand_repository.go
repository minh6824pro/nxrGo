package repositories

import (
	"context"

	"github.com/minh6824pro/nxrGO/models"
)

type BrandRepository interface {
	Create(ctx context.Context, brand *models.Brand) (*models.Brand, error)
	GetByID(ctx context.Context, id uint) (*models.Brand, error)
	Update(ctx context.Context, brand *models.Brand) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Brand, error)
	GetByName(ctx context.Context, name string) (*models.Brand, error)
}
