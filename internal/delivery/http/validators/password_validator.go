package validators

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func PasswordStrengthValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func NoSpacesValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`\s`).MatchString(fl.Field().String())
}

func RegisterValidators(v *validator.Validate) {
	v.RegisterValidation("password_strength", PasswordStrengthValidator)
	v.RegisterValidation("no_spaces", NoSpacesValidator)
}
