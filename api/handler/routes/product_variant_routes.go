package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

func RegisterProductVariantRoutes(rg *gin.RouterGroup, productVariantModule *modules.ProductVariantModule) {

	productVariants := rg.Group("/product_variants")
	{
		productVariants.POST("", productVariantModule.Controller.Create)
		productVariants.PATCH("/:id/increase_stock", productVariantModule.Controller.IncreaseStock)
		productVariants.PATCH("/:id/decrease_stock", productVariantModule.Controller.DecreaseStock)
		productVariants.POST("/listbyids", productVariantModule.Controller.ListByIds)
	}

}
