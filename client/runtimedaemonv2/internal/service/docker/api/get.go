package api

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

var (
	ErrContainerNotFound = fmt.Errorf("container not found")
)

func (s service) GetContainer(ctx context.Context, name string) (types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("name", name)

	containers, err := s.dockerApi.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})

	if err != nil {
		return types.Container{}, err
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if containerName == "/"+name {
				return container, nil
			}
		}
	}

	return types.Container{}, ErrContainerNotFound
}
