package modules

import (
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/middleware"
	"github.com/minh6824pro/nxrGO/services"
)

type OrderModule struct {
	Controller                 *controllers.OrderController
	Service                    services.OrderService
	AuthMiddleware             *middleware.AuthMiddleware
	ProductVariantRedisService cache.ProductVariantRedis
}
