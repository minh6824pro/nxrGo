package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type DraftOrderRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, order *models.DraftOrder) (*models.DraftOrder, error)
	Create(ctx context.Context, order *models.DraftOrder) error
	Delete(ctx context.Context, id uint) error
	GetById(ctx context.Context, orderID uint, userID uint) (*models.DraftOrder, error)
}
