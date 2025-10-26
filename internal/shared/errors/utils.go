package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ─── ERROR TYPE CHECKING HELPERS ─────────────────────────────────────────────
// These helpers maintain backward compatibility for error checking

// IsUserNotFound checks if error is user not found
func IsUserNotFound(err error) bool {
	return IsBusinessErrorCode(err, CodeUserNotFound)
}

// IsEntityNotFound checks if error is entity not found
func IsEntityNotFound(err error) bool {
	return IsBusinessErrorCode(err, CodeEntityNotFound)
}

// IsRoleNotFound checks if error is role not found
func IsRoleNotFound(err error) bool {
	return IsBusinessErrorCode(err, CodeRoleNotFound)
}

// IsTenantNotFound checks if error is tenant not found
func IsTenantNotFound(err error) bool {
	return IsBusinessErrorCode(err, CodeTenantNotFound)
}

// IsInvalidCredentials checks if error is invalid credentials
func IsInvalidCredentials(err error) bool {
	return IsBusinessErrorCode(err, CodeInvalidCredentials)
}

// IsUnauthorized checks if error is unauthorized
func IsUnauthorized(err error) bool {
	return IsBusinessErrorCode(err, CodeUnauthorized)
}

// IsForbidden checks if error is forbidden
func IsForbidden(err error) bool {
	return IsBusinessErrorCode(err, CodeForbidden)
}

// IsConflict checks if error is a conflict (exists) error
func IsConflict(err error) bool {
	conflictCodes := []string{
		CodeUserExists, CodeEntityNameExists, CodeEntityCodeExists,
		CodeRoleExists, CodeTenantExists, CodeEmailExists, CodeUsernameExists,
		CodeSubdomainExists,
	}

	for _, code := range conflictCodes {
		if IsBusinessErrorCode(err, code) {
			return true
		}
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if _, ok := err.(ValidationErrors); ok {
		return true
	}
	if _, ok := err.(ValidationError); ok {
		return true
	}
	if be, ok := err.(*BusinessError); ok {
		return be.Category == CategoryValidation
	}
	return false
}

// IsTemporaryError checks if error is temporary/retryable
func IsTemporaryError(err error) bool {
	return IsTemporary(err)
}

// ─── MIGRATION HELPERS ─────────────────────────────────────────────
// These help migrate from simple errors to enhanced errors

// WrapSimpleError wraps a simple error with enhanced context
func WrapSimpleError(simpleErr error, code string, httpStatus int) *BusinessError {
	return NewBusinessError(code, simpleErr.Error()).
		WithHTTPStatus(httpStatus).
		WithDetail("original_error", simpleErr.Error())
}

// UpgradeError upgrades a simple error to enhanced error if possible
func UpgradeError(err error) error {
	if err == nil {
		return nil
	}

	// Already enhanced
	if _, ok := err.(*BusinessError); ok {
		return err
	}
	if _, ok := err.(*RepositoryError); ok {
		return err
	}
	if _, ok := err.(ValidationErrors); ok {
		return err
	}

	// Map common simple errors to enhanced ones
	errMsg := err.Error()
	switch errMsg {
	case "user not found":
		return ErrUserNotFound
	case "entity not found":
		return ErrEntityNotFound
	case "role not found":
		return ErrRoleNotFound
	case "tenant not found":
		return ErrTenantNotFound
	case "invalid credentials":
		return ErrInvalidCredentials
	case "unauthorized":
		return ErrUnauthorized
	case "forbidden":
		return ErrForbidden
	default:
		// Generic upgrade
		return WrapSimpleError(err, CodeUnknownError, http.StatusInternalServerError)
	}
}

// ─── ERROR CATEGORIZATION HELPERS ─────────────────────────────────────────────

// GetErrorsByCategory returns all errors of a specific category from a collection
func GetErrorsByCategory(errs []error, category Category) []*BusinessError {
	var businessErrors []*BusinessError
	for _, err := range errs {
		if be, ok := err.(*BusinessError); ok && be.Category == category {
			businessErrors = append(businessErrors, be)
		}
	}
	return businessErrors
}

// HasCriticalErrors checks if any errors have critical severity
func HasCriticalErrors(errs []error) bool {
	for _, err := range errs {
		if be, ok := err.(*BusinessError); ok && be.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// GetMaxSeverity returns the highest severity level from a collection of errors
func GetMaxSeverity(errs []error) Severity {
	maxSeverity := SeverityInfo

	for _, err := range errs {
		if be, ok := err.(*BusinessError); ok {
			switch be.Severity {
			case SeverityCritical:
				return SeverityCritical // Critical is highest, return immediately
			case SeverityError:
				if maxSeverity != SeverityCritical {
					maxSeverity = SeverityError
				}
			case SeverityWarning:
				if maxSeverity == SeverityInfo {
					maxSeverity = SeverityWarning
				}
			}
		}
	}

	return maxSeverity
}

// ─── ERROR FILTERING HELPERS ─────────────────────────────────────────────

// FilterRetryableErrors returns only retryable errors from a collection
func FilterRetryableErrors(errs []error) []error {
	var retryable []error
	for _, err := range errs {
		if IsTemporary(err) {
			retryable = append(retryable, err)
		}
	}
	return retryable
}

// FilterNonRetryableErrors returns only non-retryable errors from a collection
func FilterNonRetryableErrors(errs []error) []error {
	var nonRetryable []error
	for _, err := range errs {
		if !IsTemporary(err) {
			nonRetryable = append(nonRetryable, err)
		}
	}
	return nonRetryable
}

// ─── ERROR FORMATTING HELPERS ─────────────────────────────────────────────

// FormatErrorCodes returns a comma-separated list of error codes
func FormatErrorCodes(errs []error) string {
	var codes []string
	seen := make(map[string]bool)

	for _, err := range errs {
		code := GetErrorCode(err)
		if !seen[code] {
			codes = append(codes, code)
			seen[code] = true
		}
	}

	return strings.Join(codes, ", ")
}

// FormatErrorMessages returns a formatted string of error messages
func FormatErrorMessages(errs []error, separator string) string {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, separator)
}

// ─── BUSINESS LOGIC HELPERS ─────────────────────────────────────────────

// IsNotFoundError checks if error represents any type of "not found" condition
func IsNotFoundError(err error) bool {
	notFoundCodes := []string{
		CodeUserNotFound, CodeEntityNotFound, CodeRoleNotFound,
		CodeTenantNotFound, CodeInvitationNotFound, CodeAPIKeyNotFound,
		CodePolicyNotFound, CodeAttributeNotFound, CodeNotFound,
	}

	for _, code := range notFoundCodes {
		if IsBusinessErrorCode(err, code) {
			return true
		}
	}
	return false
}

// IsAuthenticationError checks if error is authentication-related
func IsAuthenticationError(err error) bool {
	authCodes := []string{
		CodeInvalidCredentials, CodeAuthenticationFailed,
		CodeUnauthorized, CodeForbidden, CodeAccountLocked,
	}

	for _, code := range authCodes {
		if IsBusinessErrorCode(err, code) {
			return true
		}
	}
	return false
}

// IsSecurityError checks if error is security-related
func IsSecurityError(err error) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Category == CategorySecurity
	}
	return false
}

// IsTenantError checks if error is tenant-related
func IsTenantError(err error) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Category == CategoryTenant
	}
	return false
}

// ─── DEBUGGING HELPERS ─────────────────────────────────────────────

// GetErrorSummary returns a structured summary of an error for debugging
func GetErrorSummary(err error) map[string]any {
	summary := map[string]any{
		"type":    fmt.Sprintf("%T", err),
		"message": err.Error(),
		"code":    GetErrorCode(err),
		"status":  GetHTTPStatus(err),
	}

	switch e := err.(type) {
	case *BusinessError:
		summary["category"] = e.Category
		summary["severity"] = e.Severity
		summary["retryable"] = e.Retryable
		summary["tenant_id"] = e.TenantID
		summary["user_id"] = e.UserID
		summary["details"] = e.Details
		summary["suggestions"] = e.Suggestions

	case *RepositoryError:
		summary["operation"] = e.Operation
		summary["table"] = e.Table
		summary["tenant_id"] = e.TenantID
		summary["details"] = e.Details

	case ValidationErrors:
		summary["field_count"] = len(e)
		summary["fields"] = e.ToMap()

	case ValidationError:
		summary["field"] = e.Field
		summary["value"] = e.Value
	}

	return summary
}

// GetErrorChain returns the chain of wrapped errors
func GetErrorChain(err error) []string {
	var chain []string
	current := err

	for current != nil {
		chain = append(chain, current.Error())
		current = errors.Unwrap(current)
	}

	return chain
}
