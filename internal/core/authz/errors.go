package authz

import "fmt"

// Error is the authz package's self-contained error type.
// It intentionally mirrors the project's BusinessError pattern without
// importing shared/errors so the package stays fully self-contained.
type Error struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return fmt.Sprintf("[authz] %s: %s", e.Code, e.Message)
}

// Sentinel errors returned by Service methods.
var (
	ErrForbidden      = &Error{"AUTHZ_FORBIDDEN", "access denied", 403}
	ErrUnauthorized   = &Error{"AUTHZ_UNAUTHORIZED", "authentication required", 401}
	ErrInvalidRequest = &Error{"AUTHZ_INVALID", "subject/domain/obj/act required", 400}
	ErrPolicyConflict = &Error{"AUTHZ_DUPLICATE", "policy already exists", 409}
)
