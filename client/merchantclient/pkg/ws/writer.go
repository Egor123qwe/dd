package ws

import (
	"context"

	"github.com/gorilla/websocket"
)

const (
	msgType = websocket.TextMessage
)

type Writer interface {
	Write(ctx context.Context, msg []byte) error
}

type writer struct{ conn *websocket.Conn }

func (c conn) Writer() Writer {
	return writer{conn: c.conn}
}

func (w writer) Write(ctx context.Context, msg []byte) error {
	if err := w.conn.WriteMessage(msgType, msg); err != nil {
		if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
			return ErrDisconnected
		}

		return err
	}

	return nil
}
