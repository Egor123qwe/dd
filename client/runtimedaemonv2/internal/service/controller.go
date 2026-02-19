package service

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util"
)

func (s service) changeMode(ctx context.Context, req runtime.Configuration) error {
	approveSettings := util.Ptr(req.Settings.Mode == runtime.ModeDisable)
	defer s.approveChangeMode(ctx, approveSettings)

	s.mutex.mode.Lock()
	defer s.mutex.mode.Unlock()

	// block health checker
	if lock := s.mutex.health.TryLock(); !lock {
		s.healthCheckController.Cancel()
		s.mutex.health.Lock()
	}

	defer s.mutex.health.Unlock()

	settings := runtime.Settings{
		Mode: req.Settings.Mode,
	}

	if err := s.storage.Runtime().Settings().Set(settings); err != nil {
		return fmt.Errorf("failed to update runtime health: %w", err)
	}

	// runtime requirements by chosen mode
	components := s.resolveSettings(settings.Mode)

	dockerSettings := docker.Settings{
		Running: components.docker,
	}

	if req.Components.Docker != nil {
		dockerSettings.Options = &docker.Options{
			Usage:     s.resolveContainerUsage(settings.Mode),
			ForUserID: req.Components.Docker.ForUserID,
			Container: req.Components.Docker.Container,
			Auth:      req.Components.Docker.Auth,
		}
	}

	if err := s.docker.ChangeMode(ctx, dockerSettings); err != nil {
		log.Errorf("failed to configure container: %s", err)

		return fmt.Errorf("failed to configure container: %w", err)
	}

	networkSettings := model.Settings{
		Mode:    components.network,
		Connect: req.Components.Network,
	}

	appNetwork, err := s.resolveNetworkState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get app net mode: %w", err)
	}

	if err := s.network.ChangeMode(ctx, networkSettings, appNetwork); err != nil {
		log.Errorf("failed to configure network: %s", err)

		return fmt.Errorf("failed to configure network: %w", err)
	}

	*approveSettings = true

	return nil
}

// it's like commit/rollback realization
func (s service) approveChangeMode(ctx context.Context, approve *bool) error {
	if *approve {
		return nil
	}

	return s.rollbackSettings(ctx)
}

func (s service) rollbackSettings(ctx context.Context) error {
	runtimeDisabler := runtime.Configuration{Settings: runtime.Settings{Mode: runtime.ModeDisable}}

	if err := s.changeMode(context.Background(), runtimeDisabler); err != nil {
		log.Errorf("failed to rollback settings: %s", err)

		return fmt.Errorf("failed to rollback settings: %w", err)
	}

	return nil
}

func (s service) getCurrentConfiguration() (runtime.Configuration, error) {
	result := runtime.Configuration{}
	var err error

	result.Settings, err = s.storage.Runtime().Settings().Get()
	if err != nil {
		return result, fmt.Errorf("failed to get runtime config: %w", err)
	}

	network, err := s.storage.Network().Settings().Get()
	if err != nil {
		return result, fmt.Errorf("failed to get network config: %w", err)
	}

	result.Components.Network = network.Connect

	docker, err := s.storage.Docker().Settings().Get()
	if err != nil {
		return result, fmt.Errorf("failed to get container config: %w", err)
	}

	if docker.Options != nil {
		result.Components.Docker = &runtime.DockerOptions{
			ForUserID: docker.Options.ForUserID,
			Container: docker.Options.Container,
			Auth:      docker.Options.Auth,
		}
	}

	return result, nil
}
