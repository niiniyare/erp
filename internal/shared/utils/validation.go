package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/errors"
)

// validate holds the singleton validator instance.
var validate *validator.Validate

func init() {
	validate = validator.New()
	// Register a custom validation function for UUIDs
	_ = validate.RegisterValidation("uuid", validateUUID)
	// Customize how the 'json' tag is used for field names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct based on 'validate' tags.
// It returns a ValidationErrors collection, which is nil if there are no errors.
func ValidateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors errors.ValidationErrors
	// Cast the error to validator.ValidationErrors
	for _, err := range err.(validator.ValidationErrors) {
		// Create a new validation error
		verr := errors.ValidationError{
			Field:   err.Field(),
			Message: formatValidationMessage(err),
			Code:    err.Tag(),
			Value:   err.Value(),
		}
		validationErrors = append(validationErrors, verr)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

// validateUUID is a custom validation function for UUIDs.
func validateUUID(fl validator.FieldLevel) bool {
	// Allows for UUIDs to be optional. Use 'required,uuid' for mandatory UUIDs.
	if fl.Field().Kind() == reflect.String {
		// Handle string UUIDs if necessary, though we primarily use uuid.UUID
		return true
	}
	// Check if the field is a valid, non-nil uuid.UUID
	if u, ok := fl.Field().Interface().(uuid.UUID); ok {
		return u != uuid.Nil
	}
	if u, ok := fl.Field().Interface().(*uuid.UUID); ok {
		return u != nil && *u != uuid.Nil
	}
	return true // Pass if not a UUID type, other validators can handle it
}

// formatValidationMessage creates a user-friendly error message.
func formatValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required.", err.Field())
	case "email":
		return fmt.Sprintf("The %s field must be a valid email address.", err.Field())
	case "min":
		return fmt.Sprintf("The %s field must be at least %s.", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("The %s field may not be greater than %s.", err.Field(), err.Param())
	case "uuid":
		return fmt.Sprintf("The %s field must be a valid UUID.", err.Field())
	default:
		return fmt.Sprintf("The %s field is invalid.", err.Field())
	}
}
