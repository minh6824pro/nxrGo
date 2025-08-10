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

type draftOrderGormRepository struct {
	db *gorm.DB
}

func NewDraftOrderGormRepository(db *gorm.DB) repositories.DraftOrderRepository {
	return &draftOrderGormRepository{db}
}

func (d draftOrderGormRepository) CreateTx(ctx context.Context, tx *gorm.DB, order *models.DraftOrder) (*models.DraftOrder, error) {
	if err := tx.Create(order).Error; err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error while create order", http.StatusInternalServerError, err)
	}
	return order, nil
}

func (d draftOrderGormRepository) Create(ctx context.Context, order *models.DraftOrder) error {
	if err := d.db.Create(order).Error; err != nil {
		return customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error while create order", http.StatusInternalServerError, err)
	}
	return nil
}

func (d draftOrderGormRepository) Save(ctx context.Context, order *models.DraftOrder) error {
	if err := d.db.Save(order).Error; err != nil {
		return customErr.NewError(
			customErr.INTERNAL_ERROR,
			"Unexpected error while save order",
			http.StatusInternalServerError,
			err,
		)
	}
	return nil
}

func (d draftOrderGormRepository) Delete(ctx context.Context, id uint) error {
	if err := d.db.WithContext(ctx).Delete(&models.Order{}, id).Error; err != nil {
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

func (d draftOrderGormRepository) GetById(ctx context.Context, orderID uint) (*models.DraftOrder, error) {

	var m models.DraftOrder
	if err := d.db.WithContext(ctx).
		Where("id = ? ", orderID).
		Preload("OrderItems").
		Preload("PaymentInfo").
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.FORBIDDEN, "Order not found", http.StatusNotFound, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}

	return &m, nil
}
