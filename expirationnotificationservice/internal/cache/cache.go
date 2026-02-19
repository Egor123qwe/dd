package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	queryTimeout = 3 * time.Second
)

type Cache struct {
	redis *redis.Client
}

func New(redis *redis.Client) *Cache {
	cache := &Cache{
		redis: redis,
	}

	return cache
}

func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	return cache.redis.Get(c, key).Result()
}

func (cache *Cache) SetXX(ctx context.Context, key string, value any, exp time.Duration) (bool, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	return cache.redis.SetXX(c, key, value, exp).Result()
}

func (cache *Cache) Del(ctx context.Context, key string) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	return cache.redis.Del(c, key).Err()
}

func (cache *Cache) Expire(ctx context.Context, key string, exp time.Duration) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	return cache.redis.Expire(c, key, exp).Err()
}
