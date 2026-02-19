package repo

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/fileStorage"
)

const (
	networkSettingsStoragePath = "/network_settings.json"
)

type NetworkRepo interface {
	Settings() NetworkSettings
}

type NetworkSettings interface {
	Get() (network.Settings, error)
	Set(state network.Settings) error
}

type networkRepo struct {
	settings fileStorage.Settings[network.Settings]
}

func NewNetwork(path string) NetworkRepo {
	return networkRepo{
		settings: fileStorage.New[network.Settings](path + networkSettingsStoragePath),
	}

}

func (r networkRepo) Settings() NetworkSettings {
	return r.settings
}
