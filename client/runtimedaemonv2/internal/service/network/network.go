package network

import (
	"context"
	"fmt"
	"sync"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	iptablesModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/iptables"
	pikoModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	tailscaleModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/network/iptables"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/network/piko"
	tailscale "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/network/tailscale"
	storage "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("network", logger.DefaultWithSentry())

const (
	versionUnknown = "unknown"
)

type Service interface {
	ChangeMode(ctx context.Context, req network.Settings, appNetState network.AppNetworkState) error

	CheckHealth(ctx context.Context, appNetState network.AppNetworkState) error

	State(ctx context.Context, appNetState network.AppNetworkState) (network.State, error)
	Info(ctx context.Context) (network.Info, error)
}

type service struct {
	tailscale tailscale.Service
	piko      piko.Service
	iptables  iptables.Service

	storage storage.NetworkRepo
	mutex   *sync.Mutex
}

func New(storage storage.NetworkRepo) Service {
	srv := service{
		tailscale: tailscale.New(),
		piko:      piko.New(),
		iptables:  iptables.New(),

		storage: storage,
		mutex:   &sync.Mutex{},
	}

	return srv
}

func (s service) State(ctx context.Context, appNetworkState network.AppNetworkState) (network.State, error) {
	log.Info("[network state]: called")
	// you can also find mutex using in ChangeMode function
	// case both of them are use state settings concurrently
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Info("[network state]: started")

	result := network.State{}

	settings, err := s.storage.Settings().Get()
	if err != nil {
		return network.State{}, fmt.Errorf("failed to get network settings: %w", err)
	}

	components := s.resolveSettings(settings.Mode)

	log.Info("[network state]: try get state from tailscale")

	// get tailscale state
	tailscaleState, err := s.tailscale.State(ctx)
	if err != nil {
		return network.State{}, fmt.Errorf("failed to get tailscale state: %w", err)
	}

	result.ConnectionState.Tailscale = tailscaleState

	log.Info("[network state]: Success get state from tailscale")

	// get piko state
	var pikoPorts []pikoModel.EndpointSettings

	if settings.Connect != nil && settings.Connect.Piko != nil {
		pikoPorts = settings.Connect.Piko.Endpoints
	}

	log.Info("[network state]: try get state from piko")

	pikoEndpoints := s.getPikoEndpoints(pikoPorts, appNetworkState.ActivePorts)

	pikoState, err := s.piko.State(ctx, pikoEndpoints)
	if err != nil {
		return network.State{}, fmt.Errorf("failed to get piko_1 state: %w", err)
	}

	result.ConnectionState.Piko = pikoState

	if len(pikoEndpoints) == 0 {
		components.piko = false
	}

	log.Info("[network state]: Success get state from piko")

	log.Info("[network state]: try get state from iptables")

	// get iptables state
	iptablesCorrect, err := s.iptables.IsCorrect(ctx, s.resolveSettings(settings.Mode).iptables)
	if err != nil {
		return network.State{}, fmt.Errorf("failed to get iptables state: %w", err)
	}

	result.ConnectionState.Iptables = iptablesModel.State{Configured: iptablesCorrect}

	log.Info("[network state]: Success get state from iptables")

	result.Health = s.health(components, result.ConnectionState)

	return result, nil
}

func (s service) Info(ctx context.Context) (network.Info, error) {
	var result network.Info
	var err error

	result.Tailscale, err = s.tailscale.Info(ctx)
	if err != nil {
		result.Tailscale = tailscaleModel.Info{
			Version:      versionUnknown,
			Availability: false,
		}
	}

	result.Piko, err = s.piko.Info(ctx)
	if err != nil {
		result.Piko = pikoModel.Info{
			Version:      versionUnknown,
			Availability: false,
		}
	}

	return result, nil
}
