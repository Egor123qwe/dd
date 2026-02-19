package piko

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	pikoModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/sync/fnController"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var (
	pikoConnectApproveTime = 5 * time.Second
)

var log = logger.NewLogger("piko", logger.DefaultWithSentry())

type Service interface {
	Connect(ctx context.Context, req pikoModel.ConnectReq) error
	Disconnect(ctx context.Context) error

	State(ctx context.Context, req []pikoModel.Endpoint) (pikoModel.State, error)
	Info(ctx context.Context) (pikoModel.Info, error)
}

type state struct {
	enabled   bool
	endpoints []pikoModel.Endpoint
}

type service struct {
	state  *state
	config config

	serveController fnController.Controller

	mutex        *sync.Mutex
	connectMutex *sync.Mutex
}

func New() Service {
	return &service{
		config: newConfig(),
		state:  &state{},

		serveController: fnController.New(),

		mutex:        &sync.Mutex{},
		connectMutex: &sync.Mutex{},
	}
}

func (s service) Connect(ctx context.Context, req pikoModel.ConnectReq) error {
	log.Info("called piko proxy")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Info("started connect piko proxy")
	defer log.Info("stopped connect piko proxy")

	// stop previous proxy
	s.unsafeDisconnect()

	if 0 == len(req.Endpoints) {
		log.Warning("no ports provided in piko proxy")

		return nil
	}

	serveContext, cancel := context.WithCancel(context.Background())
	s.serveController.SetCancelFn(cancel)

	log.Info("starting piko proxy")

	s.state.enabled = true
	s.state.endpoints = req.Endpoints

	go func() {
		s.connectMutex.Lock()
		defer s.connectMutex.Unlock()

		var err error

		if err = s.serveEndpoints(serveContext, req); err != nil {
			log.Errorf("failed to serve ports: %v", err)
		}

		s.state.enabled = false
		s.state.endpoints = nil

		log.Info("piko proxy stopped")
	}()

	return nil
}

func (s service) Disconnect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.unsafeDisconnect()

	return nil
}

func (s service) unsafeDisconnect() {
	s.serveController.Cancel()

	log.Info("[PIKO] piko proxy called stopped")
	s.connectMutex.Lock()
	s.connectMutex.Unlock()

	log.Info("[PIKO] piko proxy stopped")
}

func (s service) State(ctx context.Context, req []pikoModel.Endpoint) (pikoModel.State, error) {
	log.Info("[piko state]: called")
	defer log.Info("[piko state]: stopped")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Info("[piko state]: started")

	// set default state
	result := pikoModel.State{
		Status:    pikoModel.Running,
		StatusMsg: "OK",
	}

	if !s.state.enabled {
		result.StatusMsg = "proxy not enabled"
		result.Status = pikoModel.Stopped
	}

	for _, requested := range req {
		exist := false

		for _, current := range s.state.endpoints {
			if requested.Settings.PortID == current.Settings.PortID {
				exist = true

				break
			}
		}

		if !exist {
			result.StatusMsg = fmt.Sprintf("endpoint for port [%s] not exist", requested.Port)
			result.Status = pikoModel.Error

			return result, nil
		}

		// disable availability check for endpoints
		switch requested.Protocol {
		case template.HTTP:
			if err := s.checkHTTPEndpoint(ctx, requested); err != nil {
				result.StatusMsg = err.Error()
				result.Status = pikoModel.Error

				return result, nil
			}
		}

		endpointInfo := pikoModel.EndpointInfo{
			Protocol: requested.Protocol,
			Endpoint: requested,
		}

		switch requested.Protocol {
		case template.HTTP:
			endpointInfo.Link = fmt.Sprintf(s.config.linkTemplate, requested.Settings.Name)

		case template.TCP:
			endpointInfo.Link = "use TCP proxy util"
		}

		result.Endpoints = append(result.Endpoints, endpointInfo)
	}

	if len(s.state.endpoints) != len(req) {
		result.StatusMsg = "have wrong count of endpoints"
		result.Status = pikoModel.Error

		return result, nil
	}

	return result, nil
}

// Info method is like mock now, because we use piko SDK now
func (s service) Info(ctx context.Context) (pikoModel.Info, error) {
	info := pikoModel.Info{
		Version:      s.config.version,
		Availability: true,
	}

	return info, nil
}
