package middleware

import (
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService *jwt.JWTService
}

func NewAuthMiddleware(jwtService *jwt.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth middleware xác thực JWT
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			customErr.WriteError(c,
				customErr.NewError(
					customErr.UNAUTHORIZED,
					"Authorization header required",
					http.StatusUnauthorized,
					nil))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			customErr.WriteError(c,
				customErr.NewError(
					customErr.UNAUTHORIZED,
					"Bearer token required",
					http.StatusUnauthorized,
					nil))
			c.Abort()
			return
		}

		claims, err := a.jwtService.ValidateToken(tokenString)
		if err != nil {
			customErr.WriteError(c,
				customErr.NewError(
					customErr.UNAUTHORIZED,
					"Invalid token",
					http.StatusUnauthorized,
					nil))
			c.Abort()
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole middleware kiểm tra quyền
func (a *AuthMiddleware) RequireRole(role models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			customErr.WriteError(c,
				customErr.NewError(
					customErr.UNAUTHORIZED,
					"Unauthorized",
					http.StatusUnauthorized,
					nil))
			c.Abort()
			return
		}

		if userRole != role {
			customErr.WriteError(c,
				customErr.NewError(
					customErr.FORBIDDEN,
					"Insufficient permissions",
					http.StatusForbidden,
					nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
