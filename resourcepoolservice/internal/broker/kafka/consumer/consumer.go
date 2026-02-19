package consumer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Consume(ctx context.Context) (kafka.Message, error)
	Close() error
}

type consumer struct {
	reader *kafka.Reader
}

func New(dialer *kafka.Dialer, brokers []string, topic, groupID string) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:   dialer,
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MaxBytes: 10e6,
	})

	consumer := consumer{
		reader: reader,
	}

	return consumer
}

func (c consumer) Consume(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("failed to read message: %w", err)
	}

	return msg, nil
}

func (c consumer) Close() error {
	return c.reader.Close()
}
