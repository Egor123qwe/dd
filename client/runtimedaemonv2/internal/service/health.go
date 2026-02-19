package service

import (
	"context"
	"fmt"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/health"
)

const (
	healthCheckInterval = 10 * time.Second
)

func (s service) healthChecker(ctx context.Context) {
	for {
		select {
		case <-time.After(healthCheckInterval):

		case <-ctx.Done():
			return
		}

		s.mutex.health.Lock()

		checkCtx, cancel := context.WithCancel(context.Background())
		s.healthCheckController.SetCancelFn(cancel)

		// check docker state
		if err := s.docker.CheckHealth(checkCtx); err != nil {
			log.Errorf("failed to check container health: %s", err)
		}

		// check network state
		appNetwork, err := s.resolveNetworkState(checkCtx)
		if err != nil {
			log.Errorf("failed to get app net state: %s", err)

		} else {
			if err := s.network.CheckHealth(checkCtx, appNetwork); err != nil {
				log.Errorf("failed to check network health: %s", err)
			}
		}

		s.mutex.health.Unlock()
	}
}

func (s service) health(inLaunching bool, state health.Components) health.Health {
	result := health.Health{Status: health.OK}

	if inLaunching {
		return health.Health{Status: health.Launching}
	}

	switch state.Docker.Health.Status {
	case docker.Launching:
		result.Status = s.updateStatusByPriority(result.Status, health.Launching)
		result.StatusMsg += fmt.Sprintf("Docker: container is launching: [%s]. ", state.Docker.Health.StatusMsg)

	case docker.Error:
		result.Status = s.updateStatusByPriority(result.Status, health.Error)
		result.StatusMsg += fmt.Sprintf("Docker: container is not running: [%s]. ", state.Docker.Health.StatusMsg)
	}

	switch state.Network.Health.Status {
	case network.Launching:
		result.Status = s.updateStatusByPriority(result.Status, health.Launching)
		result.StatusMsg += fmt.Sprintf("Network: network is launching: [%s]. ", state.Network.Health.StatusMsg)

	case network.Error:
		result.Status = s.updateStatusByPriority(result.Status, health.Error)
		result.StatusMsg += fmt.Sprintf("Network: network is not running: [%s]. ", state.Docker.Health.StatusMsg)
	}

	return result
}

func (s service) updateStatusByPriority(current health.Status, update health.Status) health.Status {
	priority := map[health.Status]int{
		health.OK:        0,
		health.Launching: 1,
		health.Error:     2,
	}

	if priority[current] > priority[update] {
		return current
	}

	return update
}
