package repo

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/fileStorage"
)

const (
	dockerSettingsStoragePath  = "/docker_settings.json"
	dockerTemplatesStoragePath = "/template.json"
)

type DockerRepo interface {
	Settings() DockerSettings
	Downloads() DockerTemplates
}

type DockerSettings interface {
	Get() (docker.Settings, error)
	Set(state docker.Settings) error
}

type DockerTemplates interface {
	Get() ([]template.Template, error)
	Set(templates []template.Template) error
}

type dockerRepo struct {
	settings  fileStorage.Settings[docker.Settings]
	downloads fileStorage.Settings[[]template.Template]
}

func NewDocker(path string) DockerRepo {
	return dockerRepo{
		settings:  fileStorage.New[docker.Settings](path + dockerSettingsStoragePath),
		downloads: fileStorage.New[[]template.Template](path + dockerTemplatesStoragePath),
	}
}

func (r dockerRepo) Settings() DockerSettings {
	return r.settings
}

func (r dockerRepo) Downloads() DockerTemplates {
	return r.downloads
}
