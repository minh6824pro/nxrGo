package impl

import (
	"context"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type orderItemGormRepository struct {
	db *gorm.DB
}

func NewOrderItemGormRepository(db *gorm.DB) repositories.OrderItemRepository {
	return &orderItemGormRepository{db}
}

func (o orderItemGormRepository) CreateTx(ctx context.Context, tx *gorm.DB, orderItem *models.OrderItem) (*models.OrderItem, error) {
	if err := tx.WithContext(ctx).Create(orderItem).Error; err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}

	return orderItem, nil
}

func (o orderItemGormRepository) Create(ctx context.Context, orderItem *models.OrderItem) error {
	if err := o.db.Create(orderItem).Error; err != nil {
		return customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}

	return nil
}
