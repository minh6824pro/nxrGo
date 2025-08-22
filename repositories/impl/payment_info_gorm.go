package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type paymentInfoGormRepository struct {
	db *gorm.DB
}

func NewPaymentInfoGormImpl(db *gorm.DB) repositories.PaymentInfoRepository {
	return &paymentInfoGormRepository{db}
}

func (p paymentInfoGormRepository) Create(ctx context.Context, payment *models.PaymentInfo) error {
	if err := p.db.Create(&payment).Error; err != nil {
		return customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}
	return nil
}

func (p paymentInfoGormRepository) Save(ctx context.Context, payment *models.PaymentInfo) error {
	if err := p.db.WithContext(ctx).Save(payment).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return customErr.NewError(customErr.DUPLICATED_ERROR, "Merchant already exists", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (p paymentInfoGormRepository) GetByID(ctx context.Context, paymentInfoID int64) (*models.PaymentInfo, error) {
	var pm models.PaymentInfo
	if err := p.db.WithContext(ctx).First(&pm, paymentInfoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Merchant not found", http.StatusBadRequest, nil)
		}

		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return &pm, nil
}

func (p paymentInfoGormRepository) GetByIdAndUserIdAndOrderId(
	c *gin.Context,
	paymentId int64, userId, orderId uint,
) (models.PaymentInfo, *models.Order, *models.DraftOrder, error) {

	var pm models.PaymentInfo
	var order models.Order
	var draft models.DraftOrder

	// Get payment_info
	if err := p.db.WithContext(c).
		Table("payment_infos").
		Where("id = ? AND order_id = ?", paymentId, orderId).
		First(&pm).Error; err != nil {
		return pm, nil, nil, err
	}

	// Get order/draft by order_type
	switch pm.OrderType {
	case "order":
		if err := p.db.WithContext(c).
			Table("orders").
			Preload("OrderItems").
			Preload("PaymentInfos", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at DESC")
			}).
			Preload("Delivery").
			Where("id = ? AND user_id = ?", pm.OrderID, userId).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return pm, nil, nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Order not found", http.StatusBadRequest, nil)
			}
			return pm, nil, nil, err
		}
		return pm, &order, nil, nil

	case "draft_order":
		if err := p.db.WithContext(c).
			Table("draft_orders").
			Preload("OrderItems").
			Preload("PaymentInfos", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at DESC")
			}).
			Preload("Delivery").
			Where("id = ? AND user_id = ?", pm.OrderID, userId).
			First(&draft).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return pm, nil, nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Order not found", http.StatusBadRequest, nil)
			}

			return pm, nil, nil, err
		}
		return pm, nil, &draft, nil

	default:
		return pm, nil, nil, fmt.Errorf("unknown order type: %s", pm.OrderType)
	}
}
