package tenant

import (
    "context"
    "github.com/google/uuid"
)

// Repository defines the interface for tenant data access
type Repository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, offset, limit int) ([]*Tenant, error)
    Exists(ctx context.Context, subdomain string) (bool, error)
}