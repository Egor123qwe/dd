package tariff

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"
)

type Service struct {
	tariffRepo TariffRepository
}

type TariffRepository interface {
	List(ctx context.Context, log slog.Logger) ([]model.Tariff, error)
}

func New(tariffRepo TariffRepository) Service {
	s := Service{
		tariffRepo: tariffRepo,
	}

	return s
}

func (s Service) List(ctx context.Context, log slog.Logger) ([]model.Tariff, error) {
	return s.tariffRepo.List(ctx, log)
}
