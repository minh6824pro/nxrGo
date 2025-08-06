package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterVariantRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	repo := repoImpl.NewVariantOptionGormRepository(db)
	service := serviceImpl.NewVariantOptionService(repo)
	controller := controllers.NewVariantOptionController(service)

	variantOption := rg.Group("/variant_options")
	{
		variantOption.GET("", controller.List)
		variantOption.POST("", controller.Create)
		variantOption.GET("/:id", controller.GetByID)
		variantOption.DELETE("/:id", controller.Delete)
		variantOption.PATCH("/:id", controller.Patch)
	}
}
