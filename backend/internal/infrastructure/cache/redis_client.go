package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a Redis client from a URL.
// Uses redis.ParseURL for standard format: redis://host:port/db or redis://:password@host:port/db.
func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
