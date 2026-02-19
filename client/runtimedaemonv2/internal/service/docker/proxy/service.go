package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	auth "github.com/abbot/go-http-auth"
	proxyAuth "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"golang.org/x/sync/errgroup"
)

func (s service) servePorts(ctx context.Context, creeds proxyAuth.Credentials, ports []proxyAuth.Port) error {
	auth := auth.NewBasicAuthenticator(
		authLabel, s.createAuth(creeds),
	)

	var wg sync.WaitGroup
	wg.Add(len(ports))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, port := range ports {
		port := port

		gr.Go(func() error {
			defer wg.Done()

			return s.servePort(grCtx, auth, port.InPort, port.OutPort)
		})
	}

	wg.Wait()

	return gr.Wait()

}

func (s service) servePort(ctx context.Context, auth *auth.BasicAuth, inPort, outPort string) error {
	handler := http.NewServeMux()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", outPort),
		Handler: handler,
	}

	handler.HandleFunc("/", auth.Wrap(s.createHandle(inPort)))

	errCh := make(chan error)

	go func() {
		var err error

		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server error: %v", err)
		}

		errCh <- err
	}()

	select {
	case <-ctx.Done():
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to shutdown HTTP server: %v", err)
		}

		return nil

	case err := <-errCh:
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
}
