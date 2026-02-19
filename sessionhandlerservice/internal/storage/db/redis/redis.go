package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	cashRep "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/redis/cache"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/cache"
)

type Store interface {
	Cache() cache.Repo
	Close() error
}

type store struct {
	cache cache.Repo
	db    *redis.Client
}

func configure(db *redis.Client) Store {
	return store{
		cache: cashRep.New(db),
		db:    db,
	}
}

func New(config Config) (Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return configure(client), nil
}

func (s store) Close() error {
	return s.db.Close()
}

func (s store) Cache() cache.Repo {
	return s.cache
}
