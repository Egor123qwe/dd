package rent

import (
	"time"

	"github.com/spf13/viper"
)

type config struct {
	pikoSecretKey string
	pikoTokenExp  time.Duration

	linkTemplate string
}

func newConfig() config {
	config := config{
		pikoSecretKey: viper.GetString("rent.settings.piko.secret_key"),
		pikoTokenExp:  viper.GetDuration("rent.settings.piko.token_exp"),

		linkTemplate: viper.GetString("rent.settings.piko.link_template"),
	}

	return config
}
