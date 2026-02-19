package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/go-iptables/iptables"
	dockerAPI "github.com/docker/docker/client"
	dockerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
	networkModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/event"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/health"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/system"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/software"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/sync/fnController"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("service", logger.DefaultWithSentry())

type Service interface {
	// ChangeMode attention CheckHealth will be automatically ended if ChangeMode was called
	ChangeMode(ctx context.Context, req runtime.Configuration) error
	ExecuteEvent(ctx context.Context, event event.Event) error

	HardReset()
	State(ctx context.Context) (health.State, error)

	SysInfo(ctx context.Context) (hardware.Info, error)
	Info(ctx context.Context) (health.Info, error)

	Docker() docker.Service
	Network() network.Service
}

type mutex struct {
	health *sync.Mutex
	mode   *sync.Mutex
	state  *sync.Mutex
	event  *sync.Mutex
}

type service struct {
	network network.Service
	docker  docker.Service
	system  system.Service

	storage storage.Storage
	config  Config

	mutex                 mutex
	healthCheckController fnController.Controller
}

func New(dockerAPI *dockerAPI.Client, iptablesAPI *iptables.IPTables, storage storage.Storage) (Service, error) {
	docker, err := docker.New(dockerAPI, storage.Docker())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker service: %w", err)
	}

	srv := service{
		network: network.New(iptablesAPI, storage.Network()),
		docker:  docker,
		system:  system.New(),

		storage: storage,
		config:  newConfig(),

		mutex: mutex{
			health: &sync.Mutex{},
			mode:   &sync.Mutex{},
			state:  &sync.Mutex{},
			event:  &sync.Mutex{},
		},

		healthCheckController: fnController.New(),
	}

	// you can add here fnController from utils (and stop method for service),
	// but I skipped, because we don't need it now
	go srv.eventLoop(context.Background())
	go srv.healthChecker(context.Background())

	return srv, nil
}

func (s service) ChangeMode(ctx context.Context, req runtime.Configuration) error {
	if err := s.discardEvents(ctx); err != nil {
		return fmt.Errorf("failed to discard events: %w", err)
	}

	return s.changeMode(ctx, req)
}

func (s service) State(ctx context.Context) (health.State, error) {
	log.Info("[state stream] called")

	s.mutex.state.Lock()
	defer s.mutex.state.Unlock()

	log.Info("[state stream] started")
	defer log.Info("[state stream] finished")

	result := health.State{}

	var lock bool

	if lock = s.mutex.mode.TryLock(); lock {
		defer s.mutex.mode.Unlock()
	}

	inLaunching := !lock

	worker, err := s.storage.Runtime().Settings().Get()
	if err != nil {
		return health.State{}, fmt.Errorf("failed to get runtime config: %w", err)
	}

	result.Worker = runtime.Settings{
		Mode: worker.Mode,
	}

	result.System, err = s.system.InfoFromCache(ctx)
	if err != nil {
		return health.State{}, fmt.Errorf("failed to get command health: %w", err)
	}

	log.Info("[state stream] Successfully received hardware state")

	result.Components.Docker, err = s.docker.State(ctx)
	if err != nil {
		return health.State{}, fmt.Errorf("failed to get container health: %w", err)
	}

	log.Info("[state stream] Successfully received container state")

	// if container status not "OK", we can't get network state (docker give exposed ports as results)
	if result.Components.Docker.Health.Status != dockerModel.OK {
		result.Health = s.health(inLaunching, result.Components)

		return result, nil
	}

	result.AppNetwork, err = s.resolveNetworkState(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to get app net state (can't get network state): %w", err)
	}

	log.Info("[state stream] Successfully received app network state")

	result.Components.Network, err = s.network.State(ctx, result.AppNetwork)
	if err != nil {
		return health.State{}, fmt.Errorf("failed to get network health: %w", err)
	}

	log.Info("[state stream] Successfully received network state")

	// Generate status from the received health
	result.Health = s.health(inLaunching, result.Components)

	return result, nil
}

func (s service) SysInfo(ctx context.Context) (hardware.Info, error) {
	result, err := s.system.Info(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to get system info: %w", err)
	}

	return result, nil
}

func (s service) Info(ctx context.Context) (health.Info, error) {
	uniqueMachineId, err := software.UniqueMachineID()
	if err != nil {
		return health.Info{}, fmt.Errorf("failed to get unique machine id: %w", err)
	}

	docker, err := s.docker.Info(ctx)
	if err != nil {
		return health.Info{}, fmt.Errorf("failed to get container info: %w", err)
	}

	network, err := s.network.Info(ctx)
	if err != nil {
		return health.Info{}, fmt.Errorf("failed to get network info: %w", err)
	}

	result := health.Info{
		UniqueMachineId: uniqueMachineId,
		Docker:          docker,
		Network:         network,

		Version: s.config.version,
	}

	return result, nil
}

func (s service) HardReset() {
	s.storage.Runtime().Settings().Set(runtime.Settings{Mode: runtime.ModeDisable})
	s.storage.Docker().Settings().Set(dockerModel.Settings{Running: false})
	s.storage.Network().Settings().Set(networkModel.Settings{Mode: networkModel.ModeDisable})
}

func (s service) Docker() docker.Service {
	return s.docker
}

func (s service) Network() network.Service {
	return s.network
}
