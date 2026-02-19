package network

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/iptables"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
)

type Mode int32

const (
	ModeDisable Mode = iota
	ModeP2P
	ModeProxy
)

type Status int32

const (
	OK Status = iota
	Launching
	Error
)

type Settings struct {
	Mode    Mode
	Connect *Options
}

type Options struct {
	Tailscale *tailscale.Settings
	Piko      *piko.Settings
}

type ActivePort struct {
	PortID string

	Title    string
	Port     string
	Protocol template.Protocol
}

type AppNetworkState struct {
	ActivePorts []ActivePort
}

type State struct {
	ConnectionState Connection

	Health Health
}

type Connection struct {
	Tailscale tailscale.State
	Piko      piko.State
	Iptables  iptables.State
}

type Info struct {
	Tailscale tailscale.Info
	Piko      piko.Info
}

type Health struct {
	Status    Status
	StatusMsg string
}
