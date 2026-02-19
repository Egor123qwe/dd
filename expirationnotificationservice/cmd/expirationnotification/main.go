package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/cache"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/handler/keepalive"
	mq "gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/pool"
	skeepalive "gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/service/keepalive"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/service/notifier"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	dialerTimeout = 10 * time.Second
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err)
	}

	redisOptions := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	redis := redis.NewClient(&redis.Options{
		Addr:     redisOptions,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	cache := cache.New(redis)

	kafkaUser, kafkaPass := cfg.Kafka.Username, cfg.Kafka.Password
	if v, ok := os.LookupEnv("KAFKA_USERNAME"); ok {
		kafkaUser = v
	}
	if v, ok := os.LookupEnv("KAFKA_PASSWORD"); ok {
		kafkaPass = v
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * dialerTimeout,
		DualStack: true,
	}
	if kafkaUser != "" && kafkaPass != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, kafkaUser, kafkaPass)
		if err != nil {
			log.Error().Err(err).Msg("Can not create scram mechanism")
		} else {
			dialer.SASLMechanism = mechanism
		}
	}

	consumer := mq.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Consumer.Topic, cfg.Kafka.Consumer.GroupID, dialer)
	kaService := skeepalive.New(cache, cfg.Redis)
	kaHandler := keepalive.New(consumer, kaService)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	producer := mq.NewProducer(cfg.Kafka.Brokers, dialer)

	pool := pool.NewRedisPool(cfg)

	notifier := notifier.New(pool, producer, redis, cfg)

	pool.Start()

	wg.Add(1)
	go func() {
		defer wg.Done()
		notifier.ExpireNotify(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kaHandler.KeepAlive(ctx)
	}()

	log.Info().Msg("Expiration notification service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info().Msg("Shutting down gracefully...")
	consumer.Close()
	cancel()
	pool.Shutdown()
	redis.Close()
	producer.Close()
	log.Info().Msg("Shutdown complete")
}
