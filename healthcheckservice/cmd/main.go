package main

import (
	"log"
	"log/slog"
	"os"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/app"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

//@Title API Documentation
//@Version 1.0

// @host 0.0.0.0:9000
func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)

	logger.Info("Starting")

	app, err := app.New(*cfg, *logger)

	if err != nil {
		log.Fatal(err)
	}

	app.Start()
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
