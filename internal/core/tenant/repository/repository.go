package repository

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"

	"github.com/google/uuid"
	"awo/internal/core/tenant/domain"
)

// Repository defines the interface for tenant data access.
type Repository interface {
	// CRUD
	Create(ctx context.Context, t *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.TenantFilter) ([]*domain.Tenant, int64, error)

	// Checks
	SubdomainExists(ctx context.Context, subdomain string) (bool, error)
	Exists(ctx context.Context, tenantID uuid.UUID) (bool, error)

	// Config & Usage
	CreateDefaultConfig(ctx context.Context, tenantID uuid.UUID) error
	GetConfig(ctx context.Context, tenantID uuid.UUID) (*domain.TenantConfiguration, error)
	GetUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error)
	InitUsage(ctx context.Context, tenantID uuid.UUID) error

	// Provisioning (calls DB function)
	Provision(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error)

	// Bulk
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error
	BulkSoftDelete(ctx context.Context, ids []uuid.UUID) error

	// Analytics
	GetGrowthStats(ctx context.Context, days int) ([]domain.GrowthStat, error)
	GetStatusDistribution(ctx context.Context) (*domain.StatusCount, error)

	// RLS context
	GetCurrentTenantID(ctx context.Context) (uuid.UUID, error)
	GetCurrentTenant(ctx context.Context) (*domain.Tenant, error)
	ValidateCurrentTenant(ctx context.Context) error
}
