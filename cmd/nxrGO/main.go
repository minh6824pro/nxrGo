// main.go
package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/api/handler/routes"
	"github.com/minh6824pro/nxrGO/docs"
	"github.com/minh6824pro/nxrGO/internal/cache"
	"github.com/minh6824pro/nxrGO/internal/config"
	"github.com/minh6824pro/nxrGO/internal/database"
	"github.com/minh6824pro/nxrGO/internal/elastic"
	"github.com/minh6824pro/nxrGO/internal/event"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/wire"
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
	config.LoadEnv()

	// Connect & auto create DB
	database.ConnectDatabase()
	config.AutoMigrate()

	db := database.DB

	// Init snowflake id
	config.GetSnowflakeNode()
	// Create cache
	config.InitRedis()
	// Init elastic
	config.InitElastic()
	elasticClient := elastic.NewElasticClient()
	err = elasticClient.EnsureProductIndex(context.Background())
	if err != nil {
		log.Println(err)
	}
	// Convert DB to Elastic Document
	elasticRepo := elastic.NewProductElasticRepo()
	elasticRepo.DBToElastic(context.Background())

	// Init necessary dependency
	eventPub := event.NewChannelEventPublisher()
	updateStockAgg := event.NewUpdateStockAggregator()

	//configs.InitRabbitMQ()
	//defer configs.CloseRabbitMQ()
	//
	//consumers.StartOrderConsumer()

	//orderRepo := repoImpl.NewOrderGormRepository(db)
	//consumers.ConsumeOrderDLQ(orderRepo)

	r := gin.Default()

	// Add CORS middleware

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")

	auth := wire.InitAuthModule(db)
	merchant := wire.InitMerchantModule(db)
	brand := wire.InitBrandModule(db)
	category := wire.InitCategoryModule(db)
	product := wire.InitProductModule(db, config.RedisClient, config.EsClient, config.RedisCtx, updateStockAgg)
	variant := wire.InitVariantModule(db)
	order := wire.InitOrderModule(db, config.RedisClient, config.RedisCtx, eventPub, updateStockAgg)
	productVariant := wire.InitProductVariantModule(db, config.RedisClient, config.RedisCtx, updateStockAgg)
	payOsModule := wire.InitPayOSModule(db, config.RedisClient, config.RedisCtx, eventPub, updateStockAgg)
	// Register auth routes FIRST
	routes.RegisterAuthRoutes(api, auth)

	// Existing routes
	routes.RegisterMerchantRoutes(api, merchant)
	routes.RegisterBrandRoutes(api, brand)
	routes.RegisterCategoryRoutes(api, category)
	routes.RegisterProductRoutes(api, product)
	routes.RegisterVariantRoutes(api, variant)
	routes.RegisterOrderRoutes(api, order)
	routes.RegisterPayOSRoutes(api, payOsModule)
	routes.RegisterProductVariantRoutes(api, productVariant)
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

	// Go routine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			data := updateStockAgg.Flush()
			log.Println("Flushed:", data)
			updateStocks(db, data, order.ProductVariantRedisService)
			err := order.Service.UpdateQuantity(context.Background())
			if err != nil {
				return
			}

		}
	}()

	<-ready
	fmt.Println("Server ready, init PayOS...")
	config.InitPayOS()

	// Publish payOS payment event not yet handle because of app crash
	go func() {
		err := processPendingDraftOrders(db, eventPub)
		if err != nil {
			log.Printf("Error processing pending draft orders: %v", err)
		}
	}()

	select {}
}

func updateStocks(db *gorm.DB, data map[uint]int, productVariantCache cache.ProductVariantRedis) {
	for key, value := range data {
		err := db.Model(&models.ProductVariant{}).
			Where("id = ?", key).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", value)).
			Error
		if err != nil {
			log.Printf("Error updating product variant: %d: %v", key, err)
		}
		err = productVariantCache.DeleteProductVariantHash(key)
		if err != nil {
			log.Printf("Error deleting product variant cache: %v", err)
		}
	}
	log.Print("Update stocks successfully")
}

func processPendingDraftOrders(db *gorm.DB, eventPub *event.ChannelEventPublisher) error {
	var paymentInfos []models.PaymentInfo

	latestSub := db.
		Table("payment_infos").
		Select("order_id, MAX(created_at) as max_created_at").
		//	Where("order_type = ?", "draft_order").
		Where("status = ?", "PENDING").
		Where("payment_link <> ?", "").
		Group("order_id")

	err := db.
		Table("payment_infos p").
		Joins("JOIN (?) latest ON p.order_id = latest.order_id AND p.created_at = latest.max_created_at", latestSub).
		Scan(&paymentInfos).Error
	if err != nil {
		log.Printf("Error scanning pending bank payment: %v", err)
		return err
	}
	for _, p := range paymentInfos {
		payOSEvent := event.PayOSPaymentCreatedEvent{
			Id:            p.ID,
			OrderID:       p.OrderID,
			PaymentLink:   p.PaymentLink,
			Total:         p.Total,
			PaymentMethod: "BANK",
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
