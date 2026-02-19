package service

import "github.com/spf13/viper"

type config struct {
	input     string
	keepalive string
}

func newConfig() config {
	cfg := config{
		input:     viper.GetString("broker.producer.input"),
		keepalive: viper.GetString("broker.producer.keepalive"),
	}

	return cfg
}
