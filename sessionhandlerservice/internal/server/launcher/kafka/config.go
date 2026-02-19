package kafka

import (
	"github.com/spf13/viper"
)

type consumer struct {
	topic   string
	groupID string
}

type config struct {
	consumers []consumer
}

func newConfig() config {
	config := config{}

	ws := consumer{
		topic:   viper.GetString("broker.consumer.ws.topic"),
		groupID: viper.GetString("broker.consumer.ws.group_id"),
	}

	ttl := consumer{
		topic:   viper.GetString("broker.consumer.ttl.topic"),
		groupID: viper.GetString("broker.consumer.ttl.group_id"),
	}

	config.consumers = append(config.consumers, ws, ttl)

	return config
}
