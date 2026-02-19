package conn

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/ws"
)

type Conn interface {
	ID() ws.IDs
	// Reader returns a channel of read messages
	Reader() Reader

	// Writer returns a channel to Write messages
	Writer() Writer

	// Close closes the connection
	Close() error

	SetSessionID(sessionID string)
}

type conn struct {
	ids  ws.IDs
	conn *websocket.Conn
}

func New(wsConn *websocket.Conn, userID, sessionID string) Conn {
	conn := &conn{
		ids: ws.IDs{
			UserID: userID,
			ConnID: uuid.New().String(),
			SessionID: sessionID,
		},

		conn: wsConn,
	}

	return conn
}

func (c conn) Close() error {
	return c.conn.Close()
}

func (c conn) ID() ws.IDs {
	return c.ids
}

func (c *conn) SetSessionID(sessionID string) {
	c.ids.SessionID = sessionID
}
