package repo

import (
	"context"
	"errors"

	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/domain/model"
)

var ErrNoHistory = errors.New("user has no history")

type History interface {
	RentHistory(ctx context.Context, userID string) ([]model.Rent, error)
	SessionHistory(ctx context.Context, userID string) ([]model.Rent, error)
	AllRents(ctx context.Context) ([]model.AdminRent, error)
}
