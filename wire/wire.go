//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"github.com/google/wire"
	controllers2 "github.com/minh6824pro/nxrGO/api/handler/controllers"
	"github.com/minh6824pro/nxrGO/api/middleware"
	cache2 "github.com/minh6824pro/nxrGO/internal/cache"
	event2 "github.com/minh6824pro/nxrGO/internal/event"
	"github.com/minh6824pro/nxrGO/internal/jwt"
	modules2 "github.com/minh6824pro/nxrGO/internal/modules"
	"github.com/minh6824pro/nxrGO/internal/repositories/impl"
	impl2 "github.com/minh6824pro/nxrGO/internal/services/impl"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitAuthModule(db *gorm.DB) *modules2.AuthModule {
	wire.Build(
		impl.NewAuthRepository,
		impl2.NewAuthService,
		jwt.NewJWTService,
		middleware.NewAuthMiddleware,
		controllers2.NewAuthController,
		wire.Struct(new(modules2.AuthModule), "*"))
	return nil
}

func InitMerchantModule(db *gorm.DB) *modules2.MerchantModule {
	wire.Build(
		impl.NewMerchantGormRepository,
		impl2.NewMerchantService,
		controllers2.NewMerchantController,
		wire.Struct(new(modules2.MerchantModule), "*"))
	return nil
}

func InitBrandModule(db *gorm.DB) *modules2.BrandModule {
	wire.Build(
		impl.NewBrandGormRepository,
		impl2.NewBrandService,
		controllers2.NewBrandController,
		wire.Struct(new(modules2.BrandModule), "*"))
	return nil
}

func InitCategoryModule(db *gorm.DB) *modules2.CategoryModule {
	wire.Build(
		impl.NewCategoryGormRepository,
		impl2.NewCategoryService,
		controllers2.NewCategoryController,
		wire.Struct(new(modules2.CategoryModule), "*"))
	return nil
}

func InitProductModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, updateStockAgg *event2.UpdateStockAggregator) *modules2.ProductModule {
	wire.Build(
		impl.NewProductGormRepository,
		impl.NewMerchantGormRepository,
		impl.NewCategoryGormRepository,
		impl.NewBrandGormRepository,
		impl.NewProductVariantGormRepository,
		impl.NewVariantOptionGormRepository,
		impl.NewVariantOptionValueGormRepository,
		cache2.NewProductVariantRedisService,
		cache2.NewProductCacheService,
		impl2.NewProductVariantService,
		impl2.NewProductService,
		controllers2.NewProductController,
		wire.Struct(new(modules2.ProductModule), "*"))
	return nil
}

func InitVariantModule(db *gorm.DB) *modules2.VariantModule {
	wire.Build(
		impl.NewVariantOptionGormRepository,
		impl2.NewVariantOptionService,
		controllers2.NewVariantOptionController,
		wire.Struct(new(modules2.VariantModule), "*"))
	return nil
}

func InitOrderModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, eventBus event2.EventPublisher, updateStockAgg *event2.UpdateStockAggregator) *modules2.OrderModule {
	wire.Build(
		impl.NewProductVariantGormRepository,
		impl.NewOrderItemGormRepository,
		impl.NewOrderGormRepository,
		impl.NewDraftOrderGormRepository,
		impl.NewPaymentInfoGormImpl,
		impl.NewMerchantGormRepository,
		cache2.NewProductVariantRedisService,
		impl2.NewOrderService,
		controllers2.NewOrderController,
		jwt.NewJWTService,
		middleware.NewAuthMiddleware,
		wire.Struct(new(modules2.OrderModule), "*"))
	return nil
}

func InitProductVariantModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, updateStockAgg *event2.UpdateStockAggregator) *modules2.ProductVariantModule {
	wire.Build(
		impl.NewProductVariantGormRepository,
		impl.NewProductGormRepository,
		cache2.NewProductVariantRedisService,
		impl2.NewProductVariantService,
		controllers2.NewProductVariantController,
		wire.Struct(new(modules2.ProductVariantModule), "*"))

	return nil
}

func InitPayOSModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, eventBus event2.EventPublisher, updateStockAgg *event2.UpdateStockAggregator) *modules2.PayOsModule {
	wire.Build(
		impl.NewProductVariantGormRepository,
		impl.NewOrderItemGormRepository,
		impl.NewOrderGormRepository,
		impl.NewPaymentInfoGormImpl,
		impl.NewMerchantGormRepository,
		impl.NewDraftOrderGormRepository,
		cache2.NewProductVariantRedisService,
		impl2.NewOrderService,
		controllers2.NewWebhookController,
		wire.Struct(new(modules2.PayOsModule), "*"))
	return nil
}
