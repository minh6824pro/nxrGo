package modules

import (
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/middleware"
)

type AuthModule struct {
	AuthController *controllers.AuthController
	AuthMiddleware *middleware.AuthMiddleware
}
