package hardware

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"
)

type Service struct {
	hardwareRepo HardwareRepository
}

type HardwareRepository interface {
	GPUList(ctx context.Context, log slog.Logger, filter model.FilterRepo) ([]model.GPU, error)
}

func New(hardwareRepo HardwareRepository) Service {
	s := Service{
		hardwareRepo: hardwareRepo,
	}

	return s
}

func (s Service) GPUList(ctx context.Context, log slog.Logger) ([]model.GPU, error) {
	filter := model.FilterRepo{}

	return s.hardwareRepo.GPUList(ctx, log, filter)
}

func (s Service) GPUByID(ctx context.Context, log slog.Logger, gpuID string) ([]model.GPU, error) {
	filter := model.FilterRepo{
		ID:   gpuID,
		Type: model.TypeGPU,
	}

	return s.hardwareRepo.GPUList(ctx, log, filter)
}
