package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) error
	Delete(ctx context.Context, id uint) error
	Save(ctx context.Context, order *models.Order) error
	GetByIdAndUserId(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetById(ctx context.Context, orderID uint) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	GetsByStatusAndUserId(ctx context.Context, status models.OrderStatus, userId uint) ([]*models.Order, error)
	ListByUserId(ctx context.Context, userID uint) ([]*models.Order, error)
	ListByAdmin(ctx context.Context) ([]*models.Order, error)
}
