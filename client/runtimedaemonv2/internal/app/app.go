// Package app handles starting and monitoring the server for graceful shutdown.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	dockerAPI "github.com/docker/docker/client"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler"
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/server"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/server/launcher"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var resetSettingsTTL = 10 * time.Second

var log = logger.NewLogger("app", logger.DefaultWithSentry())

type App struct {
	server launcher.Server
	srv    service.Service
}

func New() (*App, error) {
	storage, err := storage.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	dockerAPI, err := dockerAPI.NewClientWithOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to create container client: %w", err)
	}

	srv, err := service.New(dockerAPI, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to init service: %w", err)
	}

	app := &App{
		server: server.New(handler.New(srv)),
		srv:    srv,
	}

	return app, nil
}

func (a *App) Run() error {
	ctx, stop := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		errCh <- a.server.Serve(ctx)

		close(errCh)
	}()

	var err error

	select {
	case <-a.getExitSignal():
	case err = <-errCh:
	}

	stop()

	log.Infof("app: shutting down the server...")
	<-errCh

	log.Info("app: resetting settings...")

	resetDoneCh := make(chan struct{})

	resetCtx, cancelReset := context.WithTimeout(context.Background(), resetSettingsTTL)

	go func() {
		// reset runtime settings
		disabler := model.Configuration{Settings: model.Settings{Mode: model.ModeDisable}}

		defer cancelReset()

		if err := a.srv.ChangeMode(resetCtx, disabler); err != nil {
			log.Errorf("failed to discard settings: %v", err, logger.DefaultWithSentry())
		}

		close(resetDoneCh)
	}()

	select {
	case <-resetCtx.Done():
		a.srv.HardReset()
		log.Error("app: failed to reset settings. Hard reset was used. FIXME!!!")

	case <-resetDoneCh:
	}

	return err
}

func (a *App) getExitSignal() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	return quit
}
