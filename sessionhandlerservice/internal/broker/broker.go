package broker

import (
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker/kafka"
)

type Broker struct {
	Kafka kafka.Service
}

func New() (Broker, error) {
	kafka, err := kafka.New()
	if err != nil {
		return Broker{}, fmt.Errorf("failed to create kafka broker: %w", err)
	}

	broker := Broker{
		Kafka: kafka,
	}

	return broker, nil
}
