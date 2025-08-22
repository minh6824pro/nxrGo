package cache

import (
	"context"
	"fmt"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"log"
	"time"

	"github.com/minh6824pro/nxrGO/models"
	"github.com/redis/go-redis/v9"
)

type productVariantRedisService struct {
	client             *redis.Client
	ctx                context.Context
	productVariantRepo repositories.ProductVariantRepository
	db                 *gorm.DB
}

const ProductVariantKeyPattern = "productVariant:%d"

func NewProductVariantRedisService(client *redis.Client, ctx context.Context,
	productVariantRepo repositories.ProductVariantRepository, db *gorm.DB) ProductVariantRedis {
	return &productVariantRedisService{
		client:             client,
		ctx:                ctx,
		productVariantRepo: productVariantRepo,
		db:                 db,
	}
}

func (r *productVariantRedisService) SaveProductVariantHash(pv models.ProductVariant) error {
	ttl := 30 * time.Minute
	key := fmt.Sprintf(ProductVariantKeyPattern, pv.ID)

	err := r.client.HSet(r.ctx, key, map[string]interface{}{
		"id":          pv.ID,
		"quantity":    pv.Quantity,
		"price":       pv.Price,
		"image":       pv.Image,
		"productName": pv.Product.Name,
		"productId":   pv.Product.ID,
	}).Err()
	if err != nil {
		return err
	}

	return r.client.Expire(r.ctx, key, ttl).Err()
}

func (r *productVariantRedisService) GetProductVariantHash(id uint) (map[string]string, error) {
	key := fmt.Sprintf(ProductVariantKeyPattern, id)
	return r.client.HGetAll(r.ctx, key).Result()
}

func (r *productVariantRedisService) GetOrCreateProductVariantHash(id uint) (map[string]string, error) {
	hash, err := r.GetProductVariantHash(id)
	if err != nil {
		variant, err := r.productVariantRepo.GetByIDForRedisCache(context.Background(), id)
		if err != nil {
			return nil, err
		}
		if err := r.SaveProductVariantHash(*variant); err != nil {
			return nil, err
		}
		hash, err = r.GetProductVariantHash(id)
		return hash, err
	}
	return hash, nil
}

func (r *productVariantRedisService) EvalLua(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

func (r *productVariantRedisService) DecrementStock(orderItems []models.OrderItem) error {
	for _, oi := range orderItems {
		key := fmt.Sprintf(ProductVariantKeyPattern, oi.ProductVariantID)
		delta := -oi.Quantity
		err := r.client.HIncrBy(r.ctx, key, "quantity", int64(delta)).Err()
		if err != nil {
			return fmt.Errorf("failed to decrement stock for key %s: %w", key, err)
		}
		r.DeleteMiniProduct(oi.ProductVariantID)
	}
	return nil
}

func (r *productVariantRedisService) IncrementStock(orderItems []models.OrderItem) error {
	for _, oi := range orderItems {
		key := fmt.Sprintf(ProductVariantKeyPattern, oi.ProductVariantID)
		delta := oi.Quantity
		err := r.client.HIncrBy(r.ctx, key, "quantity", int64(delta)).Err()
		if err != nil {
			return fmt.Errorf("failed to increment stock for key %s: %w", key, err)
		}
		r.DeleteMiniProduct(oi.ProductVariantID)
	}
	return nil
}

func (r *productVariantRedisService) DeleteProductVariantHash(id uint) error {
	key := fmt.Sprintf(ProductVariantKeyPattern, id)

	r.DeleteMiniProduct(id)
	return r.client.Del(context.Background(), key).Err()
}

func (r *productVariantRedisService) PingRedis(ctx context.Context) error {
	// Set timeout riêng cho lệnh ping
	healthCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	// Thực hiện ping
	if err := r.client.Ping(healthCtx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

func (r *productVariantRedisService) DeleteMiniProduct(variantId uint) error {
	var productID uint
	err := r.db.Table("product_variants").
		Select("product_id").
		Where("id = ?", variantId).
		Pluck("product_id", &productID).Error

	if err != nil {
		log.Println("Error while delete cache: ", err)
	}
	log.Println(productID)

	key := fmt.Sprintf(productMiniCacheKeyPattern, productID)

	// Thực hiện xoá key
	return r.client.Del(r.ctx, key).Err()
}
