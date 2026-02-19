package health

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
)

type Status int32

const (
	OK Status = iota
	Launching
	Error
)

type State struct {
	Worker runtime.Settings

	Components Components

	System     hardware.Info
	AppNetwork network.AppNetworkState

	Health Health
}

type Components struct {
	Docker  docker.State
	Network network.State
}

type Health struct {
	Status    Status
	StatusMsg string
}

type Info struct {
	UniqueMachineId string
	Docker          docker.Info
	Network         network.Info

	Version string
}
