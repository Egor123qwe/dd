package network

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
)

func (s service) configureTailscale(ctx context.Context, connected bool, settings *tailscale.Settings) error {
	switch connected {
	case true:
		if settings == nil {
			return fmt.Errorf("failed to connect to network: connect credentials not provided")
		}

		if err := s.tailscale.Connect(ctx, *settings); err != nil {
			return fmt.Errorf("failed to connect to network: %w", err)
		}

	case false:
		if err := s.tailscale.Disconnect(ctx); err != nil {
			log.Errorf("failed to disconnect from network: %s", err)
		}
	}

	return nil
}

func (s service) configurePiko(ctx context.Context, connected bool, settings *piko.Settings, appPorts []network.ActivePort) error {
	switch connected {
	case true:
		if settings == nil {
			return fmt.Errorf("failed to connect to network: connect credentials not provided")
		}

		connectReq := piko.ConnectReq{
			Endpoints: s.getPikoEndpoints(settings.Endpoints, appPorts),
			AuthKey:   settings.AuthKey,
		}

		if err := s.piko.Connect(ctx, connectReq); err != nil {
			return fmt.Errorf("failed to connect to network: %w", err)
		}

	case false:
		if err := s.piko.Disconnect(ctx); err != nil {
			log.Errorf("failed to disconnect from network: %s", err)
		}
	}

	return nil
}

func (s service) configureIptables(ctx context.Context, configured bool) error {
	switch configured {
	case true:
		return s.iptables.Set(ctx)

	case false:
		return s.iptables.Discard(ctx)
	}

	return nil
}
