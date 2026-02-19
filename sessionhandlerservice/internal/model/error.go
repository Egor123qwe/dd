package model

import (
	"errors"
)

var (
	ErrInvalidContent = errors.New("invalid content")

	ErrFailedToStartSession = errors.New("failed to start session")
	ErrFailedToStopSession  = errors.New("failed to stop session")

	ErrMerchantAlreadyReserved = errors.New("merchant is already used by another user")
	ErrMerchantNotReady        = errors.New("merchant is not ready to rent")

	ErrFailedToGetRentSettings = errors.New("failed to get rent settings")
	ErrFailedToGetClientRents  = errors.New("failed to get client rents")
	ErrFailedToGetMerchantRent = errors.New("failed to get merchant rent")
)
