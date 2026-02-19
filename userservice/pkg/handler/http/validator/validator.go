package validator

import (
	"errors"
	"reflect"
	"strings"

	baseValidator "github.com/Interpuls/ifc2-service-farm/pkg/validator"
)

type Validator struct {
	validator *baseValidator.Validator
}

func New(v *baseValidator.Validator) Validator {
	s := Validator{
		validator: v,
	}
	
	return s
}

func (v *Validator) ValidateStruct(s any) error {
	err := v.validator.ValidateStruct(s)

	var validationErr *baseValidator.ValidationError

	if errors.As(err, &validationErr) {
		newValidationErr := baseValidator.NewValidationErr()

		for fieldName, fieldErrs := range validationErr.Errors {
			jsonFieldTag := v.getJSONFieldTag(s, fieldName)

			for _, fieldErr := range fieldErrs {
				newValidationErr.AddError(jsonFieldTag, fieldErr)
			}
		}

		return newValidationErr
	}

	return err
}

func (v *Validator) getJSONFieldTag(structValue any, fieldName string) string {
	val := reflect.ValueOf(structValue)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fieldName
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Name == fieldName {
			jsonTag := field.Tag.Get("json")

			if jsonTag != "" {
				jsonFieldName := strings.Split(jsonTag, ",")[0]

				if jsonFieldName != "" && jsonFieldName != "-" {
					return jsonFieldName
				}
			}

			break
		}
	}

	return fieldName
}
