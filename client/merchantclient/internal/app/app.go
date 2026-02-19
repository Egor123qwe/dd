// Package app handles starting and monitoring the server for graceful shutdown.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	httpServer "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/http"
	wsServer "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/ws"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/handler"
	msgModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/rent"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/ws"
)

var log = logging.MustGetLogger("app")

const (
	startTimeout = 20 * time.Second
	stopTimeout  = 20 * time.Second
)

type App struct {
	srv service.Service
	rd  runtimedaemon.API

	server server.Server
	conn   ws.Conn
}

type Options struct {
	CheatMode bool
}

func New(token string) (App, error) {
	rd, err := runtimedaemon.New()
	if err != nil {
		return App{}, err
	}
	return NewWithRD(token, "", rd)
}

// NewWithRD creates an app using an existing runtime daemon (e.g. for web UI connect).
// connectionURL is optional; if empty, config state_machine.connection_url is used.
func NewWithRD(token string, connectionURL string, rd runtimedaemon.API) (App, error) {
	log.Infof("initializing merchant v%s", viper.GetString("version"))

	conn, err := newConnWithURL(token, connectionURL)
	if err != nil {
		return App{}, err
	}

	log.Infof("auth success")

	wsConn := wsServer.Conn{
		Conn:    conn,
		MsgConn: msg.NewConn(msgModel.NewIDParser()),
	}

	srv, err := service.New(wsConn, rd, token)
	if err != nil {
		_ = conn.Close()
		return App{}, fmt.Errorf("failed to create service: %w", err)
	}

	a := App{
		srv: srv,
		rd:  rd,

		server: wsServer.New(conn, wsConn.MsgConn, handler.New(srv).Event),
		conn:   conn,
	}

	return a, nil
}

// ServeWS runs only the WebSocket message server (for use when rd is already running elsewhere).
func (a App) ServeWS(ctx context.Context) error {
	return a.server.Serve(ctx)
}

// Close closes the WebSocket connection.
func (a App) Close() error {
	return a.conn.Close()
}

// Rent returns the rent usecase for this app.
func (a App) Rent() rent.Usecase {
	return a.srv.Usecase().Rent()
}

func (a App) Start(ctx context.Context, opts Options) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()

	// start serving
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		var err error

		if err = server.NewServeGr(a.server, a.rd).Serve(ctx); err != nil {
			log.Errorf("failed to serve: %s", err)
		}

		stop()
		errCh <- err
	}()

	rdWaitCtx, cancel := context.WithTimeout(ctx, startTimeout)
	defer cancel()

	if err := a.rd.WaitEnabled(rdWaitCtx); err != nil {
		return fmt.Errorf("failed to wait for runtime daemon started: %w", err)
	}

	log.Info("runtime daemon enabled")

	web := httpServer.New(RentOnlyBackend{RentUC: a.srv.Usecase().Rent(), RD: a.rd})
	go func() {
		_ = web.Serve(ctx)
	}()

	log.Infof("client started. Web UI available (see config http.port, default :8080)")

	var runErr error
	select {
	case runErr = <-errCh:
		if runErr != nil {
			log.Errorf("server stopped with error: %v", runErr)
		}
	case <-ctx.Done():
		log.Infof("shutdown signal received")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	if err := a.srv.Usecase().Rent().StopRequest(stopCtx); err != nil {
		log.Warningf("failed to stop merchant: %s", err)
	} else {
		log.Infof("merchant stopped successfully")
	}

	return runErr
}
