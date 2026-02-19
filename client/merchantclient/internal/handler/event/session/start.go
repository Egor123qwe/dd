package session

import (
	"context"
	"fmt"

	msgModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session"
)

func (h handler) Start(ctx context.Context, m []byte) error {
	req, _ := msgModel.Unmarshal(m)

	if req.Meta.Err != nil && req.Meta.Status != string(msgModel.Ok) {
		return fmt.Errorf("execution of the %s event failed with error: %s", req.Type, req.Meta.Err.Message)
	}

	var reqData session.MerchantRentStartReq

	if err := req.UnmarshalContent(&reqData); err != nil {
		return fmt.Errorf("failed to parse message content: %w", err)
	}

	if err := h.usecase.Session().Start(ctx, reqData); err != nil {
		return fmt.Errorf("failed to start client session: %w", err)
	}

	return nil
}
