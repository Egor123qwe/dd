package conn

import (
	"context"

	"github.com/gorilla/websocket"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

const (
	msgType = websocket.TextMessage
)

type Writer interface {
	Write(ctx context.Context, msg msg.MSG) error
}

type writer struct{ conn *websocket.Conn }

func (c conn) Writer() Writer {
	return writer{conn: c.conn}
}

func (w writer) Write(ctx context.Context, msg msg.MSG) error {
	return w.conn.WriteMessage(msgType, msg.Data)
}
