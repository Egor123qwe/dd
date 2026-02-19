package handler

import (
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/handler/history"
)

type Handler interface {
	History() history.History
}

type handler struct {
	history history.History
}

func New(cfg *config.Config, log *slog.Logger) Handler {
	return handler{
		history: history.New(cfg, log),
	}
}

func (h handler) History() history.History {
	return h.history
}
