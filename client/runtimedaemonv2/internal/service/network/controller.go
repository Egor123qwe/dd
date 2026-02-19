package network

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
)

func (s service) ChangeMode(ctx context.Context, req network.Settings, appNetState network.AppNetworkState) error {
	// you can also find mutex using in healthChecker function
	// case both of them are change state and settings concurrently
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if req.Mode == network.ModeDisable {
		req.Connect = &network.Options{}
	}

	if req.Connect == nil {
		return fmt.Errorf("failed to configure network: connect credentials not provided")
	}

	// update settings
	if err := s.storage.Settings().Set(req); err != nil {
		return fmt.Errorf("failed to update network settings: %w", err)
	}

	components := s.resolveSettings(req.Mode)

	if err := s.configureTailscale(ctx, components.tailscale, req.Connect.Tailscale); err != nil {
		return fmt.Errorf("failed to configure tailscale: %w", err)
	}

	if err := s.configurePiko(ctx, components.piko, req.Connect.Piko, appNetState.ActivePorts); err != nil {
		return fmt.Errorf("failed to configure piko: %w", err)
	}

	if err := s.configureIptables(ctx, components.iptables); err != nil {
		return fmt.Errorf("failed to configure iptables: %w", err)
	}

	return nil
}
