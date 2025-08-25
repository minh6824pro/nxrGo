package modules

import (
	"github.com/minh6824pro/nxrGO/api/handler/controllers"
	"github.com/minh6824pro/nxrGO/api/middleware"
)

type AuthModule struct {
	AuthController *controllers.AuthController
	AuthMiddleware *middleware.AuthMiddleware
}
