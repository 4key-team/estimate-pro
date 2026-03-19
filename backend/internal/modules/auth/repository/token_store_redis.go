package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore stores refresh tokens in Redis with TTL-based auto-expiry.
type RedisTokenStore struct {
	client *redis.Client
}

func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{client: client}
}

func (s *RedisTokenStore) Save(ctx context.Context, userID, tokenID string, ttl time.Duration) error {
	if err := s.client.Set(ctx, s.key(userID, tokenID), "1", ttl).Err(); err != nil {
		return fmt.Errorf("tokenStore.Save: %w", err)
	}
	return nil
}

func (s *RedisTokenStore) Exists(ctx context.Context, userID, tokenID string) (bool, error) {
	n, err := s.client.Exists(ctx, s.key(userID, tokenID)).Result()
	if err != nil {
		return false, fmt.Errorf("tokenStore.Exists: %w", err)
	}
	return n > 0, nil
}

func (s *RedisTokenStore) Delete(ctx context.Context, userID, tokenID string) error {
	if err := s.client.Del(ctx, s.key(userID, tokenID)).Err(); err != nil {
		return fmt.Errorf("tokenStore.Delete: %w", err)
	}
	return nil
}

func (s *RedisTokenStore) DeleteAll(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("refresh_token:%s:*", userID)
	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		s.client.Del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("tokenStore.DeleteAll: %w", err)
	}
	return nil
}

func (s *RedisTokenStore) key(userID, tokenID string) string {
	return fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
}
