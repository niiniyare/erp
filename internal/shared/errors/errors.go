// Package errors provides a comprehensive error handling system for the Awo ERP platform.
// This package maintains backward compatibility while organizing errors into logical files.
//
// Key Features:
// - Rich error types with HTTP status, suggestions, and context
// - Predefined business errors for common scenarios
// - Validation error collection and handling
// - Repository error wrapping for database operations
// - HTTP error conversion for API responses
// - Context-aware error construction
// - Backward compatible exports
//
// File Organization:
// - types.go: Core error types and interfaces
// - business.go: Predefined business errors and constructors
// - codes.go: Error code constants
// - context.go: Context helper functions
// - http.go: HTTP error handling
// - tenant.go: Tenant-specific errors
// - utils.go: Utility and checking functions
// - errors.go: Main export file (this file)

package errors

// ─── RE-EXPORTS FOR BACKWARD COMPATIBILITY ─────────────────────────────────────────────
// This file ensures that all existing imports continue to work exactly as before

// Types from types.go
// Core error types are automatically available from types.go

// Constants from types.go and codes.go
// All constants are automatically available from their respective files

// Predefined errors from business.go
// All predefined error variables are automatically available from business.go

// Functions from utils.go
// All utility functions are automatically available from utils.go

// Functions from context.go
// All context helper functions are automatically available from context.go

// Functions from http.go
// HTTP error handling functions are automatically available from http.go

// Functions from tenant.go (existing file)
// Tenant-specific functions are automatically available from tenant.go

// ─── PACKAGE DOCUMENTATION ─────────────────────────────────────────────

/*
Basic Usage:

	import "project/erp/internal/shared/errors"

	// Check for specific errors (backward compatible)
	if errors.IsUserNotFound(err) {
		// Handle user not found
	}

	// Use predefined errors (backward compatible)
	return errors.ErrUserNotFound

	// Create contextual errors
	return errors.NewUserNotFoundError(userID)

	// Create business errors with rich context
	return errors.NewBusinessError("CUSTOM_ERROR", "Custom message").
		WithHTTPStatus(http.StatusBadRequest).
		WithCategory(errors.CategoryBusiness).
		WithSuggestion("Try again later")

	// Handle validation errors
	var validationErrs errors.ValidationErrors
	validationErrs.Add("email", "Invalid email format")
	return validationErrs

	// Convert to HTTP response
	httpErr := errors.ToHTTPError(err)
	w.WriteHeader(httpErr.Status)
	json.NewEncoder(w).Encode(httpErr)

Error Categories:

- CategoryValidation: Input validation errors
- CategoryRepository: Database/storage errors
- CategoryBusiness: Business logic errors
- CategorySecurity: Authentication/authorization errors
- CategoryIntegration: External service errors
- CategorySystem: System/infrastructure errors
- CategoryTenant: Multi-tenant specific errors

Error Severity Levels:

- SeverityInfo: Informational
- SeverityWarning: Warning level
- SeverityError: Error level
- SeverityCritical: Critical system errors

Predefined Errors:

Tenant Errors:
- ErrTenantExists, ErrTenantNotFound, ErrSubdomainAlreadyExists
- ErrFeatureNotEnabled, ErrTenantIDNotInContext

User Errors:
- ErrUserNotFound, ErrUserExists, ErrEmailAlreadyExists
- ErrInvalidCredentials, ErrAccountLocked

Entity Errors:
- ErrEntityNotFound, ErrEntityNameExists, ErrCircularReference

Security Errors:
- ErrUnauthorized, ErrForbidden, ErrRoleNotFound

ABAC Errors:
- ErrPolicyNotFound, ErrAttributeInvalid, ErrEvaluationFailed

Migration Guide:

Your existing code continues to work unchanged:

	// ✅ All these patterns still work
	if err == errors.ErrUserNotFound { ... }
	if errors.Is(err, errors.ErrUserNotFound) { ... }
	if errors.IsUserNotFound(err) { ... }

New enhanced features are available:

	// 🆕 Rich error context
	err := errors.NewUserNotFoundError(userID).
		WithDetail("lookup_method", "email").
		WithSuggestion("Check the email address")

	// 🆕 HTTP integration
	httpErr := errors.ToHTTPError(err)
	// Returns structured HTTP response with status, suggestions, etc.

Testing:

	func TestErrors(t *testing.T) {
		err := errors.NewUserNotFoundError("123")
		assert.True(t, errors.IsUserNotFound(err))
		assert.Equal(t, http.StatusNotFound, errors.GetHTTPStatus(err))
	}
*/
