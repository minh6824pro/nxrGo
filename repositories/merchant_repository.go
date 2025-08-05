package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant *models.Merchant) (*models.Merchant, error)
	GetByID(ctx context.Context, id uint) (*models.Merchant, error)
	Update(ctx context.Context, merchant *models.Merchant) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Merchant, error)
	GetByName(ctx context.Context, name string) (*models.Merchant, error)
}
