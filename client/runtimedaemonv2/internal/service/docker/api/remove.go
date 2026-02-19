package api

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util"
)

const (
	dockerTimeLayout = "2006-01-02T15:04:05.999999999Z"
)

func (s service) RemoveContainer(ctx context.Context, name string) error {
	if err := s.StopContainer(ctx, name); err != nil {
		if docker.IsErrNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to stop container: %w", err)
	}

	if err := s.dockerApi.ContainerRemove(ctx, name, container.RemoveOptions{}); err != nil {
		if docker.IsErrNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// RemoveExpiredContainers removes expired containers from docker if they are not running for ttl time
// returns list of removed containers
// if filter [container names] is not empty, specified containers will be skipped!!!
func (s service) RemoveExpiredContainers(ctx context.Context, ttl time.Duration, filter []string) ([]string, error) {
	containers, err := s.dockerApi.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get container containers: %w", err)
	}

	removedNames := make([]string, 0)

	for _, container := range containers {
		containerInfo, err := s.dockerApi.ContainerInspect(ctx, container.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container: %w", err)
		}

		if !containerInfo.State.Running {
			var name string

			if len(container.Names) == 0 {
				continue
			}

			name = container.Names[0][1:] // remove leading slash prefix

			if util.InList(name, filter) {
				continue
			}

			t, err := time.Parse(dockerTimeLayout, containerInfo.State.FinishedAt)
			if err != nil {
				log.Errorf("failed to parse time: %v", err)

				continue
			}

			if time.Now().UTC().After(t.Add(ttl)) {
				if err := s.RemoveContainer(ctx, container.ID); err != nil {
					log.Errorf("failed to remove container: %v", err)

					continue
				}

				removedNames = append(removedNames, name)
			}
		}
	}

	return removedNames, nil
}
