package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type ProductRepository interface {
	Create(ctx context.Context, c *models.Product) (*models.Product, error)
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	Update(ctx context.Context, c *models.Product) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Product, error)
}
