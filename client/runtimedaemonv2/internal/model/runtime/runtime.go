package runtime

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
)

type Mode int32

const (
	ModeDisable Mode = iota
	ModeLocal
	ModeHostP2PVM
	ModeHostProxyVM
	ModeClientP2PVM
)

var modeNames = []string{
	"disable",
	"local",
	"host_p2p_vm",
	"host_proxy_vm",
	"client_p2p_vm",
}

func (m Mode) String() string {
	return modeNames[m]
}

type Configuration struct {
	Settings   Settings
	Components Components
}

type Settings struct {
	Mode Mode
}

type Components struct {
	Docker  *DockerOptions
	Network *network.Options
}

type DockerOptions struct {
	ForUserID string // who will use container

	Container docker.ContainerSettings
	Auth      auth.Settings
}
