package ws

import (
	"context"
	"errors"
	"fmt"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/ws"
)

var log = logging.MustGetLogger("server")

type Conn struct {
	ws.Conn
	MsgConn msg.Conn
}

func New(conn ws.Conn, msgConn msg.Conn, msgHandler msg.Resolver) server.Server {
	srv := msg.NewServer(
		serverFn(msgConn, msgHandler),
		func(ctx context.Context) ([]byte, error) { return conn.Reader().Read() },
	)

	return srv
}

func serverFn(msgConn msg.Conn, msgHandler msg.Resolver) msg.HandleFunc {
	serveMSG := func(ctx context.Context, m []byte) error {
		returnFn := func(err error) error {
			if err != nil {
				if errors.Is(err, msg.ErrUnknownResp) || errors.Is(err, msg.UnknownEventErr) {
					log.Debugf("failed to serve message: %s", err)
				} else {
					log.Error(err)
				}
			}

			return err
		}

		err := msgConn.Serve(m)

		if err != nil && !errors.Is(err, msg.ErrUnknownResp) {
			return returnFn(fmt.Errorf("failed to serve response: %w", err))
		}

		// ErrUnknownResp mean that there was no response, and this is server generated request
		if errors.Is(err, msg.ErrUnknownResp) {
			return returnFn(msgHandler.Serve(ctx, m))
		}

		return nil
	}

	return serveMSG
}
