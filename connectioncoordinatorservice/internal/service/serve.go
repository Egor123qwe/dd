package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
	parcer "gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/util/msg"
	"golang.org/x/sync/errgroup"
)

func (s service) serveToBroker(ctx context.Context, req serve) error {
	gr, grCtx := errgroup.WithContext(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	gr.Go(func() error {
		defer wg.Done()
		return s.readLoop(grCtx, req.wsConn.Reader().Read, req.toBroker)
	})

	gr.Go(func() error {
		defer wg.Done()

		writeFn := func(ctx context.Context, m msg.MSG) error {
			dest := msg.Connection{
				ConnID:    req.wsConn.ID().ConnID,
				UserID:    req.wsConn.ID().UserID,
				SessionID: req.wsConn.ID().SessionID,
				Type:      msg.AllID,
			}

			m, err := parcer.New(m).WithConnection(dest)
			if err != nil {
				// use goroutine to avoid locks in case of error in toWS channel re
				go func() { req.toWS <- parcer.ErrorResponse(fmt.Errorf("failed to parse message: %w", err)) }()

				return nil
			}

			var parsedMessage msg.Full

			if err := json.Unmarshal(m.Data, &parsedMessage); err != nil {
				go func() { req.toWS <- parcer.ErrorResponse(fmt.Errorf("failed to unmarshal message: %w", err)) }()

				return nil
			}

			if parsedMessage.Type == s.config.keepalive {
				return s.broker.Producer().Produce(ctx, m, s.config.keepalive)
			}

			return s.broker.Producer().Produce(ctx, m, s.config.input)
		}

		return s.writeLoop(grCtx, writeFn, req.toBroker)
	})

	wg.Wait()
	return gr.Wait()
}

func (s service) serveToWS(ctx context.Context, req serve) error {
	gr, grCtx := errgroup.WithContext(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	gr.Go(func() error {
		defer wg.Done()

		brokerListener := s.listener.Subscribe(req.wsConn.ID().ConnID, req.wsConn.ID().UserID)
		defer brokerListener.Close()

		readFn := func() (msg.MSG, error) {
			message, err := brokerListener.Read()
			if err != nil {
				return msg.MSG{}, fmt.Errorf("failed to read message: %w", err)
			}

			var parsedMsg msg.InitResponse

			if err := json.Unmarshal(message.Data, &parsedMsg); err != nil {
				return msg.MSG{}, fmt.Errorf("failed to parse consumed message: %v", err)
			}

			if parsedMsg.Type == "share-p2p-init" {
				req.wsConn.SetSessionID(parsedMsg.Meta.SessionID)
			}

			return message, nil
		}

		return s.readLoop(grCtx, readFn, req.toWS)
	})

	gr.Go(func() error {
		defer wg.Done()
		return s.writeLoop(grCtx, req.wsConn.Writer().Write, req.toWS)
	})

	wg.Wait()
	return gr.Wait()
}
