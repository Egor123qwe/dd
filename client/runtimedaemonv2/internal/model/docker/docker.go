package docker

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
)

type Status int32

const (
	OK Status = iota
	Launching
	Error
)

type Settings struct {
	Running bool

	// can be nil in [Running = false]
	Options *Options
}

type Options struct {
	Usage     container.Usage
	ForUserID string // who will use container

	Container ContainerSettings
	Auth      auth.Settings
}

type ContainerSettings struct {
	SharedVolume *SharedVolume

	TemplateID string
}

type SharedVolume struct {
	AccessKeyID     string
	SecretAccessKey string

	BucketName string
	Mount      *string
}

type State struct {
	Container ContainerState
	Auth      auth.State

	Health Health
}

type ContainerState struct {
	// template can be "unset" in [Running = false]
	UsedTemplateID string
	SharedVolume   volume.SharedVolumeState

	State container.State
}

type Health struct {
	Status    Status
	StatusMsg string
}

type SystemInfo struct {
	TotalImgMem float64
}

type Info struct {
	Availability bool
	Version      string
}
