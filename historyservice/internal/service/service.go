package service

import (
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/service/history"
)

type Sevice interface {
	History() history.History
}

type service struct {
	history history.History
}

func New(cfg *config.Config, log *slog.Logger) Sevice {
	return &service{
		history: history.New(cfg, log),
	}
}

func (s service) History() history.History {
	return s.history
}
