package proxy

import (
	"context"
	"fmt"
	"sync"

	proxyAuth "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/net"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/sync/fnController"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	authLabel = "container auth"
)

var log = logger.NewLogger("proxy", logger.DefaultWithSentry())

type Service interface {
	Start(ctx context.Context, settings proxyAuth.Credentials, ports []string) error
	Stop(ctx context.Context)

	State(ctx context.Context) (proxyAuth.State, error)
}

type service struct {
	state *proxyAuth.State

	serveController fnController.Controller

	mutex      *sync.Mutex
	serveMutex *sync.Mutex
}

func New() Service {
	srv := &service{
		state: &proxyAuth.State{},

		serveController: fnController.New(),

		mutex:      &sync.Mutex{},
		serveMutex: &sync.Mutex{},
	}

	return srv
}

func (s service) Start(ctx context.Context, creeds proxyAuth.Credentials, ports []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// stop previous proxy
	s.unsafeStop()

	if 0 == len(ports) {
		log.Errorf("no ports provided in local proxy")

		return nil
	}

	serveContext, cancel := context.WithCancel(context.Background())
	s.serveController.SetCancelFn(cancel)

	freePorts, err := net.GetFreePorts(ctx, len(ports))
	if err != nil {
		return fmt.Errorf("failed to get free ports: %w", err)
	}

	var portsToProxy []proxyAuth.Port

	for i, port := range ports {
		hostPort := freePorts[i]

		portsToProxy = append(portsToProxy, proxyAuth.Port{
			InPort:  port,
			OutPort: hostPort,
		})
	}

	go func() {
		s.serveMutex.Lock()
		defer s.serveMutex.Unlock()

		if err := s.servePorts(serveContext, creeds, portsToProxy); err != nil {
			log.Errorf("failed to serve ports: %v", err)
		}

		s.state.Enabled = false
		s.state.Ports = nil
	}()

	s.state.Enabled = true
	s.state.Ports = portsToProxy

	return nil
}

func (s service) Stop(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.unsafeStop()
}

func (s service) unsafeStop() {
	s.serveController.Cancel()

	// wait for proxy to stop
	s.serveMutex.Lock()
	defer s.serveMutex.Unlock()
}

func (s service) State(ctx context.Context) (proxyAuth.State, error) {
	log.Info("[proxy state] called")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Info("[proxy state] started")

	return *s.state, nil
}
