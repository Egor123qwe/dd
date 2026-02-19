package docker

import (
	"time"

	"github.com/spf13/viper"
)

type config struct {
	stoppedContainerTTL time.Duration
}

func newConfig() config {
	return config{
		stoppedContainerTTL: viper.GetDuration("docker.stopped_container_ttl"),
	}
}
