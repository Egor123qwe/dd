package container

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	volumeModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
)

func (s service) Start(ctx context.Context, settings container.Settings) (string, error) {
	var containerId string
	template := settings.Template

	// try to find this container
	currentContainer, err := s.api.GetContainer(ctx, Name(template.ID))
	if err != nil && !errors.Is(err, api.ErrContainerNotFound) {
		return "", fmt.Errorf("failed in trying find container: %w", err)
	}

	containerExist := !errors.Is(err, api.ErrContainerNotFound)

	// create container if container not found
	if !containerExist {
		var volumes []api.Volume

		for _, volume := range template.Configuration.Volumes {
			volumes = append(
				volumes, s.volume.Volume(volumeModel.Usage(settings.Usage), template.ID, settings.ForUserID, volume),
			)
		}

		if settings.Options.SharedVolume != nil {
			volumes = append(
				volumes, s.volume.Volume(volumeModel.Shared, "", "", settings.Options.SharedVolume.Mount),
			)
		}

		createReq := api.CreateContainerReq{
			UseGPU:  template.Configuration.UseGPU,
			Image:   fmt.Sprintf("%s:%s", template.ImageName, template.ImageTag),
			Name:    Name(template.ID),
			Volumes: volumes,
			Envs:    template.Configuration.Envs,
		}

		containerId, err = s.api.CreateContainer(ctx, createReq)
		if err != nil {
			return "", fmt.Errorf("failed to create container: %w", err)
		}

	} else {
		// save container id if container found
		containerId = currentContainer.ID
	}

	// start container [container skip start if container already running]
	if err := s.api.StartContainer(ctx, containerId); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return containerId, nil
}
