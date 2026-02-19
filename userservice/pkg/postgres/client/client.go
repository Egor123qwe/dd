package client

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const driverName = "postgres"

type Config struct {
	// URL format: "postgres://user:password@host:5432/dbname?sslmode=disable"
	URL string
}

func New(config Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driverName, config.URL)
	if err != nil {
		return nil, err
	}
	return db, nil
}
