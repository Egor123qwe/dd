package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model"
)

var log = logging.MustGetLogger("ws")

func (h handler) initWSRoutes() {
	h.router.HandleFunc("/api/v1/ws", h.init)
}

func (h handler) init(w http.ResponseWriter, r *http.Request) {
	conn, err := h.srv.WS().Init(w, r)
	if err != nil {
		log.Errorf("websocket init error: %s", err)

		switch {
		case errors.Is(err, model.ErrMissingAuth):
			http.Error(w, err.Error(), http.StatusUnauthorized)

		case errors.Is(err, model.ErrInvalidAuth):
			http.Error(w, err.Error(), http.StatusUnauthorized)

		case errors.Is(err, model.ErrBadClaims):
			http.Error(w, err.Error(), http.StatusUnauthorized)

		case errors.Is(err, model.ErrWSConn):
			http.Error(w, err.Error(), http.StatusInternalServerError)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	go func() {
		defer conn.Close()

		if err := h.srv.Serve(context.Background(), conn); err != nil {
			log.Errorf("websocket serve error: %s", err, h.srv.WS())
		}
	}()
}
