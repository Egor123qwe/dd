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
	sessionKey := fmt.Sprintf("%s %s", event.SessionIDType, message.Meta.Conn.SessionID)
	clientKey := fmt.Sprintf("%s %s", event.ClientIDType, message.Meta.Conn.UserID)
	
	if message.Meta.Conn.SessionID != "" {
		err := s.cache.Expire(ctx, sessionKey, time.Minute*time.Duration(s.cfg.TTL))
		if err != nil {
			log.Error().Err(err).Msg("Can not update TTL")
			return
		}
	}

	err := s.cache.Expire(ctx, clientKey, time.Minute*time.Duration(s.cfg.TTL))
	if err != nil {
		log.Error().Err(err).Msg("Can not update TTL")
		return
	}
}
