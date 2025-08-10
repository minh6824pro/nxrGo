package consumers

import (
	"context"
	"encoding/json"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/repositories"
	"log"
	"time"
)

func ConsumeOrderDLQ(orderRepo repositories.OrderRepository) {
	msgs, err := configs.RMQChannel.Consume(
		configs.OrderDLQ, // Tên DLQ
		"",               // consumer
		true,             // auto-ack
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("❌ Failed to consume from DLQ: %v", err)
	}
	log.Printf("Successfully consumed from DLQ")

	go func() {
		for msg := range msgs {
			var failedMsg dto.CreateOrderMessage
			if err := json.Unmarshal(msg.Body, &failedMsg); err != nil {
				log.Println("❌ Failed to unmarshal DLQ message:", err)
				continue
			}
			log.Printf("🪦 DLQ Received - DraftOrderID: %d", failedMsg.OrderID)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := orderRepo.Delete(ctx, failedMsg.OrderID)
			if err != nil {
				parseErr := customErr.ParseError(err)
				log.Println("Order DLQ can't delete", parseErr.Message, parseErr.Code)
				return
			}

			log.Printf("🪦 DLQ Order Deleted - DraftOrderID: %d", failedMsg.OrderID)

		}
	}()
}
