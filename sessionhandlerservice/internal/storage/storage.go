package storage

import (
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/psql"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/redis"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/balance"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/cache"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/template"
)

type Storage interface {
	Session() session.Session
	Rent() rent.Rent
	Template() template.Template
	Balance() balance.Balance

	Cache() cache.Repo

	Close() error
}

type storage struct {
	psql  psql.Store
	redis redis.Store
}

func New() (Storage, error) {
	var err error
	var storage storage

	storage.psql, err = psql.New(psql.NewConfig())
	if err != nil {
		return nil, err
	}

	storage.redis, err = redis.New(redis.NewConfig())
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s storage) Close() error {
	if err := s.redis.Close(); err != nil {
		return err
	}

	if err := s.psql.Close(); err != nil {
		return err
	}

	return nil
}
