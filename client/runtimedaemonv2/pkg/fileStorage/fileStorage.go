package fileStorage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"
)

type Settings[T any] interface {
	Get() (T, error)
	Set(state T) error
}

type settings[T any] struct {
	path  string
	mutex *sync.RWMutex
}

func New[T any](path string) Settings[T] {
	return settings[T]{
		path:  path,
		mutex: &sync.RWMutex{},
	}
}

func (s settings[T]) Get() (T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result T

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return result, nil
		}

		return result, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s settings[T]) Set(state T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = os.WriteFile(s.path, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
