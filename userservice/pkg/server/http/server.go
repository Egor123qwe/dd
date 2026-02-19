package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/server"
)

type srv struct {
	srv    *http.Server
	config Config
}

type Config struct {
	Port         int
	ShutdownTime time.Duration
	ReadTime     time.Duration
}

func New(handler http.Handler, config Config) server.Server {
	srv := &srv{
		srv: &http.Server{
			Addr:        fmt.Sprintf(":%d", config.Port),
			Handler:     handler,
			ReadTimeout: config.ReadTime,
		},

		config: config,
	}

	return srv
}

func (s *srv) Serve(ctx context.Context) error {
	errCh := make(chan error)

	go func() {
		errCh <- s.srv.ListenAndServe()

		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTime)
		defer cancel()

		err := s.srv.Shutdown(ctx)
		<-errCh

		return err
	}
}
