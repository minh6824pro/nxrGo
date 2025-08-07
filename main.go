// main.go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/consumers"
	"github.com/minh6824pro/nxrGO/database"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	"github.com/minh6824pro/nxrGO/routes"
	"log"
	"time"
)

func main() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatalf("cannot load location: %v", err)
	}
	time.Local = loc

	configs.LoadEnv()
	database.ConnectDatabase()

	// Auto
	//configs.AutoMigrate()

	db := database.DB

	configs.InitRabbitMQ()
	defer configs.CloseRabbitMQ()

	consumers.StartOrderConsumer()

	orderRepo := repoImpl.NewOrderGormRepository(db)
	consumers.ConsumeOrderDLQ(orderRepo)

	configs.InitPayOS()

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

	api := r.Group("/api")

	// Register auth routes FIRST
	routes.RegisterAuthRoutes(api, db)

	// Existing routes
	routes.RegisterMerchantRoutes(api, db)
	routes.RegisterBrandRoutes(api, db)
	routes.RegisterCategoryRoutes(api, db)
	routes.RegisterProductRoutes(api, db)
	routes.RegisterVariantRoutes(api, db)
	routes.RegisterOrderRoutes(api, db)

	r.Run(":8080")
}
