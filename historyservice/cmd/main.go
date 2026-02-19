package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/server"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

//@Title API Documentation
//@Version 1.0

// @host 0.0.0.0:8899
func main() {

	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)

	server := server.NewServer(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		server.Start()
	}()

	<-stop

	server.Stop(ctx)

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	}

	return log
}
