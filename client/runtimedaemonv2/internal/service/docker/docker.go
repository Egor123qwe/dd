package docker

import (
	"context"
	"fmt"
	"sync"

	dockerAPI "github.com/docker/docker/client"
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/proxy"
	tempaltesrv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/template"
	templateController "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume"
	storage "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	unsetTemplateID = "unset"
)

var log = logger.NewLogger("container", logger.DefaultWithSentry())

type Container interface {
	Logs(ctx context.Context, req containerModel.LogReq) (<-chan string, error)
}

type Service interface {
	ChangeMode(ctx context.Context, req model.Settings) error
	CheckHealth(ctx context.Context) error

	State(ctx context.Context) (model.State, error)

	Info(ctx context.Context) (model.Info, error)

	Template() tempaltesrv.Service
	Container() Container
}

type service struct {
	container container.Service
	proxy     proxy.Service
	template  templateController.Service
	volume    volume.Service

	api api.Api

	config  config
	storage storage.DockerRepo

	stateMutex  *sync.Mutex
	modeMutex   *sync.Mutex
	healthMutex *sync.Mutex
}

func New(dockerApi *dockerAPI.Client, storage storage.DockerRepo) (Service, error) {
	api := api.New(dockerApi)

	volume, err := volume.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init volumes: %w", err)
	}

	srv := service{
		container: container.New(api, volume, storage),
		proxy:     proxy.New(),
		volume:    volume,
		template:  templateController.New(api, volume, storage),

		api: api,

		config:  newConfig(),
		storage: storage,

		stateMutex:  &sync.Mutex{},
		modeMutex:   &sync.Mutex{},
		healthMutex: &sync.Mutex{},
	}

	return srv, nil
}

func (s service) State(ctx context.Context) (model.State, error) {
	log.Info("[docker state] called")
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	log.Info("[docker state] started")
	var state model.State

	var lock bool

	if lock = s.modeMutex.TryLock(); lock {
		defer s.modeMutex.Unlock()
	}

	inLaunching := !lock

	healthParams := healthParams{inLaunching: inLaunching}

	settings, err := s.storage.Settings().Get()
	if err != nil {
		return model.State{}, fmt.Errorf("failed to get container settings: %w", err)
	}

	// set proxy state
	proxy, err := s.proxy.State(ctx)
	if err != nil {
		return model.State{}, fmt.Errorf("failed to get auth proxy state: %w", err)
	}

	state.Container.SharedVolume, err = s.volume.Shared().State(ctx)
	if err != nil {
		return model.State{}, fmt.Errorf("failed to get shared volume state: %w", err)
	}

	log.Info("[docker state] Successfully got proxy state")

	state.Auth = proxy

	state.Container.UsedTemplateID = unsetTemplateID

	if settings.Options != nil {
		state.Container.UsedTemplateID = settings.Options.Container.TemplateID
	}

	switch state.Container.UsedTemplateID {
	case unsetTemplateID:
		state.Container.State = containerModel.State{
			Status:    api.StoppedContainer,
			StatusMsg: "template settings not found",
		}

	default:
		log.Info("[docker state] try to get container state")

		state.Container.State, err = s.container.State(ctx, settings.Options.Container.TemplateID)
		if err != nil {
			state.Container.State = containerModel.State{
				Status:    api.ErrorStatus,
				StatusMsg: fmt.Errorf("failed to get container state: %w", err).Error(),
			}
		}

		log.Info("[docker state] Successfully got container state")
	}

	state.Health = s.health(
		ctx, healthParams, settings, state.Container, state.Auth,
	)

	return state, nil
}

func (s service) Info(ctx context.Context) (model.Info, error) {
	info, err := s.api.Info(ctx)
	if err != nil {
		return model.Info{}, fmt.Errorf("failed to get container info: %w", err)
	}

	result := model.Info{
		Availability: info.Availability,
		Version:      info.Version,
	}

	return result, nil
}

func (s service) Container() Container {
	return s.container
}

func (s service) Template() tempaltesrv.Service {
	return s.template
}
