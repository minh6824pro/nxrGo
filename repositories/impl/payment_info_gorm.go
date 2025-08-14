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
