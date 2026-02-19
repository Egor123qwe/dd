package service

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	stopSharingTimeout time.Duration
	version            string
}

func newConfig() Config {
	return Config{
		stopSharingTimeout: viper.GetDuration("event.stop_sharing_timeout"),
		version:            viper.GetString("version"),
	}
}
