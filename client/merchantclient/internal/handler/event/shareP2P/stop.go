package shareP2P

import (
	"context"
	"fmt"

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

	if err := h.usecase.Rent().StopEvent(ctx, req.Meta.SessionID); err != nil {
		return fmt.Errorf("failed to stop merchant by %s event: %w", req.Type, err)
	}

	return fmt.Errorf("merchant stoped by %s event: %w", req.Type, model.FatalErr)
}
