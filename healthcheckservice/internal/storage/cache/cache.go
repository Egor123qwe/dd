package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/cache/repo"
)

type Storage interface {
	Repo() repo.Cache

	Close() error
}

type storage struct {
	cache  repo.Cache
	client *redis.Client
}

func New(cfg config.Config, log slog.Logger) (Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisConfig.Host, cfg.RedisConfig.Port),
		Password: cfg.RedisConfig.Password,
		DB: 3,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return storage{
		cache:  repo.New(client, cfg, log),
		client: client,
	}, nil
}

func (s storage) Repo() repo.Cache {
	return s.cache
}

func (s storage) Close() error {
	return s.client.Close()
}
