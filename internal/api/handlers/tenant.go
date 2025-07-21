package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	goaTenant "github.com/niiniyare/erp/gen/tenant"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	service coreTenant.Service
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(service coreTenant.Service) *TenantHandler {
	return &TenantHandler{service: service}
}

// CreateTenant handles tenant creation requests
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req coreTenant.CreateTenantRequest

	logger.InfoContext(c.Request.Context(), "Received tenant creation request", logger.Fields{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"client_ip": c.ClientIP(),
	})

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnContext(c.Request.Context(), "Invalid request payload", logger.Fields{
			"error":        err.Error(),
			"content_type": c.GetHeader("Content-Type"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create tenant
	newTenant, err := h.service.CreateTenant(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, sharedErrors.ErrSubdomainAlreadyExists):
			logger.WarnContext(c.Request.Context(), "Tenant creation failed - subdomain exists", logger.Fields{
				"subdomain":   req.Subdomain,
				"tenant_name": req.Name,
			})
			c.JSON(http.StatusConflict, gin.H{"error": "Subdomain already exists"})
		default:
			logger.ErrorContext(c.Request.Context(), "Tenant creation failed", logger.Fields{
				"error":       err.Error(),
				"tenant_name": req.Name,
				"subdomain":   req.Subdomain,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	logger.InfoContext(c.Request.Context(), "Tenant created successfully", logger.Fields{
		"tenant_id":   newTenant.ID.String(),
		"tenant_name": newTenant.Name,
		"subdomain":   newTenant.Subdomain,
		"status_code": http.StatusCreated,
	})

	c.JSON(http.StatusCreated, newTenant)
}

// GetTenant handles tenant retrieval requests
func (h *TenantHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	t, err := h.service.GetTenant(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, sharedErrors.ErrTenantNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, t)
}

// GetTenantBySubdomain handles tenant retrieval by subdomain
func (h *TenantHandler) GetTenantBySubdomain(c *gin.Context) {
	subdomain := c.Param("subdomain")

	t, err := h.service.GetTenantBySubdomain(c.Request.Context(), subdomain)
	if err != nil {
		switch {
		case errors.Is(err, sharedErrors.ErrTenantNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, t)
}

// UpdateTenant handles tenant update requests
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req coreTenant.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateTenant(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, sharedErrors.ErrTenantNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant updated successfully"})
}

// DeleteTenant handles tenant deletion requests
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	err = h.service.DeactivateTenant(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, sharedErrors.ErrTenantNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deactivated successfully"})
}

// ListTenants handles tenant listing requests
func (h *TenantHandler) ListTenants(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	tenants, err := h.service.ListTenants(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"offset":  offset,
		"limit":   limit,
		"count":   len(tenants),
	})
}

/*
// Goa-generated service implementation
type tenantService struct {
    handler *TenantHandler
}

// NewTenantService creates a Goa service implementation
func NewTenantService(handler *TenantHandler) gen.Service {
    return &tenantService{handler: handler}
}
*/

/*
// Create implements the generated service interface
func (s *tenantService) Create(ctx context.Context, p *gen.CreatePayload) (*gen.Tenant, error) {
    req := tenant.CreateTenantRequest{
        Name:      p.Name,
        Subdomain: p.Subdomain,
        PlanType:  tenant.PlanType(p.PlanType),
        Settings:  p.Settings,
    }

    newTenant, err := s.handler.service.CreateTenant(ctx, req)
    if err != nil {
        // Convert to Goa errors
        switch {
        case errors.Is(err, sharedErrors.ErrSubdomainAlreadyExists):
            return nil, gen.MakeConflict(err)
        default:
            return nil, gen.MakeInternalError(err)
        }
    }

    // Convert to Goa result type
    res := &gen.Tenant{
        ID:        newTenant.ID.String(),
        Name:      newTenant.Name,
        Subdomain: newTenant.Subdomain,
        PlanType:  string(newTenant.PlanType),
        Status:    string(newTenant.Status),
        CreatedAt: newTenant.CreatedAt.Format(time.RFC3339),
        UpdatedAt: newTenant.UpdatedAt.Format(time.RFC3339),
    }

    return res, nil
}

func (s *tenantService) Get(ctx context.Context, p *gen.GetPayload) (*gen.Tenant, error) {
    id, err := uuid.Parse(p.ID)
    if err != nil {
        return nil, gen.MakeBadRequest(err)
    }

    t, err := s.handler.service.GetTenant(ctx, id)
    if err != nil {
        switch {
        case errors.Is(err, sharedErrors.ErrTenantNotFound):
            return nil, gen.MakeNotFound(err)
        default:
            return nil, gen.MakeInternalError(err)
        }
    }

    res := &gen.Tenant{
        ID:        t.ID.String(),
        Name:      t.Name,
        Subdomain: t.Subdomain,
        PlanType:  string(t.PlanType),
        Status:    string(t.Status),
        CreatedAt: t.CreatedAt.Format(time.RFC3339),
        UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
    }

    return res, nil
}

func (s *tenantService) List(ctx context.Context, p *gen.ListPayload) (gen.TenantCollection, error) {
    tenants, err := s.handler.service.ListTenants(ctx, *p.Offset, *p.Limit)
    if err != nil {
        return nil, gen.MakeInternalError(err)
    }

    res := make(gen.TenantCollection, len(tenants))
    for i, t := range tenants {
        res[i] = &gen.Tenant{
            ID:        t.ID.String(),
            Name:      t.Name,
            Subdomain: t.Subdomain,
            PlanType:  string(t.PlanType),
            Status:    string(t.Status),
            CreatedAt: t.CreatedAt.Format(time.RFC3339),
            UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
        }
    }

    return res, nil
}
*/

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

	logger.Info("Tenant create called", logger.Fields{
		"name": p.Name,
	})

	// TODO: Integrate with existing tenant creation logic using h.tenantService.CreateTenant
	// This should convert p (GOA payload) to domain request, call the service, and convert back

	// For now, return a mock response
	tenantResult := &goaTenant.Tenant{
		ID:        "mock-tenant-id",
		Name:      p.Name,
		Status:    "active",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("tenant_create_total", metrics.Fields{})
	span.SetAttributes(attribute.String("result.tenant_id", tenantResult.ID))

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

	// TODO: Integrate with existing tenant retrieval logic using h.tenantService.GetTenantByID
	tenantResult := &goaTenant.Tenant{
		ID:        p.ID,
		Name:      "Mock Tenant",
		Status:    "active",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
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

	// TODO: Integrate with existing tenant listing logic using h.tenantService.ListTenants
	result := &goaTenant.ListResult{
		Data:       []*goaTenant.Tenant{},
		Pagination: &goaTenant.PaginationMeta{CurrentPage: 1, PageSize: 20, TotalItems: 0, TotalPages: 0, HasNext: false, HasPrev: false},
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
