package handler

import (
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/broker"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/server"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

type Handler interface {
	Broker() broker.Handler
	Server() server.Handler
}

type handler struct {
	broker broker.Handler
	server server.Handler
}

func New(cfg config.Config, log slog.Logger, storage storage.Storage) Handler {
	brokerHandler := broker.New(cfg, log, storage)

	return handler{
		broker: brokerHandler,
	}
}

func (h handler) Broker() broker.Handler {
	return h.broker
}

func (h handler) Server() server.Handler {
    return h.server
}
