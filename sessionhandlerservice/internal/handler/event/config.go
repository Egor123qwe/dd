package event

import "github.com/spf13/viper"

type producerConfig struct {
	topic string
}

type config struct {
	ws producerConfig
}

func newConfig() config {
	config := config{}

	config.ws = producerConfig{
		topic: viper.GetString("broker.producer.ws"),
	}

	return config
}
