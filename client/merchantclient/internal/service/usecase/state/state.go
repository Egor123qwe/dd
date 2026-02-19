package state

import (
	"sync"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("state")

type Status int32

const (
	Disabled Status = iota
	Configuring
	Ready
	InRent
)

var statusName = []string{"Disabled", "Configuring", "Ready", "InRent"}

func (s Status) String() string { return statusName[s] }

type State interface {
	SetStatus(status Status)
	GetStatus() Status

	SetSessionID(sessionID string)
	GetSessionID() string

	SetRequestID(sessionID string)
	GetRequestID() string

	SetRentStartedAt(t *time.Time)
	GetRentStartedAt() *time.Time

	SetTotalPrice(price float64)
	GetTotalPrice() float64

	Mutex() *sync.Mutex
	Reset()
}

type state struct {
	status        Status
	sessionID     string
	requestID     string
	rentStartedAt *time.Time
	totalPrice    float64

	usecaseMutex *sync.Mutex
	m            *sync.Mutex
}

func New() State {
	return &state{
		status:    Disabled,
		sessionID: "",

		usecaseMutex: &sync.Mutex{},
		m:            &sync.Mutex{},
	}
}

func (s *state) Mutex() *sync.Mutex {
	return s.usecaseMutex
}

func (s *state) Reset() {
	s.SetStatus(Disabled)
	s.SetSessionID("")
	s.SetRequestID("")
	s.SetRentStartedAt(nil)
	s.SetTotalPrice(0)
}

func (s *state) SetTotalPrice(price float64) {
	s.m.Lock()
	defer s.m.Unlock()
	s.totalPrice = price
}

func (s *state) GetTotalPrice() float64 {
	s.m.Lock()
	defer s.m.Unlock()
	return s.totalPrice
}

func (s *state) SetStatus(status Status) {
	s.m.Lock()
	defer s.m.Unlock()

	log.Infof("status updated: %s", status)
	s.status = status
}

func (s *state) GetStatus() Status {
	s.m.Lock()
	defer s.m.Unlock()
	return s.status
}

func (s *state) SetSessionID(sessionID string) {
	s.m.Lock()
	defer s.m.Unlock()

	log.Infof("merchant sessionID updated: \"%s\"", sessionID)
	s.sessionID = sessionID
}

func (s *state) GetSessionID() string {
	s.m.Lock()
	defer s.m.Unlock()
	return s.sessionID
}

func (s *state) SetRequestID(requestID string) {
	s.m.Lock()
	defer s.m.Unlock()

	log.Infof("merchant requestID updated: \"%s\"", requestID)
	s.requestID = requestID
}

func (s *state) GetRequestID() string {
	s.m.Lock()
	defer s.m.Unlock()
	return s.requestID
}

func (s *state) SetRentStartedAt(t *time.Time) {
	s.m.Lock()
	defer s.m.Unlock()
	s.rentStartedAt = t
}

func (s *state) GetRentStartedAt() *time.Time {
	s.m.Lock()
	defer s.m.Unlock()
	return s.rentStartedAt
}
