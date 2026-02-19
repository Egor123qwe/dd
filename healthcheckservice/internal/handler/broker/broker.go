package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/service"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

var ErrParseMessage = errors.New("Can`t parse message")

const (
	ConnectionTypeConnID = "conn_id"
	ConnectionTypeUserID = "user_id"
)

type Handler interface {
	HandleOutput(ctx context.Context, msg []byte) error
	HandleClientStartRent(ctx context.Context, msg []byte) error
	HandleStatusUpdated(ctx context.Context, msg []byte) error
	HandleMerchantStartRent(ctx context.Context, msg []byte) error
	HandleExpiredSession(ctx context.Context, msg []byte) error
}

type handler struct {
	service service.Service
	log     slog.Logger
}

func New(cfg config.Config, log slog.Logger, storage storage.Storage) Handler {
	service := service.New(cfg, log, storage)

	handler := handler{
		service: service,
		log:     log,
	}

	return handler
}

func (h handler) HandleOutput(ctx context.Context, msg []byte) error {
	const op = "output"
	var fullMsg message.FullMessage

	err := json.Unmarshal(msg, &fullMsg)
	if err != nil {
		return fmt.Errorf("%s : can`t parse message: %v", op, err)
	}

	h.log.Debug("message received")

	switch fullMsg.Type {
		
	case "client-start-rent":
		return h.HandleClientStartRent(ctx, msg)

	case "session-status-updated":
		return h.HandleStatusUpdated(ctx, msg)

	case "merchant-start-rent":
		return h.HandleMerchantStartRent(ctx, msg)

	case "expired-session":
		return h.HandleExpiredSession(ctx, msg)

	default:
		return fmt.Errorf("unknown message type: %s", fullMsg.Type)
	}
}

func (h handler) HandleStatusUpdated(ctx context.Context, msg []byte) error {
	op := "Handles status update"
	var (
		ClientReq   message.ClientRent
		MerchantReq message.MerchantRent
		fullMsg     message.FullMessage
	)

	err := json.Unmarshal(msg, &fullMsg)
	if err != nil {
		return fmt.Errorf("%s : can`t parse message: %v", op, err)
	}

	switch fullMsg.Meta.Conn.Type {

	case ConnectionTypeConnID:
		err = json.Unmarshal(msg, &MerchantReq)
		if err != nil {
			return ErrParseMessage
		}
		err = h.service.Status().SetStatusMerchant(ctx, MerchantReq)
		if err != nil {
			return err
		}

	case ConnectionTypeUserID:
		err = json.Unmarshal(msg, &ClientReq)
		if err != nil {
			return ErrParseMessage
		}

		err = h.service.Status().RentClient(ctx, ClientReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h handler) HandleExpiredSession(ctx context.Context, msg []byte) error {
	op := "Handles expired session"
	var fullMsg message.FullMessage

	err := json.Unmarshal(msg, &fullMsg)
	if err != nil {
		return fmt.Errorf("%s : can`t parse message: %v", op, err)
	}

	err = h.service.Status().SessionExpiredMerchant(ctx, fullMsg)
	if err != nil {
		return err
	}

	return nil
}
