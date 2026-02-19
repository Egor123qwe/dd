package network

import "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"

type requiredComponents struct {
	tailscale bool
	piko      bool
	iptables  bool
}

func (s service) resolveSettings(mode network.Mode) requiredComponents {
	switch mode {
	case network.ModeP2P:
		return requiredComponents{
			iptables:  true,
			tailscale: true,
		}

	case network.ModeProxy:
		return requiredComponents{
			piko: true,
		}
	}

	return requiredComponents{}
}
