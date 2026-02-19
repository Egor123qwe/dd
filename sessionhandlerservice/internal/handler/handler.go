package handler

import (
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/broker"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/event"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/msghandler"
)

type Handler struct {
	Event msghandler.Server
}

func New(srv service.Service, broker broker.Broker) Handler {
	handler := Handler{
		Event: event.New(srv, broker.Kafka),
	}

	return handler
}
