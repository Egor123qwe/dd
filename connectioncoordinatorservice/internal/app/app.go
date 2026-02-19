// Package app handles starting and monitoring the server for graceful shutdown.
package app

import (
	"context"
	"fmt"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/server"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service"
)

var log = logging.MustGetLogger("app")

type App struct {
	srv service.Service
}

func New(debug bool) (*App, error) {
	api, err := api.New()
	if err != nil {
		return nil, err
	}

	srv, err := service.New(api, debug)
	if err != nil {
		return nil, err
	}

	app := &App{
		srv: srv,
	}

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	if err := server.New(a.srv).Serve(ctx); err != nil {
		return fmt.Errorf("server stopped with error: %w\n", err)
	}

	return nil
}
