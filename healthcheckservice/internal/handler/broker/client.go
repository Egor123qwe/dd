package broker

import (
	"context"
	"encoding/json"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
)

func (h handler) HandleClientStartRent(ctx context.Context, msg []byte) error {
	var ClientReq message.ClientRent

	err := json.Unmarshal(msg, &ClientReq)

	if err != nil {
		return ErrParseMessage
	}

	err = h.service.Status().RentClient(ctx, ClientReq)
	if err != nil {
		return err
	}

	return nil
}
