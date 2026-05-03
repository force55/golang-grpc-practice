package redisbus

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := rdb.Ping(context.Background()); err != nil {
		t.Skip("Redis is not running")
	}

	return rdb
}
