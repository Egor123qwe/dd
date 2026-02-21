package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	broker "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/cache"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/client/sessionhandler"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	handler "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/handler/broker"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/handler/merchant"
	resthandler "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/handler/rest"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/handler/ttlnotification"
	pbp "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/proto/gen/pricing.v1"
	rmerchant "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/repository/merchant"
	smerchant "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/service/merchant"
	snotification "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/service/ttlnotification"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err)
	}

	dbOptions := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	redisOptions := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	db, err := sql.Open("postgres", dbOptions)
	if err != nil {
		log.Error().Err(err)
	}
	defer db.Close()

	redis := redis.NewClient(&redis.Options{
		Addr:     redisOptions,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	cache := cache.New(redis)

	broker, err := broker.New(cfg.Kafka)
	if err != nil {
		log.Error().Err(err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	priceServiceURL := fmt.Sprintf("%s:%s", cfg.PriceService.Host, cfg.PriceService.Port)
	conn, err := grpc.NewClient(priceServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("Can not connect to grpc server")
	}
	defer conn.Close()

	priceServiceClient := pbp.NewPriceClient(conn)

	merchantRepository := rmerchant.New(db, cache, cfg.Redis)
	merchantService := smerchant.New(merchantRepository, cache, cfg.Redis, priceServiceClient)
	merchantHandler := merchant.New(merchantService, broker.Producer(), cfg)

	ttlNotificationService := snotification.New(merchantRepository)
	ttlNotificationHandler := ttlnotification.New(ttlNotificationService, broker.Producer(), cfg)

	brokerConsumer := handler.New(broker, cfg, merchantHandler, ttlNotificationHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go brokerConsumer.ConsumeInput(ctx)
	go brokerConsumer.ConsumeNotification(ctx)

	var rentStopper resthandler.RentStopper
	if cfg.SessionHandler.URL != "" {
		rentStopper = sessionhandler.New(cfg.SessionHandler.URL)
	}
	restServer := resthandler.New(cfg.Server.Port, merchantRepository, rentStopper)
	go func() {
		if err := restServer.Run(ctx); err != nil {
			log.Error().Err(err).Msg("REST server stopped")
		}
	}()

	log.Info().Msg("Service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()

	log.Info().Msg("Shutdown complete")
}
