package worker

import (
	"context"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
)

func (h Handler) GetSysInfo(ctx context.Context, req *model.SystemInfoReq) (*model.SystemInfo, error) {
	resp, err := h.srv.SysInfo(ctx)
	if err != nil {
		return nil, err
	}

	result := convertHardwareToHandlerType(resp)

	return result, nil
}
