package repo

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/fileStorage"
)

const runtimeSettingsStoragePath = "/runtime_settings.json"

type RuntimeRepo interface {
	Settings() RuntimeSettings
}

type RuntimeSettings interface {
	Get() (runtime.Settings, error)
	Set(state runtime.Settings) error
}

type runtimeRepo struct {
	settings fileStorage.Settings[runtime.Settings]
}

func NewRuntime(path string) RuntimeRepo {
	return runtimeRepo{
		settings: fileStorage.New[runtime.Settings](path + runtimeSettingsStoragePath),
	}
}

func (r runtimeRepo) Settings() RuntimeSettings {
	return r.settings
}
