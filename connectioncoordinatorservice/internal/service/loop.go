package service

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

type readFn func() (msg.MSG, error)

type writeFn func(ctx context.Context, msg msg.MSG) error

func (s service) writeLoop(ctx context.Context, fn writeFn, ch <-chan msg.MSG) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg := <-ch:
			if err := fn(ctx, msg); err != nil {
				return err
			}
		}
	}
}

func (s service) readLoop(ctx context.Context, fn readFn, ch chan<- msg.MSG) error {
	reader := func(fn readFn) (<-chan msg.MSG, <-chan error) {
		errCh := make(chan error)
		msgCh := make(chan msg.MSG)

		go func() {
			defer close(errCh)
			defer close(msgCh)

			msg, err := fn()
			if err != nil {
				errCh <- err
				return
			}

			msgCh <- msg
		}()

		return msgCh, errCh
	}

	for {
		msgCh, errCh := reader(fn)

		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg := <-msgCh:
			ch <- msg

		case err := <-errCh:
			return err
		}
	}
}
