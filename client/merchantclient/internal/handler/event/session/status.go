package session

import (
	"context"
	"fmt"

	msgModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	session "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session"
)

func (h handler) StatusUpdated(ctx context.Context, m []byte) error {
	req, _ := msgModel.Unmarshal(m)

	if req.Meta.Err != nil && req.Meta.Status != string(msgModel.Ok) {
		return fmt.Errorf("%s event with error: %s", req.Type, req.Meta.Err.Message)
	}

	var reqData session.SessionStatusResp

	if err := req.UnmarshalContent(&reqData); err != nil {
		return fmt.Errorf("failed to parse message content: %w", err)
	}

	switch reqData.Status {
	case session.StoppedStatus:
		log.Infof("client requested to stop session [%s]", reqData.RequestID)

		if err := h.usecase.Session().StopEvent(ctx, reqData.RequestID); err != nil {
			return fmt.Errorf("failed to stop merchant by %s event: %w", req.Type, err)
		}

		log.Infof("session [%s] stoped by %s event", reqData.RequestID, req.Type)

	case session.RunningStatus:
		{
			// this status implemented in services layer (by request-response pattern)
			log.Debugf("impossible event params: %s with status \"running\" in handlers", req.Type)

			return nil
		}
	}

	return nil
}
