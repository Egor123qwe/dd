package safeMap

import "sync"

type Map[K comparable, V ~chan []byte] struct {
	requests map[K]V
	m        *sync.Mutex
}

func New[K comparable, V ~chan []byte]() *Map[K, V] {
	return &Map[K, V]{
		requests: make(map[K]V),
		m:        &sync.Mutex{},
	}
}

func (m *Map[K, V]) Add(id K, ch V) {
	m.m.Lock()
	defer m.m.Unlock()

	m.requests[id] = ch
}

func (m *Map[K, V]) Delete(id K) {
	m.m.Lock()
	defer m.m.Unlock()

	delete(m.requests, id)
}

func (m *Map[K, V]) Get(id K) (V, bool) {
	m.m.Lock()
	defer m.m.Unlock()

	ch, ok := m.requests[id]
	return ch, ok
}
