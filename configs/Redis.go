package configs

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	RedisClient *redis.Client
	RedisCtx    context.Context
)

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	RedisCtx = context.Background()

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB value '%s', default to 0", dbStr)
		db = 0
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Kiểm tra kết nối Redis
	err = RedisClient.Ping(RedisCtx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")
}

// Ví dụ hàm tiện ích Set với TTL
func Set(key string, value interface{}, ttl time.Duration) error {
	return RedisClient.Set(RedisCtx, key, value, ttl).Err()
}

// Ví dụ hàm tiện ích Get
func Get(key string) (string, error) {
	return RedisClient.Get(RedisCtx, key).Result()
}

// Xóa key
func Delete(key string) error {
	return RedisClient.Del(RedisCtx, key).Err()
}
