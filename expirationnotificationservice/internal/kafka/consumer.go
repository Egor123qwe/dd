package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/expirationnotificationservice/internal/domain/message"

	"github.com/segmentio/kafka-go"
)

// cunsumer config
const (
	readBatchTimeot = 500 * time.Millisecond
	maxWait         = 500 * time.Millisecond
)

type MessageHandler func(msg kafka.Message) error

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string, dialer *kafka.Dialer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          brokers,
		GroupID:          groupID,
		Topic:            topic,
		Dialer:           dialer,
		MinBytes:         10e3,
		MaxBytes:         10e6,
		ReadBatchTimeout: readBatchTimeot,
		MaxWait:          maxWait,
	})

	consumer := &Consumer{
		reader: reader,
	}

	return consumer
}

func (c *Consumer) Consume(ctx context.Context) (message.Message, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return message.Message{}, err
	}

	var msg message.Message
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		return message.Message{}, err
	}

	log.Print(msg)

	return msg, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
