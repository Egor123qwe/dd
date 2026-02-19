package broker

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	brokers []string

	username string
	password string
}

func newConfig() config {
	username := viper.GetString("broker.username")
	password := viper.GetString("broker.password")
	if v, ok := os.LookupEnv("BROKER_USERNAME"); ok {
		username = v
	}
	if v, ok := os.LookupEnv("BROKER_PASSWORD"); ok {
		password = v
	}

	brokers := viper.GetStringSlice("broker.URLs")
	if v, ok := os.LookupEnv("BROKER_URLS"); ok && v != "" {
		brokers = strings.Split(strings.TrimSpace(v), ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}
	}

	return config{
		brokers:  brokers,
		username: username,
		password: password,
	}
}
