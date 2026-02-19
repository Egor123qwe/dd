package event

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	msgcontent "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content"
	rentModel "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
)

func (h handler) UpdateRentStatus(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requesterMeta := reqMSG.Meta
	requesterMeta.Conn.Type = msg.ConnectionID

	var content msgcontent.RentRequestStatusUpdatedReq

	// parse content data (request)
	if err := reqMSG.ParseContent(&content); err != nil {
		err := fmt.Errorf("%w: %w", model.ErrInvalidContent, err)

		return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
	}

	switch content.Status {
	case msgcontent.RunningMerchantRentStatus:
		{
			rentSettings, err := h.srv.Rent().GetRentSettings(ctx, content.RequestID)
			if err != nil {
				err := fmt.Errorf("%w: %w", model.ErrFailedToGetRentSettings, err)

				return h.respondent.ws.Produce(ctx, createErrorResp(reqMSG.Type, requesterMeta, err))
			}

			content := msgcontent.ClientRentStartReq{
				RequestID: content.RequestID,
				Settings:  rentModel.ConvertToClientSettings(rentSettings),
			}

			return h.StartSession(ctx, createResp(string(event.RentRequestStatusUpdatedEvent), requesterMeta, content))
		}

	case msgcontent.ErrorMerchantRentStatus:
		{
			content := msgcontent.StopSessionReq{
				RequestID: content.RequestID,
				Reason:    content.StatusMsg,
			}

			return h.StopSession(ctx, createResp(string(event.RentRequestStatusUpdatedEvent), requesterMeta, content))
		}
	}

	return nil
}
