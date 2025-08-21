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
	if err := d.db.WithContext(ctx).Delete(&models.DraftOrder{}, id).Error; err != nil {
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
		Where("id = ?", orderID).
		Preload("OrderItems").
		Preload("PaymentInfos", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.FORBIDDEN, "Order not found", http.StatusNotFound, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}

	return &m, nil
}

func (d draftOrderGormRepository) GetsForDbUpdate(ctx context.Context) ([]models.DraftOrder, error) {
	var draftOrders []models.DraftOrder
	err := d.db.WithContext(ctx).
		Where("to_order != ?", 0).
		Find(&draftOrders).Error
	if err != nil {
		return nil, err
	}
	return draftOrders, nil
}

func (d draftOrderGormRepository) CleanDraft(ctx context.Context) error {
	// Get drafts to remove
	var drafts []models.DraftOrder
	if err := d.db.WithContext(ctx).
		Where("to_order = ?", 0).
		Find(&drafts).Error; err != nil {
		return err
	}

	// Delete related order items
	for _, draft := range drafts {
		if err := d.db.WithContext(ctx).
			Where("order_type = ? AND order_id = ?", "draft_order", draft.ID).
			Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
	}

	// Xóa draft orders trước
	if err := d.db.WithContext(ctx).
		Where("to_order = ?", 0).
		Delete(&models.DraftOrder{}).Error; err != nil {
		return err
	}

	//// Sau đó xóa payment info liên quan
	//for _, draft := range drafts {
	//	if err := d.db.WithContext(ctx).
	//		Where("id = ?", draft.PaymentInfoID).
	//		Delete(&models.PaymentInfo{}).Error; err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (d draftOrderGormRepository) ListByUserIdToOrderNull(ctx context.Context, userID uint) ([]*models.DraftOrder, error) {
	var orders []*models.DraftOrder

	err := d.db.WithContext(ctx).
		Preload("OrderItems").
		Preload("PaymentInfos").
		Preload("OrderItems.Variant.Product").
		Preload("OrderItems.Variant.OptionValues").
		Where("user_id = ? AND to_order IS NULL", userID).
		//	Where("user_id = ? AND to_order IS NULL AND (parent_id IS NULL OR parent_id != ?)", userID, 0).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}
func (d draftOrderGormRepository) ListByAdmin(ctx context.Context) ([]*models.DraftOrder, error) {

	var orders []*models.DraftOrder
	err := d.db.WithContext(ctx).
		Preload("OrderItems").
		Preload("PaymentInfos").
		Preload("OrderItems.Variant.Product").
		Preload("OrderItems.Variant.OptionValues").
		Where("to_order IS NULL").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
