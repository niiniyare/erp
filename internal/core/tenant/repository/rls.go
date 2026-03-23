package repository

import (
	"context"

	"github.com/google/uuid"
	db "awo.so/db/sqlc"
)

// SetTenantContext sets the RLS tenant context on the database session.
func SetTenantContext(ctx context.Context, store db.Store, tenantID uuid.UUID) error {
	return store.SetTenantContext(ctx, tenantID)
}

// ResetTenantContext clears the RLS tenant context.
func ResetTenantContext(ctx context.Context, store db.Store) error {
	return store.ResetTenantContext(ctx)
}

// WithTenantContext executes fn inside a tenant-scoped transaction.
func WithTenantContext(ctx context.Context, store db.Store, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
	return store.WithTenant(ctx, tenantID, fn)
}
