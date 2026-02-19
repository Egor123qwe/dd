package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

const (
	batchTimeout = 400 * time.Millisecond
)

type Producer interface {
	Produce(ctx context.Context, m msg.MSG, topic string) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

type Config struct {
	Brokers []string
}

func New(dialer *kafka.Dialer, brokers []string) Producer {
	config := newConfig(brokers)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Dialer:       dialer,
		Brokers:      config.brokers,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: batchTimeout,
	})

	return producer{writer: writer}
}

func (p producer) Produce(ctx context.Context, m msg.MSG, topic string) error {
	err := p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: m.Data,
		},
	)

	if err != nil {
		err = fmt.Errorf("failed to produce message: %w", err)
	}

	return err
}

func (p producer) Close() error {
	return p.writer.Close()
}
