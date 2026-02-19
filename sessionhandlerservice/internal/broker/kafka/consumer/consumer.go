package consumer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Consume(ctx context.Context) ([]byte, error)
	Close() error
}

type consumer struct {
	reader *kafka.Reader
}

func New(dialer *kafka.Dialer, brokers []string, topic string, groupID string) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:   dialer,
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MaxBytes: 10e6, // 10MB
	})

	return consumer{reader: reader}
}

func (c consumer) Consume(ctx context.Context) ([]byte, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	return m.Value, nil
}

func (c consumer) Close() error {
	return c.reader.Close()
}
