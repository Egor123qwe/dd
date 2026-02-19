package client

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	driverName = "mysql"
)

type Config struct {
	// format: "username:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	URL string
}

func New(config Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driverName, config.URL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
