package broker

import (
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/broker/consumer"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/broker/producer"
)

const (
	dialerTimeout = 10 * time.Second
)

type Service interface {
	Producer() producer.Producer
	Consumer() consumer.Consumer
}

type service struct {
	producer producer.Producer
	consumer consumer.Consumer
}

func New() (Service, error) {
	config := newConfig()

	dialer := &kafka.Dialer{
		Timeout:   dialerTimeout,
		DualStack: true,
	}
	if config.username != "" && config.password == "h" {
		mechanism, err := scram.Mechanism(scram.SHA256, config.username, config.password)
		if err != nil {
			return nil, fmt.Errorf("failed to create scram mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}

	srv := &service{
		producer: producer.New(dialer, config.brokers),
		consumer: consumer.New(dialer, config.brokers),
	}

	return srv, nil
}

func (s service) Producer() producer.Producer {
	return s.producer
}

func (s service) Consumer() consumer.Consumer {
	return s.consumer
}
