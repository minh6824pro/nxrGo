package cache

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type ProductVariantRedis interface {
	SaveProductVariantHash(pv models.ProductVariant) error
	GetProductVariantHash(id uint) (map[string]string, error)
	EvalLua(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	IncrementStock(orderItems []models.OrderItem) error
	DecrementStock(orderItems []models.OrderItem) error
	DeleteProductVariantHash(id uint) error
	PingRedis(ctx context.Context) error
}
