package errs

import "errors"

var (
	ErrDomain                 = errors.New("domain error")
	ErrDomainValidationFailed = errors.New("validation errors in request data")
	ErrInvalidRequest         = errors.New("malformed request or invalid input")
	ErrUnauthorized           = errors.New("authentication required or failed")
	ErrForbidden              = errors.New("access denied - insufficient permissions")
	ErrResourceNotFound       = errors.New("requested resource does not exist")
	ErrRateLimitExceeded      = errors.New("too many requests - rate limit exceeded")
	ErrInternalError          = errors.New("unexpected internal error")
	ErrServiceUnavailable     = errors.New("service temporarily unavailable")
	ErrInsufficientBalance    = errors.New("insufficient balance")
)
