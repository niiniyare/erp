package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	goaTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/timeutil"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// TenantHandler implements the unified tenant service interface
type TenantHandler struct {
	tenantService       tenant.Service
	provisioningService tenant.ProvisioningService
	tracing             tracing.TracingService
	metrics             *metrics.MetricsService
	logger              logger.Logger
}

// NewUnifiedTenantHandler creates a new unified tenant handler
func NewUnifiedTenantHandler(
	tenantService tenant.Service,
	provisioningService tenant.ProvisioningService,
	tracingService tracing.TracingService,
	metricsService *metrics.MetricsService,
) goaTenant.Service {
	return &TenantHandler{
		tenantService:       tenantService,
		provisioningService: provisioningService,
		tracing:             tracingService,
		metrics:             metricsService,
		logger:              logger.WithFields(logger.Fields{"component": "unified_tenant_handler"}),
	}
}

// ============================================================================
// BASIC TENANT OPERATIONS (from original tenant handler)
// ============================================================================

// Create implements tenant creation
func (h *TenantHandler) Create(ctx context.Context, p *goaTenant.CreateTenantPayload) (*goaTenant.CreateTenantResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Create")
	defer span.End()

	timer := h.metrics.Timer("tenant_create_duration", metrics.Fields{})
	defer timer.Stop()

	span.SetAttributes(
		attribute.String("tenant.name", p.Name),
		attribute.String("tenant.plan_type", p.PlanType),
	)

	h.logger.InfoContext(ctx, "Creating tenant", logger.Fields{
		"name":      p.Name,
		"plan_type": p.PlanType,
	})

	// Extract user ID from context for audit logging
	// _, _ = shared.GetUserID(ctx)

	// Convert GOA payload to service request
	createReq := tenant.CreateTenantRequest{
		Name:               p.Name,
		Email:              p.Email, // Default email, should be provided in payload
		Subdomain:          p.Subdomain,
		CountryCode:        p.CountryCode,  // Default country code, should be provided in payload
		CurrencyCode:       p.CurrencyCode, // Currency code from payload
		Status:             tenant.Status(strings.ToUpper(p.Status)),
		Industry:           p.Industry,
		CompanySize:        p.CompanySize,
		TaxID:              p.TaxID,
		RegistrationNumber: p.RegistrationNumber,
		LegalEntityType:    p.LegalEntityType,
	}
	if p.Slug != nil {
		createReq.Slug = *p.Slug
	}

	// Set contact email if provided
	if p.Contact != nil && p.Contact.Email != nil {
		createReq.Email = *p.Contact.Email
	}

	// Set settings if provided
	if p.Settings != nil {
		settings := make(map[string]any)
		settings["timezone"] = p.Settings.Timezone
		settings["currency"] = p.Settings.Currency
		settings["date_format"] = p.Settings.DateFormat
		settings["language"] = p.Settings.Language
		settings["features"] = p.Settings.Features

		if p.Settings.Limits != nil {
			limits := map[string]any{
				"max_users":              uintValue(p.Settings.Limits.MaxUsers),
				"max_storage_mb":         uintValue(p.Settings.Limits.MaxStorageMb),
				"max_api_calls_per_hour": uintValue(p.Settings.Limits.MaxAPICallsPerHour),
			}
			settings["limits"] = limits
		}

		createReq.Settings = settings
	}

	// Call the tenant service
	result, err := h.tenantService.CreateTenant(ctx, createReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant creation failed", logger.Fields{
			"error":     err.Error(),
			"name":      p.Name,
			"plan_type": p.PlanType,
		})

		h.metrics.IncrementCounter("tenant_create_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return &goaTenant.CreateTenantResult{
			Tenant:  nil,
			Status:  "FAILED",
			Message: fmt.Sprintf("Tenant creation failed: %s", err.Error()),
		}, convertServiceError(err)
	}

	// Convert service result to GOA response
	goaResultTenant := convertTenantToGOA(result)

	h.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
		"tenant_id": result.ID.String(),
		"name":      result.Name,
		"subdomain": result.Subdomain,
	})

	h.metrics.IncrementCounter("tenant_create_total", metrics.Fields{
		"status": "success",
	})

	return &goaTenant.CreateTenantResult{
		Tenant:  goaResultTenant,
		Status:  "SUCCESS",
		Message: "Tenant created successfully",
	}, nil
}

// Get implements tenant retrieval by ID
func (h *TenantHandler) Get(ctx context.Context, p *goaTenant.GetPayload) (*goaTenant.Tenant, string, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Get")
	defer span.End()

	timer := h.metrics.Timer("tenant_get_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, "", goaTenant.BadRequest("Invalid tenant ID format")
	}

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	h.logger.InfoContext(ctx, "Retrieving tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "handler.Get",
	})

	// Call the tenant service
	result, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant retrieval failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_get_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, "", convertServiceError(err)
	}

	// Convert service result to GOA response
	goaResult := convertTenantToGOA(result)

	h.logger.InfoContext(ctx, "Tenant retrieved successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"name":      result.Name,
	})

	h.metrics.IncrementCounter("tenant_get_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, "default", nil
}

// List implements tenant listing with pagination
func (h *TenantHandler) List(ctx context.Context, p *goaTenant.ListPayload) (*goaTenant.ListResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.List")
	defer span.End()

	timer := h.metrics.Timer("tenant_list_duration", metrics.Fields{})
	defer timer.Stop()

	h.logger.InfoContext(ctx, "Listing tenants", logger.Fields{
		"page":      p.Page,
		"page_size": p.PageSize,
		"operation": "handler.List",
	})

	// Calculate offset from page and page size
	offset := int(p.Page-1) * int(p.PageSize)
	limit := int(p.PageSize)

	// Call the tenant service
	tenants, err := h.tenantService.ListTenants(ctx, offset, limit)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant listing failed", logger.Fields{
			"error": err.Error(),
		})

		h.metrics.IncrementCounter("tenant_list_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert to GOA response
	goaResult := &goaTenant.ListResult{
		Data: make([]*goaTenant.Tenant, len(tenants)),
		Pagination: &goaTenant.PaginationMeta{
			CurrentPage: p.Page,
			PageSize:    p.PageSize,
			TotalItems:  uint(len(tenants)), // Simple count, should be improved with total count
			TotalPages:  uint(len(tenants)/int(p.PageSize)) + 1,
			HasNext:     len(tenants) == int(p.PageSize),
			HasPrev:     p.Page > 1,
		},
	}

	for i, t := range tenants {
		goaResult.Data[i] = convertTenantToGOA(t)
	}

	h.logger.InfoContext(ctx, "Tenants listed successfully", logger.Fields{
		"count": len(tenants),
	})

	h.metrics.IncrementCounter("tenant_list_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// Update implements tenant updates
func (h *TenantHandler) Update(ctx context.Context, p *goaTenant.UpdateTenantPayload) (*goaTenant.Tenant, string, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Update")
	defer span.End()

	timer := h.metrics.Timer("tenant_update_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, "", goaTenant.BadRequest("Invalid tenant ID format")
	}

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	h.logger.InfoContext(ctx, "Updating tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "handler.Update",
	})

	// Extract user ID from context for audit logging
	_, _ = shared.GetUserID(ctx)

	// Build update request
	updateReq := tenant.UpdateTenantRequest{
		Name: p.Name,
	}

	// Set email if provided in contact
	if p.Contact != nil && p.Contact.Email != nil {
		updateReq.Email = p.Contact.Email
	}

	// Set subdomain if provided
	if p.Settings != nil {
		// For now, just store settings as a simple map
		// This should be improved to match the actual settings structure
		settings := make(map[string]any)
		if p.Settings.Timezone != "" {
			settings["timezone"] = p.Settings.Timezone
		}
		if p.Settings.Currency != "" {
			settings["currency"] = p.Settings.Currency
		}
		if p.Settings.DateFormat != "" {
			settings["date_format"] = p.Settings.DateFormat
		}
		if p.Settings.Language != "" {
			settings["language"] = p.Settings.Language
		}
		if len(p.Settings.Features) > 0 {
			settings["features"] = p.Settings.Features
		}
		updateReq.Settings = settings
	}

	// Call the tenant service
	result, err := h.tenantService.UpdateTenant(ctx, tenantID, updateReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant update failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_update_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, "", convertServiceError(err)
	}

	// Convert service result to GOA response
	goaResult := convertTenantToGOA(result)

	h.logger.InfoContext(ctx, "Tenant updated successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"name":      result.Name,
	})

	h.metrics.IncrementCounter("tenant_update_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, "default", nil
}

// Delete implements tenant deletion (soft delete)
func (h *TenantHandler) Delete(ctx context.Context, p *goaTenant.DeletePayload) error {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Delete")
	defer span.End()

	timer := h.metrics.Timer("tenant_delete_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return goaTenant.BadRequest("Invalid tenant ID format")
	}

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	h.logger.InfoContext(ctx, "Deleting tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "handler.Delete",
	})

	// Extract user ID from context for audit logging
	_, _ = shared.GetUserID(ctx)

	// Call the tenant service
	err = h.tenantService.DeleteTenant(ctx, tenantID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant deletion failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_delete_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return convertServiceError(err)
	}

	h.logger.InfoContext(ctx, "Tenant deleted successfully", logger.Fields{
		"tenant_id": tenantID.String(),
	})

	h.metrics.IncrementCounter("tenant_delete_total", metrics.Fields{
		"status": "success",
	})

	return nil
}

// Health implements health check
func (h *TenantHandler) Health(ctx context.Context) (*goaTenant.HealthResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Health")
	defer span.End()

	// Simple health check
	result := &goaTenant.HealthResult{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	h.logger.InfoContext(ctx, "Health check completed", logger.Fields{
		"status": result.Status,
	})

	return result, nil
}

// ============================================================================
// TENANT MANAGEMENT OPERATIONS (from tenant management adapter)
// ============================================================================

// Provision implements tenant provisioning
func (h *TenantHandler) Provision(ctx context.Context, p *goaTenant.ProvisionPayload) (*goaTenant.ProvisionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Provision")
	defer span.End()

	timer := h.metrics.Timer("tenant_provision_duration", metrics.Fields{})
	defer timer.Stop()

	span.SetAttributes(
		attribute.String("tenant.name", p.Name),
		attribute.String("tenant.subdomain", p.Subdomain),
	)

	h.logger.InfoContext(ctx, "Processing tenant provisioning request", logger.Fields{
		"tenant_name": p.Name,
		"subdomain":   p.Subdomain,
		"admin":       p.AdminEmail,
		"operation":   "handler.Provision",
	})

	// Convert simple GOA payload to comprehensive service request
	provisionReq := tenant.ComprehensiveProvisionRequest{
		Name:        p.Name,
		Subdomain:   p.Subdomain,
		Description: fmt.Sprintf("Tenant provisioned for %s", p.ContactEmail),

		// Contact information
		Contact: tenant.ContactInfo{
			Email: p.ContactEmail,
			Name:  p.ContactEmail, // Use email as name
		},

		// Admin user setup
		AdminUser: tenant.AdminUserRequest{
			Email:     p.AdminEmail,
			FirstName: p.AdminFirstName,
			LastName:  p.AdminLastName,
			Language:  "en-US",
			Timezone:  "UTC",
		},

		// Default settings
		InitialSettings: tenant.TenantInitialSettings{
			Currency:     "USD",
			Timezone:     "UTC",
			Language:     "en-US",
			DateFormat:   "MM/DD/YYYY",
			NumberFormat: "US",
		},

		// Default modules
		EnabledModules: []string{"finance"},

		// Default limits
		InitialLimits: tenant.TenantLimits{
			MaxUsers:           100,
			MaxStorageMB:       10240, // 10GB
			MaxAPICallsPerHour: 5000,
		},
	}

	// Call the provisioning service
	result, err := h.provisioningService.ProvisionTenantComplete(ctx, provisionReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant provisioning failed", logger.Fields{
			"error":     err.Error(),
			"name":      p.Name,
			"subdomain": p.Subdomain,
		})

		h.metrics.IncrementCounter("tenant_provision_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert service result to simple GOA response
	goaResult := &goaTenant.ProvisionResult{
		TenantID: result.TenantID.String(),
		Status:   "SUCCESS",
		Message:  "Tenant provisioned successfully",
	}

	h.logger.InfoContext(ctx, "Tenant provisioned successfully", logger.Fields{
		"tenant_id": result.TenantID.String(),
	})

	h.metrics.IncrementCounter("tenant_provision_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// Suspend implements tenant suspension
func (h *TenantHandler) Suspend(ctx context.Context, p *goaTenant.SuspendPayload) (*goaTenant.SuspendResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Suspend")
	defer span.End()

	timer := h.metrics.Timer("tenant_suspend_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, goaTenant.BadRequest("Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Suspending tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	// Create suspend request
	suspendReq := tenant.SuspendTenantRequest{
		TenantID:    tenantID,
		Reason:      p.Reason,
		ActorID:     getActorIDFromContext(ctx),
		NotifyUsers: true,
	}

	// Call the provisioning service
	result, err := h.provisioningService.SuspendTenant(ctx, suspendReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant suspension failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_suspend_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert to simple GOA response
	goaResult := &goaTenant.SuspendResult{
		TenantID: result.TenantID.String(),
		Action:   result.Action,
		Status:   result.Status,
		Message:  result.Message,
	}

	h.logger.InfoContext(ctx, "Tenant suspended successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	h.metrics.IncrementCounter("tenant_suspend_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// Reactivate implements tenant reactivation
func (h *TenantHandler) Reactivate(ctx context.Context, p *goaTenant.ReactivatePayload) (*goaTenant.ReactivateResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.Reactivate")
	defer span.End()

	timer := h.metrics.Timer("tenant_reactivate_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, goaTenant.BadRequest("Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Reactivating tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	// Create reactivate request
	reactivateReq := tenant.ReactivateTenantRequest{
		TenantID: tenantID,
		Reason:   p.Reason,
		ActorID:  getActorIDFromContext(ctx),
	}

	// Call the provisioning service
	result, err := h.provisioningService.ReactivateTenant(ctx, reactivateReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant reactivation failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_reactivate_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert to simple GOA response
	goaResult := &goaTenant.ReactivateResult{
		TenantID: result.TenantID.String(),
		Action:   result.Action,
		Status:   result.Status,
		Message:  result.Message,
	}

	h.logger.InfoContext(ctx, "Tenant reactivated successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	h.metrics.IncrementCounter("tenant_reactivate_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// UpdateConfiguration implements tenant configuration updates
func (h *TenantHandler) UpdateConfiguration(ctx context.Context, p *goaTenant.UpdateConfigurationPayload) (*goaTenant.UpdateConfigurationResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.UpdateConfiguration")
	defer span.End()

	timer := h.metrics.Timer("tenant_config_update_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, goaTenant.BadRequest("Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Updating tenant configuration", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	// Build update request
	updateReq := tenant.UpdateTenantConfigurationRequest{
		TenantID: tenantID,
		Reason:   p.Reason,
		ActorID:  getActorIDFromContext(ctx),
	}

	// Add limits if provided
	if p.MaxUsers != nil || p.MaxStorageMb != nil || p.MaxAPICallsPerHour != nil {
		limits := &tenant.TenantLimits{}

		if p.MaxUsers != nil {
			limits.MaxUsers = *p.MaxUsers
		}
		if p.MaxStorageMb != nil {
			limits.MaxStorageMB = *p.MaxStorageMb
		}
		if p.MaxAPICallsPerHour != nil {
			limits.MaxAPICallsPerHour = *p.MaxAPICallsPerHour
		}

		updateReq.Limits = limits
	}

	// Call the provisioning service
	_, err = h.provisioningService.UpdateTenantConfiguration(ctx, updateReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Tenant configuration update failed", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_config_update_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert to simple GOA response
	goaResult := &goaTenant.UpdateConfigurationResult{
		TenantID: tenantID.String(),
		Status:   "SUCCESS",
		Message:  "Configuration updated successfully",
	}

	h.logger.InfoContext(ctx, "Tenant configuration updated successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"reason":    p.Reason,
	})

	h.metrics.IncrementCounter("tenant_config_update_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// GetUsageAnalytics implements usage analytics retrieval
func (h *TenantHandler) GetUsageAnalytics(ctx context.Context, p *goaTenant.GetUsageAnalyticsPayload) (*goaTenant.GetUsageAnalyticsResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant.handler.GetUsageAnalytics")
	defer span.End()

	timer := h.metrics.Timer("tenant_analytics_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, goaTenant.BadRequest("Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Retrieving tenant usage analytics", logger.Fields{
		"tenant_id": tenantID.String(),
		"period":    p.Period,
	})

	// Create analytics request
	analyticsReq := tenant.UsageAnalyticsRequest{
		TenantID: tenantID,
		Period:   p.Period,
		Metrics:  []string{"users", "storage", "api_calls"}, // Default metrics
	}

	// Call the provisioning service
	result, err := h.provisioningService.GetTenantUsageAnalytics(ctx, analyticsReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to retrieve usage analytics", logger.Fields{
			"error":     err.Error(),
			"tenant_id": tenantID.String(),
		})

		h.metrics.IncrementCounter("tenant_analytics_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		return nil, convertServiceError(err)
	}

	// Convert to simple GOA response
	goaResult := &goaTenant.GetUsageAnalyticsResult{
		TenantID: tenantID.String(),
		Period:   result.Period,
	}

	// Set metrics if available
	if result.UserMetrics != nil {
		userCount := uint(result.UserMetrics.ActiveUsers)
		goaResult.UserCount = &userCount
	}

	if result.StorageMetrics != nil {
		storageUsed := uint64(result.StorageMetrics.TotalUsedMB)
		goaResult.StorageUsedMb = &storageUsed
	}

	if result.APIMetrics != nil {
		apiCalls := uint64(result.APIMetrics.TotalCalls)
		goaResult.APICalls = &apiCalls
	}

	h.logger.InfoContext(ctx, "Usage analytics retrieved successfully", logger.Fields{
		"tenant_id": tenantID.String(),
		"period":    p.Period,
	})

	h.metrics.IncrementCounter("tenant_analytics_total", metrics.Fields{
		"status": "success",
	})

	return goaResult, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// convertTenantToGOA converts a service tenant to GOA tenant
func convertTenantToGOA(t *tenant.Tenant) *goaTenant.Tenant {
	result := &goaTenant.Tenant{
		ID:        t.ID.String(),
		Name:      t.Name,
		Slug:      t.Slug,
		Status:    string(t.Status), // Convert Status enum to string
		PlanType:  "basic",          // Default plan type since it's not in tenant model
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}

	if t.Subdomain != nil && *t.Subdomain != "" {
		result.Subdomain = t.Subdomain
	}

	// TODO: Add description field to tenant model if needed
	// if t.Description != "" {
	// 	result.Description = &t.Description
	// }

	// TODO: Add CreatedBy/UpdatedBy fields to tenant model if needed
	// if t.CreatedBy != uuid.Nil {
	// 	createdBy := t.CreatedBy.String()
	// 	result.CreatedBy = &createdBy
	// }

	// Convert settings from map[string]any to structured format
	if t.Settings != nil {
		settings := &goaTenant.TenantSettings{
			Timezone:   t.Timezone,             // Use direct fields from tenant
			Currency:   t.CurrencyCode,         // Use direct fields from tenant
			DateFormat: timeutil.HumanDateTime, // Default
			Language:   "en",                   // Default
			Features:   []string{},             // Default empty
		}

		// Extract settings from the map if they exist
		if timezone, ok := t.Settings["timezone"].(string); ok {
			settings.Timezone = timezone
		}
		if currency, ok := t.Settings["currency"].(string); ok {
			settings.Currency = currency
		}
		if dateFormat, ok := t.Settings["date_format"].(string); ok {
			settings.DateFormat = dateFormat
		}
		if language, ok := t.Settings["language"].(string); ok {
			settings.Language = language
		}
		if features, ok := t.Settings["features"].([]string); ok {
			settings.Features = features
		}

		result.Settings = settings
	}

	return result
}

func uintValue(u *uint) uint {
	if u == nil {
		return 0
	}
	return *u
}

// getActorIDFromContext extracts the user ID from the request context
func getActorIDFromContext(_ context.Context) uuid.UUID {
	// TODO: Extract from JWT claims when authentication is implemented
	// For now, return a system actor ID
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// classifyError determines the error type for metrics
func classifyError(err error) string {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid") || strings.Contains(errStr, "required"):
		return "validation"
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist"):
		return "not_found"
	case strings.Contains(errStr, "business") || strings.Contains(errStr, "conflict") || strings.Contains(errStr, "constraint"):
		return "business"
	default:
		return "internal"
	}
}

// convertServiceError converts service errors to GOA errors
func convertServiceError(err error) error {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid") || strings.Contains(errStr, "required"):
		return goaTenant.BadRequest(err.Error())
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist"):
		return goaTenant.NotFound(err.Error())
	case strings.Contains(errStr, "business") || strings.Contains(errStr, "conflict") || strings.Contains(errStr, "constraint"):
		return goaTenant.BadRequest(err.Error())
	default:
		return goaTenant.InternalError("Internal server error")
	}
}
