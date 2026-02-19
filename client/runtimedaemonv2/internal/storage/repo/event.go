package repo

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/event"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/fileStorage"
)

const (
	eventSharingStoragePath = "/sharing_event.json"
)

type EventRepo interface {
	Sharing() EventSharing
}

type EventSharing interface {
	Get() (*event.StopSharingData, error)
	Set(info *event.StopSharingData) error
}
type eventRepo struct {
	sharing fileStorage.Settings[*event.StopSharingData]
}

func NewEvent(path string) EventRepo {
	return eventRepo{
		sharing: fileStorage.New[*event.StopSharingData](path + eventSharingStoragePath),
	}
}

func (r eventRepo) Sharing() EventSharing {
	return r.sharing
}
