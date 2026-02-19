package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/grpcprice"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler"
	pricingv1 "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/pricingv1"
	categoryrepo "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/repository/category"
	hardwarerepo "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/repository/hardware"
	sessionrepo "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/repository/session"
	tariffrepo "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/repository/tariff"
	categorysvc "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/service/category"
	hardwaresvc "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/service/hardware"
	sessionsvc "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/service/session"
	tariffsvc "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/service/tariff"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/telemetry/tracer"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	shutdownTimeout = 5 * time.Second
)

// @BasePath		/api/v1
// @title			Swagger MachinesAggregation Service Api
// @version		1.0
// @description	This is an API for MachinesAggregation service
func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to connect to db", "error: ", err)
	}
	defer db.Close()

	tp, err := tracer.InitTracer(cfg.Tracer)
	if err != nil {
		logger.Error("failed to init tracer: ", "error", err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer: ", "error", err)
		}
	}()

	tracer := tp.Tracer(cfg.Tracer.ServiceName)

	tariffRepo := tariffrepo.New(db)
	sessionRepo := sessionrepo.New(db)
	categoryRepo := categoryrepo.New(db)
	hardwareRepo := hardwarerepo.New(db)

	tariffService := tariffsvc.New(tariffRepo)
	sessionService := sessionsvc.New(sessionRepo)
	categoryService := categorysvc.New(categoryRepo)
	hardwareService := hardwaresvc.New(hardwareRepo)

	router := handler.Router(*logger, *cfg, tariffService, sessionService, categoryService, hardwareService, tracer)

	srv := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Listen error: ", err.Error())
		}
	}()
	logger.Info("Server started on ", "addr", srv.Addr)

	// gRPC Price service (для resourcepoolservice)
	lis, err := net.Listen("tcp", cfg.GrpcServer.Address)
	if err != nil {
		log.Fatal("gRPC listen error: ", err)
	}
	grpcSrv := grpc.NewServer()
	pricingv1.RegisterPriceServer(grpcSrv, grpcprice.New())
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal("gRPC serve error: ", err)
		}
	}()
	logger.Info("gRPC Price server started on ", "addr", cfg.GrpcServer.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Shutting down server: ", err.Error())
	}

	logger.Info("Gracefully shutting down the http server")
}

func setupLogger(env string) *slog.Logger {
	var loggger *slog.Logger

	switch env {
	case envDev:
		loggger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		loggger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		loggger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return loggger
}
