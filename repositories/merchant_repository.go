package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant *models.Merchant) (*models.Merchant, error)
	GetByID(ctx context.Context, id uint) (*models.Merchant, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*models.Merchant, error)
	Update(ctx context.Context, merchant *models.Merchant) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Merchant, error)
	GetByName(ctx context.Context, name string) (*models.Merchant, error)
	GetByNameTx(ctx context.Context, tx *gorm.DB, name string) (*models.Merchant, error)
	CreateTx(ctx context.Context, tx *gorm.DB, merchant *models.Merchant) (*models.Merchant, error)
	GetDeliveriesInfo(ctx context.Context) ([]*models.Delivery, error)
	GetDeliveryInfo(ctx context.Context, id uint) (*models.Delivery, error)
}
