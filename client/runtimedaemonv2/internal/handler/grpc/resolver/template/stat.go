package template

import (
	"context"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
)

func (h Handler) GetStat(ctx context.Context, req *model.StatReq) (*model.StatResp, error) {
	stat, err := h.srv.GetStat(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.StatResp{
		TotalImgMem: float32(stat.TotalImgMem),
	}

	return result, nil
}
