package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minh6824pro/nxrGO/internal/models/CacheModel"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/redis/go-redis/v9"
	"time"
)

type productCacheServiceImpl struct {
	rdb                        *redis.Client
	productRepository          repositories.ProductRepository
	productVariantRedisService ProductVariantRedis
}

const (
	productMiniCacheKeyPattern = "productMiniInfo:%d"
	ListProductCacheKeyPattern = "priceMin:%s|priceMax:%s|priceAsc:%s|totalBuyDesc:%s|page:%d|pageSize:%d"
)
const ttl = 1 * time.Hour

// Hàm khởi tạo
func NewProductCacheService(rdb *redis.Client, repo repositories.ProductRepository,
	productVariantRedisService ProductVariantRedis) ProductCacheService {
	return &productCacheServiceImpl{
		rdb:                        rdb,
		productRepository:          repo,
		productVariantRedisService: productVariantRedisService,
	}
}

// Get key
func (s *productCacheServiceImpl) getCacheKey(id uint) string {
	return fmt.Sprintf(productMiniCacheKeyPattern, id)
}

func (s *productCacheServiceImpl) GetAllProductId(ctx context.Context) ([]uint, error) {
	return s.productRepository.GetAllProductId(ctx)
}

func (s *productCacheServiceImpl) GenerateListProductCacheKey(priceMin, priceMax *float64, priceAsc, totalBuyDesc *bool, page, pageSize int) string {
	minStr := "nil"
	maxStr := "nil"
	if priceMin != nil {
		minStr = fmt.Sprintf("%.2f", *priceMin)
	}
	if priceMax != nil {
		maxStr = fmt.Sprintf("%.2f", *priceMax)
	}

	priceAscStr := "nil"
	if priceAsc != nil {
		if *priceAsc {
			priceAscStr = "asc"
		} else {
			priceAscStr = "desc"
		}
	}

	totalBuyDescStr := "nil"
	if totalBuyDesc != nil {
		totalBuyDescStr = "desc"
	}

	return fmt.Sprintf(ListProductCacheKeyPattern, minStr, maxStr, priceAscStr, totalBuyDescStr, page, pageSize)
}

func (s *productCacheServiceImpl) GetListProductCache(ctx context.Context, key string) ([]CacheModel.ListProductQueryCache, error) {
	cachedData, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err // nil nếu key không tồn tại
	}

	var result []CacheModel.ListProductQueryCache
	if err := json.Unmarshal([]byte(cachedData), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *productCacheServiceImpl) CacheListProduct(ctx context.Context, key string, data []CacheModel.ListProductQueryCache) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ttl := time.Hour // TTL 1 giờ
	return s.rdb.Set(ctx, key, jsonData, ttl).Err()
}

// Lấy cache theo product ID
func (s *productCacheServiceImpl) GetProductMiniCache(ctx context.Context, productID uint) (*CacheModel.ProductMiniCache, error) {
	key := fmt.Sprintf(productMiniCacheKeyPattern, productID)
	cachedData, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// key không tồn tại
			return nil, nil
		}
		return nil, err
	}

	var result CacheModel.ProductMiniCache
	if err := json.Unmarshal([]byte(cachedData), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *productCacheServiceImpl) GetProductMiniCacheBulk(
	ctx context.Context,
	list []CacheModel.ListProductQueryCache,
) ([]*CacheModel.ProductMiniCache, []CacheModel.ListProductQueryCache, error) {

	ids := make([]uint, len(list))
	for i, item := range list {
		ids[i] = item.ProductID
	}

	results := make([]*CacheModel.ProductMiniCache, len(ids))
	var missing []CacheModel.ListProductQueryCache

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf(productMiniCacheKeyPattern, id)
	}

	cachedValues, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	for i, val := range cachedValues {
		if val == nil {
			missing = append(missing, list[i])
			continue
		}

		var cacheItem CacheModel.ProductMiniCache
		if err := json.Unmarshal([]byte(val.(string)), &cacheItem); err != nil {
			missing = append(missing, list[i])
			continue
		}

		results[i] = &cacheItem
	}

	return results, missing, nil
}

// Lưu cache cho product ID
func (s *productCacheServiceImpl) CacheMiniProduct(ctx context.Context, product *CacheModel.ProductMiniCache) error {
	key := fmt.Sprintf(productMiniCacheKeyPattern, product.ID)
	jsonData, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, key, jsonData, ttl).Err()
}

func (s *productCacheServiceImpl) CacheMiniProducts(ctx context.Context, products []*CacheModel.ProductMiniCache) error {
	pipe := s.rdb.Pipeline()

	for _, product := range products {
		key := fmt.Sprintf(productMiniCacheKeyPattern, product.ID)
		jsonData, err := json.Marshal(product)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, jsonData, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *productCacheServiceImpl) PingRedis(ctx context.Context) error {
	// Set timeout riêng cho lệnh ping
	healthCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	// Thực hiện ping
	if err := s.rdb.Ping(healthCtx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
