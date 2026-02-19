package notifier

import (
	"context"
	"encoding/json"
	"strings"

	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/message"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/kafka"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Service struct {
	pool        RedisPool
	producer    *kafka.Producer
	cfg         config.Config
	redisClient *redis.Client
}

type RedisPool interface {
	GetMessages() <-chan string
}

func New(pool RedisPool, producer *kafka.Producer, redisClient *redis.Client, cfg config.Config) *Service {
	service := &Service{
		pool:        pool,
		producer:    producer,
		cfg:         cfg,
		redisClient: redisClient,
	}

	return service
}

func (s *Service) ExpireNotify(ctx context.Context) {
	for msg := range s.pool.GetMessages() {

		lockKey := "lock:" + msg
		lockValue := uuid.NewString()

		// Try to acquire lock
		locked, err := s.redisClient.SetNX(ctx, lockKey, lockValue, 0).Result()
		if err != nil || !locked {
			log.Warn().Msgf("Failed to acquire lock for %s", msg)
			continue
		}

		log.Info().Msgf("Expired %s", msg)

		response := strings.Split(msg, " ")
		idType := response[0]
		id := response[1]

		eventType, exists := event.EventTypeMap[event.IDType(idType)]
		if !exists {
			log.Error().Msgf("Unknown event type: %s", eventType)
			continue
		}

		message := &message.Message{
			Type: string(eventType),
			Meta: message.Meta{
				Status: "ok",
			},
			Content: message.Content{},
		}

		switch eventType {
		case event.ExpiredSession:
			message.Meta.Conn.SessionID = id
			message.Meta.SessionID = id
			message.Content.SessionID = id

		case event.ExpiredClient:
			message.Meta.Conn.UserID = id

		case event.ExpiredRequest:
			message.Content.RequestID = id

		case event.ExpiredPaidRequest:
			message.Content.RequestID = id

		case event.ExpiredDeal:
			message.Content.DealID = id
		}

		messageBytes, _ := json.Marshal(message)

		err = s.producer.Produce(
			ctx,
			s.cfg.Kafka.Producer.Topic,
			[]byte(msg),
			messageBytes,
		)
		if err != nil {
			log.Error().Err(err).Msg("can not produce message")
		}

		log.Printf("Message produced: %s", msg)

		// Unlock
		s.redisClient.Del(ctx, lockKey)
	}
}
