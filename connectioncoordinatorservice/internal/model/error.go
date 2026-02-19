package model

import (
	"errors"
)

var (
	ErrWSConn = errors.New("failed to create web socket connection")

	ErrMissingAuth = errors.New("missing Authorization header")
	ErrInvalidAuth = errors.New("invalid Bearer token format")
	ErrBadClaims   = errors.New("bad token")

	ErrFailedCreateRequest    = errors.New("failed create request to validate token")
	ErrFailedSendRequest      = errors.New("failed send request to validate token")
	ErrFailedToCreateResponse = errors.New("failed to create response")

	ErrInternalServerError = errors.New("internal server error")

	ErrDestinationNotFound = errors.New("destination id not found")
)
