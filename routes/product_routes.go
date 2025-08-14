package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/event"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterProductRoutes(rg *gin.RouterGroup, db *gorm.DB, redisClient *redis.Client, updateStockAgg *event.UpdateStockAggregator) {
	productRepo := repoImpl.NewProductGormRepository(db)
	merchantRepo := repoImpl.NewMerchantGormRepository(db)
	categoryRepo := repoImpl.NewCategoryGormRepository(db)
	brandRepo := repoImpl.NewBrandGormRepository(db)
	productVariantRepo := repoImpl.NewProductVariantGormRepository(db)
	variantOptionRepo := repoImpl.NewVariantOptionGormRepository(db)
	variantOptionValueRepo := repoImpl.NewVariantOptionValueGormRepository(db)

	productVariantCache := cache.NewProductVariantRedisService(configs.RedisClient, configs.RedisCtx, productVariantRepo)
	productCache := cache.NewProductCacheService(redisClient, productRepo, productVariantCache)
	productVariantService := serviceImpl.NewProductVariantService(productRepo, productVariantRepo, productVariantCache, updateStockAgg)
	productService := serviceImpl.NewProductService(db, productRepo, brandRepo, merchantRepo, categoryRepo, productVariantRepo, variantOptionValueRepo, variantOptionRepo, productCache, productVariantService)

	productController := controllers.NewProductController(productService)

	product := rg.Group("/products")
	{
		product.POST("", productController.Create)
		product.GET("", productController.List)
		product.GET("/:id", productController.GetByID)
		product.DELETE("/:id", productController.Delete)
		product.GET("/query/", productController.ListProductQuery)
	}
}
