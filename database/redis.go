package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// rdb is the Redis client instance
var rdb *redis.Client
var ctx = context.Background()

// initRedis initializes a connection to Redis
func initRedis() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Default to localhost if not set
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379" // Default to 6379 if not set
	}
	addr := fmt.Sprintf("%s:%s", host, port)
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr, // Redis server address
		Password: "",   // no password set
		DB:       0,    // use default DB
	})
}

// SetValue sets a key-value pair in Redis with an expiration time
func SetValue(key string, value string, expire time.Duration) error {
	if rdb == nil {
		initRedis()
	}
	return rdb.Set(ctx, key, value, expire).Err()
}

// GetValue retrieves a value from Redis by key
func GetValue(key string) (string, error) {
	if rdb == nil {
		initRedis()
	}
	return rdb.Get(ctx, key).Result()
}

// DeleteValue removes a key-value pair from Redis
func DeleteValue(key string) error {
	if rdb == nil {
		initRedis()
	}
	return rdb.Del(ctx, key).Err()
}
