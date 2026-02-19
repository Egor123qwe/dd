package container

import (
	"context"
	"fmt"

	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume"
	storage "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	defaultContainerName = "container"
)

var log = logger.NewLogger("container", logger.DefaultWithSentry())

// Service is abstract layer on container api to use container in container
type Service interface {
	Start(ctx context.Context, settings containerModel.Settings) (string, error)
	Stop(ctx context.Context, templateID string) error

	State(ctx context.Context, templateID string) (containerModel.State, error)

	Logs(ctx context.Context, req containerModel.LogReq) (<-chan string, error)
}

type service struct {
	api    api.Api
	volume volume.Service

	storage storage.DockerRepo
}

func New(api api.Api, volume volume.Service, storage storage.DockerRepo) Service {
	srv := &service{
		api:     api,
		volume:  volume,
		storage: storage,
	}

	return srv
}

// Name return container name in container
func Name(templateID string) string {
	return fmt.Sprintf("%s-%s", defaultContainerName, templateID)
}

// ParseID Returns template ID from name
func ParseID(name string) string {
	if len(name) < len(defaultContainerName)+1 {
		return ""
	}

	return name[len(defaultContainerName)+1:]
}

func (s service) State(ctx context.Context, templateID string) (containerModel.State, error) {
	log.Info("[container state] started")

	result := containerModel.State{}
	var err error

	result.Status, err = s.api.GetContainerStatus(ctx, Name(templateID))
	if err != nil {
		result.Status = api.StoppedContainer
		result.StatusMsg = fmt.Errorf("failed to get container status: %w", err).Error()

		return result, nil
	}

	log.Info("[container state] Successful get container status")

	if result.Status == api.RunningContainer {
		log.Info("[container state] try to get container ports")

		result.Ports, err = s.api.GetContainerPorts(ctx, Name(templateID))
		if err != nil {
			result.Status = api.ErrorStatus
			result.StatusMsg = fmt.Errorf("failed to get container ports: %w", err).Error()

			return result, nil
		}

		log.Info("[container state] Successful get container ports")
	}

	return result, nil
}
