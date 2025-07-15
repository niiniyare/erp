package errors

import "errors"

var (
	// Tenant Errors
	ErrTenantExists           = errors.New("tenant already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserExists             = errors.New("user already exists")
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleExists             = errors.New("role already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden")
	ErrInvalidInput           = errors.New("invalid input")
	ErrInvitationNotFound     = errors.New("invitation not found")
	ErrInvitationExpired      = errors.New("invitation expired")
	ErrAPIKeyNotFound         = errors.New("api key not found")
	ErrFeatureNotEnabled      = errors.New("feature not enabled")
	ErrSubdomainAlreadyExists = errors.New("subdomain already exists")
	ErrTenantNotFound         = errors.New("tenant not found")
	ErrNotFound               = errors.New("not found")
	ErrTenantIDNotInContext   = errors.New("tenant id not in context")

	// Entity errors
	ErrEntityNotFound      = errors.New("entity not found")
	ErrEntityNameExists    = errors.New("entity name already exists")
	ErrEntityCodeExists    = errors.New("entity code already exists")
	ErrInvalidParentEntity = errors.New("invalid parent entity")
	ErrCircularReference   = errors.New("circular reference detected")
	ErrEntityHasChildren   = errors.New("entity has children")
	ErrInvalidEntityType   = errors.New("invalid entity type")

	//User errors
	ErrInvalidUserType      = errors.New("Invalid User Type")
	ErrInvalidAccountStatus = errors.New("Invalid Account Status")
	ErrAccountLocked        = errors.New("Account is tLocked")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)
