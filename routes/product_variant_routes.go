package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/event"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterProductVariantRoutes(rg *gin.RouterGroup, db *gorm.DB, productVariantCache cache.ProductVariantRedis, updateStockAgg *event.UpdateStockAggregator) {
	productVariantRepo := repoImpl.NewProductVariantGormRepository(db)
	productRepo := repoImpl.NewProductGormRepository(db)
	productVariantService := serviceImpl.NewProductVariantService(productRepo, productVariantRepo, productVariantCache, updateStockAgg)
	productVariantController := controllers.NewProductVariantController(productVariantService)

	productVariants := rg.Group("/product_variants")
	{
		productVariants.POST("/", productVariantController.Create)
		productVariants.PATCH("/:id/increase_stock", productVariantController.IncreaseStock)
		productVariants.PATCH("/:id/decrease_stock", productVariantController.DecreaseStock)

	}

}
