package container

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
)

type Usage int32

const (
	Local Usage = iota
	Host
)

type Settings struct {
	Usage     Usage
	ForUserID string // who will use container

	Options Options

	Template template.Template
}

type Options struct {
	SharedVolume *volume.SharedVolume

	MemoryLimitBytes  int64
	StorageLimitBytes int64
	CPULimit          int64
}

type State struct {
	Status    api.ContainerStatus
	StatusMsg string

	Ports []api.Port
}
