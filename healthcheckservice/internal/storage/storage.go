package storage

import (
	"fmt"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/cache"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db"
)

type Storage interface {
	Close() error
	Cache() cache.Storage
	DB() db.Storage
}

type storage struct {
	cache cache.Storage
	db    db.Storage
}

func New(cfg config.Config, log slog.Logger) (Storage, error) {
	sqlStorage := db.New(cfg, log)

	cache, err := cache.New(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("error creating Storage : %v", err)
	}

	return storage{
		cache: cache,
		db:    sqlStorage,
	}, nil

}

func (s storage) Close() error {
	err := s.cache.Close()
	if err != nil {
		return fmt.Errorf("failed to close cache storage")
	}

	err = s.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close db storage")
	}

	return nil
}

func (s storage) Cache() cache.Storage {
	return s.cache
}

func (s storage) DB() db.Storage {
	return s.db
}
