// main.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/database"
	"github.com/minh6824pro/nxrGO/routes"
	"log"
	"time"
)

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
	routes.RegisterPayOSRoutes(api, db)

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
	select {}
}
