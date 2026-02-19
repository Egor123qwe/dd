package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/server/kafka"
	api "gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/server/rest"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

type Server struct {
	wg        *sync.WaitGroup
	broker    kafka.BrokerServer
	apiServer api.RestServer
	log       slog.Logger
	storage   storage.Storage
}

func New(log slog.Logger, cfg config.Config, storage storage.Storage) (*Server, error) {
	broker, err := kafka.New(log, cfg, storage)
	if err != nil {
		return nil, err
	}

	apiServer := api.New(cfg, log, storage)

	wg := &sync.WaitGroup{}

	return &Server{
		broker:    broker,
		apiServer: apiServer,
		log:       log,
		wg:        wg,
		storage:   storage,
	}, nil
}

func (s *Server) Start(ctx context.Context) {
	s.wg.Add(1)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer s.wg.Done()
		s.broker.Start(ctx)
	}()

	s.apiServer.Start(ctx)

	<-stop

	s.apiServer.Stop(ctx)

	cancel()

	err := s.storage.Close()
	if err != nil {
		s.log.Error("Failed to close storage: ", err.Error(), nil)
	}

	s.log.Info("Storage closed")

	s.log.Info("Gracefully stopped")
	s.wg.Wait()
}
