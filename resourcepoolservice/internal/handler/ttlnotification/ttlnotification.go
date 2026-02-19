package ttlnotification

import (
	"context"
	"encoding/json"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka/producer"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/message"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/response"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/status"
)

type Handler struct {
	ttlNotificationService TTLNotificationService
	producer               producer.Producer
	cfg                    config.Config
}

type TTLNotificationService interface {
	Delete(ctx context.Context, sessionID string) (string, string, error)
}

func New(tllNotificationService TTLNotificationService, producer producer.Producer, cfg config.Config) *Handler {
	handler := &Handler{
		ttlNotificationService: tllNotificationService,
		producer:               producer,
		cfg:                    cfg,
	}

	return handler
}

func (h *Handler) ExpireSession(ctx context.Context, msg message.FullMessage) []byte {
	userID, connectionID, err := h.ttlNotificationService.Delete(ctx, msg.Meta.Conn.SessionID)
	if err != nil {
		return response.ErrorSession(msg.Type, msg.Meta.SessionID, connectionID, msg.Meta.MessageID, status.ERR, err.Error())
	}

	msg.Meta.Conn.UserID = userID
	msg.Meta.Conn.Type = "user_id"

	respBytes, _ := json.Marshal(msg)

	return respBytes
}
