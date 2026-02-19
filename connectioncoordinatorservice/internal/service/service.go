package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/broker"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/listener"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/ws"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/ws/conn"
	"golang.org/x/sync/errgroup"
)

var log = logging.MustGetLogger("service")

type Service interface {
	Serve(ctx context.Context, conn conn.Conn) error
	WS() ws.Service
}
	
type service struct {
	ws       ws.Service
	broker   broker.Service
	listener listener.Listener

	config config
}

func New(api api.Service, debug bool) (Service, error) {
	broker, err := broker.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create broker: %w", err)
	}

	listener := listener.NewListener(broker)

	srv := &service{
		ws:       ws.New(api, debug),
		broker:   broker,
		listener: listener,

		config: newConfig(),
	}

	return srv, nil
}

type serve struct {
	wsConn   conn.Conn
	toWS     chan msg.MSG
	toBroker chan msg.MSG
}

func (s service) Serve(ctx context.Context, wsConn conn.Conn) error {
	gr, grCtx := errgroup.WithContext(ctx)

	serve := serve{
		wsConn:   wsConn,
		toWS:     make(chan msg.MSG),
		toBroker: make(chan msg.MSG),
	}

	defer close(serve.toWS)
	defer close(serve.toBroker)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	gr.Go(func() error {
		defer wg.Done()
		return s.serveToWS(grCtx, serve)
	})

	gr.Go(func() error {
		defer wg.Done()
		return s.serveToBroker(grCtx, serve)
	})

	wg.Wait()
	return gr.Wait()
}

func (s service) WS() ws.Service {
	return s.ws
}
