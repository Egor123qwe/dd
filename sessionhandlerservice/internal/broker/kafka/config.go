package kafka

import (
	"os"

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
	return config{
		brokers:  viper.GetStringSlice("broker.URLs"),
		username: username,
		password: password,
	}
}
