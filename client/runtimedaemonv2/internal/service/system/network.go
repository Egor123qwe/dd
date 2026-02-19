package system

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
)

const (
	loading = "loading..."
)

const (
	networkInfoUpdateTime = 10 * time.Second
)

type networkSrv interface {
	infoFromCache(ctx context.Context) hardware.Network
	info(ctx context.Context) hardware.Network
	networkInfo(ctx context.Context) (hardware.Network, error)
}

type network struct {
	state hardware.Network

	// running bool need to optimize network monitoring
	// and use network monitoring only if it needed
	running *atomic.Bool

	cacheLocker *sync.Mutex
	m           *sync.Mutex
}

func newNetwork() networkSrv {
	srv := &network{
		state: hardware.Network{
			Ping:     0,
			Download: 0,
			Upload:   0,
		},

		running: &atomic.Bool{},

		cacheLocker: &sync.Mutex{},
		m:           &sync.Mutex{},
	}

	// turn on network monitoring
	srv.running.Store(true)
	srv.cacheLocker.Lock()

	// start network checker
	go srv.networkChecker(context.Background())

	return srv
}

func (s *network) info(ctx context.Context) hardware.Network {
	// turn on network monitoring
	//s.running.Store(true)

	// used to wait available data from cache
	s.cacheLocker.Lock()
	s.cacheLocker.Unlock()

	s.m.Lock()
	defer s.m.Unlock()

	return s.state
}

func (s *network) infoFromCache(ctx context.Context) hardware.Network {
	// turn on network monitoring
	//s.running.Store(true)

	s.m.Lock()
	defer s.m.Unlock()

	return s.state
}

func (s *network) networkChecker(ctx context.Context) {
	for {
		if s.running.Load() {
			hardware, err := s.networkInfo(ctx)
			if err != nil {
				log.Infof("error while network infoFromCache upadate: %s", err)
			}

			s.m.Lock()
			s.state = hardware
			s.m.Unlock()

			s.cacheLocker.TryLock()
			s.cacheLocker.Unlock()
		}

		// stop network monitoring
		s.running.Store(false)

		select {
		case <-time.After(networkInfoUpdateTime):

		case <-ctx.Done():
			return
		}
	}
}

func (s *network) networkInfo(ctx context.Context) (hardware.Network, error) {
	speedtestClient := speedtest.New()

	serverList, err := speedtestClient.FetchServers()
	if err != nil {
		return hardware.Network{}, fmt.Errorf("failed to fetch servers: %v", err)
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil {
		return hardware.Network{}, fmt.Errorf("failed to find servers: %v", err)
	}

	result := hardware.Network{}
	checksCount := 0

	for _, s := range targets {
		if s.Latency > 100*time.Millisecond || s.Distance > 500.0 {
			continue
		}

		if err := s.PingTestContext(ctx, nil); err != nil {
			return hardware.Network{}, fmt.Errorf("ping test error: %v", err)
		}

		if err := s.DownloadTestContext(ctx); err != nil {
			return hardware.Network{}, fmt.Errorf("download test error: %v", err)
		}

		if err := s.UploadTestContext(ctx); err != nil {
			return hardware.Network{}, fmt.Errorf("upload test error: %v", err)
		}

		checksCount++
		result.Ping += float64(s.Latency.Milliseconds())
		result.Download += s.DLSpeed.Mbps()
		result.Upload += s.ULSpeed.Mbps()
	}

	if checksCount == 0 {
		return hardware.Network{}, fmt.Errorf("no valid test results")
	}

	result.Ping /= float64(checksCount)
	result.Download /= float64(checksCount)
	result.Upload /= float64(checksCount)

	return result, nil
}
