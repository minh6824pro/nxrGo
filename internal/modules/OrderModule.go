package modules

import (
	"github.com/minh6824pro/nxrGO/api/handler/controllers"
	"github.com/minh6824pro/nxrGO/api/middleware"
	"github.com/minh6824pro/nxrGO/internal/cache"
	"github.com/minh6824pro/nxrGO/internal/services"
)

type OrderModule struct {
	Controller                 *controllers.OrderController
	Service                    services.OrderService
	AuthMiddleware             *middleware.AuthMiddleware
	ProductVariantRedisService cache.ProductVariantRedis
}
