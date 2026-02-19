package keepalive

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/message"
	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/kafka"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	consumer         *kafka.Consumer
	keepAliveService KeepAliveService
}

type KeepAliveService interface {
	UpdateTTL(ctx context.Context, message message.Message)
}

func New(consumer *kafka.Consumer, service KeepAliveService) *Handler {
	handler := &Handler{
		consumer:         consumer,
		keepAliveService: service,
	}

	return handler
}

func (h *Handler) KeepAlive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			msg, err := h.consumer.Consume(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Error consuming message")
				continue
			}

			go h.keepAliveService.UpdateTTL(ctx, msg)
		}
	}
}
