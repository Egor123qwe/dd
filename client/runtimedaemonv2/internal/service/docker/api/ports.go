package api

import (
	"context"
	"fmt"

	docker "github.com/docker/docker/client"
)

type Port struct {
	Host  string
	Local string
}

func (s service) GetContainerPorts(ctx context.Context, name string) ([]Port, error) {
	data, err := s.dockerApi.ContainerInspect(ctx, name)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return nil, ErrContainerNotFound
		}

		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	var result []Port

	for local, host := range data.NetworkSettings.Ports {
		if len(host) == 0 {
			continue
		}

		port := Port{
			Local: local.Port(),
			Host:  host[0].HostPort,
		}

		result = append(result, port)
	}

	return result, nil
}
