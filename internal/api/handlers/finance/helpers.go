package finance

// helpers.go — small utilities shared across finance handler files.
//
// Rules for what belongs here:
//   - Used by two or more handler files
//   - Pure functions with no side effects
//   - Not business logic (that lives in the service layer)

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	sharedErrors "awo.so/internal/shared/errors"
)

// ============================================================================
// Decimal conversion
// ============================================================================

// decimalFromString converts a string amount to decimal.Decimal.
//
// Use this for all monetary values that arrive as JSON strings.
// Never use decimalFromFloat — float64 cannot represent all decimal fractions
// exactly (e.g. 0.1 + 0.2 ≠ 0.3 in IEEE 754), which causes silent rounding
// errors in financial calculations.
//
// Request structs should declare monetary fields as string:
//
//	Amount string `json:"amount" validate:"required,decimal"`
//
// Then call decimalFromString(req.Amount) in the handler.
func decimalFromString(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, sharedErrors.NewBusinessError("INVALID_AMOUNT", "amount must be a valid decimal number").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("value", s).
			WithSuggestion(`Use a string representation of the decimal, e.g. "123.45"`)
	}
	return d, nil
}

// decimalFromFloat converts float64 → decimal.Decimal using the string path
// to avoid IEEE 754 representation errors.
//
// Only use this when the source value is already a float64 with no further
// decimal precision (e.g. integer-valued amounts from legacy fields).
// Prefer decimalFromString for all new code.
func decimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// ============================================================================
// UUID helpers
// ============================================================================

// parseUUID parses a string into a uuid.UUID and returns a ready-to-use
// BusinessError when the format is invalid.
//
// Centralising this removes the repeated 4-line parse+error block that
// previously appeared in every handler.
//
// Example:
//
//	id, err := parseUUID(c.Params("id"), "account ID")
//	if err != nil {
//	    return h.fail(c, err)
//	}
func parseUUID(s, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, sharedErrors.NewBusinessError("INVALID_UUID", "invalid "+fieldName+" format").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("field", fieldName).
			WithDetail("value", s).
			WithSuggestion("Provide a valid UUID v4, e.g. 550e8400-e29b-41d4-a716-446655440000")
	}
	return id, nil
}

// ============================================================================
// Context value extraction
// ============================================================================

// extractUserID reads the user ID that auth middleware stored in Locals as a
// string, then parses it to uuid.UUID.
//
// Auth middleware stores user_id as string (consistent with all other string
// Locals). A type assertion to uuid.UUID would silently yield uuid.Nil when
// the stored type is string — this helper makes the conversion explicit and
// safe.
//
// Returns uuid.Nil when the local is absent (unauthenticated path) without
// panicking. Callers that require a valid user ID should check for uuid.Nil.
func extractUserID(c *fiber.Ctx) uuid.UUID {
	s, ok := c.Locals("user_id").(string)
	if !ok || s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// ============================================================================
// Validation error constructor
// ============================================================================

// newValidationError builds a 400 BusinessError for invalid field values.
//
// Use this when a handler detects an invalid enum/code value after binding,
// e.g. an unrecognised rejection reason or period status.
//
//	return newValidationError("INVALID_FOO", "foo '"+v+"' is not valid", "foo", "Valid values: a, b, c")
func newValidationError(code, message, field, suggestion string) error {
	return sharedErrors.NewBusinessError(code, message).
		WithHTTPStatus(fiber.StatusBadRequest).
		WithCategory(sharedErrors.CategoryValidation).
		WithDetail("field", field).
		WithSuggestion(suggestion)
}

// ============================================================================
// Custom validator rules
// ============================================================================

// registerCustomValidators adds AWO-specific validation tags to the shared
// validator instance. Called once in NewFinanceHandler.
//
// Available tags after registration:
//
//	`validate:"decimal"`  — value must be parseable as a decimal number
func registerCustomValidators(v *validator.Validate) {
	// "decimal" tag: validates that a string field can be parsed as a decimal.
	// Use on all monetary amount fields in request structs.
	_ = v.RegisterValidation("decimal", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if s == "" {
			return true // let "required" handle empty check
		}
		_, err := decimal.NewFromString(s)
		return err == nil
	})
}
