package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	batchTimeout = 300 * time.Millisecond
)

type Producer interface {
	Produce(ctx context.Context, m []byte) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

func New(dialer *kafka.Dialer, brokers []string, topic string) Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Dialer:   dialer,
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		BatchTimeout: batchTimeout,
	})

	return producer{writer: writer}
}

func (p producer) Produce(ctx context.Context, m []byte) error {
	err := p.writer.WriteMessages(ctx,
		kafka.Message{Value: m},
	)

	if err != nil {
		err = fmt.Errorf("failed to produce message: %w", err)
	}

	return err
}

func (p producer) Close() error {
	return p.writer.Close()
}
