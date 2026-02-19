package db

import (
	"log"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db/repo"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db/repo/rent"
)

type Storage interface {
	Rent() repo.Rent
	Close() error
}

type storage struct {
	rentRepo repo.Rent
	db       *sqlx.DB
}

func New(cfg config.Config, logger slog.Logger) Storage {
	dsn := "postgres://" + cfg.DbConfig.DbUser + ":" + cfg.DbConfig.DbPassword + "@" + cfg.DbConfig.DbHost +
		":" + cfg.DbConfig.DbPort + "/" + cfg.DbConfig.DbName + "?sslmode=disable"
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("db Connection init")

	return storage{
		rentRepo: rent.New(db),
		db:       db,
	}
}

func (s storage) Rent() repo.Rent {
	return s.rentRepo
}

func (s storage) Close() error {
	return s.db.Close()
}
