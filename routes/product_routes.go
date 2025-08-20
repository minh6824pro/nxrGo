package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/modules"
)

func RegisterProductRoutes(rg *gin.RouterGroup, productModule *modules.ProductModule) {

	product := rg.Group("/products")
	{
		product.POST("", productModule.Controller.Create)
		product.GET("", productModule.Controller.List)
		product.GET("/:id", productModule.Controller.GetByID)
		product.DELETE("/:id", productModule.Controller.Delete)
		product.GET("/query", productModule.Controller.ListProductQuery)
		product.GET("/admin", productModule.Controller.ListProductManagement)

	}
}
