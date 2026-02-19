package system

import (
	"sync"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
)

type storageCache struct {
	Types []hardware.DiskType
}

type cache struct {
	storage *storageCache
	m       *sync.Mutex
}

func (s *cache) setStorageInfo(storage storageCache) {
	s.m.Lock()
	s.storage = &storage
	s.m.Unlock()
}

func (s *cache) getStorageInfo() *storageCache {
	s.m.Lock()
	defer s.m.Unlock()

	return s.storage
}
