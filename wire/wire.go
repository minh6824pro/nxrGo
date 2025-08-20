//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"github.com/google/wire"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/controllers"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/middleware"
	"github.com/minh6824pro/nxrGO/modules"
	repoImpl "github.com/minh6824pro/nxrGO/repositories/impl"
	serviceImpl "github.com/minh6824pro/nxrGO/services/impl"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitAuthModule(db *gorm.DB) *modules.AuthModule {
	wire.Build(
		repoImpl.NewAuthRepository,
		serviceImpl.NewAuthService,
		jwt.NewJWTService,
		middleware.NewAuthMiddleware,
		controllers.NewAuthController,
		wire.Struct(new(modules.AuthModule), "*"))
	return nil
}

func InitMerchantModule(db *gorm.DB) *modules.MerchantModule {
	wire.Build(
		repoImpl.NewMerchantGormRepository,
		serviceImpl.NewMerchantService,
		controllers.NewMerchantController,
		wire.Struct(new(modules.MerchantModule), "*"))
	return nil
}

func InitBrandModule(db *gorm.DB) *modules.BrandModule {
	wire.Build(
		repoImpl.NewBrandGormRepository,
		serviceImpl.NewBrandService,
		controllers.NewBrandController,
		wire.Struct(new(modules.BrandModule), "*"))
	return nil
}

func InitCategoryModule(db *gorm.DB) *modules.CategoryModule {
	wire.Build(
		repoImpl.NewCategoryGormRepository,
		serviceImpl.NewCategoryService,
		controllers.NewCategoryController,
		wire.Struct(new(modules.CategoryModule), "*"))
	return nil
}

func InitProductModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, updateStockAgg *event.UpdateStockAggregator) *modules.ProductModule {
	wire.Build(
		repoImpl.NewProductGormRepository,
		repoImpl.NewMerchantGormRepository,
		repoImpl.NewCategoryGormRepository,
		repoImpl.NewBrandGormRepository,
		repoImpl.NewProductVariantGormRepository,
		repoImpl.NewVariantOptionGormRepository,
		repoImpl.NewVariantOptionValueGormRepository,
		cache.NewProductVariantRedisService,
		cache.NewProductCacheService,
		serviceImpl.NewProductVariantService,
		serviceImpl.NewProductService,
		controllers.NewProductController,
		wire.Struct(new(modules.ProductModule), "*"))
	return nil
}

func InitVariantModule(db *gorm.DB) *modules.VariantModule {
	wire.Build(
		repoImpl.NewVariantOptionGormRepository,
		serviceImpl.NewVariantOptionService,
		controllers.NewVariantOptionController,
		wire.Struct(new(modules.VariantModule), "*"))
	return nil
}

func InitOrderModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, eventBus event.EventPublisher, updateStockAgg *event.UpdateStockAggregator) *modules.OrderModule {
	wire.Build(
		repoImpl.NewProductVariantGormRepository,
		repoImpl.NewOrderItemGormRepository,
		repoImpl.NewOrderGormRepository,
		repoImpl.NewDraftOrderGormRepository,
		repoImpl.NewPaymentInfoGormImpl,
		cache.NewProductVariantRedisService,
		serviceImpl.NewOrderService,
		controllers.NewOrderController,
		jwt.NewJWTService,
		middleware.NewAuthMiddleware,
		wire.Struct(new(modules.OrderModule), "*"))
	return nil
}

func InitProductVariantModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, updateStockAgg *event.UpdateStockAggregator) *modules.ProductVariantModule {
	wire.Build(
		repoImpl.NewProductVariantGormRepository,
		repoImpl.NewProductGormRepository,
		cache.NewProductVariantRedisService,
		serviceImpl.NewProductVariantService,
		controllers.NewProductVariantController,
		wire.Struct(new(modules.ProductVariantModule), "*"))

	return nil
}

func InitPayOSModule(db *gorm.DB, redisClient *redis.Client, redisContext context.Context, eventBus event.EventPublisher, updateStockAgg *event.UpdateStockAggregator) *modules.PayOsModule {
	wire.Build(
		repoImpl.NewProductVariantGormRepository,
		repoImpl.NewOrderItemGormRepository,
		repoImpl.NewOrderGormRepository,
		repoImpl.NewPaymentInfoGormImpl,
		repoImpl.NewDraftOrderGormRepository,
		cache.NewProductVariantRedisService,
		serviceImpl.NewOrderService,
		controllers.NewWebhookController,
		wire.Struct(new(modules.PayOsModule), "*"))
	return nil
}
