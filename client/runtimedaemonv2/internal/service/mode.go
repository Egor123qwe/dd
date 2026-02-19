package service

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
)

type settings struct {
	docker  bool
	network network.Mode
}

func (s service) resolveSettings(mode model.Mode) settings {
	switch mode {
	case model.ModeHostP2PVM:
		return settings{
			docker:  true,
			network: network.ModeP2P,
		}

	case model.ModeLocal:
		return settings{
			docker:  true,
			network: network.ModeDisable,
		}

	case model.ModeClientP2PVM:
		return settings{
			docker:  false,
			network: network.ModeP2P,
		}

	case model.ModeHostProxyVM:
		return settings{
			docker:  true,
			network: network.ModeProxy,
		}

	case model.ModeDisable:
		return settings{
			docker:  false,
			network: network.ModeDisable,
		}
	}

	return settings{}
}

func (s service) resolveContainerUsage(mode model.Mode) container.Usage {
	switch mode {
	case model.ModeHostP2PVM, model.ModeHostProxyVM:
		return container.Host

	default:
		return container.Local
	}
}
