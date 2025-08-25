package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

func RegisterVariantRoutes(rg *gin.RouterGroup, variantModule *modules.VariantModule) {

	variantOption := rg.Group("/variant_options")
	{
		variantOption.GET("", variantModule.Controller.List)
		variantOption.POST("", variantModule.Controller.Create)
		variantOption.GET("/:id", variantModule.Controller.GetByID)
		variantOption.DELETE("/:id", variantModule.Controller.Delete)
		variantOption.PATCH("/:id", variantModule.Controller.Patch)
	}
}
