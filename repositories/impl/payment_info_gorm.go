package impl

import (
	"context"
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
