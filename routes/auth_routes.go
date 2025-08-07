package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/middleware"
	"github.com/minh6824pro/nxrGO/models"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"gorm.io/gorm"
)

//func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController, authMiddleware *middleware.AuthMiddleware) {

func RegisterAuthRoutes(rg *gin.RouterGroup, db *gorm.DB) {

	repo := repoImpl.NewAuthRepository(db)
	jwtService := jwt.NewJWTService()
	service := serviceImpl.NewAuthService(repo, jwtService)
	authController := controllers.NewAuthController(service)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	auth := rg.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
	}

	// Protected routes
	protected := rg.Group("/user")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.GET("/profile", authController.GetProfile)

	}

	// Admin only routes
	admin := rg.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		// Thêm các routes admin ở đây
	}
}
