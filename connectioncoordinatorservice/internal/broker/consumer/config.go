package consumer

import "github.com/spf13/viper"

type config struct {
	brokers []string
	groupID string
	topic   string
}

func newConfig(brokers []string) config {
	return config{
		brokers: brokers,
		groupID: viper.GetString("broker.consumer.group_id"),
		topic:   viper.GetString("broker.consumer.topic"),
	}
}
