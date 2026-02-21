package container

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	volumeModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
)

// proportionalStorage распределяет totalAvailable по пропорциям минимумов:
// контейнер (rootfs) и каждый том получают долю (min / totalUnits) * totalAvailable.
// Если totalUnits == 0, весь объём отдаётся контейнеру, томам — 0.
func proportionalStorage(totalAvailable int64, minContainer uint64, minVolumeBytes []uint64, numVolumes int) (containerStorage int64, volumeSizes []int64) {
	var totalUnits uint64 = minContainer
	for _, v := range minVolumeBytes {
		totalUnits += v
	}
	if totalUnits == 0 {
		return totalAvailable, make([]int64, numVolumes)
	}
	containerStorage = int64(uint64(totalAvailable) * minContainer / totalUnits)
	volumeSizes = make([]int64, numVolumes)
	for i := 0; i < numVolumes && i < len(minVolumeBytes); i++ {
		volumeSizes[i] = int64(uint64(totalAvailable) * minVolumeBytes[i] / totalUnits)
	}
	return containerStorage, volumeSizes
}

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

		totalAvailable := settings.Options.StorageLimitBytes
		cfg := template.Configuration
		containerStorage, volumeSizes := proportionalStorage(
			totalAvailable,
			cfg.MinStorageBytes,
			cfg.MinVolumeStorageBytes,
			len(template.Configuration.Volumes),
		)

		for i := range volumes {
			if i < len(volumeSizes) {
				volumes[i].SizeLimit = volumeSizes[i]
			}
		}

		createReq := api.CreateContainerReq{
			UseGPU:   template.Configuration.UseGPU,
			Image:    fmt.Sprintf("%s:%s", template.ImageName, template.ImageTag),
			Name:     Name(template.ID),
			Volumes:  volumes,
			Envs:     template.Configuration.Envs,
			CPUs:     settings.Options.CPULimit,
			Memory:   settings.Options.MemoryLimitBytes,
			Storage:  containerStorage,
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
