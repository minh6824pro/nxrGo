package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, orderModule *modules.OrderModule) {

	order := rg.Group("/orders")
	order.Use(orderModule.AuthMiddleware.RequireAuth())
	{
		order.POST("", orderModule.Controller.Create)
		order.GET("/:id", orderModule.Controller.GetById)
		order.POST("/updatedb", orderModule.Controller.UpdateDb)
		order.GET("/status", orderModule.Controller.GetByStatus)
		order.GET("", orderModule.Controller.List)
		order.POST("/changepaymentmethod", orderModule.Controller.ChangePaymentMethod)
		order.GET("/shippingFee", orderModule.Controller.GetShippingFee)
		order.GET("/mockpayos/:id", orderModule.Controller.PaymentSuccessMock)

	}
	order.Use(orderModule.AuthMiddleware.RequireRole(models.RoleAdmin))
	{
		order.PATCH("/:id", orderModule.Controller.UpdateOrderStatus)
		order.GET("/admin", orderModule.Controller.ListByAdmin)
	}

}
