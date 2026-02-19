package handler

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	broker "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/broker/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/message"
)

type Broker struct {
	broker                 broker.Broker
	cfg                    config.Config
	merhcantHandler        MerchantHandler
	ttlNotificationHandler TTLNotificationHandler
}

type MerchantHandler interface {
	ShareP2PInit(ctx context.Context, mt string, msg []byte) []byte
	ShareP2PReady(ctx context.Context, mt string, msg []byte) []byte
	ShareP2PStop(ctx context.Context, mt string, msg []byte) []byte
}

type TTLNotificationHandler interface {
	ExpireSession(ctx context.Context, msg message.FullMessage) []byte
}

func New(broker broker.Broker, cfg config.Config, merhcantHandler MerchantHandler, ttlNotificationHandler TTLNotificationHandler) Broker {
	b := Broker{
		broker:                 broker,
		cfg:                    cfg,
		merhcantHandler:        merhcantHandler,
		ttlNotificationHandler: ttlNotificationHandler,
	}

	return b
}

func (b Broker) ConsumeInput(ctx context.Context) {
	inputTopic := b.cfg.Kafka.Consumer.Topics.Input
	outputTopic := b.cfg.Kafka.Producer.Topic
	groupID := b.cfg.Kafka.Consumer.ConsumerGroup

	consumer := b.broker.Consumer(inputTopic, groupID)

	for {
		select {
		case <-ctx.Done():
			consumer.Close()
			return

		default:
			msg, err := consumer.Consume(ctx)
			if err != nil {
				log.Error().Err(err).Msgf("Error reading message from topic %s", outputTopic)

				continue
			}

			var fullMsg message.FullMessage
			if err := json.Unmarshal(msg.Value, &fullMsg); err != nil {
				log.Error().Err(err).Msg("error unmarshalling message")
				continue
			}

			var res []byte

			log.Info().Msgf("Received message: %v", string(msg.Value))

			go func() {
				switch fullMsg.Type {
				case string(event.ShareP2PInit):
					res = b.merhcantHandler.ShareP2PInit(ctx, fullMsg.Type, msg.Value)

				case string(event.ShareP2PReady):
					res = b.merhcantHandler.ShareP2PReady(ctx, fullMsg.Type, msg.Value)

				case string(event.ShareP2PStop):
					res = b.merhcantHandler.ShareP2PStop(ctx, fullMsg.Type, msg.Value)

				default:
					log.Error().Msg("unknown event type")
				}

				err = b.broker.Producer().Produce(ctx, outputTopic, []byte(fullMsg.Meta.MessageID), res)
				if err != nil {
					log.Error().Err(err).Msg("error producing message")
				}
			}()
		}
	}
}

func (b Broker) ConsumeNotification(ctx context.Context) {
	ttlNotificationTopic := b.cfg.Kafka.Consumer.Topics.TTLNotification
	outputTopic := b.cfg.Kafka.Producer.Topic
	groupID := b.cfg.Kafka.Consumer.ConsumerGroup

	consumer := b.broker.Consumer(ttlNotificationTopic, groupID)
	for {
		select {
		case <-ctx.Done():
			consumer.Close()
			return

		default:
			msg, err := consumer.Consume(ctx)
			if err != nil {
				log.Error().Err(err).Msgf("Error reading message from topic %s", outputTopic)

				continue
			}

			var fullMsg message.FullMessage
			if err := json.Unmarshal(msg.Value, &fullMsg); err != nil {
				log.Error().Err(err).Msg("error unmarshalling message")

				continue
			}

			var res []byte

			log.Info().Msgf("Received message: %v", string(msg.Value))

			go func() {
				switch fullMsg.Type {
				case string(event.Expired):
					res = b.ttlNotificationHandler.ExpireSession(ctx, fullMsg)

				default:
					log.Error().Msg("unknown event type")
				}

				err = b.broker.Producer().Produce(ctx, outputTopic, []byte(fullMsg.Meta.SessionID), res)
				if err != nil {
					log.Error().Err(err).Msg("error producing message")
				}
			}()
		}
	}
}
