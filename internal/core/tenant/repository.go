package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
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

// repository implements Repository interface
type repository struct {
	store db.Store
}

// NewRepository creates a new tenant repository
func NewRepository(store db.Store) Repository {
	return &repository{store: store}
}

// Create implements Repository.Create
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
	logger.DebugContext(ctx, "Creating tenant in database", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"subdomain":   tenant.Subdomain,
	})

	params := db.CreateTenantParams{
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Email:     tenant.Email,
		Subdomain: tenant.Subdomain,
		Status:    string(tenant.Status),
		Industry:  tenant.Industry,
	}

	_, err := r.store.CreateTenant(ctx, params)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create tenant in database", logger.Fields{
			"tenant_id": tenant.ID.String(),
			"error":     err.Error(),
		})
	}
	return err
}

// GetByID implements Repository.GetByID
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	sqlcTenant, err := r.store.GetTenantByID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return FromSQLCTenant(sqlcTenant)
}

// GetBySubdomain implements Repository.GetBySubdomain
func (r *repository) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	sqlcTenant, err := r.store.GetTenantByUUID(ctx, &subdomain)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	return FromSQLCTenant(sqlcTenant)
}

// Update implements Repository.Update
func (r *repository) Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error {
	var status *string
	if updates.Status != nil {
		statusStr := string(*updates.Status)
		status = &statusStr
	}

	params := db.UpdateTenantParams{
		Name:      updates.Name,
		Subdomain: updates.Subdomain,
		Status:    status,
		Industry:  updates.Industry,
		ID:        id,
	}

	_, err := r.store.UpdateTenant(ctx, params)
	return err
}

// Delete implements Repository.Delete
func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	// Use soft delete (bulk soft delete for single tenant)
	return r.store.BulkSoftDeleteTenants(ctx, []uuid.UUID{id})
}

// List implements Repository.List
func (r *repository) List(ctx context.Context, offset, limit int) ([]*Tenant, error) {
	// For now, get all active tenants since SQLC doesn't have a paginated list method
	// In production, you would add a paginated query to your SQL files
	sqlcTenants, err := r.store.GetActiveTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Apply manual pagination (ideally this should be done in SQL)
	start := offset
	end := offset + limit
	if start > len(sqlcTenants) {
		return []*Tenant{}, nil
	}
	if end > len(sqlcTenants) {
		end = len(sqlcTenants)
	}

	paginatedSQLCTenants := sqlcTenants[start:end]

	var tenants []*Tenant
	for _, sqlcTenant := range paginatedSQLCTenants {
		tenant, err := FromSQLCTenant(sqlcTenant)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// Exists implements Repository.Exists
func (r *repository) Exists(ctx context.Context, subdomain string) (bool, error) {
	// Try to get tenant by subdomain, if no error then it exists
	_, err := r.store.GetTenantByUUID(ctx, &subdomain)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}

	return true, nil
}
