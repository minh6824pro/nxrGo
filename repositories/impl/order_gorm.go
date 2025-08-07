package impl

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type orderGormRepository struct {
	db *gorm.DB
}

func NewOrderGormRepository(db *gorm.DB) repositories.OrderRepository {
	return &orderGormRepository{db}
}

func (o orderGormRepository) CreateTx(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error) {
	if err := tx.Create(order).Error; err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error while create order", http.StatusInternalServerError, err)
	}
	return order, nil
}

func (o orderGormRepository) Create(ctx context.Context, order *models.Order) error {
	if err := o.db.Create(order).Error; err != nil {
		return customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error while create order", http.StatusInternalServerError, err)
	}
	return nil
}

func (o orderGormRepository) Delete(ctx context.Context, id uint) error {
	if err := o.db.WithContext(ctx).Delete(&models.Order{}, id).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1451 {
				return customErr.NewError(customErr.FOREIGN_KEY_CONSTRAINT_ERROR, "Cannot delete order because it is associated with order items", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)

	}
	return nil
}

func (o orderGormRepository) GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {

	var m models.Order
	if err := o.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", orderID, userID).
		Preload("OrderItems").
		Preload("OrderItems.Variant").
		Preload("PaymentInfo").
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.FORBIDDEN, "Order not found", http.StatusNotFound, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}

	return &m, nil
}
