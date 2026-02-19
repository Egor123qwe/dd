package custom

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/server"
	"github.com/Interpuls/ifc2-service-farm/pkg/util/function"
)

type Reader interface {
	Read(ctx context.Context) ([]byte, error)
}

type srv struct {
	handler Handler
	reader  Reader
	options Options
}

type Options struct {
	ErrHandler func(err error)
}

func NewServer(handler Handler, reader Reader, options Options) server.Server {
	srv := &srv{
		handler: handler,
		reader:  reader,
		options: options,
	}

	return srv
}

func (s srv) Serve(ctx context.Context) error {
	errCh := make(chan error)
	defer close(errCh)

	msgCh := make(chan []byte)
	defer close(msgCh)

	for {
		go func() {
			m, err := s.reader.Read(ctx)
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
				if err := s.handler.Handle(ctx, m); err != nil {
					function.SafeCall(s.options.ErrHandler, err)
				}
			}()
		}
	}
}
