package worker

import (
	"context"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/event"
)

func (h Handler) ExecuteEvent(ctx context.Context, req *model.ExecuteEventReq) (*model.ExecuteEventResp, error) {
	if err := h.srv.ExecuteEvent(ctx, event.Event(req.Event)); err != nil {
		return nil, err
	}

	return nil, nil
}
