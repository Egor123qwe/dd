package resp_processor

import (
	"errors"
	"net/http"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

func resolveError(err error) (messages []string, status int) {
	switch {
	case errors.Is(err, errs.ErrInternalError):
		return []string{err.Error()}, http.StatusInternalServerError

	case errors.Is(err, errs.ErrServiceUnavailable):
		return []string{err.Error()}, http.StatusServiceUnavailable

	case errors.Is(err, errs.ErrDomain):
		return []string{err.Error()}, http.StatusUnprocessableEntity

	case errors.Is(err, errs.ErrDomainValidationFailed):
		return []string{err.Error()}, http.StatusBadRequest

	case errors.Is(err, errs.ErrInvalidRequest):
		return []string{err.Error()}, http.StatusBadRequest

	case errors.Is(err, errs.ErrUnauthorized):
		return []string{err.Error()}, http.StatusUnauthorized

	case errors.Is(err, errs.ErrForbidden):
		return []string{err.Error()}, http.StatusForbidden

	case errors.Is(err, errs.ErrResourceNotFound):
		return []string{err.Error()}, http.StatusNotFound

	case errors.Is(err, errs.ErrInsufficientBalance):
		return []string{err.Error()}, http.StatusBadRequest

	case errors.Is(err, errs.ErrRateLimitExceeded):
		return []string{err.Error()}, http.StatusTooManyRequests

	default:
		return []string{errs.ErrInternalError.Error()}, http.StatusInternalServerError
	}
}
