package validator

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Errors map[string][]string
}

func NewValidationErr() *ValidationError {
	return &ValidationError{
		Errors: make(map[string][]string),
	}
}

func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	var errorMessages []string

	for field, errors := range ve.Errors {
		if len(errors) == 1 {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, errors[0]))
		} else {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: [%s]", field, strings.Join(errors, ", ")))
		}
	}

	return fmt.Sprintf("validation failed: %s", strings.Join(errorMessages, "; "))
}

func (ve *ValidationError) AddError(field, message string) {
	ve.Errors[field] = append(ve.Errors[field], message)
}

func (ve *ValidationError) HasErrors() bool {
	return len(ve.Errors) > 0
}

func (ve *ValidationError) GetErrors(field string) []string {
	return ve.Errors[field]
}

func (ve *ValidationError) GetAllErrors() map[string][]string {
	return ve.Errors
}
