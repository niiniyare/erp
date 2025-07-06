package errors

import "errors"

var (
	ErrSubdomainAlreadyExists = errors.New("subdomain already exists")
	ErrTenantNotFound         = errors.New("tenant not found")
)
