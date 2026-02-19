package kafka

import (
	"context"
	"sync"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/server/launcher"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/msghandler"
	"golang.org/x/sync/errgroup"
)

var log = logging.MustGetLogger("kafka")

type server struct {
	handler msghandler.Server
	broker  kafka.Service
	config  config
}

func New(broker kafka.Service, handler msghandler.Server) launcher.Server {
	server := &server{
		handler: handler,
		broker:  broker,
		config:  newConfig(),
	}

	return server
}

func (s server) Serve(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(len(s.config.consumers))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, c := range s.config.consumers {
		c := c

		gr.Go(func() error {
			defer wg.Done()

			return s.serve(grCtx, c)
		})
	}

	wg.Wait()

	return gr.Wait()
}

func (s server) serve(ctx context.Context, consumer consumer) error {
	c := s.broker.Consumer(consumer.topic, consumer.groupID)

	log.Infof("kafka listener started for [%s] topic with group id: %s", consumer.topic, consumer.groupID)

	for {
		if err := ctx.Err(); err != nil {
			log.Infof("kafka listener stopped [%s]", consumer.topic)

			return nil
		}

		m, err := c.Consume(ctx)
		if err != nil {
			log.Errorf("failed to consume message [%s]: %v", consumer.topic, err)

			continue
		}

		go func() {
			err := s.handler.Serve(ctx, m)
			if err != nil {
				log.Errorf("failed to handle message: %v", err)

				return
			}
		}()
	}
}
