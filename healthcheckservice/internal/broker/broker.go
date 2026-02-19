package broker

import (
	"fmt"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/broker/kafka"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
)

type Broker interface {
	Kafka() kafka.BrokerService
}

type broker struct {
	kafka kafka.BrokerService
}

func New(cfg config.Config, log slog.Logger) (Broker, error) {
	kafkaService, err := kafka.New(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka broker: %w", err)
	}

	return broker{kafka: kafkaService}, nil
}

func (b broker) Kafka() kafka.BrokerService {
	return b.kafka
}
