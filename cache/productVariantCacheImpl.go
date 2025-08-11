package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/minh6824pro/nxrGO/models"
	"github.com/redis/go-redis/v9"
)

type productVariantRedisService struct {
	client *redis.Client
	ctx    context.Context
}

const ProductVariantKeyPattern = "productVariant:%d"

func NewProductVariantRedisService(client *redis.Client, ctx context.Context) ProductVariantRedis {
	return &productVariantRedisService{
		client: client,
		ctx:    ctx,
	}
}

func (r *productVariantRedisService) SaveProductVariantHash(pv models.ProductVariant) error {
	ttl := 12 * time.Hour
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
	}
	return nil
}
