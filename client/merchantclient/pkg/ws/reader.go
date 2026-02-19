package ws

import (
	"github.com/gorilla/websocket"
)

type Reader interface {
	Read() ([]byte, error)
}

type reader struct{ conn *websocket.Conn }

func (c conn) Reader() Reader {
	return reader{conn: c.conn}
}

func (r reader) Read() ([]byte, error) {
	_, data, err := r.conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
			return nil, ErrDisconnected
		}

		return nil, err
	}

	return data, nil
}
