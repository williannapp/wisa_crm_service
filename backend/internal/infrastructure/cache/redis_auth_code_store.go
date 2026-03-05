package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/service"
)

const authCodeKeyPrefix = "auth_code:"

// RedisAuthCodeStore implements service.AuthCodeStore using Redis.
var _ service.AuthCodeStore = (*RedisAuthCodeStore)(nil)

// RedisAuthCodeStore stores authorization codes in Redis with TTL.
type RedisAuthCodeStore struct {
	client *redis.Client
}

// NewRedisAuthCodeStore creates a new RedisAuthCodeStore.
func NewRedisAuthCodeStore(client *redis.Client) *RedisAuthCodeStore {
	return &RedisAuthCodeStore{client: client}
}

// Store saves the auth code with the given data and TTL.
func (s *RedisAuthCodeStore) Store(ctx context.Context, code string, data *service.AuthCodeData, ttlSeconds int) error {
	key := authCodeKeyPrefix + code
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal auth code data: %w", err)
	}
	expiration := time.Duration(ttlSeconds) * time.Second
	return s.client.Set(ctx, key, val, expiration).Err()
}

// GetAndDelete retrieves the auth code data and removes it atomically (single-use).
func (s *RedisAuthCodeStore) GetAndDelete(ctx context.Context, code string) (*service.AuthCodeData, error) {
	key := authCodeKeyPrefix + code
	val, err := s.client.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrCodeInvalidOrExpired
		}
		return nil, err
	}
	var data service.AuthCodeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal auth code data: %w", err)
	}
	return &data, nil
}
