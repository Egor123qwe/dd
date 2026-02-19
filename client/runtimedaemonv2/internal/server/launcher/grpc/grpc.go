package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	handler "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/server/launcher"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcStopTimeout = 10 * time.Second
)

var log = logger.NewLogger("grpc", logger.DefaultWithSentry())

type server struct {
	config Config
	srv    *grpc.Server
}

func New(config Config, handler handler.Handler) launcher.Server {
	srv := grpc.NewServer()
	reflection.Register(srv)

	handler.Subscribe(srv)

	return &server{
		config: config,
		srv:    srv,
	}
}

func (s *server) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		log.Fatalf("Failed to listen grpc server: %v", err)
	}

	log.Infof("serving gRPC on http://localhost:%d\n", s.config.Port)

	errCh := make(chan error)

	go func() {
		errCh <- s.srv.Serve(lis)

		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpc-server: %w", err)

	case <-ctx.Done():
		doneCh := make(chan struct{})

		go func() {
			s.srv.GracefulStop()
			close(doneCh)
		}()

		select {
		case <-doneCh:
		case <-time.After(grpcStopTimeout):
			s.srv.Stop()
		}

		log.Infof("grpc-server: server stopped successfully.")
	}

	return nil
}
