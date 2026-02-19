package api

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

func (s service) StartContainer(ctx context.Context, name string) error {
	if err := s.dockerApi.ContainerStart(ctx, name, container.StartOptions{}); err != nil {
		if docker.IsErrNotFound(err) {
			return ErrContainerNotFound
		}

		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}
