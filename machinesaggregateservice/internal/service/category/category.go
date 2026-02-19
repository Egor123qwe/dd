package category

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"
)

type Service struct {
	categoryRepo CategoryRepository
}

type CategoryRepository interface {
	GPUDictList(ctx context.Context, log slog.Logger, vramFrom, vramTo int) ([]model.GPUDict, error)
}

func New(categoryRepository CategoryRepository) Service {
	s := Service{
		categoryRepo: categoryRepository,
	}

	return s
}

func (s Service) GPUDictList(ctx context.Context, log slog.Logger, vramFrom, vramTo int) ([]model.GPUDict, error) {
	return s.categoryRepo.GPUDictList(ctx, log, vramFrom, vramTo)
}
