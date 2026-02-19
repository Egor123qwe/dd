// Package server configures and starts servers for handling incoming requests.
package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler"
	httplauncher "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/server/launcher/http"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/server/launcher"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/server/launcher/kafka"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service"
	"golang.org/x/sync/errgroup"
)

var log = logging.MustGetLogger("server")

type server struct {
	servers []launcher.Server
}

func New(srv service.Service) (launcher.Server, error) {
	broker, err := broker.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create broker: %w", err)
	}

	handler := handler.New(srv, broker)

	httpSrv, err := httplauncher.New(srv)
	if err != nil {
		return nil, fmt.Errorf("failed to create http launcher: %w", err)
	}

	result := &server{
		servers: []launcher.Server{
			kafka.New(broker.Kafka, handler.Event),
			httpSrv,
		},
	}

	return result, nil
}

func (s *server) Serve(ctx context.Context) error {
	ctx, stop := context.WithCancel(ctx)

	errCh := make(chan error)

	gr, grCtx := errgroup.WithContext(ctx)

	// start server
	gr.Go(func() error {
		return s.serve(grCtx)
	})

	go func() {
		defer close(errCh)
		errCh <- gr.Wait()
	}()

	var err error

	select {
	case <-getExitSignal():

	case err = <-errCh:
		if err != nil {
			log.Criticalf("app error: %v", err)
		}
	}

	stop()

	log.Infof("app: shutting down the server...")
	<-errCh

	return err
}

func (s *server) serve(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(len(s.servers))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, server := range s.servers {
		server := server

		gr.Go(func() error {
			defer wg.Done()

			return server.Serve(grCtx)
		})
	}

	wg.Wait()

	return gr.Wait()
}

func getExitSignal() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	return quit
}
