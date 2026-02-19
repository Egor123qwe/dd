package api

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

func (s service) GetContainerLogsReader(ctx context.Context, options container.LogsOptions, name string) (io.ReadCloser, error) {
	logsReader, err := s.dockerApi.ContainerLogs(ctx, name, options)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return nil, ErrContainerNotFound
		}

		return nil, fmt.Errorf("failed to get container logs scanner: %w", err)
	}

	return logsReader, nil
}

func (s service) GetContainerTTY(ctx context.Context, name string) (bool, error) {
	runningContainer, err := s.GetContainer(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed in trying to find container: %w", err)
	}

	container, err := s.dockerApi.ContainerInspect(ctx, runningContainer.ID)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return false, ErrContainerNotFound
		}

		return false, fmt.Errorf("failed to inspect container: %w", err)
	}

	return container.Config.Tty, nil
}
