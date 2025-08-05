package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterCategoryRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	repo := repoImpl.NewCategoryGormRepository(db)
	service := serviceImpl.NewCategoryService(repo)
	controller := controllers.NewCategoryController(service)

	category := rg.Group("/categories")
	{
		category.POST("", controller.Create)
		category.GET("/:id", controller.GetByID)
		category.DELETE("/:id", controller.Delete)
		category.GET("", controller.List)
		category.PATCH("/:id", controller.Patch)
	}

}
