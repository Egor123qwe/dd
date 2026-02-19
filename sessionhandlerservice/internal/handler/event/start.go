package event

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	msgcontent "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content"
	rentModel "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
)

func (h handler) InitSession(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requesterMeta := reqMSG.Meta
	requesterMeta.Conn.Type = msg.ConnectionID

	messageID := reqMSG.Meta.MessageID

	var content msgcontent.InitSessionReq

	// parse content data (request)
	if err := reqMSG.ParseContent(&content); err != nil {
		err := fmt.Errorf("%w: %w", model.ErrInvalidContent, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	requesterMeta.SessionID = content.SessionID

	// request to session service to start session with settings from request
	// we can use meta from request, because only client can use start session content
	srvReq := session.InitReq{
		Client: session.Client{
			UserID: reqMSG.Meta.Conn.UserID,
		},

		SessionID: content.SessionID,

		Settings: rent.RequestSettings{
			Mode:       rent.MerchantMode(content.Settings.Mode),
			TemplateID: content.Settings.TemplateID,
		},
	}

	resp, err := h.srv.Session().Init(ctx, srvReq)
	if err != nil {
		err := fmt.Errorf("%w: %w", model.ErrFailedToStartSession, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	// request to merchant to start session with settings from request
	merchantMeta := msg.Meta{
		MessageID: messageID,
		SessionID: content.SessionID,
		Status:    string(msg.Ok),

		Conn: msg.Connection{
			ConnID: resp.Merchant.ConnID,
			Type:   msg.ConnectionID,
		},
	}

	merchantContent := msgcontent.MerchantRentStartReq{
		RequestID: resp.RequestID,
		SessionID: content.SessionID,
		ClientID:  requesterMeta.Conn.UserID,

		Settings: rentModel.ConvertToMerchantSettings(resp.Settings),
	}

	merchantResp := createResp(string(event.MerchantStartRentEvent), merchantMeta, merchantContent)

	if err := h.respondent.ws.Produce(ctx, merchantResp); err != nil {
		return fmt.Errorf("failed to produce message to merchant: %w", err)
	}

	clientMeta := msg.Meta{
		MessageID: messageID,
		Status:    string(msg.Ok),

		Conn: msg.Connection{
			UserID: reqMSG.Meta.Conn.UserID,
			Type:   msg.UserID,
		},
	}

	// response to client with "PendingStatus" status of session
	clientContent := msgcontent.SessionStatusResp{
		RequestID: resp.RequestID,
		SessionID: content.SessionID,

		CreatedAt: &resp.CreatedAt,

		Status: msgcontent.PendingStatus,
	}

	// in this content requester will be client
	clientResp := createResp(string(event.SessionStatusUpdatedEvent), clientMeta, clientContent)

	if err := h.respondent.ws.Produce(ctx, clientResp); err != nil {
		return fmt.Errorf("failed to produce message to client: %w", err)
	}

	serviceContent := msgcontent.WatchRentReq{
		RequestID: resp.RequestID,
	}

	// response to check requestID
	serviceResp := createResp(string(event.InitRent), msg.Meta{}, serviceContent)

	// produce this message to ws (output topic) because it will be used in service by checking this topic
	if err := h.respondent.ws.Produce(ctx, serviceResp); err != nil {
		return fmt.Errorf("failed to produce message to client: %w", err)
	}

	return nil
}

// StartSession starts rent session by merchant ready content
func (h handler) StartSession(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requesterMeta := reqMSG.Meta
	requesterMeta.Conn.Type = msg.ConnectionID

	messageID := reqMSG.Meta.MessageID

	var content msgcontent.ClientRentStartReq

	// parse content data (request)
	if err := reqMSG.ParseContent(&content); err != nil {
		err := fmt.Errorf("%w: %w", model.ErrInvalidContent, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	requesterMeta.SessionID = content.SessionID

	// request to session service to start session with settings from request
	srvReq := session.StartReq{
		RequestID: content.RequestID,
	}

	resp, err := h.srv.Session().Start(ctx, srvReq)
	if err != nil {
		err := fmt.Errorf("%w: %w", model.ErrFailedToStartSession, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	// request to merchant to start session with settings from request
	// take ids from request from merchant, because only merchant can start session
	merchantMeta := msg.Meta{
		MessageID: messageID,
		Status:    string(msg.Ok),

		Conn: msg.Connection{
			ConnID: reqMSG.Meta.Conn.ConnID,
			Type:   msg.ConnectionID,
		},
	}

	merchantContent := msgcontent.SessionStatusResp{
		RequestID: content.RequestID,
		SessionID: resp.SessionID,

		Status: msgcontent.RunningStatus,
	}

	merchantResp := createResp(string(event.SessionStatusUpdatedEvent), merchantMeta, merchantContent)

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

	clientContent := content
	clientContent.SessionID = resp.SessionID

	// in this content requester will be client
	clientResp := createResp(string(event.ClientStartRentEvent), clientMeta, clientContent)

	if err := h.respondent.ws.Produce(ctx, clientResp); err != nil {
		return fmt.Errorf("failed to produce message to client: %w", err)
	}

	return nil
}
