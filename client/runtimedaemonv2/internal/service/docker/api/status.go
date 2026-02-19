package api

import (
	"context"

	docker "github.com/docker/docker/client"
)

type ContainerStatus int32

const (
	RunningContainer ContainerStatus = iota
	LaunchingContainer
	StoppedContainer
	ErrorStatus
)

const (
	createdStatus    = "created"
	restartingStatus = "restarting"
	runningStatus    = "running"
	removingStatus   = "removing"
	pausedStatus     = "paused"
	exitedStatus     = "exited"
)

func (s service) GetContainerStatus(ctx context.Context, name string) (ContainerStatus, error) {
	containerInfo, err := s.GetContainer(ctx, name)
	if err != nil {
		return 0, err
	}

	container, err := s.dockerApi.ContainerInspect(ctx, containerInfo.ID)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return 0, ErrContainerNotFound
		}

		return 0, err
	}

	return s.statusResolver(container.State.Status), nil
}

func (s service) statusResolver(status string) ContainerStatus {
	switch status {
	case runningStatus:
		return RunningContainer

	case restartingStatus:
		return LaunchingContainer

	case pausedStatus, removingStatus, exitedStatus, createdStatus:
		return StoppedContainer
	}

	return StoppedContainer
}
