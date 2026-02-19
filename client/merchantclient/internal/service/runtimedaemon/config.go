package runtimedaemon

import "github.com/spf13/viper"

type config struct {
	port int
}

func newConfig() config {
	return config{
		port: viper.GetInt("runtimedaemon.port"),
	}
}
