package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/minh6824pro/nxrGO/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

type orderItemGormRepository struct {
	db *gorm.DB
}

func NewOrderItemGormRepository(db *gorm.DB) repositories.OrderItemRepository {
	return &orderItemGormRepository{db}
}

func (o *orderItemGormRepository) CreateTx(ctx context.Context, tx *gorm.DB, orderItem *models.OrderItem) (*models.OrderItem, error) {
	if err := tx.WithContext(ctx).Create(orderItem).Error; err != nil {
		return nil, errors.NewError(errors.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}

	return orderItem, nil
}

func (o *orderItemGormRepository) Create(ctx context.Context, orderItem *models.OrderItem) error {
	if err := o.db.Create(orderItem).Error; err != nil {
		return errors.NewError(errors.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}

	return nil
}
func (o *orderItemGormRepository) Save(ctx context.Context, orderItem *models.OrderItem) error {
	return o.db.WithContext(ctx).Save(orderItem).Error

}
