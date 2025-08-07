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

			err := processOrderMessage(orderMsg)
			if err != nil {
				orderMsg.RetryCount++
				if orderMsg.RetryCount > 3 {
					log.Printf("⚠️ Dropping order %d after 3 retries -> send to DLQ", orderMsg.OrderID)
					d.Nack(false, false) // 👈 gửi sang DLQ
				} else {
					retryMessage(orderMsg)
					d.Ack(false) // đã retry nên vẫn ack
				}
			} else {
				d.Ack(false) // xử lý thành công
			}
		}
	}()
}

func processOrderMessage(msg dto.CreateOrderMessage) error {
	db := database.DB
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, item := range msg.Items {
		var pv models.ProductVariant
		if err := tx.First(&pv, item.ProductVariantID).Error; err != nil || pv.Quantity < item.Quantity {
			tx.Rollback()
			return err // báo lỗi để retry hoặc gửi DLQ
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
		//tx.Rollback()
		//return customErr.NewError(customErr.VERSION_CONFLICT, "Version Conflict", http.StatusConflict, nil)
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
			Price:    int(oi.Price), // đảm bảo giá là số nguyên VND
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
