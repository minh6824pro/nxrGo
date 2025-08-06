// main.go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/config"
	"github.com/minh6824pro/nxrGO/database"
	"github.com/minh6824pro/nxrGO/routes"
)

func main() {

	config.LoadEnv()
	database.ConnectDatabase()

	//config.AutoMigrate()

	r := gin.Default()

	api := r.Group("/api")

	db := database.DB

	routes.RegisterMerchantRoutes(api, db)
	routes.RegisterBrandRoutes(api, db)
	routes.RegisterCategoryRoutes(api, db)
	routes.RegisterProductRoutes(api, db)
	r.Run(":8080")
}
