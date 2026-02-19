package ws

import (
	"fmt"
	"net/http"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/ws/auth"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/ws/conn"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/ws/connector"
)

var log = logging.MustGetLogger("ws.service")

type Service interface {
	Init(w http.ResponseWriter, r *http.Request) (conn.Conn, error)
}

type service struct {
	auth      auth.Auth
	connector connector.Connector
}

func New(api api.Service, debug bool) Service {
	return &service{
		auth:      auth.New(api.Auth(), debug),
		connector: connector.New(),
	}
}

func (s service) Init(w http.ResponseWriter, r *http.Request) (conn.Conn, error) {
	userID, err := s.auth.Auth(w, r)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	sessionID := r.Header.Get("X-Session")

	wsConn, err := s.connector.Connect(w, r)
	if err != nil {
		return nil, err
	}

	log.Infof("user connected: user_id=%s session_id=%s", userID, sessionID)
	return conn.New(wsConn, userID, sessionID), nil
}
