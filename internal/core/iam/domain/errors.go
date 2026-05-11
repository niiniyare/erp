package domain

import "fmt"

// Domain Error Type

// Error is the IAM domain error type.
// It intentionally does not import shared/errors so the domain stays
// fully self-contained — only stdlib dependencies.
type Error struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return fmt.Sprintf("[iam] %s: %s", e.Code, e.Message)
}

// Authorization Sentinels

var (
	ErrForbidden           = &Error{"AUTHZ_FORBIDDEN", "access denied", 403}
	ErrUnauthorized        = &Error{"AUTHZ_UNAUTHORIZED", "authentication required", 401}
	ErrInvalidRequest      = &Error{"AUTHZ_INVALID", "subject/domain/obj/act required", 400}
	ErrPolicyConflict      = &Error{"AUTHZ_DUPLICATE", "policy already exists", 409}
	ErrPolicyLimitExceeded = &Error{"AUTHZ_POLICY_LIMIT", "policy limit per domain exceeded", 429}
)

// Identity Validation Errors

// ErrInvalidIdentity returns a 400 domain error for invalid identity inputs.
// Used by CreateUserRequest.Validate() and similar domain-level checks.
func ErrInvalidIdentity(msg string) *Error {
	return &Error{
		Code:       "IDENTITY_INVALID",
		Message:    msg,
		HTTPStatus: 400,
	}
}
