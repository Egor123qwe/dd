package docker

import (
	"context"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume/shared"
)

func (s service) ChangeMode(ctx context.Context, req model.Settings) error {
	s.modeMutex.Lock()
	defer s.modeMutex.Unlock()

	// block health checker
	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()

	if err := s.storage.Settings().Set(req); err != nil {
		return fmt.Errorf("failed to update container settings: %w", err)
	}

	switch req.Running {
	case true:
		if req.Options == nil {
			return fmt.Errorf("settings required, in request with param [running = true]")
		}

		// configure container
		template, err := s.template.Get(ctx, req.Options.Container.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get downloaded template: %w", err)
		}

		startContainerSettings := containerModel.Settings{
			Usage:     req.Options.Usage,
			ForUserID: req.Options.ForUserID,

			Template: template,
		}

		if req.Options.Container.SharedVolume != nil {
			startContainerSettings.Options.SharedVolume = &volume.SharedVolume{
				AccessKeyID:     req.Options.Container.SharedVolume.AccessKeyID,
				SecretAccessKey: req.Options.Container.SharedVolume.SecretAccessKey,

				BucketName: req.Options.Container.SharedVolume.BucketName,
				Mount:      shared.DefaultSharedVolumeMount,
			}

			if req.Options.Container.SharedVolume.Mount != nil {
				startContainerSettings.Options.SharedVolume.Mount = *req.Options.Container.SharedVolume.Mount
			}

			if err := s.volume.Shared().Connect(ctx, *startContainerSettings.Options.SharedVolume); err != nil {
				return fmt.Errorf("failed to connect shared volume: %w", err)
			}
		}

		_, err = s.container.Start(ctx, startContainerSettings)
		if err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		if err := s.template.AddUsage(ctx, req.Options.Container.TemplateID, req.Options.Usage); err != nil {
			return fmt.Errorf("failed to add usage: %w", err)
		}

		if req.Options.Auth.Credentials == nil {
			return nil
		}

		// configure auth proxy
		container, err := s.container.State(ctx, req.Options.Container.TemplateID)
		if err != nil {
			return fmt.Errorf("falied to get container state")
		}

		if container.Status != api.RunningContainer {
			return fmt.Errorf("container not started. nothing to proxy")
		}

		portsToProxy := s.filterContainerPortsByProxability(container.Ports, template.Configuration.Ports)

		if err := s.proxy.Start(ctx, *req.Options.Auth.Credentials, portsToProxy); err != nil {
			return fmt.Errorf("failed to auth proxy: %w", err)
		}

	case false:
		if err := s.api.StopAllContainers(ctx); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}

		if err := s.volume.Shared().Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect shared volume: %w", err)
		}

		s.proxy.Stop(ctx)
	}

	return nil
}
