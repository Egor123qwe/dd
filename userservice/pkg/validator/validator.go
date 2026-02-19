package validator

import (
	"errors"

	baseValidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *baseValidator.Validate
}

func New(customValidators ...CustomValidator) *Validator {
	v := baseValidator.New()

	registerBasicCustomTypes(v)

	for _, customValidator := range customValidators {
		_ = v.RegisterValidation(customValidator.Name, customValidator.Fn)
	}

	s := &Validator{
		validator: v,
	}

	return s
}

func (v *Validator) ValidateStruct(s any) error {
	validationErr := NewValidationErr()

	// baseValidator.FieldLevel() - эта строка была ошибочной

	if err := v.validator.Struct(s); err != nil {
		var validationErrors baseValidator.ValidationErrors

		if errors.As(err, &validationErrors) {
			for _, fieldErr := range validationErrors {
				validationErr.AddError(fieldErr.Field(), fieldErr.Error())
			}
		}
	}

	if validationErr.HasErrors() {
		return validationErr
	}

	return nil
}
