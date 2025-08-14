package cache

import (
	"context"
	"github.com/minh6824pro/nxrGO/models/CacheModel"
)

type ProductCacheService interface {
	GetProductMiniCache(ctx context.Context, productID uint) (*CacheModel.ProductMiniCache, error)
	GetProductMiniCacheBulk(ctx context.Context, list []CacheModel.ListProductQueryCache) ([]*CacheModel.ProductMiniCache, []CacheModel.ListProductQueryCache, error)
	CacheMiniProduct(ctx context.Context, product *CacheModel.ProductMiniCache) error
	CacheMiniProducts(ctx context.Context, products []*CacheModel.ProductMiniCache) error
	GenerateListProductCacheKey(priceMin, priceMax *float64, priceAsc, totalBuyDesc *bool, page, pageSize int) string
	GetListProductCache(ctx context.Context, key string) ([]CacheModel.ListProductQueryCache, error)
	CacheListProduct(ctx context.Context, key string, data []CacheModel.ListProductQueryCache) error
	PingRedis(ctx context.Context) error
}
