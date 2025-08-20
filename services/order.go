package services

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/utils"
)

type OrderService interface {
	Create(ctx context.Context, input dto.CreateOrderInput) (*dto.CreateOrderResponse, error)
	GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	PayOSPaymentSuccess(ctx context.Context, paymentInfoID int64)
	UpdateQuantity(ctx context.Context) error
	GetsByStatus(ctx context.Context, status models.OrderStatus, userId uint) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderId uint, status utils.OrderEvent) (*models.Order, error)
	ReBuy(c context.Context, orderID uint, userId uint) (*dto.CreateOrderResponse, error)
	ListByUserId(ctx context.Context, userID uint) ([]*dto.OrderData, error)
	ChangePaymentMethod(c *gin.Context, payment dto.ChangePaymentMethodRequest, u uint) (*models.Order, error)
	ListByAdmin(c *gin.Context) ([]*dto.OrderData, error)
}
