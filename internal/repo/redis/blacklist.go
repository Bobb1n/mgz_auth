package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BlacklistRepository struct {
	client *redis.Client
	prefix string
}

func NewBlacklistRepository(client *redis.Client, keyPrefix string) *BlacklistRepository {
	if keyPrefix == "" {
		keyPrefix = "auth:blacklist:"
	}
	return &BlacklistRepository{client: client, prefix: keyPrefix}
}

func (r *BlacklistRepository) Add(ctx context.Context, tokenID string, ttlSeconds int) error {
	key := r.prefix + tokenID
	return r.client.Set(ctx, key, "1", time.Duration(ttlSeconds)*time.Second).Err()
}

func (r *BlacklistRepository) Exists(ctx context.Context, tokenID string) (bool, error) {
	key := r.prefix + tokenID
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return n > 0, nil
}
