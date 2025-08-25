package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/modules"
)

//func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController, authMiddleware *middleware.AuthMiddleware) {

func RegisterAuthRoutes(rg *gin.RouterGroup, authModule *modules.AuthModule) {

	auth := rg.Group("/auth")
	{
		auth.POST("/register", authModule.AuthController.Register)
		auth.POST("/login", authModule.AuthController.Login)
		auth.POST("/refresh", authModule.AuthController.RefreshToken)
	}

	// Protected routes
	protected := rg.Group("/user")
	protected.Use(authModule.AuthMiddleware.RequireAuth())
	{
		protected.GET("/profile", authModule.AuthController.GetProfile)

	}

	// Admin only routes
	admin := rg.Group("/admin")
	admin.Use(authModule.AuthMiddleware.RequireAuth())
	admin.Use(authModule.AuthMiddleware.RequireRole(models.RoleAdmin))
	{
		// Thêm các routes admin ở đây
	}
}
