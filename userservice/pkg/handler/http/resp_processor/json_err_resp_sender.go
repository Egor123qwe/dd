package resp_processor

import (
	"errors"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/Interpuls/ifc2-service-farm/pkg/validator"
	"github.com/gin-gonic/gin"
)

func JsonErrRespSender(c *gin.Context, messages []string, err error) {
	result := Response{
		Messages:   messages,
		Validation: make(map[string][]string),
	}

	// for debug
	result.Messages = append(result.Messages, err.Error())

	var validationErr *validator.ValidationError

	if errors.As(err, &validationErr) {
		err = fmt.Errorf("%w: %w", errs.ErrDomainValidationFailed, err)

		result.Validation = validationErr.Errors
	}

	_, status := resolveError(err)
	logError(err)

	c.JSON(status, result)
}

func logError(err error) {
	switch {
	case errors.Is(err, errs.ErrInternalError),
		errors.Is(err, errs.ErrServiceUnavailable),
		errors.Is(err, errs.ErrDomain):

		log := logger.Get()
		log.Error().Err(err).Msg("handler caught an error")
	}
}
