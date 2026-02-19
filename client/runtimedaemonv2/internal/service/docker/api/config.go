package api

import "github.com/spf13/viper"

type config struct {
	registry string
	username string
	password string
}

func newConfig() config {
	return config{
		registry: viper.GetString("docker.registry_name"),
		username: viper.GetString("docker.username"),
		password: viper.GetString("docker.password"),
	}
}
