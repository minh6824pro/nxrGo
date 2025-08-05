package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *models.Category) (*models.Category, error)
	GetByID(ctx context.Context, id uint) (*models.Category, error)
	Update(ctx context.Context, c *models.Category) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Category, error)
	GetByName(ctx context.Context, name string) (*models.Category, error)
}
