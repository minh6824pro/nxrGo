package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/modules"
)

func RegisterMerchantRoutes(rg *gin.RouterGroup, merchantModule *modules.MerchantModule) {

	merchant := rg.Group("/merchants")
	{
		merchant.POST("", merchantModule.Controller.Create)
		merchant.GET("", merchantModule.Controller.List)
		merchant.GET("/:id", merchantModule.Controller.GetByID)
		merchant.DELETE("/:id", merchantModule.Controller.Delete)
		merchant.PATCH("/:id", merchantModule.Controller.Patch)
	}
}
