package server

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type serveGroup struct {
	servers []Server
}

func NewServeGroup(servers ...Server) Server {
	return serveGroup{servers: servers}
}

func (s serveGroup) Serve(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(len(s.servers))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, s := range s.servers {
		s := s

		gr.Go(func() error {
			defer wg.Done()

			return s.Serve(grCtx)
		})
	}

	wg.Wait()

	return gr.Wait()
}
