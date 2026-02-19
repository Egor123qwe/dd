package merchant

import (
	"context"
	"encoding/json"
	"errors"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka/producer"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/message"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/response"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/sharep2p"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/status"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	merchantService MerchantService
	producer        producer.Producer
	cfg             *config.Config
}

type MerchantService interface {
	ShareP2PInit(ctx context.Context, req sharep2p.InitMerchantRequest) (string, float32, error)
	ShareP2Pready(ctx context.Context, sessionID string) error
	ShareP2PStop(ctx context.Context, sessionID string) error
}

func New(merchantService MerchantService, producer producer.Producer, cfg config.Config) *Handler {
	handler := &Handler{
		merchantService: merchantService,
		producer:        producer,
		cfg:             &cfg,
	}
	return handler
}

func (h *Handler) ShareP2PInit(ctx context.Context, mt string, msg []byte) []byte {
	var req sharep2p.InitMerchantRequest

	if err := json.Unmarshal(msg, &req); err != nil {
		return response.BaseError(mt, status.ERR, sharep2p.ErrInavlidRequest.Error())
	}

	sessionID, totalPrice, err := h.merchantService.ShareP2PInit(ctx, req)
	if err != nil {
		msg := sharep2p.ErrCanNotCreateSession.Error()
		if errors.Is(err, sharep2p.ErrUnidentifiedHardware) {
			msg = err.Error()
		} else {
			log.Error().Err(err).Msg(sharep2p.ErrCanNotCreateSession.Error())
		}
		return response.ErrorWS(
			mt,
			req.Meta.Conn.ConnectionID,
			req.Meta.MessageID,
			status.ERR,
			msg,
		)
	}

	resp := &sharep2p.InitMerchantResponse{
		Type: mt,
		Meta: message.Meta{
			Status:    status.OK,
			SessionID: sessionID,
			MessageID: req.Meta.MessageID,
			Conn: message.Connection{
				ConnectionID: req.Meta.Conn.ConnectionID,
				Type:         "conn_id",
			},
		},
		Content: sharep2p.ContentInitMerchantResponse{
			SessionID: sessionID,
			Price:     totalPrice,
		},
	}

	responseBytes, _ := json.Marshal(resp)

	return responseBytes
}

func (h *Handler) ShareP2PReady(ctx context.Context, mt string, msg []byte) []byte {
	var req sharep2p.ReadyMerchantRequest

	if err := json.Unmarshal(msg, &req); err != nil {
		return response.BaseError(mt, status.ERR, sharep2p.ErrInavlidRequest.Error())
	}

	err := h.merchantService.ShareP2Pready(ctx, req.Content.SessionID)
	if err != nil {
		log.Error().Err(err).Msg(sharep2p.ErrCanNotSetSessionStatus.Error())
		return response.ErrorSession(
			mt,
			req.Content.SessionID,
			req.Meta.Conn.ConnectionID,
			req.Meta.MessageID,
			status.ERR,
			sharep2p.ErrCanNotSetSessionStatus.Error(),
		)
	}

	resp := &sharep2p.ReadyMerchantResponse{
		Type: mt,
		Meta: message.Meta{
			Status:    status.OK,
			MessageID: req.Meta.MessageID,
			Conn: message.Connection{
				ConnectionID: req.Meta.Conn.ConnectionID,
				Type:         "conn_id",
			},
		},
		Content: sharep2p.ContentReadyMerchant{
			SessionID: req.Content.SessionID,
		},
	}

	responseBytes, _ := json.Marshal(resp)

	return responseBytes
}

func (h *Handler) ShareP2PStop(ctx context.Context, mt string, msg []byte) []byte {
	var req sharep2p.Stop
	if err := json.Unmarshal(msg, &req); err != nil {
		return response.BaseError(mt, status.ERR, sharep2p.ErrInavlidRequest.Error())
	}

	err := h.merchantService.ShareP2PStop(ctx, req.Content.SessionID)
	if err != nil {
		log.Error().Err(err).Msg(sharep2p.ErrCanNotDeleteSession.Error())
		return response.ErrorSession(
			mt,
			req.Content.SessionID,
			req.Meta.Conn.ConnectionID,
			req.Meta.MessageID,
			status.ERR,
			sharep2p.ErrCanNotDeleteSession.Error(),
		)
	}

	req.Meta.Status = status.OK
	req.Meta.Conn.Type = "conn_id"
	req.Meta.SessionID = req.Content.SessionID

	responseBytes, _ := json.Marshal(req)

	return responseBytes
}
