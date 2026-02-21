package keepalive

import (
	"context"
	"fmt"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/message"

	"github.com/rs/zerolog/log"
)

type Service struct {
	cache Cache
	cfg   config.RedisConfig
}

type Cache interface {
	Expire(ctx context.Context, key string, exp time.Duration) error
}

func New(cache Cache, cfg config.RedisConfig) *Service {
	service := &Service{
		cache: cache,
		cfg:   cfg,
	}

	return service
}

func (s *Service) UpdateTTL(ctx context.Context, message message.Message) {
	ttl := time.Minute * time.Duration(s.cfg.TTL)

	// Продлеваем client_user_id (по keepalive от покупателя или мерчанта)
	if message.Meta.Conn.UserID != "" {
		clientKey := fmt.Sprintf("%s %s", event.ClientIDType, message.Meta.Conn.UserID)
		if err := s.cache.Expire(ctx, clientKey, ttl); err != nil {
			log.Error().Err(err).Str("key", clientKey).Msg("can not update TTL")
		}
	}

	// Продлеваем session_id (по keepalive от мерчанта — в Conn приходит session_id узла)
	sessionID := message.Meta.Conn.SessionID
	if sessionID == "" {
		sessionID = message.Content.SessionID
	}
	if sessionID != "" {
		sessionKey := fmt.Sprintf("%s %s", event.SessionIDType, sessionID)
		if err := s.cache.Expire(ctx, sessionKey, ttl); err != nil {
			log.Error().Err(err).Str("key", sessionKey).Msg("can not update TTL")
		}
	}

	// Продлеваем request_id (по keepalive от покупателя с content.request_id)
	if message.Content.RequestID != "" {
		requestKey := fmt.Sprintf("%s %s", event.RequestIDType, message.Content.RequestID)
		if err := s.cache.Expire(ctx, requestKey, ttl); err != nil {
			log.Error().Err(err).Str("key", requestKey).Msg("can not update TTL")
		}
	}
}
