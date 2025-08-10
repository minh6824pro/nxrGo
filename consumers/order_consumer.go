package consumers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/database"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"github.com/payOSHQ/payos-lib-golang"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func StartOrderConsumer() {
	msgs, err := configs.RMQChannel.Consume(configs.OrderCreateQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume from order.create: %v", err)
	}

	go func() {
		for d := range msgs {
			var orderMsg dto.CreateOrderMessage
			if err := json.Unmarshal(d.Body, &orderMsg); err != nil {
				log.Println("Invalid message format")
				d.Ack(false)
				continue
			}

			// Xử lý với retry logic ngay trong consumer
			success, finalErr := processOrderWithRetry(orderMsg)

			// Gửi reply sau khi hoàn tất tất cả retry attempts
			result := dto.OrderProcessingResult{
				OrderID: orderMsg.OrderID,
				Success: success,
			}

			if success {
				// Load order đầy đủ thông tin để gửi về
				var fullOrder models.Order
				db := database.DB
				if loadErr := db.Where("id = ?", orderMsg.OrderID).
					Preload("OrderItems").
					Preload("OrderItems.Variant").
					Preload("OrderItems.Variant.Product").
					Preload("PaymentInfo").
					First(&fullOrder).Error; loadErr == nil {
					result.OrderData = &fullOrder
				}
			} else {
				result.ErrorMessage = finalErr.Error()

				// Gửi message vào DLQ để cleanup
				dlqMsg, _ := json.Marshal(orderMsg)
				configs.RMQChannel.Publish("", configs.OrderDLQ, false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        dlqMsg,
				})
			}

			sendReply(d.ReplyTo, d.CorrelationId, result)
			d.Ack(false)
		}
	}()
}

// Thêm hàm xử lý retry
func processOrderWithRetry(msg dto.CreateOrderMessage) (bool, error) {
	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 Retrying order %d (attempt %d/%d)", msg.OrderID, attempt, maxRetries)
			time.Sleep(2 * time.Second) // Delay giữa các retry
		}

		err := processOrderMessage(msg)
		if err == nil {
			log.Printf("✅ Successfully processed order %d on attempt %d", msg.OrderID, attempt+1)
			return true, nil
		}

		lastErr = err
		log.Printf("❌ Order %d failed on attempt %d: %v", msg.OrderID, attempt+1, err)
	}

	log.Printf("⚠️ Order %d failed after %d attempts -> will be sent to DLQ", msg.OrderID, maxRetries+1)
	return false, lastErr
}

func sendReply(replyTo, correlationID string, result dto.OrderProcessingResult) {
	if replyTo == "" || correlationID == "" {
		return // Không có reply queue hoặc correlation ID
	}

	body, _ := json.Marshal(result)
	err := configs.RMQChannel.Publish(
		"",
		replyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			CorrelationId: correlationID,
		},
	)
	if err != nil {
		log.Printf("Failed to send reply: %v", err)
	}
}

func processOrderMessage(msg dto.CreateOrderMessage) error {
	db := database.DB
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//tx.Rollback()
	//return customErr.NewError(customErr.VERSION_CONFLICT, "Test order retry", http.StatusConflict, nil)
	for _, item := range msg.Items {
		var pv models.ProductVariant
		if err := tx.First(&pv, item.ProductVariantID).Error; err != nil || pv.Quantity < item.Quantity {
			tx.Rollback()
			return err
		}
		pv.Quantity -= item.Quantity

		result := tx.Model(&models.ProductVariant{}).
			Where("id = ? AND version = ?", pv.ID, pv.Version).
			Updates(map[string]interface{}{
				"quantity": pv.Quantity,
				"version":  pv.Version + 1,
			})

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Println("Product variant : ", pv.ID, pv.Version, "Order", msg.OrderID, "conflicted")
			return customErr.NewError(customErr.VERSION_CONFLICT, "Version Conflict", http.StatusConflict, nil)
		}

		orderItem := &models.OrderItem{
			OrderID:          msg.OrderID,
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.Price * float64(item.Quantity),
		}
		if err := tx.Create(orderItem).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Commit error:", err)
		return err
	}

	log.Printf("✅ Processed order: %d", msg.OrderID)

	createPayment(msg.OrderID)
	return nil
}

func retryMessage(msg dto.CreateOrderMessage) {
	body, _ := json.Marshal(msg)
	err := configs.RMQChannel.Publish(
		"",
		configs.OrderRetryQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Println("Failed to retry publish:", err)
	}
}

func createPayment(orderID uint) {
	var order models.Order
	db := database.DB
	if err := db.Where("id= ?", orderID).
		Preload("OrderItems").
		Preload("OrderItems.Variant").
		Preload("OrderItems.Variant.Product").
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Print("Order not found to create payment link")
			return
		}
		log.Printf(err.Error(), "while creating payment link")
		return
	}

	paymentLink := ""
	if order.PaymentMethod == models.PaymentMethodBank {
		paymentData, err := serviceImpl.CreatePayOSPayment(int(order.ID), order.Total+order.ShippingFee, MapOrderItemsToPayOSItems(order), "Thanh toán đơn hàng", "returnURL", "cancelURL")
		if err != nil {
			log.Print(err.Error(), "while creating payment link")
			return
		}
		paymentLink = paymentData.CheckoutUrl
	}

	var paymentInfo = &models.PaymentInfo{
		Amount:      order.Total,
		Status:      models.PaymentPending,
		PaymentLink: paymentLink,
	}
	if err := db.Create(&paymentInfo).Error; err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return
	}

	order.PaymentInfo = paymentInfo
	if err := db.Save(&order).Error; err != nil {
		log.Printf(err.Error(), "while saving payment info to order")
	}
	return
}

func MapOrderItemsToPayOSItems(order models.Order) []payos.Item {
	var items []payos.Item

	for _, oi := range order.OrderItems {
		item := payos.Item{
			Name:     fmt.Sprintf("%s (Variant #%d)", oi.Variant.Product.Name, oi.ProductVariantID),
			Price:    int(oi.Price),
			Quantity: int(oi.Quantity),
		}
		items = append(items, item)
	}

	items = append(items, payos.Item{
		Name:     "Shipping Fee",
		Price:    int(order.ShippingFee),
		Quantity: 1,
	})
	return items
}
