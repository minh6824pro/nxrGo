package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/middleware"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	productVariantRepo := repoImpl.NewProductVariantGormRepository(db)
	orderItemRepo := repoImpl.NewOrderItemGormRepository(db)
	orderRepo := repoImpl.NewOrderGormRepository(db)
	orderService := serviceImpl.NewOrderService(db, productVariantRepo, orderItemRepo, orderRepo)
	orderController := controllers.NewOrderController(orderService)

	jwtService := jwt.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	order := rg.Group("/orders")
	order.Use(authMiddleware.RequireAuth())
	{
		order.POST("", orderController.Create)
		order.GET("/:id", orderController.GetById)
	}
}
