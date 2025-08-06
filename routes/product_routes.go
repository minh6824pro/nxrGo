package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterProductRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	productRepo := repoImpl.NewProductGormRepository(db)
	merchantRepo := repoImpl.NewMerchantGormRepository(db)
	categoryRepo := repoImpl.NewCategoryGormRepository(db)
	brandRepo := repoImpl.NewBrandGormRepository(db)
	productVariantRepo := repoImpl.NewProductVariantGormRepository(db)
	variantOptionRepo := repoImpl.NewVariantOptionGormRepository(db)
	variantOptionValueRepo := repoImpl.NewVariantOptionValueGormRepository(db)

	productService := serviceImpl.NewProductService(db, productRepo, brandRepo, merchantRepo, categoryRepo, productVariantRepo, variantOptionValueRepo, variantOptionRepo)

	productController := controllers.NewProductController(productService)

	product := rg.Group("/products")
	{
		product.POST("", productController.Create)
		product.GET("", productController.List)
		product.GET("/:id", productController.GetByID)
		product.DELETE("/:id", productController.Delete)

	}
}
