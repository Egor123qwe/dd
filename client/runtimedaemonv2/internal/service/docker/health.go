package docker

import (
	"context"
	"errors"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume/shared"
)

func (s service) CheckHealth(ctx context.Context) error {
	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()

	var removeSkipList []string

	templates, err := s.template.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get templates: %w", err)
	}

	for _, t := range templates {
		if _, ok := t.Usages[int32(containerModel.Local)]; ok {
			removeSkipList = append(removeSkipList, container.Name(t.ID))
		}
	}

	// clean expired containers
	removedNames, err := s.api.RemoveExpiredContainers(ctx, s.config.stoppedContainerTTL, removeSkipList)
	if err != nil {
		return fmt.Errorf("failed to clean container pull: %w", err)
	}

	// clean host usage volumes
	for _, name := range removedNames {
		id := container.ParseID(name)
		opts := template.RemoveVolumeOptions{HostUsageVolumes: true}

		if err := s.template.RemoveVolumes(ctx, id, opts); err != nil {
			log.Errorf("failed to remove host usage volumes: %v", err)
		}
	}

	// optimize templates mem usage
	if err := s.template.OptimizeMemUsage(ctx); err != nil {
		log.Errorf("failed to optimize templates mem usage")
	}

	// get current settings
	settings, err := s.storage.Settings().Get()
	if err != nil {
		return fmt.Errorf("failed to get container settings: %w", err)
	}

	// get components state
	proxy, err := s.proxy.State(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth proxy state: %w", err)
	}

	switch settings.Running {
	case true:
		{
			if settings.Options == nil {
				return fmt.Errorf("settings required, in request with param [running = true]")
			}

			// check container health
			template, err := s.template.Get(ctx, settings.Options.Container.TemplateID)
			if err != nil {
				return fmt.Errorf("failed to get downloaded template: %w", err)
			}

			containerState, err := s.container.State(ctx, settings.Options.Container.TemplateID)
			if err != nil && !errors.Is(err, api.ErrContainerNotFound) {
				return fmt.Errorf("failed to get container container state: %w", err)
			}

			exist := !errors.Is(err, api.ErrContainerNotFound)

			if containerState.Status == api.StoppedContainer || containerState.Status == api.ErrorStatus || !exist {
				startContainerSettings := containerModel.Settings{
					Usage:     settings.Options.Usage,
					ForUserID: settings.Options.ForUserID,

					Template: template,
				}

				_, err = s.container.Start(ctx, startContainerSettings)
				if err != nil {
					return fmt.Errorf("failed to start container: %w", err)
				}

				containerState, err = s.container.State(ctx, container.Name(settings.Options.Container.TemplateID))
				if err != nil && !errors.Is(err, api.ErrContainerNotFound) {
					return fmt.Errorf("failed to get container container state: %w", err)
				}

				if containerState.Status != api.RunningContainer {
					return fmt.Errorf("falid to start container")
				}
			}

			sharedFolderState, err := s.volume.Shared().State(ctx)
			if err != nil {
				return fmt.Errorf("failed to get shared volume state: %w", err)
			}

			if sharedFolderState.Enabled != (settings.Options.Container.SharedVolume != nil) {
				switch sharedFolderState.Enabled {
				case true:
					{
						if err := s.volume.Shared().Disconnect(ctx); err != nil {
							return fmt.Errorf("failed to disconnect shared volume: %w", err)
						}
					}

				case false:
					{
						shareFolderOpts := volume.SharedVolume{
							AccessKeyID:     settings.Options.Container.SharedVolume.AccessKeyID,
							SecretAccessKey: settings.Options.Container.SharedVolume.SecretAccessKey,

							BucketName: settings.Options.Container.SharedVolume.BucketName,
							Mount:      shared.DefaultSharedVolumeMount,
						}

						if settings.Options.Container.SharedVolume.Mount != nil {
							shareFolderOpts.Mount = *settings.Options.Container.SharedVolume.Mount
						}

						if err := s.volume.Shared().Connect(ctx, shareFolderOpts); err != nil {
							return fmt.Errorf("failed to connect shared volume: %w", err)
						}
					}

				}
			}

			if settings.Options.Auth.Credentials == nil {
				// this means that auth is disabled
				return nil
			}

			// check auth proxy health
			portsToProxy := s.filterContainerPortsByProxability(containerState.Ports, template.Configuration.Ports)

			if !s.isProxyPortsCorrect(
				s.containerPortsFromProxyPorts(proxy.Ports),
				portsToProxy,
			) {
				log.Info("proxy ports are not correct, restarting proxy")

				if err := s.proxy.Start(ctx, *settings.Options.Auth.Credentials, portsToProxy); err != nil {
					return fmt.Errorf("failed to start auth proxy: %w", err)
				}
			}
		}

	case false:
		if proxy.Enabled {
			s.proxy.Stop(ctx)
		}
	}

	return nil
}

type healthParams struct {
	inLaunching bool
}

func (s service) health(
	ctx context.Context,
	params healthParams,
	settings model.Settings,
	container model.ContainerState,
	proxy auth.State,
) model.Health {
	if params.inLaunching {
		return model.Health{
			Status:    model.Launching,
			StatusMsg: "Launching...",
		}
	}

	switch settings.Running {
	case true:
		{
			if settings.Options == nil {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "settings required, in request with param [running = true]",
				}
			}

			if container.State.Status == api.ErrorStatus {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "container have error state. " + container.State.StatusMsg,
				}
			}

			if container.State.Status == api.StoppedContainer {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "container stopped. " + container.State.StatusMsg,
				}
			}

			if settings.Options.Container.SharedVolume != nil && !container.SharedVolume.Enabled {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "shared volume is not enabled",
				}
			}

			if settings.Options.Auth.Credentials != nil {
				template, err := s.template.Get(ctx, settings.Options.Container.TemplateID)
				if err != nil {
					return model.Health{
						Status:    model.Error,
						StatusMsg: "failed to find template in storage: " + err.Error(),
					}
				}

				portsToProxy := s.filterContainerPortsByProxability(container.State.Ports, template.Configuration.Ports)

				if !proxy.Enabled && len(portsToProxy) > 0 {
					return model.Health{
						Status:    model.Error,
						StatusMsg: "proxy proxy is disabled",
					}
				}

				if !s.isProxyPortsCorrect(
					s.containerPortsFromProxyPorts(proxy.Ports), portsToProxy,
				) {
					return model.Health{
						Status:    model.Error,
						StatusMsg: "proxy used incorrect ports",
					}
				}
			}
		}

	case false:
		{
			if container.State.Status == api.RunningContainer {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "container running (but mode = stopped)",
				}
			}

			if proxy.Enabled {
				return model.Health{
					Status:    model.Error,
					StatusMsg: "proxy proxy is enabled (but mode = stopped)",
				}
			}
		}
	}

	return model.Health{Status: model.OK, StatusMsg: "OK"}
}
