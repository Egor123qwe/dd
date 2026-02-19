package storage

import (
	"fmt"
	"os"

	docker "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage/repo"
)

const (
	storagePath = "data"
)

type Storage interface {
	Runtime() docker.RuntimeRepo
	Docker() docker.DockerRepo
	Network() docker.NetworkRepo
	Event() docker.EventRepo
}

type storage struct {
	runtime docker.RuntimeRepo
	docker  docker.DockerRepo
	network docker.NetworkRepo
	event   docker.EventRepo
}

func New() (Storage, error) {
	err := os.Mkdir(storagePath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("falid to init storage: %w", err)
	}

	storage := storage{
		runtime: docker.NewRuntime(storagePath),
		docker:  docker.NewDocker(storagePath),
		network: docker.NewNetwork(storagePath),
		event:   docker.NewEvent(storagePath),
	}

	return storage, nil
}

func (s storage) Runtime() docker.RuntimeRepo {
	return s.runtime
}

func (s storage) Docker() docker.DockerRepo {
	return s.docker
}

func (s storage) Network() docker.NetworkRepo {
	return s.network
}

func (s storage) Event() docker.EventRepo {
	return s.event
}
