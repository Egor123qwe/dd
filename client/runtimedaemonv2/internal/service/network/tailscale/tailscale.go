package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/command"
	storage "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
)

const (
	tailscaleStatusConnected = "Running"
)

type Service interface {
	Connect(ctx context.Context, req tailscale.Settings) error
	Disconnect(ctx context.Context) error

	State(ctx context.Context) (tailscale.State, error)
	Info(ctx context.Context) (tailscale.Info, error)
}

type service struct {
	config  config
	storage storage.NetworkRepo
	system  command.Service

	mutex *sync.Mutex
}

func New() Service {
	return &service{
		config: newConfig(),
		system: command.New(),

		mutex: &sync.Mutex{},
	}
}

func (s service) Connect(ctx context.Context, req tailscale.Settings) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx, _ = context.WithTimeout(ctx, timeout)

	params := []string{
		commandPrefix,
		"up",
		"--login-server=" + s.config.loginServer,
		"--accept-routes",
		"--force-reauth",
		"--authkey=" + req.AuthKey,
		"--timeout=" + fmt.Sprintf("%ds", timeout/time.Second),
		"--hostname=" + s.hostname(req.ClientID),
	}

	_, err := s.system.Run(ctx, params)
	if err != nil {
		return err
	}

	return err
}

func (s service) Disconnect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx, _ = context.WithTimeout(ctx, timeout)

	_, err := s.system.Run(ctx, []string{commandPrefix, "logout"})
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	return nil
}

func (s service) State(ctx context.Context) (tailscale.State, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx, _ = context.WithTimeout(ctx, timeout)

	currentState, err := s.status(ctx)
	if err != nil {
		// failed to get status - assume that tailscale is not available and tailscale is not connected
		return tailscale.State{
			Status:    tailscale.Stopped,
			StatusMsg: "Tailscale is not available",
		}, err
	}

	connected := (currentState.BackendState == tailscaleStatusConnected) && currentState.Self.Online
	if !connected {
		return tailscale.State{
			Status:    tailscale.Stopped,
			StatusMsg: "Tailscale is available, but not connected",
		}, nil
	}

	// connected case
	result := tailscale.State{
		Status: tailscale.Running,

		Connection: tailscale.ConnectionState{
			PeerHostnames: s.peerHostname(currentState.Peer),
			IPs:           s.ip(currentState.TailscaleIPs),
		},

		StatusMsg: "OK",
	}

	return result, nil
}

func (s service) Info(ctx context.Context) (tailscale.Info, error) {
	version, err := s.version(ctx)
	if err != nil {
		return tailscale.Info{}, fmt.Errorf("failed to get version: %w", err)
	}

	return tailscale.Info{
		Availability: true,
		Version:      version,
	}, nil
}
