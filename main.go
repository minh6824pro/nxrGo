// main.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/database"
	"github.com/minh6824pro/nxrGO/docs"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/routes"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"log"
	"time"
)

// @title           nxrGO
// @version         1.0
// @description     This is an ecommerce API server
// @host            localhost:8080
// @BasePath        /api
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {

	// Set Time location
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatalf("cannot load location: %v", err)
	}
	time.Local = loc

	// Load env var
	configs.LoadEnv()

	// Connect & auto create DB
	database.ConnectDatabase()
	//configs.AutoMigrate()

	db := database.DB

	configs.InitRedis()
	//configs.InitRabbitMQ()
	//defer configs.CloseRabbitMQ()
	//
	//consumers.StartOrderConsumer()

	//orderRepo := repoImpl.NewOrderGormRepository(db)
	//consumers.ConsumeOrderDLQ(orderRepo)

	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})
	eventPub := event.NewChannelEventPublisher()

	api := r.Group("/api")

	// Register auth routes FIRST
	routes.RegisterAuthRoutes(api, db)

	// Existing routes
	routes.RegisterMerchantRoutes(api, db)
	routes.RegisterBrandRoutes(api, db)
	routes.RegisterCategoryRoutes(api, db)
	routes.RegisterProductRoutes(api, db)
	routes.RegisterVariantRoutes(api, db)
	routes.RegisterOrderRoutes(api, db, eventPub)
	routes.RegisterPayOSRoutes(api, db)

	// setup swagger info
	docs.SwaggerInfo.Title = "nxrGO"
	docs.SwaggerInfo.Description = "This is an ecommerce API server"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	ready := make(chan bool)

	go func() {
		go func() {
			time.Sleep(2 * time.Second)
			ready <- true
		}()

		if err := r.Run(":8080"); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	<-ready
	fmt.Println("Server ready, init PayOS...")
	configs.InitPayOS()

	// Publish payOS payment event not yet handle because of app crash
	go func() {
		err := processPendingDraftOrders(db, eventPub)
		if err != nil {
			log.Printf("Error processing pending draft orders: %v", err)
		}
	}()
	select {}
}

func processPendingDraftOrders(db *gorm.DB, eventPub *event.ChannelEventPublisher) error {
	var draftOrders []models.DraftOrder

	err := db.Model(&models.DraftOrder{}).
		Joins("JOIN payment_infos ON payment_infos.id = draft_orders.payment_info_id").
		Where("draft_orders.payment_method = ? AND payment_infos.status = ?", "BANK", "PENDING").
		Preload("PaymentInfo").
		Find(&draftOrders).Error
	if err != nil {
		return err
	}
	for _, d := range draftOrders {
		payOSEvent := event.PayOSPaymentCreatedEvent{
			DraftOrderID:  d.ID,
			PaymentLink:   d.PaymentInfo.PaymentLink,
			Amount:        d.PaymentInfo.Amount,
			PaymentMethod: string(d.PaymentMethod),
			CreatedAt:     time.Now(),
		}

		err := eventPub.PublishPaymentCreated(payOSEvent)
		if err != nil {
			log.Printf("Error publishing draft order: %v", err)
			continue
		}
		log.Println("draft order published")

	}

	return nil
}
