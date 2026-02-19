package session

import (
	"time"

	"github.com/spf13/viper"
)

type config struct {
	clientTTL time.Duration
	rentTTL   time.Duration
}

func newConfig() config {
	config := config{
		clientTTL: viper.GetDuration("rent.client_ttl"),
		rentTTL:   viper.GetDuration("rent.rent_ttl"),
	}

	return config
}
