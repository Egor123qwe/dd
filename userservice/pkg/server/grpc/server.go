package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Interpuls/ifc2-service-farm/pkg/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type srv struct {
	config     Config
	grpcServer *grpc.Server
}

type Config struct {
	Port int
}

func New(handler Handler, config Config) server.Server {
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	handler.Register(grpcServer)

	srv := &srv{
		config:     config,
		grpcServer: grpcServer,
	}

	return srv
}

func (s *srv) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return err
	}

	errCh := make(chan error)

	go func() {
		errCh <- s.grpcServer.Serve(lis)

		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		s.grpcServer.GracefulStop()
	}

	return nil
}
