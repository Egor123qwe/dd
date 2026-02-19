package iptables

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"github.com/coreos/go-iptables/iptables"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("iptables", logger.DefaultWithSentry())

const (
	availableOS = "linux"
)

type Service interface {
	IsCorrect(ctx context.Context, mustBeConfigured bool) (bool, error)

	Set(ctx context.Context) error
	Discard(ctx context.Context) error
}

type service struct {
	iptablesAPI *iptables.IPTables
	config      Config

	mutex *sync.Mutex
}

// used if OS != linux
type mockService struct{}

func NewClient() (*iptables.IPTables, error) {
	if !isIptablesAvailable() {
		return nil, nil
	}

	return iptables.New()
}

func New(iptablesAPI *iptables.IPTables) Service {
	if !isIptablesAvailable() {
		return mockService{}
	}

	config, err := newConfig()
	if err != nil {
		log.Errorf("failed to get iptables config: %v", err)
	}

	srv := &service{
		iptablesAPI: iptablesAPI,
		config:      config,
		mutex:       &sync.Mutex{},
	}

	return srv
}

func isIptablesAvailable() bool {
	return strings.HasPrefix(runtime.GOOS, availableOS)
}
