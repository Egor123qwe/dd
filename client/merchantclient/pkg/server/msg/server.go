package msg

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/model"
)

type ReaderFn func(ctx context.Context) ([]byte, error)

func NewServer(fn HandleFunc, reader ...ReaderFn) server.Server {
	var gr []server.Server

	for _, r := range reader {
		srv := &service{
			handler: fn,
			reader:  r,
		}

		gr = append(gr, srv)
	}

	return server.NewServeGr(gr...)
}

type service struct {
	reader  ReaderFn
	handler HandleFunc
}

func (s service) Serve(ctx context.Context) error {
	errCh := make(chan error)
	msgCh := make(chan []byte)

	for {
		go func() {
			m, err := s.reader(ctx)
			if err != nil {
				errCh <- fmt.Errorf("failed to read message: %w", err)

				return
			}

			msgCh <- m
		}()

		select {
		case <-ctx.Done():
			return nil

		case err := <-errCh:
			return err

		case m := <-msgCh:
			go func() {
				if err := s.handler(ctx, m); err != nil {
					if errors.Is(err, model.FatalErr) {
						errCh <- err
					}
				}
			}()
		}

	}
}
