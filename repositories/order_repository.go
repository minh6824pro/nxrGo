package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) error
	Delete(ctx context.Context, id uint) error
	GetByIdAndUserId(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetById(ctx context.Context, orderID uint) (*models.Order, error)
	Save(ctx context.Context, order *models.Order) error
}
