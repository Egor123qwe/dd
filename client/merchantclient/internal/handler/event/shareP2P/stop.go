package shareP2P

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	msgModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/model"
)

func (h handler) Stop(ctx context.Context, m []byte) error {
	req, _ := msgModel.Unmarshal(m)

	if req.Meta.Err != nil && req.Meta.Status != string(msgModel.Ok) {
		if req.Meta.Err.Message == "" {
			req.Meta.Err.Message = "unknown error"
		}

		return fmt.Errorf("failed to stop merchant by %s event: %s", req.Type, req.Meta.Err.Message)
	}

	// share-p2p-stop при остановке аренды покупателем: переводим в Ready, сессия остаётся активной.
	// expired-session: полный сброс (Disabled) и отключение.
	if req.Type == string(event.ShareP2PStop) {
		if err := h.usecase.Session().TransitionToReadyAfterRentEnd(ctx, req.Meta.SessionID); err != nil {
			return fmt.Errorf("failed to transition to ready by %s event: %w", req.Type, err)
		}
		return nil
	}

	// ExpiredSession или неизвестное событие — полный сброс
	if err := h.usecase.Rent().StopEvent(ctx, req.Meta.SessionID); err != nil {
		return fmt.Errorf("failed to stop merchant by %s event: %w", req.Type, err)
	}

	return fmt.Errorf("merchant stoped by %s event: %w", req.Type, model.FatalErr)
}
