package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/domain/tenant"
)

type TenantRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{
		db:      pool,
		queries: db.New(pool),
	}
}

func (r *TenantRepository) Create(ctx context.Context, tenantReq *tenant.CreateTenantRequest) (*tenant.Tenant, error) {
	params := db.CreateTenantParams{
		Name:      tenantReq.Name,
		Subdomain: tenantReq.Subdomain,
		Status:    string(tenantReq.Status),
		Industry:  tenantReq.Industry,
	}

	dbTenant, err := r.queries.CreateTenant(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) GetByID(ctx context.Context, id int32) (*tenant.Tenant, error) {
	dbTenant, err := r.queries.GetTenantByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*tenant.Tenant, error) {
	dbTenant, err := r.queries.GetTenantByUUID(ctx, uuid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant by UUID: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) GetBySubdomain(ctx context.Context, subdomain string) (*tenant.Tenant, error) {
	dbTenant, err := r.queries.GetTenantBySubdomain(ctx, &subdomain)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) Update(ctx context.Context, id int32, req *tenant.UpdateTenantRequest) (*tenant.Tenant, error) {
	params := db.UpdateTenantParams{
		ID:        id,
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Status:    (*string)(req.Status),
		Industry:  req.Industry,
	}

	dbTenant, err := r.queries.UpdateTenant(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) Delete(ctx context.Context, id int32) error {
	err := r.queries.SoftDeleteTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

func (r *TenantRepository) List(ctx context.Context, limit, offset int32) ([]*tenant.Tenant, error) {
	params := db.ListTenantsParams{
		Limit:  limit,
		Offset: offset,
	}

	dbTenants, err := r.queries.ListTenants(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	tenants := make([]*tenant.Tenant, len(dbTenants))
	for i, dbTenant := range dbTenants {
		tenants[i] = r.mapToTenant(&dbTenant)
	}

	return tenants, nil
}

func (r *TenantRepository) GetActiveTenants(ctx context.Context) ([]*tenant.Tenant, error) {
	dbTenants, err := r.queries.GetActiveTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tenants: %w", err)
	}

	tenants := make([]*tenant.Tenant, len(dbTenants))
	for i, dbTenant := range dbTenants {
		tenants[i] = r.mapToTenant(&dbTenant)
	}

	return tenants, nil
}

func (r *TenantRepository) UpdateStatus(ctx context.Context, id int32, status tenant.TenantStatus) (*tenant.Tenant, error) {
	params := db.UpdateTenantStatusParams{
		ID:     id,
		Status: string(status),
	}

	dbTenant, err := r.queries.UpdateTenantStatus(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to update tenant status: %w", err)
	}

	return r.mapToTenant(&dbTenant), nil
}

func (r *TenantRepository) CheckExists(ctx context.Context, id int32) (bool, error) {
	exists, err := r.queries.CheckTenantExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	return exists, nil
}

func (r *TenantRepository) CheckSubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	exists, err := r.queries.CheckSubdomainExists(ctx, &subdomain)
	if err != nil {
		return false, fmt.Errorf("failed to check subdomain existence: %w", err)
	}
	return exists, nil
}

func (r *TenantRepository) CheckNameExists(ctx context.Context, name string) (bool, error) {
	exists, err := r.queries.CheckTenantNameExists(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant name existence: %w", err)
	}
	return exists, nil
}

func (r *TenantRepository) SearchByName(ctx context.Context, name string, limit, offset int32) ([]*tenant.Tenant, error) {
	params := db.SearchTenantsByNameParams{
		Name:   &name,
		Limit:  limit,
		Offset: offset,
	}

	dbTenants, err := r.queries.SearchTenantsByName(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search tenants by name: %w", err)
	}

	tenants := make([]*tenant.Tenant, len(dbTenants))
	for i, dbTenant := range dbTenants {
		tenants[i] = r.mapToTenant(&dbTenant)
	}

	return tenants, nil
}

func (r *TenantRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountTenants(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	return count, nil
}

func (r *TenantRepository) GetStats(ctx context.Context) (*tenant.TenantStats, error) {
	stats, err := r.queries.GetTenantStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}

	return &tenant.TenantStats{
		TotalTenants:     stats.TotalTenants,
		ActiveTenants:    stats.ActiveTenants,
		SuspendedTenants: stats.SuspendedTenants,
		PendingTenants:   stats.PendingTenants,
	}, nil
}

func (r *TenantRepository) Filter(ctx context.Context, filter *tenant.TenantFilter) ([]*tenant.Tenant, error) {
	params := db.FilterTenantsParams{
		NameFilter:     filter.NameFilter,
		StatusFilter:   filter.StatusFilter,
		IndustryFilter: filter.IndustryFilter,
		SortBy:         filter.SortBy,
		LimitCount:     filter.Limit,
		OffsetCount:    filter.Offset,
	}

	dbTenants, err := r.queries.FilterTenants(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to filter tenants: %w", err)
	}

	tenants := make([]*tenant.Tenant, len(dbTenants))
	for i, dbTenant := range dbTenants {
		tenants[i] = r.mapToTenant(&dbTenant)
	}

	return tenants, nil
}

func (r *TenantRepository) BulkUpdateStatus(ctx context.Context, ids []int32, status tenant.TenantStatus) error {
	params := db.BulkUpdateTenantStatusParams{
		Status: string(status),
		ID:     ids,
	}

	err := r.queries.BulkUpdateTenantStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to bulk update tenant status: %w", err)
	}
	return nil
}

func (r *TenantRepository) mapToTenant(dbTenant *db.Tenant) *tenant.Tenant {
	t := &tenant.Tenant{
		ID:       dbTenant.ID,
		Uuid:     dbTenant.Uuid,
		Name:     dbTenant.Name,
		Status:   tenant.TenantStatus(dbTenant.Status),
		Industry: r.stringPtrToString(dbTenant.Industry),
	}

	if dbTenant.Subdomain != nil {
		t.Subdomain = *dbTenant.Subdomain
	}

	if dbTenant.CreatedAt.Valid {
		t.CreatedAt = dbTenant.CreatedAt.Time
	}

	if dbTenant.UpdatedAt.Valid {
		t.UpdatedAt = dbTenant.UpdatedAt.Time
	}

	if dbTenant.DeletedAt.Valid {
		t.DeletedAt = dbTenant.DeletedAt.Time
	}

	return t
}

func (r *TenantRepository) stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}