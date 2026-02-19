package app

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/server"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

type App struct {
	server *server.Server
}

func New(cfg config.Config, log slog.Logger) (*App, error) {
	storage, err := storage.New(cfg, log)
	if err != nil {
		return nil, err
	}

	srv, err := server.New(log, cfg, storage)
	if err != nil {
		return nil, err
	}

	return &App{
		server: srv,
	}, nil
}

func (a *App) Start() {
	ctx := context.Background()

	a.server.Start(ctx)
}
