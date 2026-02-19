package kafka

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/broker/kafka/consumer"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
)

const dialerTimeout = 20 * time.Second

type BrokerService interface {
	Consumer() consumer.Consumer
}

type service struct {
	consumer consumer.Consumer
	log      slog.Logger
}

func New(cfg config.Config, log slog.Logger) (BrokerService, error) {
	username, password := cfg.KafkaConfig.Username, cfg.KafkaConfig.Password
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
	if username != "" && password != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create scram mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}

	consumer := consumer.New(dialer, cfg.KafkaConfig.Brokers, cfg.KafkaConfig.Consumer.Topic, cfg.KafkaConfig.Consumer.ConsumerGroup)
	log.Info("consumer created")

	service := &service{
		consumer: consumer,
		log:      log,
	}

	return service, nil
}

func (s service) Consumer() consumer.Consumer {
	return s.consumer
}
