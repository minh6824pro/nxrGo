package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type orderService struct {
	db                 *gorm.DB
	productVariantRepo repositories.ProductVariantRepository
	orderItemRepo      repositories.OrderItemRepository
	orderRepo          repositories.OrderRepository
}

func NewOrderService(db *gorm.DB, productVariantRepo repositories.ProductVariantRepository, orderItemRepo repositories.OrderItemRepository,
	orderRepo repositories.OrderRepository) services.OrderService {
	return &orderService{
		db:                 db,
		productVariantRepo: productVariantRepo,
		orderRepo:          orderRepo,
		orderItemRepo:      orderItemRepo,
	}
}

func (o orderService) Create(ctx context.Context, input dto.CreateOrderInput) (*models.Order, error) {
	var total float64 = 0

	// Kiểm tra số lượng & tồn tại từng orderItems
	for _, orderItem := range input.OrderItems {

		productVariant, err := o.productVariantRepo.GetByIDNoPreload(ctx, orderItem.ProductVariantID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product variant not found: %d", orderItem.ProductVariantID), http.StatusBadRequest, nil)

			}
			return nil, err
		}

		if orderItem.Quantity > productVariant.Quantity {
			return nil, customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product variant : %d Insufficient stock", orderItem.ProductVariantID), http.StatusBadRequest, nil)

		}
		if orderItem.Price != productVariant.Price {
			return nil, customErr.NewError(customErr.INVALID_PRICE, fmt.Sprintf("Product variant: %d Price Invalid", orderItem.ProductVariantID), http.StatusBadRequest, nil)
		}

		total += productVariant.Price * float64(orderItem.Quantity)
	}

	// Kiểm tra total price
	if total != input.Total {
		fmt.Println(input.Total)
		fmt.Println(total)
		return nil, customErr.NewError(customErr.INVALID_PRICE, "Total Price Invalid", http.StatusBadRequest, nil)
	}
	// Tạo order pending
	order := &models.Order{
		UserID:          input.UserID,
		Status:          models.OrderStatePending,
		Total:           total,
		ShippingAddress: input.ShippingAddress,
		ShippingFee:     o.CalculateShippingFee(ctx, input.ShippingAddress),
		PaymentMethod:   input.PaymentMethod,
		PhoneNumber:     input.PhoneNumber,
	}

	if err := o.orderRepo.Create(ctx, order); err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, fmt.Sprintf("Failed to create order : %s", err.Error()), http.StatusBadRequest, err)
	}

	msg := dto.CreateOrderMessage{
		OrderID:    order.ID,
		UserID:     order.UserID,
		Items:      input.OrderItems,
		RetryCount: 0,
	}

	body, _ := json.Marshal(msg)
	err := configs.RMQChannel.Publish(
		"",
		configs.OrderCreateQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish order: %v", err)
	}

	return order, nil

}
func (o orderService) GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	return o.orderRepo.GetById(ctx, orderID, userID)
}

func (o orderService) CalculateShippingFee(ctx context.Context, shippingAdress string) float64 {
	return float64(10000)
}
