package postgres

import (
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/postgres/client"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type Store interface {
	DB() *sqlx.DB
	UpdateSchema() error
	RunTestSeeders() error
	Close() error
}

type store struct {
	db  *sqlx.DB
	cfg Config
}

type Config struct {
	URL            string
	MigrationsDir  string
	TestSeedersDir string
	RunTestSeeders bool
}

func New(cfg Config) (Store, error) {
	db, err := client.New(client.Config{URL: cfg.URL})
	if err != nil {
		return nil, err
	}
	return &store{db: db, cfg: cfg}, nil
}

func (s *store) DB() *sqlx.DB {
	return s.db
}

func (s *store) UpdateSchema() error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	return goose.Up(s.db.DB, s.cfg.MigrationsDir, goose.WithAllowMissing())
}

func (s *store) RunTestSeeders() error {
	if !s.cfg.RunTestSeeders || s.cfg.TestSeedersDir == "" {
		return nil
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	return goose.Up(s.db.DB, s.cfg.TestSeedersDir, goose.WithAllowMissing())
}

func (s *store) Close() error {
	return s.db.Close()
}
