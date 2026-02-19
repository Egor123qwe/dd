package repo

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
)

var (
	ErrNoRent = errors.New("has no rent for session_id")
)

const (
	cacheTimeout = 10 * time.Second
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	GetClientRentBySession(ctx context.Context, key, sessionID string) (model.Client, error)

	Set(ctx context.Context, key, value string) (string, error)
	SetOrUpdateClient(ctx context.Context, key string, msg message.ClientRent) error
	SetOrUpdateMerchant(ctx context.Context, key string, msg message.MerchantRent) error

	DeleteMerchant(ctx context.Context, key, keyUserID string) error
	DeleteClientRent(ctx context.Context, key, sessionID string) error
	Delete(ctx context.Context, key string) error
}

type cache struct {
	client *redis.Client
	cfg    config.Config
	log    slog.Logger
}

func New(clent *redis.Client, cfg config.Config, log slog.Logger) Cache {
	return cache{
		client: clent,
		cfg:    cfg,
		log:    log,
	}
}

func (c cache) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	return c.client.Get(ctx, key).Result()
}

func (c cache) Set(ctx context.Context, key, value string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	return c.client.Set(ctx, key, value, c.cfg.RedisConfig.TTL).Result()
}

func (c cache) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	return c.client.Del(ctx, key).Err()
}
