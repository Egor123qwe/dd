package validation

import (
	"github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/validation/custom"
	"github.com/Interpuls/ifc2-service-farm/pkg/validator"

	httpValidation "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/validator"
)

func New() httpValidation.Validator {
	v := validator.New(
		validator.NewCustomValidator("password", custom.ValidatePassword),
	)

	return httpValidation.New(v)
}
