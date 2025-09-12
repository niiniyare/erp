package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	goaTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// TenantGoaHandler implements the GOA tenant service following the data flow pattern
type TenantGoaHandler struct {
	tenantService coreTenant.Service
	tracing       tracing.TracingService
	metrics       *metrics.MetricsService
}

// NewTenantGoaHandler creates a new GOA tenant handler following Clean Architecture pattern
func NewTenantGoaHandler(tenantSvc coreTenant.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) goaTenant.Service {
	return &TenantGoaHandler{
		tenantService: tenantSvc,
		tracing:       tracing,
		metrics:       metrics,
	}
}

// Create creates a new tenant following the data flow pattern
func (h *TenantGoaHandler) Create(ctx context.Context, p *goaTenant.CreateTenantPayload) (*goaTenant.Tenant, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.create",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("tenant.name", p.Name),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_create_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "Tenant create called", logger.Fields{
		"name": p.Name,
	})

	// Convert GOA payload to domain request
	req := coreTenant.CreateTenantRequest{
		Name:  p.Name,
		Email: "admin@" + p.Name + ".com", // Default email - should be provided in payload
	}

	// Set optional fields if provided
	if p.Subdomain != nil {
		req.Subdomain = p.Subdomain
	}

	// Create tenant using core service
	newTenant, err := h.tenantService.CreateTenant(ctx, req)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_create_errors", metrics.Fields{"error": err.Error()})

		// Convert domain errors to GOA errors
		switch {
		case errors.Is(err, sharedErrors.ErrSubdomainAlreadyExists):
			return nil, "", goaTenant.MakeConflict(err)
		case errors.Is(err, sharedErrors.ErrInvalidInput):
			return nil, "", goaTenant.MakeBadRequest(err)
		default:
			logger.ErrorContext(ctx, "Tenant creation failed", logger.Fields{
				"error": err.Error(),
				"name":  p.Name,
			})
			return nil, "", goaTenant.MakeUnprocessableEntity(err)
		}
	}

	// Convert domain model to GOA result
	tenantResult := &goaTenant.Tenant{
		ID:        newTenant.ID.String(),
		Name:      newTenant.Name,
		Slug:      newTenant.Slug,
		Subdomain: newTenant.Subdomain,
		Status:    string(newTenant.Status),
		CreatedAt: newTenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: newTenant.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.metrics.IncrementCounter("tenant_create_total", metrics.Fields{})
	span.SetAttributes(attribute.String("result.tenant_id", tenantResult.ID))

	logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
		"tenant_id": tenantResult.ID,
		"name":      tenantResult.Name,
	})

	return tenantResult, "default", nil
}

// Get retrieves a tenant by ID following the data flow pattern
func (h *TenantGoaHandler) Get(ctx context.Context, p *goaTenant.GetPayload) (*goaTenant.Tenant, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.get",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("tenant.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_get_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid tenant ID format", logger.Fields{
			"tenant_id": p.ID,
			"error":     err.Error(),
		})
		return nil, "", goaTenant.MakeBadRequest(errors.New("invalid tenant ID format"))
	}

	// Get tenant using core service
	tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_get_errors", metrics.Fields{"error": err.Error()})

		// Convert domain errors to GOA errors
		switch {
		case errors.Is(err, sharedErrors.ErrTenantNotFound):
			return nil, "", goaTenant.MakeNotFound(err)
		default:
			logger.ErrorContext(ctx, "Tenant retrieval failed", logger.Fields{
				"error":     err.Error(),
				"tenant_id": p.ID,
			})
			return nil, "", err
		}
	}

	// Convert domain model to GOA result
	tenantResult := &goaTenant.Tenant{
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Subdomain: tenant.Subdomain,
		Status:    string(tenant.Status),
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.metrics.IncrementCounter("tenant_get_total", metrics.Fields{})
	return tenantResult, "default", nil
}

// List retrieves tenants with pagination following the data flow pattern
func (h *TenantGoaHandler) List(ctx context.Context, p *goaTenant.ListPayload) (*goaTenant.ListResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.list",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_list_duration", metrics.Fields{})
	defer timer.Stop()

	// Set defaults for pagination
	page := int(p.Page)
	pageSize := int(p.PageSize)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get tenants using core service
	tenants, err := h.tenantService.ListTenants(ctx, offset, pageSize)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_list_errors", metrics.Fields{"error": err.Error()})

		logger.ErrorContext(ctx, "Tenant list failed", logger.Fields{
			"error":  err.Error(),
			"offset": offset,
			"limit":  pageSize,
		})
		return nil, err
	}

	// Convert domain models to GOA results
	tenantResults := make([]*goaTenant.Tenant, len(tenants))
	for i, tenant := range tenants {
		tenantResults[i] = &goaTenant.Tenant{
			ID:        tenant.ID.String(),
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			Subdomain: tenant.Subdomain,
			Status:    string(tenant.Status),
			CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Calculate pagination metadata
	totalItems := uint(len(tenants)) // Note: This is a simplified count, real implementation should get total count separately
	totalPages := (totalItems + uint(pageSize) - 1) / uint(pageSize)
	hasNext := uint(page) < totalPages
	hasPrev := page > 1

	result := &goaTenant.ListResult{
		Data: tenantResults,
		Pagination: &goaTenant.PaginationMeta{
			CurrentPage: uint(page),
			PageSize:    uint(pageSize),
			TotalItems:  totalItems,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
	}

	h.metrics.IncrementCounter("tenant_list_total", metrics.Fields{})
	return result, nil
}

// Update updates an existing tenant following the data flow pattern
func (h *TenantGoaHandler) Update(ctx context.Context, p *goaTenant.UpdateTenantPayload) (*goaTenant.Tenant, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.update",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("tenant.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_update_duration", metrics.Fields{})
	defer timer.Stop()

	// TODO: Integrate with existing tenant update logic using h.tenantService.UpdateTenant
	tenantResult := &goaTenant.Tenant{
		ID:        p.ID,
		Name:      *p.Name,
		Status:    "active",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("tenant_update_total", metrics.Fields{})
	return tenantResult, "default", nil
}

// Delete deletes a tenant following the data flow pattern
func (h *TenantGoaHandler) Delete(ctx context.Context, p *goaTenant.DeletePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.delete",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("tenant.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_delete_duration", metrics.Fields{})
	defer timer.Stop()

	logger.Info("Tenant delete called", logger.Fields{
		"id": p.ID,
	})

	// TODO: Integrate with existing tenant deletion logic using h.tenantService.DeleteTenant
	h.metrics.IncrementCounter("tenant_delete_total", metrics.Fields{})
	return nil
}

// Health returns tenant service health status following the data flow pattern
func (h *TenantGoaHandler) Health(ctx context.Context) (*goaTenant.HealthResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "tenant.health",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_health_duration", metrics.Fields{})
	defer timer.Stop()

	// Return health status
	result := &goaTenant.HealthResult{
		Status:    "healthy",
		Timestamp: "2024-01-01T00:00:00Z",
		Version:   "1.0.0",
	}

	h.metrics.IncrementCounter("tenant_health_total", metrics.Fields{})
	return result, nil
}