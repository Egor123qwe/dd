package server

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type ServeGr struct {
	Servers []Server
}

func NewServeGr(servers ...Server) ServeGr {
	gr := ServeGr{
		Servers: servers,
	}

	return gr
}

func (s ServeGr) Serve(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(len(s.Servers))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, s := range s.Servers {
		s := s

		gr.Go(func() error {
			defer wg.Done()

			return s.Serve(grCtx)
		})
	}

	wg.Wait()

	return gr.Wait()
}
