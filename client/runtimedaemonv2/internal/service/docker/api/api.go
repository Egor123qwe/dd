package api

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	versionUnknown   = "unknown"
	dockerPingTimeout = 5 * time.Second
)

var log = logger.NewLogger("api", logger.DefaultWithSentry())

type Api interface {
	CreateContainer(ctx context.Context, req CreateContainerReq) (string, error)

	RemoveContainer(ctx context.Context, name string) error
	RemoveExpiredContainers(ctx context.Context, ttl time.Duration, filter []string) ([]string, error)

	StartContainer(ctx context.Context, name string) error

	StopContainer(ctx context.Context, name string) error
	StopAllContainers(ctx context.Context) error

	GetImageList(ctx context.Context) ([]ImageInfo, error)
	PullImage(ctx context.Context, name string) (<-chan PullInfo, error)
	PullImageOld(ctx context.Context, name string) error
	RemoveImage(ctx context.Context, image string) error
	IsExistImage(ctx context.Context, image string) (bool, error)
	GetImageUsage(ctx context.Context, name string) (float64, error)

	GetContainer(ctx context.Context, name string) (types.Container, error)
	GetContainerStatus(ctx context.Context, name string) (ContainerStatus, error)
	GetContainerLogsReader(ctx context.Context, options container.LogsOptions, name string) (io.ReadCloser, error)
	GetContainerTTY(ctx context.Context, name string) (bool, error)
	GetContainerPorts(ctx context.Context, name string) ([]Port, error)

	Sys(ctx context.Context) (SysInfo, error)
	Info(ctx context.Context) (Info, error)
}

type service struct {
	dockerApi *docker.Client

	config config
}

func New(dockerApi *docker.Client) Api {
	return service{
		dockerApi: dockerApi,

		config: newConfig(),
	}
}

type Info struct {
	Availability bool
	Version      string
}

func (s service) Info(ctx context.Context) (Info, error) {
	pingCtx, cancel := context.WithTimeout(ctx, dockerPingTimeout)
	defer cancel()
	_, err := s.dockerApi.Ping(pingCtx)
	if err != nil {
		return Info{Availability: false, Version: versionUnknown}, nil
	}

	version, err := s.dockerApi.ServerVersion(context.Background())
	if err != nil {
		return Info{Availability: true, Version: versionUnknown}, nil
	}

	return Info{Availability: true, Version: version.Version}, nil
}
