package repositories

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/models"
)

type PaymentInfoRepository interface {
	Create(ctx context.Context, payment *models.PaymentInfo) error
	Save(ctx context.Context, payment *models.PaymentInfo) error
	GetByID(ctx context.Context, paymentInfoID int64) (*models.PaymentInfo, error)
	GetByIdAndUserIdAndOrderId(c *gin.Context, paymentId int64, userId, orderId uint) (models.PaymentInfo, *models.Order, *models.DraftOrder, error)
}
