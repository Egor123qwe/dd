package conn

import (
	"github.com/gorilla/websocket"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/msg"
)

type Reader interface {
	Read() (msg.MSG, error)
}

type reader struct{ conn *websocket.Conn }

func (c conn) Reader() Reader {
	return reader{conn: c.conn}
}

func (r reader) Read() (msg.MSG, error) {
	_, data, err := r.conn.ReadMessage()
	if err != nil {
		return msg.MSG{}, err
	}

	return msg.MSG{Data: data}, nil
}
