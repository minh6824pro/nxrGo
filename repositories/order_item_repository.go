package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type OrderItemRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, orderItem *models.OrderItem) (*models.OrderItem, error)
	Create(ctx context.Context, orderItem *models.OrderItem) error
	Save(ctx context.Context, orderItem *models.OrderItem) error
}
