package custom

import (
	"github.com/go-playground/validator/v10"
	"unicode"
)

func ValidatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
