package broker

import (
	"fmt"
	"os"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka/consumer"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka/producer"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	dialerTimeout = 10 * time.Second
)

type Broker interface {
	Producer() producer.Producer
	Consumer(topic string, groupID string) consumer.Consumer
}

type broker struct {
	config    config.KafkaConfig
	mechanism sasl.Mechanism
	dialer    *kafka.Dialer
}

func New(cfg config.KafkaConfig) (Broker, error) {
	username, password := cfg.Username, cfg.Password
	if v, ok := os.LookupEnv("KAFKA_USERNAME"); ok {
		username = v
	}
	if v, ok := os.LookupEnv("KAFKA_PASSWORD"); ok {
		password = v
	}

	dialer := &kafka.Dialer{
		Timeout:   dialerTimeout,
		DualStack: true,
	}
	var mechanism sasl.Mechanism
	if username != "" && password != "" {
		var err error
		mechanism, err = scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create scram mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}

	b := &broker{
		config:    cfg,
		mechanism: mechanism,
		dialer:    dialer,
	}

	return b, nil
}

func (b broker) Producer() producer.Producer {
	return producer.New(b.mechanism, b.config.Brokers)
}

func (b broker) Consumer(topic string, groupID string) consumer.Consumer {
	return consumer.New(b.dialer, b.config.Brokers, topic, groupID)
}
