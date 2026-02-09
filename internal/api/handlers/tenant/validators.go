package tenant

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// slugRegex is compiled once at package level.
var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// registerCustomValidators adds custom validation rules to the validator instance.
func registerCustomValidators(v *validator.Validate) {
	// "uuid" is already built-in in validator v10 — no need to re-register.

	// Validator for URL-friendly slugs
	if err := v.RegisterValidation("slug", validateSlug); err != nil {
		panic("failed to register slug validator: " + err.Error())
	}
}

func validateSlug(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	return slugRegex.MatchString(value)
}
