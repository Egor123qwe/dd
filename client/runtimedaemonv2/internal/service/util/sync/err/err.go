package err

import "sync"

type Err interface {
	Set(err error)
	Get() error
}

func New() Err {
	return &err{
		Err:   nil,
		mutex: &sync.RWMutex{},
	}
}

type err struct {
	Err   error
	mutex *sync.RWMutex
}

func (l *err) Set(err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.Err = err
}

func (l *err) Get() error {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.Err
}
