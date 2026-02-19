package service

import (
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/service/status"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

type Service interface {
	Status() status.Status
}

type service struct {
	status status.Status
}

func New(cfg config.Config, log slog.Logger, storage storage.Storage) Service {
	return service{
		status: status.New(log, storage),
	}
}

func (s service) Status() status.Status {
	return s.status
}
