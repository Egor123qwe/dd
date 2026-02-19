package model

import (
	"errors"
)

type SessionStatus string

const (
	StatusReady SessionStatus = "ready"
)

const (
	TypeTariff  = 3
	TypeGPU     = 4
	TypeSession = 5
	TypeGPUDict = 6

	TierCommunity = "community"
	TierCloud     = "cloud"
)

var (
	ErrCatsNotFound     = errors.New("Categories not found")
	ErrGPUsNotFound     = errors.New("GPUs not found")
	ErrSessionsNotFound = errors.New("Sessions not found")
	ErrGPUNamesnotFound = errors.New("GPU Names not found")
)

type FilterRepo struct {
	Type uint
	ID   string
}

type GPUDictFilter struct {
	VramMax int
	VramMin int
}
