package consumer

import (
	"context"
	"fmt"

	"github.com/op/go-logging"
	"github.com/segmentio/kafka-go"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

var log = logging.MustGetLogger("consumer")

type Consumer interface {
	Consume(ctx context.Context) (msg.MSG, error)
	Close() error
}

type consumer struct {
	reader *kafka.Reader
}

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func New(dialer *kafka.Dialer, brokers []string) Consumer {
	config := newConfig(brokers)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:   dialer,
		Brokers:  config.brokers,
		GroupID:  config.groupID,
		Topic:    config.topic,
		MaxBytes: 10e6, // 10MB
	})

	return consumer{reader: reader}
}

func (c consumer) Consume(ctx context.Context) (msg.MSG, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return msg.MSG{}, fmt.Errorf("failed to read message: %w", err)
	}

	return msg.MSG{Data: m.Value}, nil
}

func (c consumer) Close() error {
	return c.reader.Close()
}
