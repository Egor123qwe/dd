package connector

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model"
)

type Connector interface {
	Connect(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error)
}

type connector struct {
	wsConnector websocket.Upgrader
}

func New() Connector {
	connector := connector{
		wsConnector: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}

	return connector
}

func (c connector) Connect(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := c.wsConnector.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", model.ErrWSConn, err)
	}

	return conn, nil
}
