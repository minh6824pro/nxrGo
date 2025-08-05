package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterBrandRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	// Setup Repository, Service, Controller
	repo := repoImpl.NewBrandGormRepository(db)
	service := serviceImpl.NewBrandService(repo)
	controller := controllers.NewBrandController(service)

	brand := rg.Group("/brands")
	{
		brand.POST("", controller.Create)
		brand.GET("/:id", controller.GetByID)
		brand.DELETE("/:id", controller.Delete)
		brand.GET("", controller.List)
		brand.PATCH("/:id", controller.Patch)
	}
}
