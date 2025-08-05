package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

func RegisterMerchantRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	repo := repoImpl.NewMerchantGormRepository(db)
	service := serviceImpl.NewMerchantService(repo)
	controller := controllers.NewMerchantController(service)

	merchant := rg.Group("/merchants")
	{
		merchant.POST("", controller.Create)
		merchant.GET("", controller.List)
		merchant.GET("/:id", controller.GetByID)
		merchant.DELETE("/:id", controller.Delete)
		merchant.PATCH("/:id", controller.Patch)
	}
}
