// Package server configures and starts servers for handling incoming requests.
package server

import (
	"context"
	"sync"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/server/launcher"
	grpc2 "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/server/launcher/grpc"
	"golang.org/x/sync/errgroup"
)

type server struct {
	servers []launcher.Server
}

func New(handler handler.Handler) launcher.Server {
	return &server{
		servers: []launcher.Server{
			grpc2.New(grpc2.NewConfig(), handler.Grpc),
		},
	}
}

func (s *server) Serve(ctx context.Context) error {
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
