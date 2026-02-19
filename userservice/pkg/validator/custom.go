package validator

import baseValidator "github.com/go-playground/validator/v10"

type CustomValidator struct {
	Name string
	Fn   func(fl baseValidator.FieldLevel) bool
}

func NewCustomValidator(name string, fn func(fl baseValidator.FieldLevel) bool) CustomValidator {
	cv := CustomValidator{
		Name: name,
		Fn:   fn,
	}

	return cv
}
