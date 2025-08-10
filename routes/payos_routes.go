package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
)

func RegisterPayOSRoutes(router *gin.RouterGroup) {
	controller := controllers.NewWebhookController()
	payos := router.Group("/payos")
	{
		payos.POST("/webhook", controller.HandleWebhook)
	}
}
