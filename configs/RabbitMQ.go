package configs

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RMQConn *amqp.Connection
var RMQChannel *amqp.Channel

const (
	OrderCreateQueue = "order.create"
	OrderRetryQueue  = "order.retry"
	OrderDLQ         = "order.dlq"
	OrderReplyQueue  = "order.reply" // Thêm reply queue
)

func InitRabbitMQ() {
	var err error
	url := os.Getenv("RABBITMQ_URL") // amqp://guest:guest@localhost:5672/
	RMQConn, err = amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	RMQChannel, err = RMQConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	DeclareOrderQueues()
}

func DeclareOrderQueues() {
	// DLQ
	_, _ = RMQChannel.QueueDeclare(OrderDLQ, true, false, false, false, nil)

	// Retry queue with TTL + DLX back to order.create
	_, _ = RMQChannel.QueueDeclare(OrderRetryQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": OrderCreateQueue,
		"x-message-ttl":             int32(5000), // 5s delay
	})

	// Main processing queue with DLX to DLQ
	_, _ = RMQChannel.QueueDeclare(OrderCreateQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": OrderDLQ,
	})

	// Reply queue để nhận kết quả từ consumer
	_, _ = RMQChannel.QueueDeclare(OrderReplyQueue, true, false, false, false, nil)

	log.Printf("Success to open RabbitMQ channel")
}

// Đóng kết nối khi ứng dụng shutdown
func CloseRabbitMQ() {
	if RMQChannel != nil {
		if err := RMQChannel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if RMQConn != nil {
		if err := RMQConn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
	log.Println("🛑 RabbitMQ connection closed.")
}
