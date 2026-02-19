package event

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	msgcontent "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
)

func (h handler) StopSession(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requesterMeta := reqMSG.Meta
	requesterMeta.Conn.Type = msg.ConnectionID

	messageID := reqMSG.Meta.MessageID

	// parse content if it exists
	content := &msgcontent.StopSessionReq{}

	// parse content data (request)
	if err := reqMSG.ParseContent(&content); err != nil {
		err := fmt.Errorf("%w: %w", model.ErrInvalidContent, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	params := h.stopReasonByEvent(event.Event(reqMSG.Type))

	if content.Reason != "" {
		params.reason = fmt.Sprintf("%s. %s", content.Reason, params.reason)
	}

	// request to session service to start session with settings from request
	srvReq := session.StopReq{
		RequestID: content.RequestID,
		Reason:    params.reason,
	}

	resp, err := h.srv.Session().Stop(ctx, srvReq)
	if err != nil {
		err := fmt.Errorf("%w: %w", model.ErrFailedToStopSession, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	// resolve session stopper
	if event.Event(reqMSG.Type) == event.StopSessionEvent {
		switch reqMSG.Meta.Conn.UserID {
		case resp.Client.UserID:
			params.initiator = msgcontent.ClientInitiator

		case resp.Merchant.UserID:
			params.initiator = msgcontent.MerchantInitiator
		}
	}

	params.reason += fmt.Sprintf(": %s stopped session", params.initiator)

	sessionStatus := msgcontent.SessionStatusResp{
		RequestID: content.RequestID,
		SessionID: resp.SessionID,

		Status:    msgcontent.StoppedStatus,
		StatusMsg: params.reason,

		Initiator: params.initiator,
	}

	// request to merchant to start session with settings from request
	merchantMeta := msg.Meta{
		MessageID: messageID,
		Status:    string(msg.Ok),

		Conn: msg.Connection{
			ConnID: resp.Merchant.ConnID,
			Type:   msg.ConnectionID,
		},
	}

	merchantResp := createResp(string(event.SessionStatusUpdatedEvent), merchantMeta, sessionStatus)

	if err := h.respondent.ws.Produce(ctx, merchantResp); err != nil {
		return fmt.Errorf("failed to produce message to merchant: %w", err)
	}

	clientMeta := msg.Meta{
		MessageID: messageID,
		Status:    string(msg.Ok),

		Conn: msg.Connection{
			UserID: resp.Client.UserID,
			Type:   msg.UserID,
		},
	}

	clientResp := createResp(string(event.SessionStatusUpdatedEvent), clientMeta, sessionStatus)

	if err := h.respondent.ws.Produce(ctx, clientResp); err != nil {
		return fmt.Errorf("failed to produce message to client: %w", err)
	}

	return nil
}

type stopParams struct {
	initiator msgcontent.Initiator
	reason    string
}

func (h handler) stopReasonByEvent(e event.Event) stopParams {
	var result stopParams

	switch e {
	case event.ExpiredRentEvent:
		result.reason = "rent expired"
		result.initiator = msgcontent.ClientInitiator

	case event.ExpiredClientEvent:
		result.reason = "client expired"
		result.initiator = msgcontent.ClientInitiator

	case event.ExpiredSessionEvent:
		result.reason = "merchant session expired"
		result.initiator = msgcontent.MerchantInitiator

	case event.StopSessionEvent:
		result.reason = "user stopped session"
		result.initiator = msgcontent.UnknownInitiator

	case event.RentRequestStatusUpdatedEvent:
		result.reason = "merchant error while processing request"
		result.initiator = msgcontent.MerchantInitiator

	case event.ShareP2PStop:
		result.reason = "merchant stopped session"
		result.initiator = msgcontent.MerchantInitiator

	default:
		result.reason = "not defined"
		result.initiator = msgcontent.UnknownInitiator
	}

	return result
}
