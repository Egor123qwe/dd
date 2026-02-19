package kafka

import (
	"context"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

const (
	produceTimeout = 5 * time.Second
	batchTimeout   = 300 * time.Millisecond
)

type Producer struct {
	writter *kafka.Writer
}

func NewProducer(brokers []string, dialer *kafka.Dialer) *Producer {
	producer := &Producer{
		writter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      brokers,
			Balancer:     &kafka.LeastBytes{},
			Dialer:       dialer,
			BatchTimeout: batchTimeout,
		}),
	}

	return producer
}

func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	c, cancel := context.WithTimeout(ctx, produceTimeout)
	defer cancel()

	err := p.writter.WriteMessages(c, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})

	return err
}

func (p *Producer) Close() error {
	return p.writter.Close()
}
