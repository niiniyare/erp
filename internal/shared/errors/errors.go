package errors

import "errors"

var (
	ErrSubdomainAlreadyExists = errors.New("subdomain already exists")
	ErrTenantNotFound         = errors.New("tenant not found")
	ErrNotFound               = errors.New("not found")
	ErrTenantIDNotInContext   = errors.New("tenant id not in context")
	
	// Entity errors
	ErrEntityNotFound         = errors.New("entity not found")
	ErrEntityNameExists       = errors.New("entity name already exists")
	ErrEntityCodeExists       = errors.New("entity code already exists")
	ErrInvalidParentEntity    = errors.New("invalid parent entity")
	ErrCircularReference      = errors.New("circular reference detected")
	ErrEntityHasChildren      = errors.New("entity has children")
	ErrInvalidEntityType      = errors.New("invalid entity type")
)
