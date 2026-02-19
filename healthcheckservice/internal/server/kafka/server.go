package kafka

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/broker"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

var ErrContextCanceled = errors.New("context canceled")

type BrokerServer interface {
	Start(ctx context.Context)
}

type server struct {
	handler handler.Handler
	broker  broker.Broker
	log     slog.Logger
}

func New(log slog.Logger, cfg config.Config, storage storage.Storage) (BrokerServer, error) {
	broker, err := broker.New(cfg, log)

	if err != nil {
        log.Error(err.Error())
        return nil, err
    }

	handler := handler.New(cfg, log, storage)

	return &server{
		handler : handler,
		broker: broker,
		log:    log,
	}, nil
}

func (s server) Start(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := s.serve(ctx)
		if err != nil && !errors.Is(ErrContextCanceled, err){
			s.log.Error("failed to start kafka listener :", err.Error(), nil)
		}
	}()

	wg.Wait()
}

func (s server) serve(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			s.log.Info("kafka listener stopped")

			return ErrContextCanceled
		default:
			m, err := s.broker.Kafka().Consumer().Consume(ctx)
			if err != nil {
				s.log.Error("failed to consume message :", err.Error(), nil)

				continue
			}
			
			go func() {
				err := s.handler.Broker().HandleOutput(ctx, m)
				if err != nil {
					s.log.Error("failed to handle message: %v", err.Error(), nil)

					return
				}
			}()
		}
	}
}
