package organization

import (
    "context"
    "github.com/google/uuid"
)

// Repository defines the interface for organization data access
type Repository interface {
    Create(ctx context.Context, org *Organization) error
    GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
    GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Organization, error)
    Update(ctx context.Context, org *Organization) error
    Delete(ctx context.Context, id uuid.UUID) error
    UpdateManager(ctx context.Context, orgID, managerID uuid.UUID) error
}
