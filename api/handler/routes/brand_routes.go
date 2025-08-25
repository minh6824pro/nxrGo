package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

func RegisterBrandRoutes(rg *gin.RouterGroup, brandModule *modules.BrandModule) {

	brand := rg.Group("/brands")
	{
		brand.POST("", brandModule.Controller.Create)
		brand.GET("/:id", brandModule.Controller.GetByID)
		brand.DELETE("/:id", brandModule.Controller.Delete)
		brand.GET("", brandModule.Controller.List)
		brand.PATCH("/:id", brandModule.Controller.Patch)
	}
}
