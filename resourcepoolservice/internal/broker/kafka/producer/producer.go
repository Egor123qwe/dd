package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
)

const (
	produceTimeout = 5 * time.Second
	batchTimeout   = 300 * time.Millisecond
)

type Producer interface {
	Produce(ctx context.Context, topic string, key, msg []byte) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

func New(mechanism sasl.Mechanism, brokers []string) Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: batchTimeout,
	}
	if mechanism != nil {
		writer.Transport = &kafka.Transport{SASL: mechanism}
	}

	return producer{writer: writer}
}

func (p producer) Produce(ctx context.Context, topic string, key, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, produceTimeout)
	defer cancel()

	err := p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Key:   key,
			Value: msg,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (p producer) Close() error {
	return p.writer.Close()
}
