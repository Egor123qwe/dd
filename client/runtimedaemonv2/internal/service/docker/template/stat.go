package template

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
)

func (s service) GetStat(ctx context.Context) (template.Stat, error) {
	result := template.Stat{}

	info, err := s.api.Sys(ctx)
	if err != nil {
		return template.Stat{}, fmt.Errorf("failed to get system info: %w", err)
	}

	result.TotalImgMem = float64(info.DiskUsage.LayersSize) / 1000

	return result, nil
}
