package event

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg"
	msgcontent "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
)

func (h handler) ClientExpired(ctx context.Context, m []byte) error {
	reqMSG, err := msg.New(m).Parse()
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	requestIDs, err := h.srv.Session().GetClientRents(ctx, reqMSG.Meta.Conn.UserID)
	if err != nil {
		return fmt.Errorf("%w: %w", model.ErrFailedToGetClientRents, err)
	}

	for _, id := range requestIDs {
		content := msgcontent.StopSessionReq{RequestID: id}

		handlerReq := createResp(reqMSG.Type, msg.Meta{}, content)

		go func() {
			if err := h.StopSession(ctx, handlerReq); err != nil {
				log.Errorf("failed to complite stop event from client expired: %s", err)
			}
		}()
	}

	return nil
}
