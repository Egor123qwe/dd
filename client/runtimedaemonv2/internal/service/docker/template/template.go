package template

import (
	"context"
	"errors"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var NotFoundErr = errors.New("template not found")

var log = logger.NewLogger("template", logger.DefaultWithSentry())

type Service interface {
	Get(ctx context.Context, templateId string) (template.Template, error)
	GetWithStat(ctx context.Context, templateId string) (template.Info, error)

	GetAll(ctx context.Context) ([]template.Template, error)
	GetAllWithStat(ctx context.Context) ([]template.Info, error)
	GetFiltredWithStat(ctx context.Context, templateTypes []string) ([]template.Info, error)

	Download(ctx context.Context, t template.Template) (<-chan api.PullInfo, error)
	Download_OLD(ctx context.Context, t template.Template) error

	Remove(ctx context.Context, templateID string) error
	RemoveVolumes(ctx context.Context, templateID string, opts RemoveVolumeOptions) error

	AddUsage(ctx context.Context, templateID string, usage container.Usage) error

	GetStat(ctx context.Context) (template.Stat, error)

	OptimizeMemUsage(ctx context.Context) error
}

type service struct {
	api    api.Api
	volume volume.Service

	storage repo.DockerRepo
}

func New(api api.Api, volume volume.Service, storage repo.DockerRepo) Service {
	return service{
		api:     api,
		volume:  volume,
		storage: storage,
	}
}
