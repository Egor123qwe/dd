package piko

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	piko "github.com/andydunstall/piko/client"
	pikoModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"golang.org/x/sync/errgroup"
)

const (
	connectTimeout     = 15 * time.Second
	checkOnlineTimeout = 10 * time.Second
)

func (s service) serveEndpoints(ctx context.Context, req pikoModel.ConnectReq) error {
	var wg sync.WaitGroup
	wg.Add(len(req.Endpoints))

	gr, grCtx := errgroup.WithContext(ctx)

	for _, e := range req.Endpoints {
		e := e

		gr.Go(func() error {
			defer wg.Done()

			return s.serveEndpoint(grCtx, e, req.AuthKey)
		})
	}

	wg.Wait()

	return gr.Wait()

}

func (s service) serveEndpoint(ctx context.Context, endpoint pikoModel.Endpoint, authkey string) error {
	url, err := url.Parse(s.config.serverURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	upstream := &piko.Upstream{
		URL:   url,
		Token: authkey,
	}

	var forwarder *piko.Forwarder

	createForwarderCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)

	go func() {
		forwarder, err = upstream.ListenAndForward(
			createForwarderCtx, endpoint.Settings.Name, fmt.Sprintf("localhost:%s", endpoint.Port),
		)

		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to start listener: start was stopped")

	case <-time.After(connectTimeout):
		return fmt.Errorf("failed to start listener: timeout")

	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to start listener: %w", err)
		}
	}

	defer forwarder.Close()
	errCh = make(chan error)

	go func() {
		var err error

		if err := forwarder.Wait(); err != nil {
			log.Errorf("failed duirng connection: %v", err)
		}

		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return nil

	case err := <-errCh:
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
}

func (s service) checkHTTPEndpoint(ctx context.Context, endpoint pikoModel.Endpoint) error {
	ctx, cancel := context.WithTimeout(ctx, checkOnlineTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf(s.config.linkTemplate, endpoint.Settings.Name), nil,
	)

	if err != nil {
		return fmt.Errorf("failed to create request [%s]: %w", endpoint.Settings.Name, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("piko endpoint %s not online", endpoint.Settings.Name)
	}

	if resp.StatusCode == http.StatusBadGateway {
		return fmt.Errorf("piko endpoint %s not reachable", endpoint.Settings.Name)
	}

	resp.Body.Close()

	return nil
}
