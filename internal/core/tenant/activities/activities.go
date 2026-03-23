package activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/core/tenant/repository"
	"awo.so/internal/core/tenant/service"
	"awo.so/internal/shared/tracing"
)

// Activities holds all tenant Temporal activity implementations.
type Activities struct {
	tenant *service.TenantService
	prov   *service.ProvisioningService
	repo   repository.Repository
	tracer tracing.Service
}

// Deps contains dependencies for tenant activities.
type Deps struct {
	TenantService       *service.TenantService
	ProvisioningService *service.ProvisioningService
	Repo                repository.Repository
	Tracer              tracing.Service
}

// New creates a new Activities instance.
func New(deps Deps) *Activities {
	return &Activities{
		tenant: deps.TenantService,
		prov:   deps.ProvisioningService,
		repo:   deps.Repo,
		tracer: deps.Tracer,
	}
}

// ProvisionTenantActivity provisions a tenant via the DB function.
func (a *Activities) ProvisionTenantActivity(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "activity.ProvisionTenant")
	defer span.End()

	result, err := a.prov.Provision(ctx, input)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("provision tenant activity failed: %w", err)
	}
	return result, nil
}

// CreateDefaultConfigActivity creates the default tenant configuration.
func (a *Activities) CreateDefaultConfigActivity(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.CreateDefaultConfig")
	defer span.End()

	if err := a.repo.CreateDefaultConfig(ctx, tenantID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("create default config activity failed: %w", err)
	}
	return nil
}

// InitUsageActivity initializes usage stats for a tenant.
func (a *Activities) InitUsageActivity(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.InitUsage")
	defer span.End()

	if err := a.repo.InitUsage(ctx, tenantID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("init usage activity failed: %w", err)
	}
	return nil
}

// ActivateTenantActivity activates a tenant.
func (a *Activities) ActivateTenantActivity(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.ActivateTenant")
	defer span.End()

	if err := a.tenant.Activate(ctx, tenantID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("activate tenant activity failed: %w", err)
	}
	return nil
}

// CleanupTenantActivity soft-deletes a tenant (compensation).
func (a *Activities) CleanupTenantActivity(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.CleanupTenant")
	defer span.End()

	if err := a.tenant.Delete(ctx, tenantID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("cleanup tenant activity failed: %w", err)
	}
	return nil
}

// SendWelcomeNotificationActivity sends a welcome notification (placeholder).
func (a *Activities) SendWelcomeNotificationActivity(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: integrate with notification service
	return nil
}

// BulkUpdateStatusActivity updates status for multiple tenants.
func (a *Activities) BulkUpdateStatusActivity(ctx context.Context, ids []uuid.UUID, status domain.TenantStatus) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.BulkUpdateStatus")
	defer span.End()

	if err := a.repo.BulkUpdateStatus(ctx, ids, status); err != nil {
		span.RecordError(err)
		return fmt.Errorf("bulk update status activity failed: %w", err)
	}
	return nil
}

// BulkSoftDeleteActivity soft-deletes multiple tenants.
func (a *Activities) BulkSoftDeleteActivity(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "activity.BulkSoftDelete")
	defer span.End()

	if err := a.repo.BulkSoftDelete(ctx, ids); err != nil {
		span.RecordError(err)
		return fmt.Errorf("bulk soft delete activity failed: %w", err)
	}
	return nil
}
