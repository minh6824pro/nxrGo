package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

func RegisterPayOSRoutes(router *gin.RouterGroup, payOSModule *modules.PayOsModule) {

	payos := router.Group("/payos")
	{
		payos.POST("/webhook", payOSModule.Controller.HandleWebhook)
	}
}
