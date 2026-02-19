package listener

import (
	"sync"
)

type connection struct {
	listener  connListener
	userID    string
	sessionID string
}

type connections interface {
	get(connectionID string) (connListener, bool)
	getAll(userID string) []connListener

	add(userID string, connectionID string, listener connListener)
	remove(id string)

	addSession(userID, connectionID, sessionID string)
}

type state struct {
	users       map[string][]string
	connections map[string]connection
	m           *sync.Mutex
}

func newState() connections {
	return state{
		users:       make(map[string][]string),
		connections: make(map[string]connection),
		m:           &sync.Mutex{},
	}
}

func (s state) add(userID string, connectionID string, listener connListener) {
	s.m.Lock()
	defer s.m.Unlock()

	s.users[userID] = append(s.users[userID], connectionID)

	s.connections[connectionID] = connection{
		listener: listener,
		userID:   userID,
	}
}

func (s state) remove(connectionID string) {
	s.m.Lock()
	defer s.m.Unlock()

	conn, ok := s.connections[connectionID]
	if !ok {
		return
	}

	delete(s.connections, connectionID)

	user, ok := s.users[conn.userID]
	if !ok {
		return
	}

	var userConnections []string

	for _, v := range user {
		if v != connectionID {
			userConnections = append(userConnections, v)
		}
	}

	if len(userConnections) == 0 {
		delete(s.users, conn.userID)

		return
	}

	s.users[conn.userID] = userConnections
}

func (s state) get(id string) (connListener, bool) {
	s.m.Lock()
	defer s.m.Unlock()

	conn, ok := s.connections[id]

	return conn.listener, ok
}

func (s state) getAll(userID string) []connListener {
	s.m.Lock()
	defer s.m.Unlock()

	userConnections, ok := s.users[userID]
	if !ok {
		return nil
	}

	var result []connListener

	for _, id := range userConnections {
		if conn, ok := s.connections[id]; ok {
			result = append(result, conn.listener)
		}
	}

	return result
}

func (s state) addSession(userID, connectionID, sessionID string) {
	s.m.Lock()
	defer s.m.Unlock()

	conn, ok := s.connections[connectionID]
	if !ok {
		return
	}

	conn.sessionID = sessionID

	s.connections[connectionID] = conn
}
