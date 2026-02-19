package template

import (
	"context"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/template"
)

func (h Handler) CleanCache(ctx context.Context, req *model.CleanCacheReq) (*model.CleanCacheResp, error) {
	srvReqOpts := template.RemoveVolumeOptions{
		HostUsageVolumes:  req.CleanHostUsageCache,
		LocalUsageVolumes: req.CleanLocalUsageCache,
	}

	if err := h.srv.RemoveVolumes(ctx, req.TemplateID, srvReqOpts); err != nil {
		return nil, fmt.Errorf("failed to remove template cache: %w", err)
	}

	return &model.CleanCacheResp{}, nil
}
