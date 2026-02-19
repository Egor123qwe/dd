package handler

import (
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/handler/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/server/msg"
)

type Handler struct {
	Event msg.Resolver
}

func New(srv service.Service) Handler {
	handler := Handler{
		Event: event.New(srv),
	}

	return handler
}
