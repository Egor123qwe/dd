package network

import (
	"context"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	pikoModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
	tailscaleModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
)

func (s service) CheckHealth(ctx context.Context, appNetState model.AppNetworkState) error {
	log.Info("[network health]: called")
	defer log.Info("[network health]: done")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	settings, err := s.storage.Settings().Get()
	if err != nil {
		return fmt.Errorf("failed to get network settings: %w", err)
	}

	components := s.resolveSettings(settings.Mode)

	// check tailscale state
	tailscale, err := s.tailscale.State(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tailscale status: %w", err)
	}

	// fix tailscale if state not correct
	if components.tailscale && tailscale.Status == tailscaleModel.Stopped {
		if settings.Connect == nil || settings.Connect.Tailscale == nil {
			return fmt.Errorf("failed to configure tailscale: connect credentials not provided")
		}

		if err := s.configureTailscale(ctx, components.tailscale, settings.Connect.Tailscale); err != nil {
			return fmt.Errorf("failed to configure tailscale: %w", err)
		}
	}

	// fix piko if state not correct
	if components.piko {
		if settings.Connect == nil || settings.Connect.Piko == nil {
			return fmt.Errorf("failed to check piko state: connect credentials not provided")
		}

		endpoint := s.getPikoEndpoints(settings.Connect.Piko.Endpoints, appNetState.ActivePorts)

		piko, err := s.piko.State(ctx, endpoint)
		if err != nil {
			return fmt.Errorf("failed to get piko status: %w", err)
		}

		if piko.Status != pikoModel.Running && len(endpoint) > 0 || piko.Status == pikoModel.Error {
			log.Info("piko state: not correct")

			if err := s.configurePiko(ctx, components.piko, settings.Connect.Piko, appNetState.ActivePorts); err != nil {
				return fmt.Errorf("failed to configure piko: %w", err)
			}

			log.Info("piko state: successfully fixed")
		}

	}

	// check iptables state
	iptablesCorrect, err := s.iptables.IsCorrect(ctx, components.iptables)
	if err != nil {
		return fmt.Errorf("failed to get iptables status: %w", err)
	}

	// fix iptables if state not correct
	if !iptablesCorrect {
		if err := s.configureIptables(ctx, components.iptables); err != nil {
			return fmt.Errorf("failed to configure iptables: %w", err)
		}
	}

	return nil
}

func (s service) health(components requiredComponents, state model.Connection) model.Health {
	result := model.Health{Status: model.OK}

	if components.tailscale != (state.Tailscale.Status == tailscale.Running) {
		result.Status = s.updateStatusByPriority(result.Status, model.Error)

		result.StatusMsg += fmt.Sprintf("Tailscale: have wrong status: have: [%s]: [%s]. ",
			tailscale.StatusMap[state.Tailscale.Status], state.Tailscale.StatusMsg)
	}

	if components.piko != (state.Piko.Status == pikoModel.Running) {
		result.Status = s.updateStatusByPriority(result.Status, model.Error)

		result.StatusMsg += fmt.Sprintf("Piko: have wrong status: have: [%s]: [%s]. ",
			pikoModel.StatusMap[state.Piko.Status], state.Piko.StatusMsg)
	}

	if !state.Iptables.Configured {
		result.Status = s.updateStatusByPriority(result.Status, model.Error)
		result.StatusMsg += fmt.Sprintf("Iptables: iptables is not configured correctly. ")
	}

	return result
}

func (s service) updateStatusByPriority(current model.Status, update model.Status) model.Status {
	priority := map[model.Status]int{
		model.OK:        0,
		model.Launching: 1,
		model.Error:     2,
	}

	if priority[current] > priority[update] {
		return current
	}

	return update
}
