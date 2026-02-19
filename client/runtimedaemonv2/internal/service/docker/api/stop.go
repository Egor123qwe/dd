package api

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

func (s service) StopContainer(ctx context.Context, name string) error {
	if err := s.dockerApi.ContainerStop(ctx, name, container.StopOptions{}); err != nil {
		if docker.IsErrNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

func (s service) StopAllContainers(ctx context.Context) error {
	containers, err := s.dockerApi.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to get container containers: %w", err)
	}

	for _, container := range containers {
		if err := s.StopContainer(ctx, container.ID); err != nil {
			log.Errorf("failed to stop container: %v", err)
		}
	}

	return nil
}
