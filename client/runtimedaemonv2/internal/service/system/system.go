package system

import (
	"context"
	"fmt"
	"sync"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/command"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("hardware", logger.DefaultWithSentry())

type Service interface {
	InfoFromCache(ctx context.Context) (hardware.Info, error)
	Info(ctx context.Context) (hardware.Info, error)
}

type service struct {
	command command.Service
	network networkSrv

	cache *cache
}

func New() Service {
	return &service{
		command: command.New(),
		network: newNetwork(),

		cache: &cache{m: &sync.Mutex{}},
	}
}

func (s service) InfoFromCache(ctx context.Context) (hardware.Info, error) {
	GPU, err := s.GPU(ctx)
	if err != nil {
		log.Errorf("failed to get gpu infoFromCache: %v", err)
	}

	storage, err := s.Storage(ctx)
	if err != nil {
		log.Errorf("failed to get storage infoFromCache: %v", err)
	}

	return hardware.Info{
		GPU:     GPU,
		Storage: storage,
		CPU:     s.CPU(ctx),
		RAM:     s.RAM(ctx),
		Network: s.network.infoFromCache(ctx),
	}, nil
}

func (s service) Info(ctx context.Context) (hardware.Info, error) {
	GPU, err := s.GPU(ctx)
	if err != nil {
		log.Errorf("failed to get gpu infoFromCache: %v", err)
	}

	storage, err := s.Storage(ctx)
	if err != nil {
		return hardware.Info{}, fmt.Errorf("failed to get storageCache infoFromCache: %w", err)
	}

	return hardware.Info{
		GPU:     GPU,
		Storage: storage,
		CPU:     s.CPU(ctx),
		RAM:     s.RAM(ctx),
		Network: s.network.info(ctx),
	}, nil
}
