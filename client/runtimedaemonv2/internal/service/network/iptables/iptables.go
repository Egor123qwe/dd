package iptables

import (
	"context"
	"runtime"
	"strings"
	"sync"

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
	config Config

	mutex *sync.Mutex
}

// used if OS != linux
type mockService struct{}

func New() Service {

	return mockService{}
}

func isIptablesAvailable() bool {
	return strings.HasPrefix(runtime.GOOS, availableOS)
}
