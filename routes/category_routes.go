package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/modules"
)

func RegisterCategoryRoutes(rg *gin.RouterGroup, categoryModule *modules.CategoryModule) {

	category := rg.Group("/categories")
	{
		category.POST("", categoryModule.Controller.Create)
		category.GET("/:id", categoryModule.Controller.GetByID)
		category.DELETE("/:id", categoryModule.Controller.Delete)
		category.GET("", categoryModule.Controller.List)
		category.PATCH("/:id", categoryModule.Controller.Patch)
	}

}
