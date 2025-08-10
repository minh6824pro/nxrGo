package cache

import (
	"fmt"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/models"
	"time"
)

var KeyPattern = "productVariant:%d"

func SaveProductVariantHash(pv models.ProductVariant, ttl time.Duration) error {
	key := fmt.Sprintf(KeyPattern, pv.ID)

	// Lưu dưới dạng hash
	err := configs.RedisClient.HSet(configs.RedisCtx, key, map[string]interface{}{
		"id":       pv.ID,
		"quantity": pv.Quantity,
		"price":    pv.Price,
	}).Err()
	if err != nil {
		return err
	}

	// Set TTL
	return configs.RedisClient.Expire(configs.RedisCtx, key, ttl).Err()
}

func GetProductVariantHash(id uint) (map[string]string, error) {
	key := fmt.Sprintf(KeyPattern, id)
	return configs.RedisClient.HGetAll(configs.RedisCtx, key).Result()
}
