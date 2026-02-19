package service

import (
	"context"
	"errors"
	"fmt"

	dockerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
)

var (
	ErrDockerLaunching = errors.New("docker is launching")
	ErrDockerHaveError = errors.New("docker have error")
)

func (s service) resolveNetworkState(ctx context.Context) (network.AppNetworkState, error) {
	result := network.AppNetworkState{}

	log.Info("[network app state] try to get docker state")

	docker, err := s.docker.State(ctx)
	if err != nil {
		return network.AppNetworkState{}, err
	}

	// check if we even can get network state from docker
	switch docker.Health.Status {
	case dockerModel.Error:
		return network.AppNetworkState{}, fmt.Errorf("%w. Can not get ports state: %s", ErrDockerHaveError, docker.Health.StatusMsg)

	case dockerModel.Launching:
		return network.AppNetworkState{}, fmt.Errorf("%w. Can not get ports state: %s", ErrDockerLaunching, docker.Health.StatusMsg)
	}

	if docker.Container.State.Status != api.RunningContainer {
		return result, nil
	}

	log.Info("[network app state] try to get template")

	template, err := s.docker.Template().Get(ctx, docker.Container.UsedTemplateID)
	if err != nil {
		return result, fmt.Errorf("failed to get current template: %w", err)
	}

	log.Info("[network app state] try to get active ports")

	templatesPorts := template.Configuration.Ports
	containerPorts := docker.Container.State.Ports
	containerProxiedPorts := docker.Auth.Ports

	for _, template := range templatesPorts {
		// loop by container ports to find bind port from template
		for _, container := range containerPorts {
			if template.Port == container.Local {
				port := network.ActivePort{
					PortID: template.Port,

					Title:    template.Title,
					Port:     container.Host,
					Protocol: template.Protocol,
				}

				// check if port used by auth proxy
				for _, proxiedPort := range containerProxiedPorts {
					if container.Host == proxiedPort.InPort {
						port.Port = proxiedPort.OutPort
					}
				}

				result.ActivePorts = append(result.ActivePorts, port)

				break
			}
		}
	}

	return result, nil
}
