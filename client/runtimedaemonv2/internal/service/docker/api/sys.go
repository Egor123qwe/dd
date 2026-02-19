package api

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
)

type SysInfo struct {
	DiskUsage types.DiskUsage
}

func (s service) Sys(ctx context.Context) (SysInfo, error) {
	result := SysInfo{}
	var err error

	result.DiskUsage, err = s.dockerApi.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return SysInfo{}, fmt.Errorf("failed to get container system info: %w", err)
	}

	return result, nil
}
