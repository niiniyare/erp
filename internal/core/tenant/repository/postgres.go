package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/shared/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type postgres struct {
	store  db.Store
	tracer tracing.Service
}

// NewPostgres creates a new PostgreSQL-backed tenant repository.
func NewPostgres(store db.Store, tracer tracing.Service) Repository {
	return &postgres{store: store, tracer: tracer}
}

func (r *postgres) Create(ctx context.Context, t *domain.Tenant) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.Create")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.name", t.Name))

	metadataJSON, _ := json.Marshal(t.Metadata)
	settingsJSON, _ := json.Marshal(t.Settings)

	var companySize *string
	if t.CompanySize != nil {
		upper := strings.ToUpper(*t.CompanySize)
		companySize = &upper
	}

	params := db.CreateTenantCompleteParams{
		Name:               t.Name,
		Slug:               t.Slug,
		Email:              t.Email,
		Subdomain:          t.Subdomain,
		Status:             string(t.Status),
		Timezone:           t.Timezone,
		CurrencyCode:       t.CurrencyCode,
		Metadata:           metadataJSON,
		Industry:           t.Industry,
		CompanySize:        companySize,
		TaxID:              t.TaxID,
		RegistrationNumber: t.RegistrationNumber,
		LegalEntityType:    t.LegalEntityType,
		Settings:           settingsJSON,
	}

	created, err := r.store.CreateTenantComplete(ctx, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create failed")
		return parseTenantDBError(err, "create")
	}

	t.ID = created.ID
	t.Slug = created.Slug
	t.CreatedAt = created.CreatedAt
	t.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *postgres) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetByID")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", id.String()))

	row, err := r.store.GetTenantByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, parseTenantDBError(err, "get_by_id")
	}
	return toDomain(row)
}

func (r *postgres) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetBySubdomain")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	row, err := r.store.GetTenantByUUID(ctx, &subdomain)
	if err != nil {
		span.RecordError(err)
		return nil, parseTenantDBError(err, "get_by_subdomain")
	}
	return toDomain(row)
}

func (r *postgres) Update(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.Update")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", id.String()))

	var status *string
	if req.Status != nil {
		s := string(*req.Status)
		status = &s
	}

	var settingsJSON []byte
	if req.Settings != nil {
		settingsJSON, _ = json.Marshal(req.Settings)
	}

	var updateCompanySize *string
	if req.CompanySize != nil {
		upper := strings.ToUpper(*req.CompanySize)
		updateCompanySize = &upper
	}

	params := db.UpdateTenantCompleteParams{
		Name:               req.Name,
		Email:              req.Email,
		Subdomain:          req.Subdomain,
		Status:             status,
		Timezone:           req.Timezone,
		CurrencyCode:       req.CurrencyCode,
		Industry:           req.Industry,
		CompanySize:        updateCompanySize,
		TaxID:              req.TaxID,
		RegistrationNumber: req.RegistrationNumber,
		LegalEntityType:    req.LegalEntityType,
		Settings:           settingsJSON,
		ID:                 id,
	}

	_, err := r.store.UpdateTenantComplete(ctx, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update failed")
		return parseTenantDBError(err, "update")
	}
	return nil
}

func (r *postgres) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.SoftDelete")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", id.String()))

	if err := r.store.SoftDeleteTenant(ctx, id); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to soft delete tenant: %w", err)
	}
	return nil
}

func (r *postgres) List(ctx context.Context, filter domain.TenantFilter) ([]*domain.Tenant, int64, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.List")
	defer span.End()

	// Get count
	count, err := r.store.CountFilteredTenants(ctx, db.CountFilteredTenantsParams{
		NameFilter:     filter.NameFilter,
		StatusFilter:   filter.StatusFilter,
		IndustryFilter: filter.IndustryFilter,
	})
	if err != nil {
		span.RecordError(err)
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	rows, err := r.store.FilterTenants(ctx, db.FilterTenantsParams{
		NameFilter:     filter.NameFilter,
		StatusFilter:   filter.StatusFilter,
		IndustryFilter: filter.IndustryFilter,
		SortBy:         sortBy,
		OffsetCount:    filter.Offset,
		LimitCount:     limit,
	})
	if err != nil {
		span.RecordError(err)
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	tenants := make([]*domain.Tenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, filterRowToDomain(row))
	}

	span.SetAttributes(attribute.Int("result.count", len(tenants)))
	return tenants, count, nil
}

func (r *postgres) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.SubdomainExists")
	defer span.End()

	exists, err := r.store.CheckSubdomainExists(ctx, &subdomain)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check subdomain: %w", err)
	}
	return exists, nil
}

func (r *postgres) Exists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.Exists")
	defer span.End()

	exists, err := r.store.CheckTenantExists(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	return exists, nil
}

func (r *postgres) CreateDefaultConfig(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.CreateDefaultConfig")
	defer span.End()

	if err := r.store.CreateDefaultTenantConfiguration(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create default config: %w", err)
	}
	return nil
}

func (r *postgres) GetConfig(ctx context.Context, tenantID uuid.UUID) (*domain.TenantConfiguration, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetConfig")
	defer span.End()

	row, err := r.store.GetTenantConfiguration(ctx)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return configToDomain(row), nil
}

func (r *postgres) GetUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetUsage")
	defer span.End()

	row, err := r.store.GetLatestTenantUsageStats(ctx)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	return usageToDomain(row), nil
}

func (r *postgres) InitUsage(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.InitUsage")
	defer span.End()

	_, err := r.store.InitializeUsageStats(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to init usage stats: %w", err)
	}
	return nil
}

func (r *postgres) Provision(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.Provision")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.name", input.Name))

	var result *domain.ProvisioningResult

	err := r.store.WithTx(ctx, func(ctx context.Context, txStore db.Store) error {
		subdomain := ""
		if input.Subdomain != nil {
			subdomain = *input.Subdomain
		}
		industry := ""
		if input.Industry != nil {
			industry = *input.Industry
		}
		companySize := "SMALL"
		if input.CompanySize != nil {
			companySize = strings.ToUpper(*input.CompanySize)
		}

		id, err := txStore.ProvisionTenant(ctx, db.ProvisionTenantParams{
			PName:         input.Name,
			PEmail:        input.Email,
			PSubdomain:    subdomain,
			PIndustry:     industry,
			PCompanySize:  companySize,
			PCurrencyCode: input.CurrencyCode,
			PTimezone:     "UTC",
			PSettings:     []byte("{}"),
		})
		if err != nil {
			return fmt.Errorf("provision_tenant_complete failed: %w", err)
		}

		result = &domain.ProvisioningResult{
			TenantID:  id,
			Subdomain: input.Subdomain,
			Status:    "pending",
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "provisioning failed")
		return nil, err
	}

	return result, nil
}

func (r *postgres) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.BulkUpdateStatus")
	defer span.End()

	if err := r.store.BulkUpdateTenantStatus(ctx, db.BulkUpdateTenantStatusParams{
		Status: string(status),
		ID:     ids,
	}); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to bulk update status: %w", err)
	}
	return nil
}

func (r *postgres) BulkSoftDelete(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.BulkSoftDelete")
	defer span.End()

	if err := r.store.BulkSoftDeleteTenants(ctx, ids); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to bulk soft delete: %w", err)
	}
	return nil
}

func (r *postgres) GetGrowthStats(ctx context.Context, days int) ([]domain.GrowthStat, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetGrowthStats")
	defer span.End()

	cutoff := time.Now().AddDate(0, 0, -days)
	rows, err := r.store.GetTenantGrowthStats(ctx, db.GetTenantGrowthStatsParams{
		CreatedAt:   cutoff,
		CreatedAt_2: cutoff,
	})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get growth stats: %w", err)
	}

	stats := make([]domain.GrowthStat, 0, len(rows))
	for _, row := range rows {
		stats = append(stats, domain.GrowthStat{
			Month:            fmt.Sprintf("%v", row.Month),
			NewTenants:       row.NewTenants,
			ActiveNewTenants: row.ActiveNewTenants,
		})
	}
	return stats, nil
}

func (r *postgres) GetStatusDistribution(ctx context.Context) (*domain.StatusCount, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetStatusDistribution")
	defer span.End()

	row, err := r.store.GetTenantStatusDistribution(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}

	activePct, _ := row.ActivePercentage.Float64Value()
	suspPct, _ := row.SuspendedPercentage.Float64Value()
	pendPct, _ := row.PendingPercentage.Float64Value()

	return &domain.StatusCount{
		TotalTenants:     row.TotalTenants,
		ActiveTenants:    row.ActiveTenants,
		SuspendedTenants: row.SuspendedTenants,
		PendingTenants:   row.PendingTenants,
		ActivePct:        activePct.Float64,
		SuspendedPct:     suspPct.Float64,
		PendingPct:       pendPct.Float64,
	}, nil
}

func (r *postgres) GetCurrentTenantID(ctx context.Context) (uuid.UUID, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetCurrentTenantID")
	defer span.End()

	id, err := r.store.GetCurrentTenantID(ctx)
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, fmt.Errorf("failed to get current tenant ID: %w", err)
	}
	return id, nil
}

func (r *postgres) GetCurrentTenant(ctx context.Context) (*domain.Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.GetCurrentTenant")
	defer span.End()

	row, err := r.store.GetCurrentTenant(ctx)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get current tenant: %w", err)
	}
	return toDomain(row)
}

func (r *postgres) ValidateCurrentTenant(ctx context.Context) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repo.ValidateCurrentTenant")
	defer span.End()

	exists, err := r.store.CheckCurrentTenantExists(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to validate current tenant: %w", err)
	}
	if !exists {
		return domain.ErrTenantNotFound
	}
	return nil
}
