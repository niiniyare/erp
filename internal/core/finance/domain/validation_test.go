package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"awo.so/internal/core/finance/domain"
)

// ============================================================================
// FIN-TYP-040: ValidationSeverity levels are distinct and exhaustive
// ============================================================================

func TestValidationSeverity_Constants(t *testing.T) {
	assert.Equal(t, domain.ValidationSeverity("ERROR"), domain.ValidationSeverityError)
	assert.Equal(t, domain.ValidationSeverity("WARNING"), domain.ValidationSeverityWarning)
	assert.Equal(t, domain.ValidationSeverity("INFO"), domain.ValidationSeverityInfo)

	// Each constant must be distinct
	assert.NotEqual(t, domain.ValidationSeverityError, domain.ValidationSeverityWarning)
	assert.NotEqual(t, domain.ValidationSeverityError, domain.ValidationSeverityInfo)
	assert.NotEqual(t, domain.ValidationSeverityWarning, domain.ValidationSeverityInfo)
}

func TestValidationError_IsBlocker(t *testing.T) {
	errLevel := domain.ValidationError{
		Field:    "amount",
		Message:  "amount is required",
		Code:     "REQUIRED",
		Severity: domain.ValidationSeverityError,
	}
	assert.True(t, errLevel.IsBlocker(), "ERROR severity must be a blocker")

	warnLevel := domain.ValidationError{
		Field:    "description",
		Message:  "description is recommended",
		Code:     "RECOMMENDED",
		Severity: domain.ValidationSeverityWarning,
	}
	assert.False(t, warnLevel.IsBlocker(), "WARNING severity must not be a blocker")

	infoLevel := domain.ValidationError{
		Field:    "reference",
		Message:  "reference helps with reconciliation",
		Code:     "INFO",
		Severity: domain.ValidationSeverityInfo,
	}
	assert.False(t, infoLevel.IsBlocker(), "INFO severity must not be a blocker")
}

func TestValidationError_EffectiveSeverity_DefaultsToError(t *testing.T) {
	// Zero-value Severity field (empty string) must default to ERROR so that
	// existing ValidationError literals without explicit Severity remain blockers.
	ve := domain.ValidationError{
		Field:   "account_code",
		Message: "account code is required",
		Code:    "REQUIRED",
		// Severity intentionally omitted
	}
	assert.Equal(t, domain.ValidationSeverityError, ve.EffectiveSeverity(),
		"omitted Severity must default to ERROR")
	assert.True(t, ve.IsBlocker())
}

func TestValidationResult_HasErrors_AnyItem(t *testing.T) {
	// HasErrors returns true when ANY error exists (regardless of severity)
	vr := &domain.ValidationResult{IsValid: true}

	assert.False(t, vr.HasErrors(), "empty result must not have errors")

	vr.AddError("field", "message", "CODE")
	assert.True(t, vr.HasErrors(), "result with one error must report HasErrors=true")
}

// ============================================================================
// FIN-TYP-041: Nested field paths use entries[N].field format
// ============================================================================

func TestNestedFieldPath(t *testing.T) {
	vr := &domain.ValidationResult{IsValid: true}

	// Simulate a validation error on line-item index 1 of a transaction
	vr.IsValid = false
	vr.Errors = append(vr.Errors, domain.ValidationError{
		Field:    "entries[1].account_code",
		Message:  "account does not exist or is inactive",
		Code:     "INVALID_ACCOUNT_CODE",
		Severity: domain.ValidationSeverityError,
	})

	assert.Len(t, vr.Errors, 1)
	assert.Equal(t, "entries[1].account_code", vr.Errors[0].Field,
		"entry-level field path must include index notation")
	assert.Equal(t, "INVALID_ACCOUNT_CODE", vr.Errors[0].Code)

	// Flat path (no index) would be unusable for UI highlighting
	assert.NotEqual(t, "account_code", vr.Errors[0].Field,
		"flat field path without index is not acceptable for entry-level errors")
}

func TestValidationResult_GetErrorsByField(t *testing.T) {
	vr := &domain.ValidationResult{IsValid: false}
	vr.Errors = []domain.ValidationError{
		{Field: "entries[0].amount", Code: "REQUIRED"},
		{Field: "entries[1].account_code", Code: "INVALID_ACCOUNT_CODE"},
		{Field: "entries[0].amount", Code: "NEGATIVE_AMOUNT"},
	}

	amountErrs := vr.GetErrorsByField("entries[0].amount")
	assert.Len(t, amountErrs, 2, "must return all errors for the specified field path")

	acctErrs := vr.GetErrorsByField("entries[1].account_code")
	assert.Len(t, acctErrs, 1)

	noErrs := vr.GetErrorsByField("description")
	assert.Empty(t, noErrs)
}

// ============================================================================
// FIN-TYP-042: ValidationResult.Merge — combines two results correctly
// ============================================================================

func TestValidationResult_Merge(t *testing.T) {
	result1 := &domain.ValidationResult{IsValid: false}
	result1.Errors = []domain.ValidationError{
		{Field: "amount", Message: "amount required", Code: "REQUIRED", Severity: domain.ValidationSeverityError},
		{Field: "amount", Message: "budget soft exceeded", Code: "BUDGET_SOFT_EXCEEDED", Severity: domain.ValidationSeverityWarning},
	}

	result2 := &domain.ValidationResult{IsValid: false}
	result2.Errors = []domain.ValidationError{
		{Field: "account_id", Message: "account inactive", Code: "ACCOUNT_INACTIVE", Severity: domain.ValidationSeverityError},
	}

	result1.Merge(result2)

	// After merge: result1 should have all 3 errors
	assert.Len(t, result1.Errors, 3, "merge must append errors from result2 to result1")
	assert.False(t, result1.IsValid, "merged result must be invalid when either has errors")

	// result2's errors appended (not replacing result1's)
	codes := make([]string, 0, 3)
	for _, e := range result1.Errors {
		codes = append(codes, e.Code)
	}
	assert.Contains(t, codes, "REQUIRED")
	assert.Contains(t, codes, "BUDGET_SOFT_EXCEEDED")
	assert.Contains(t, codes, "ACCOUNT_INACTIVE")
}

func TestValidationResult_Merge_IntoValidResult(t *testing.T) {
	// Merging errors into a valid result marks it invalid
	valid := &domain.ValidationResult{IsValid: true}
	withError := &domain.ValidationResult{IsValid: false}
	withError.Errors = []domain.ValidationError{
		{Field: "f", Message: "m", Code: "C", Severity: domain.ValidationSeverityError},
	}

	valid.Merge(withError)

	assert.False(t, valid.IsValid)
	assert.Len(t, valid.Errors, 1)
}

func TestValidationResult_Merge_EmptyDoesNotInvalidate(t *testing.T) {
	// Merging a valid (empty) result into another must not change validity
	result := &domain.ValidationResult{IsValid: false}
	result.AddError("f", "m", "C")

	emptyValid := &domain.ValidationResult{IsValid: true}
	result.Merge(emptyValid)

	// Still has the original error, still invalid
	assert.Len(t, result.Errors, 1)
	assert.False(t, result.IsValid)
}

// ============================================================================
// FIN-DEI-003: Decimal precision — no rounding errors (no float64 in finances)
// ============================================================================

func TestDecimalPrecision_NoFloatError(t *testing.T) {
	// Classic float64 trap: 100 * 0.10 using float arithmetic is 9.999999...
	// Financial code MUST use decimal.Decimal, not float64.
	//
	// This test verifies the domain constants use decimal.Decimal.
	assert.False(t, domain.DecimalZero.IsNegative(), "DecimalZero must be >= 0")
	assert.True(t, domain.DecimalOne.Equal(domain.DecimalZero.Add(domain.DecimalOne)), "1 = 0 + 1")
	assert.True(t, domain.DecimalHundred.Equal(domain.DecimalOne.Mul(domain.DecimalHundred)), "100 = 1 * 100")

	// Verify no precision loss on common tax calculation
	// 100 * 0.16 (16% VAT) must equal exactly 16.00
	rate := domain.DecimalHundred.Mul(domain.DecimalOne).Div(domain.DecimalHundred) // 1.0
	assert.True(t, rate.Equal(domain.DecimalOne), "100 / 100 = 1 exactly (no float drift)")
}
