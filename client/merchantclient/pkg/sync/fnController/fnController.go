package fnController

import (
	"context"
	"sync"
)

type Controller interface {
	Cancel()
	SetCancelFn(cancel context.CancelFunc)
}

type state struct {
	fnCancel context.CancelFunc
	m        *sync.Mutex
}

func New() Controller {
	return &state{
		m: &sync.Mutex{},
	}
}

func (s *state) Cancel() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.fnCancel == nil {
		return
	}

	s.fnCancel()
}

func (s *state) SetCancelFn(cancel context.CancelFunc) {
	s.m.Lock()
	defer s.m.Unlock()

	s.fnCancel = cancel
}
