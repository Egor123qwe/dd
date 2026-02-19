package event

import (
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
)

type Event int32

const (
	ResumeSharing Event = iota
	StopSharing
)

type StopSharingData struct {
	LastConfiguration runtime.Configuration
	Exp               time.Time
}
