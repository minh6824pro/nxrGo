package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/middleware"
	"github.com/minh6824pro/nxrGO/models"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	"github.com/minh6824pro/nxrGO/services"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, db *gorm.DB, productVariantCache cache.ProductVariantRedis,
	eventPub *event.ChannelEventPublisher, updateStockAgg *event.UpdateStockAggregator) services.OrderService {
	productVariantRepo := repoImpl.NewProductVariantGormRepository(db)
	orderItemRepo := repoImpl.NewOrderItemGormRepository(db)
	orderRepo := repoImpl.NewOrderGormRepository(db)
	paymentInfoRepo := repoImpl.NewPaymentInfoGormImpl(db)
	draftOrderRepo := repoImpl.NewDraftOrderGormRepository(db)
	orderService := serviceImpl.NewOrderService(db, productVariantRepo, orderItemRepo, orderRepo, draftOrderRepo, paymentInfoRepo, productVariantCache, eventPub, updateStockAgg)
	orderController := controllers.NewOrderController(orderService)
	jwtService := jwt.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	order := rg.Group("/orders")
	order.Use(authMiddleware.RequireAuth())
	{
		order.POST("", orderController.Create)
		order.GET("/:id", orderController.GetById)
		order.POST("/updatedb", orderController.UpdateDb)
		order.GET("/status", orderController.GetByStatus)
		order.POST("/rebuy/:id", orderController.ReBuy)
		order.GET("", orderController.List)
		order.POST("/changepaymentmethod", orderController.ChangePaymentMethod)

	}
	order.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		order.PATCH("/:id", orderController.UpdateOrderStatus)
	}

	return orderService
}
