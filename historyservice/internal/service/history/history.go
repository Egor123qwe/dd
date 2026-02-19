package history

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/storage"
)

type History interface {
	Session(userID string) ([]model.Rent, error)
	Rent(userID string) ([]model.Rent, error)
	All() ([]model.AdminRent, error)
}

type service struct {
	storage storage.Storage
	log     *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) History {
	return &service{
		storage: storage.NewStorage(cfg, log),
		log:     log,
	}
}

func (s service) Session(userID string) ([]model.Rent, error) {
	ctx := context.Background()

	sessionHistory, err := s.storage.History().SessionHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sessionHistory, nil
}

func (s service) Rent(userID string) ([]model.Rent, error) {
	ctx := context.Background()

	rents, err := s.storage.History().RentHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	return rents, nil
}

func (s service) All() ([]model.AdminRent, error) {
	ctx := context.Background()
	return s.storage.History().AllRents(ctx)
}
