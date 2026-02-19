package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler/middleware"
	rfeedback "gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/repository/feedback"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/service/auth"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/service/feedback"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/telemetry/tracer"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	shutdownTimeout = 5 * time.Second
	httpTimeout     = 10 * time.Second
)

// @BasePath /api/v1
// @title			Swagger Feedback Service Api
// @version		1.0
// @description	This is an API for feedback service
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// @security Bearer
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("can not load config")
	}

	tp, err := tracer.InitTracer(cfg.Tracer)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracer")
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	dbOptions := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := sql.Open("postgres", dbOptions)
	if err != nil {
		log.Error().Err(err)
	}
	defer db.Close()

	httpClient := http.Client{
		Timeout: httpTimeout,
	}

	feedbackRepo := rfeedback.New(db)
	feedbackService := feedback.New(feedbackRepo)
	authService := auth.New(httpClient, cfg.Api.AuthConfig)
	authMiddleware := middleware.New(authService)

	router := handler.Register(feedbackService, authMiddleware)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msgf("failed to run the http server: %v", err.Error())
		}
	}()

	log.Info().Msgf("server started on %s", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("can not shutdown server")
	}

	log.Info().Msg("gracefully shutting down the http server")
}
