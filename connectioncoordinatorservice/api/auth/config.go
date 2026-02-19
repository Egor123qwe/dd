package auth

import (
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	URL string
}

func newConfig() config {
	return config{
		URL: strings.Trim(strings.TrimSpace(viper.GetString("api.URL")), "/"),
	}
}
