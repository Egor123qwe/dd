package producer

import "github.com/spf13/viper"

type config struct {
	brokers   []string
	input     string
	keepalive string
}

func newConfig(brokers []string) config {
	return config{
		brokers: brokers,

		input:     viper.GetString("broker.producer.input"),
		keepalive: viper.GetString("broker.producer.keepalive"),
	}
}
