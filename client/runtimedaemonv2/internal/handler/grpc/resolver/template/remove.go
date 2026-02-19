package template

import (
	"context"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
)

func (h Handler) Remove(ctx context.Context, req *model.RemoveTemplateReq) (*model.RemoveTemplateResp, error) {
	if err := h.srv.Remove(ctx, req.TemplateID); err != nil {
		return nil, fmt.Errorf("failed to remove template: %w", err)
	}

	return &model.RemoveTemplateResp{}, nil
}
