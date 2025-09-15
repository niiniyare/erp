package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/api/gen/tenant_management"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// TenantManagementHandler implements comprehensive tenant management API
type TenantManagementHandler struct {
	provisioningService tenant.ProvisioningService
	tracing            tracing.TracingService
	metrics            *metrics.MetricsService
}

// NewTenantManagementHandler creates a new tenant management handler
func NewTenantManagementHandler(
	provisioningService tenant.ProvisioningService,
	tracing tracing.TracingService,
	metrics *metrics.MetricsService,
) tenant_management.Service {
	return &TenantManagementHandler{
		provisioningService: provisioningService,
		tracing:            tracing,
		metrics:            metrics,
	}
}

// Provision implements comprehensive tenant provisioning
func (h *TenantManagementHandler) Provision(ctx context.Context, p *tenant_management.ProvisionTenantPayload) (*tenant_management.ProvisionTenantResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.Provision")
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("tenant_provision_duration", metrics.Fields{})
	defer timer.Stop()

	span.SetAttributes(
		attribute.String("tenant.name", p.Name),
		attribute.String("tenant.subdomain", p.Subdomain),
		attribute.String("plan.type", p.PlanType),
	)

	logger.InfoContext(ctx, "Processing tenant provisioning request", logger.Fields{
		"tenant_name": p.Name,
		"subdomain":   p.Subdomain,
		"plan_type":   p.PlanType,
		"operation":   "handler.Provision",
	})

	// Extract user ID from context for audit logging
	userID, _ := shared.GetUserID(ctx)

	// Convert GOA payload to service request
	serviceReq := tenant.ComprehensiveProvisionRequest{
		Name:        p.Name,
		Subdomain:   p.Subdomain,
		Description: stringValue(p.Description),
		Industry:    stringValue(p.Industry),
		CompanySize: stringValue(p.CompanySize),
		Country:     stringValue(p.Country),
		PlanType:    stringValue(p.PlanType),
		BillingCycle: stringValue(p.BillingCycle),
		Contact: tenant.ContactInfo{
			Name:  p.Contact.Name,
			Email: p.Contact.Email,
			Phone: stringValue(p.Contact.Phone),
			Title: stringValue(p.Contact.Title),
		},
		AdminUser: tenant.AdminUserRequest{
			Email:            p.AdminUser.Email,
			FirstName:        p.AdminUser.FirstName,
			LastName:         p.AdminUser.LastName,
			Phone:            stringValue(p.AdminUser.Phone),
			Timezone:         stringValue(p.AdminUser.Timezone),
			Language:         stringValue(p.AdminUser.Language),
			SendWelcomeEmail: boolValue(p.AdminUser.SendWelcomeEmail),
		},
	}

	// Convert initial settings if provided
	if p.InitialSettings != nil {
		serviceReq.InitialSettings = tenant.TenantInitialSettings{
			Timezone:         stringValue(p.InitialSettings.Timezone),
			Currency:         stringValue(p.InitialSettings.Currency),
			FiscalYearStart:  uint(uintValue(p.InitialSettings.FiscalYearStart)),
			DateFormat:       stringValue(p.InitialSettings.DateFormat),
			NumberFormat:     stringValue(p.InitialSettings.NumberFormat),
			Language:         stringValue(p.InitialSettings.Language),
			AccountingMethod: stringValue(p.InitialSettings.AccountingMethod),
		}
	}

	// Convert enabled modules
	if p.EnabledModules != nil {
		serviceReq.EnabledModules = *p.EnabledModules
	}

	// Convert initial limits if provided
	if p.InitialLimits != nil {
		serviceReq.InitialLimits = tenant.TenantLimits{
			MaxUsers:           uint(uintValue(p.InitialLimits.MaxUsers)),
			MaxStorageMB:       uint64(uintValue(p.InitialLimits.MaxStorageMb)),
			MaxAPICallsPerHour: uint(uintValue(p.InitialLimits.MaxAPICallsPerHour)),
		}
	}

	// Call provisioning service
	result, err := h.provisioningService.ProvisionTenantComplete(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_provision_errors", metrics.Fields{
			"error_type": classifyError(err),
		})

		logger.ErrorContext(ctx, "Tenant provisioning failed", logger.Fields{
			"tenant_name": p.Name,
			"subdomain":   p.Subdomain,
			"error":       err.Error(),
			"operation":   "handler.Provision",
		})

		return nil, h.convertServiceError(err)
	}

	// Convert service result to GOA response
	goaResult := &tenant_management.ProvisionTenantResult{
		Tenant: h.convertToDetailedTenantResult(result, ctx),
		SetupStatus: &tenant_management.TenantSetupStatus{
			OverallStatus:  result.SetupStatus.OverallStatus,
			CompletedSteps: result.SetupStatus.CompletedSteps,
			FailedSteps:    result.SetupStatus.FailedSteps,
			NextSteps:      result.SetupStatus.NextSteps,
		},
		AccessInfo: &tenant_management.TenantAccessInfo{
			TenantURL:        result.AccessInfo.TenantURL,
			AdminPortalURL:   result.AccessInfo.AdminPortalURL,
			APIBaseURL:       result.AccessInfo.APIBaseURL,
			DocumentationURL: stringPtr(result.AccessInfo.DocumentationURL),
		},
	}

	// Add admin user if created
	if result.AdminUser != nil {
		goaResult.AdminUser = &tenant_management.AdminUserResult{
			ID:               result.AdminUser.ID.String(),
			Email:            result.AdminUser.Email,
			FirstName:        result.AdminUser.FirstName,
			LastName:         result.AdminUser.LastName,
			Status:           result.AdminUser.Status,
			WelcomeEmailSent: result.AdminUser.WelcomeEmailSent,
			CreatedAt:        result.AdminUser.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        result.AdminUser.UpdatedAt.Format(time.RFC3339),
		}
	}

	h.metrics.IncrementCounter("tenant_provision_total", metrics.Fields{
		"status":    result.SetupStatus.OverallStatus,
		"plan_type": p.PlanType,
	})

	span.SetAttributes(
		attribute.String("result.tenant_id", result.TenantID.String()),
		attribute.String("result.status", result.SetupStatus.OverallStatus),
		attribute.Int("result.completed_steps", len(result.SetupStatus.CompletedSteps)),
	)

	logger.InfoContext(ctx, "Tenant provisioned successfully", logger.Fields{
		"tenant_id":       result.TenantID.String(),
		"tenant_name":     p.Name,
		"overall_status":  result.SetupStatus.OverallStatus,
		"completed_steps": len(result.SetupStatus.CompletedSteps),
		"failed_steps":    len(result.SetupStatus.FailedSteps),
		"operation":       "handler.Provision",
	})

	return goaResult, nil
}

// GetDetailed implements detailed tenant retrieval
func (h *TenantManagementHandler) GetDetailed(ctx context.Context, p *tenant_management.GetDetailedPayload) (*tenant_management.DetailedTenantResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.GetDetailed")
	defer span.End()

	timer := h.metrics.Timer("tenant_get_detailed_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		span.RecordError(err)
		logger.WarnContext(ctx, "Invalid tenant ID format", logger.Fields{
			"tenant_id": p.ID,
			"error":     err.Error(),
		})
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	// Get basic tenant information
	tenantInfo, err := h.provisioningService.GetTenantByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_get_detailed_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	// Build detailed result based on included data
	result := h.convertToDetailedTenantResult(nil, ctx)
	result.ID = tenantInfo.ID.String()
	result.Name = tenantInfo.Name
	result.Slug = tenantInfo.Slug
	result.Subdomain = tenantInfo.Subdomain
	result.Status = string(tenantInfo.Status)
	result.CreatedAt = tenantInfo.CreatedAt.Format(time.RFC3339)
	result.UpdatedAt = tenantInfo.UpdatedAt.Format(time.RFC3339)

	// Fetch additional data based on include parameter
	if p.Include != nil {
		for _, include := range *p.Include {
			switch include {
			case "configuration":
				config, err := h.provisioningService.GetTenantConfiguration(ctx, tenantID)
				if err != nil {
					logger.WarnContext(ctx, "Failed to get tenant configuration", logger.Fields{
						"tenant_id": tenantID.String(),
						"error":     err.Error(),
					})
				} else {
					result.Configuration = h.convertTenantConfiguration(config)
				}

			case "usage_stats":
				analytics, err := h.provisioningService.GetTenantUsageAnalytics(ctx, tenant.UsageAnalyticsRequest{
					TenantID: tenantID,
					Period:   "current_month",
					Metrics:  []string{"users", "storage", "api_calls"},
				})
				if err != nil {
					logger.WarnContext(ctx, "Failed to get usage analytics", logger.Fields{
						"tenant_id": tenantID.String(),
						"error":     err.Error(),
					})
				} else {
					result.UsageStats = h.convertUsageStats(analytics)
				}
			}
		}
	}

	h.metrics.IncrementCounter("tenant_get_detailed_total", metrics.Fields{})

	return result, nil
}

// Suspend implements tenant suspension
func (h *TenantManagementHandler) Suspend(ctx context.Context, p *tenant_management.SuspendPayload) (*tenant_management.TenantActionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.Suspend")
	defer span.End()

	timer := h.metrics.Timer("tenant_suspend_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	// Get actor ID from context
	actorID, _ := shared.GetUserID(ctx)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("actor.id", actorID.String()),
	)

	// Call service
	serviceReq := tenant.SuspendTenantRequest{
		TenantID:    tenantID,
		Reason:      p.Reason,
		ActorID:     actorID,
		NotifyUsers: boolValue(p.NotifyUsers),
	}

	result, err := h.provisioningService.SuspendTenant(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_suspend_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	h.metrics.IncrementCounter("tenant_suspend_total", metrics.Fields{})

	return h.convertActionResult(result), nil
}

// Reactivate implements tenant reactivation
func (h *TenantManagementHandler) Reactivate(ctx context.Context, p *tenant_management.ReactivatePayload) (*tenant_management.TenantActionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.Reactivate")
	defer span.End()

	timer := h.metrics.Timer("tenant_reactivate_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	// Get actor ID from context
	actorID, _ := shared.GetUserID(ctx)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("actor.id", actorID.String()),
	)

	// Call service
	serviceReq := tenant.ReactivateTenantRequest{
		TenantID: tenantID,
		Reason:   p.Reason,
		ActorID:  actorID,
	}

	result, err := h.provisioningService.ReactivateTenant(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_reactivate_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	h.metrics.IncrementCounter("tenant_reactivate_total", metrics.Fields{})

	return h.convertActionResult(result), nil
}

// Archive implements tenant archiving
func (h *TenantManagementHandler) Archive(ctx context.Context, p *tenant_management.ArchivePayload) (*tenant_management.TenantActionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.Archive")
	defer span.End()

	timer := h.metrics.Timer("tenant_archive_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	// Get actor ID from context
	actorID, _ := shared.GetUserID(ctx)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("actor.id", actorID.String()),
		attribute.Bool("immediate_deletion", boolValue(p.ImmediateDeletion)),
	)

	// Call service
	serviceReq := tenant.ArchiveTenantRequest{
		TenantID:            tenantID,
		Reason:              p.Reason,
		ActorID:             actorID,
		DataRetentionDays:   uint(uintValue(p.DataRetentionDays)),
		ImmediateDeletion:   boolValue(p.ImmediateDeletion),
	}

	result, err := h.provisioningService.ArchiveTenant(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_archive_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	h.metrics.IncrementCounter("tenant_archive_total", metrics.Fields{})

	return h.convertActionResult(result), nil
}

// UpdateConfiguration implements tenant configuration updates
func (h *TenantManagementHandler) UpdateConfiguration(ctx context.Context, p *tenant_management.UpdateConfigurationPayload) (*tenant_management.TenantConfigurationResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.UpdateConfiguration")
	defer span.End()

	timer := h.metrics.Timer("tenant_update_config_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	// Get actor ID from context
	actorID, _ := shared.GetUserID(ctx)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("actor.id", actorID.String()),
	)

	// Convert payload to service request
	serviceReq := tenant.UpdateTenantConfigurationRequest{
		TenantID: tenantID,
		Reason:   p.Reason,
		ActorID:  actorID,
	}

	// Convert limits if provided
	if p.Limits != nil {
		serviceReq.Limits = &tenant.TenantLimits{
			MaxUsers:           uint(uintValue(p.Limits.MaxUsers)),
			MaxStorageMB:       uint64(uintValue(p.Limits.MaxStorageMb)),
			MaxAPICallsPerHour: uint(uintValue(p.Limits.MaxAPICallsPerHour)),
		}
	}

	// Convert enabled features
	if p.EnabledFeatures != nil {
		serviceReq.EnabledFeatures = *p.EnabledFeatures
	}

	// Call service
	result, err := h.provisioningService.UpdateTenantConfiguration(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_update_config_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	h.metrics.IncrementCounter("tenant_update_config_total", metrics.Fields{})

	return h.convertTenantConfiguration(result), nil
}

// GetUsageAnalytics implements usage analytics retrieval
func (h *TenantManagementHandler) GetUsageAnalytics(ctx context.Context, p *tenant_management.GetUsageAnalyticsPayload) (*tenant_management.TenantUsageAnalyticsResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.GetUsageAnalytics")
	defer span.End()

	timer := h.metrics.Timer("tenant_usage_analytics_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("period", stringValue(p.Period)),
	)

	// Build service request
	serviceReq := tenant.UsageAnalyticsRequest{
		TenantID: tenantID,
		Period:   stringValue(p.Period),
	}

	if p.Metrics != nil {
		serviceReq.Metrics = *p.Metrics
	}

	// Call service
	result, err := h.provisioningService.GetTenantUsageAnalytics(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_usage_analytics_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	h.metrics.IncrementCounter("tenant_usage_analytics_total", metrics.Fields{})

	return h.convertUsageAnalytics(result), nil
}

// Helper methods for conversion and error handling

func (h *TenantManagementHandler) convertToDetailedTenantResult(provisionResult *tenant.ComprehensiveProvisionResult, ctx context.Context) *tenant_management.DetailedTenantResult {
	result := &tenant_management.DetailedTenantResult{}

	if provisionResult != nil {
		result.ID = provisionResult.TenantID.String()
	}

	return result
}

func (h *TenantManagementHandler) convertActionResult(result *tenant.TenantActionResult) *tenant_management.TenantActionResult {
	goaResult := &tenant_management.TenantActionResult{
		TenantID:      result.TenantID.String(),
		Action:        result.Action,
		Status:        result.Status,
		Message:       result.Message,
		EffectiveDate: result.EffectiveDate.Format(time.RFC3339),
	}

	if result.AuditLogID != nil {
		goaResult.AuditLogID = stringPtr(result.AuditLogID.String())
	}

	return goaResult
}

func (h *TenantManagementHandler) convertTenantConfiguration(config *tenant.TenantConfigurationResult) *tenant_management.TenantConfigurationResult {
	if config == nil {
		return nil
	}

	return &tenant_management.TenantConfigurationResult{
		Limits: &tenant_management.TenantLimits{
			MaxUsers:           uintPtr(config.Limits.MaxUsers),
			MaxStorageMb:       uint64Ptr(config.Limits.MaxStorageMB),
			MaxAPICallsPerHour: uintPtr(config.Limits.MaxAPICallsPerHour),
		},
		EnabledFeatures: config.EnabledFeatures,
	}
}

func (h *TenantManagementHandler) convertUsageStats(analytics *tenant.TenantUsageAnalyticsResult) *tenant_management.TenantUsageStatsResult {
	if analytics == nil {
		return nil
	}

	result := &tenant_management.TenantUsageStatsResult{
		PeriodStart:   time.Now().Format(time.RFC3339), // Should come from analytics
		PeriodEnd:     time.Now().Format(time.RFC3339), // Should come from analytics
	}

	if analytics.UserMetrics != nil {
		result.ActiveUsers = analytics.UserMetrics.ActiveUsers
		result.TotalEntities = analytics.UserMetrics.TotalUsers // Approximation
	}

	if analytics.StorageMetrics != nil {
		result.StorageUsedMb = analytics.StorageMetrics.TotalUsedMB
	}

	if analytics.APIMetrics != nil {
		result.APICalls = analytics.APIMetrics.TotalCalls
	}

	if analytics.FinancialMetrics != nil {
		result.TransactionsProcessed = uint64(analytics.FinancialMetrics.TransactionsCount)
	}

	return result
}

func (h *TenantManagementHandler) convertUsageAnalytics(analytics *tenant.TenantUsageAnalyticsResult) *tenant_management.TenantUsageAnalyticsResult {
	if analytics == nil {
		return nil
	}

	result := &tenant_management.TenantUsageAnalyticsResult{
		TenantID: analytics.TenantID.String(),
		Period:   analytics.Period,
	}

	// Convert user metrics
	if analytics.UserMetrics != nil {
		result.UserMetrics = &tenant_management.UserMetrics{
			TotalUsers:         analytics.UserMetrics.TotalUsers,
			ActiveUsers:        analytics.UserMetrics.ActiveUsers,
			NewUsers:           analytics.UserMetrics.NewUsers,
			AvgSessionDuration: analytics.UserMetrics.AvgSessionDuration,
		}
	}

	// Convert storage metrics
	if analytics.StorageMetrics != nil {
		result.StorageMetrics = &tenant_management.StorageMetrics{
			TotalUsedMb:    analytics.StorageMetrics.TotalUsedMB,
			DocumentsCount: analytics.StorageMetrics.DocumentsCount,
			MediaUsedMb:    analytics.StorageMetrics.MediaUsedMB,
			GrowthRate:     analytics.StorageMetrics.GrowthRate,
		}
	}

	// Convert API metrics
	if analytics.APIMetrics != nil {
		result.APIMetrics = &tenant_management.APIMetrics{
			TotalCalls:      analytics.APIMetrics.TotalCalls,
			SuccessfulCalls: analytics.APIMetrics.SuccessfulCalls,
			ErrorRate:       analytics.APIMetrics.ErrorRate,
			AvgResponseTime: analytics.APIMetrics.AvgResponseTime,
		}
	}

	// Convert financial metrics
	if analytics.FinancialMetrics != nil {
		result.FinancialMetrics = &tenant_management.FinancialMetrics{
			TransactionsCount: analytics.FinancialMetrics.TransactionsCount,
			TotalAmount: &tenant_management.Money{
				Amount:   analytics.FinancialMetrics.TotalAmount.Amount,
				Currency: analytics.FinancialMetrics.TotalAmount.Currency,
			},
			AvgTransactionAmount: &tenant_management.Money{
				Amount:   analytics.FinancialMetrics.AvgTransactionAmount.Amount,
				Currency: analytics.FinancialMetrics.AvgTransactionAmount.Currency,
			},
		}
	}

	// Convert feature usage
	if analytics.FeatureUsage != nil {
		result.FeatureUsage = &tenant_management.FeatureUsageMetrics{
			FinanceUsage:   analytics.FeatureUsage.FinanceUsage,
			InventoryUsage: analytics.FeatureUsage.InventoryUsage,
			HrUsage:        analytics.FeatureUsage.HRUsage,
			ReportingUsage: analytics.FeatureUsage.ReportingUsage,
		}
	}

	return result
}

func (h *TenantManagementHandler) convertServiceError(err error) error {
	// Convert service errors to appropriate Goa errors based on error type
	if sharedErrors.IsTenantNotFound(err) {
		return tenant_management.MakeNotFound(err)
	}

	if businessErr, ok := err.(*sharedErrors.BusinessError); ok {
		switch businessErr.Code {
		case "TENANT_ALREADY_SUSPENDED":
			return tenant_management.MakeConflict(err)
		case "TENANT_NOT_SUSPENDED":
			return tenant_management.MakeConflict(err)
		case "TENANT_ALREADY_ARCHIVED":
			return tenant_management.MakeConflict(err)
		case "INVALID_TENANT_NAME", "INVALID_SUBDOMAIN":
			return tenant_management.MakeBadRequest(err)
		case "SUBDOMAIN_ALREADY_EXISTS":
			return tenant_management.MakeConflict(err)
		}
	}

	// Default to unprocessable entity for other errors
	return tenant_management.MakeUnprocessableEntity(err)
}

// Helper functions for pointer conversions
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtr(s string) *string {
	return &s
}

func boolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func uintValue(u *uint) uint {
	if u == nil {
		return 0
	}
	return *u
}

func uintPtr(u uint) *uint {
	return &u
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}

func classifyError(err error) string {
	if sharedErrors.IsTenantNotFound(err) {
		return "not_found"
	}
	if businessErr, ok := err.(*sharedErrors.BusinessError); ok {
		return businessErr.Code
	}
	return "unknown"
}

// NotImplemented methods - will be implemented in follow-up iterations

func (h *TenantManagementHandler) BulkOperation(ctx context.Context, p *tenant_management.BulkOperationPayload) (*tenant_management.BatchResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.BulkOperation")
	defer span.End()

	timer := h.metrics.Timer("tenant_bulk_operation_duration", metrics.Fields{})
	defer timer.Stop()

	// Get actor ID from context
	actorID, _ := shared.GetUserID(ctx)

	span.SetAttributes(
		attribute.String("operation", p.Operation),
		attribute.Int("tenant_count", len(p.TenantIDs)),
		attribute.String("actor.id", actorID.String()),
	)

	// Convert tenant IDs from strings to UUIDs
	var tenantIDs []uuid.UUID
	for _, idStr := range p.TenantIDs {
		tenantID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format: %s", idStr))
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	// Build service request
	serviceReq := tenant.BulkTenantOperationRequest{
		Operation:  p.Operation,
		TenantIDs:  tenantIDs,
		Parameters: p.Parameters,
		ActorID:    actorID,
	}

	// Call service
	result, err := h.provisioningService.BulkTenantOperation(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_bulk_operation_errors", metrics.Fields{
			"error_type": classifyError(err),
			"operation":  p.Operation,
		})
		return nil, h.convertServiceError(err)
	}

	// Convert to GOA response
	goaResult := &tenant_management.BatchResult{
		OperationID: result.OperationID.String(),
		Status:      result.Status,
		Total:       result.Total,
		Successful:  result.Successful,
		Failed:      result.Failed,
		StartedAt:   result.StartedAt.Format(time.RFC3339),
	}

	if result.CompletedAt != nil {
		goaResult.CompletedAt = stringPtr(result.CompletedAt.Format(time.RFC3339))
	}

	// Convert individual results
	for _, itemResult := range result.Results {
		goaItem := tenant_management.BatchItemResult{
			TenantID: itemResult.TenantID.String(),
			Status:   itemResult.Status,
		}
		if itemResult.Message != "" {
			goaItem.Message = stringPtr(itemResult.Message)
		}
		if itemResult.Error != "" {
			goaItem.Error = stringPtr(itemResult.Error)
		}
		goaResult.Results = append(goaResult.Results, goaItem)
	}

	h.metrics.IncrementCounter("tenant_bulk_operation_total", metrics.Fields{
		"operation": p.Operation,
		"status":    result.Status,
	})

	return goaResult, nil
}

func (h *TenantManagementHandler) ListAdmin(ctx context.Context, p *tenant_management.ListAdminPayload) (*tenant_management.ListAdminResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.ListAdmin")
	defer span.End()

	timer := h.metrics.Timer("tenant_list_admin_duration", metrics.Fields{})
	defer timer.Stop()

	span.SetAttributes(
		attribute.Int("page", int(uintValue(p.Page))),
		attribute.Int("page_size", int(uintValue(p.PageSize))),
	)

	// Build advanced list request
	serviceReq := tenant.AdvancedListRequest{
		Page:     uintValue(p.Page),
		PageSize: uintValue(p.PageSize),
		SortBy:   stringValue(p.SortBy),
		SortOrder: stringValue(p.SortOrder),
		IncludeArchived: boolValue(p.IncludeArchived),
	}

	// Apply filters
	if p.StatusFilter != nil {
		serviceReq.StatusFilter = *p.StatusFilter
	}
	if p.PlanFilter != nil {
		serviceReq.PlanFilter = *p.PlanFilter
	}
	if p.Search != nil {
		serviceReq.Search = *p.Search
	}
	if p.CreatedAfter != nil {
		if parsedTime, err := time.Parse(time.RFC3339, *p.CreatedAfter); err == nil {
			serviceReq.CreatedAfter = &parsedTime
		}
	}
	if p.CreatedBefore != nil {
		if parsedTime, err := time.Parse(time.RFC3339, *p.CreatedBefore); err == nil {
			serviceReq.CreatedBefore = &parsedTime
		}
	}

	// Call service
	result, err := h.provisioningService.ListTenantsAdvanced(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_list_admin_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	// Convert to GOA response
	goaResult := &tenant_management.ListAdminResult{
		Data: make([]*tenant_management.DetailedTenantResult, len(result.Data)),
		Pagination: &tenant_management.PaginationMeta{
			CurrentPage: result.Pagination.CurrentPage,
			PageSize:    result.Pagination.PageSize,
			TotalItems:  result.Pagination.TotalItems,
			TotalPages:  result.Pagination.TotalPages,
			HasNext:     result.Pagination.HasNext,
			HasPrev:     result.Pagination.HasPrev,
		},
		Summary: &tenant_management.TenantSummaryStats{
			TotalTenants:     result.Summary.TotalTenants,
			ActiveTenants:    result.Summary.ActiveTenants,
			SuspendedTenants: result.Summary.SuspendedTenants,
			PendingTenants:   result.Summary.PendingTenants,
			ArchivedTenants:  result.Summary.ArchivedTenants,
			AvgUsersPerTenant: result.Summary.AvgUsersPerTenant,
		},
	}

	// Convert total revenue if available
	if result.Summary.TotalRevenue != nil {
		goaResult.Summary.TotalRevenue = &tenant_management.Money{
			Amount:   result.Summary.TotalRevenue.Amount,
			Currency: result.Summary.TotalRevenue.Currency,
		}
	}

	// Convert detailed tenant results
	for i, detailedTenant := range result.Data {
		goaTenant := &tenant_management.DetailedTenantResult{
			ID:        detailedTenant.ID.String(),
			Name:      detailedTenant.Name,
			Slug:      detailedTenant.Slug,
			Subdomain: detailedTenant.Subdomain,
			Status:    string(detailedTenant.Status),
			CreatedAt: detailedTenant.CreatedAt.Format(time.RFC3339),
			UpdatedAt: detailedTenant.UpdatedAt.Format(time.RFC3339),
		}

		// Add configuration if available
		if detailedTenant.Configuration != nil {
			goaTenant.Configuration = h.convertTenantConfiguration(detailedTenant.Configuration)
		}

		// Add usage stats if available
		if detailedTenant.UsageStats != nil {
			goaTenant.UsageStats = &tenant_management.TenantUsageStatsResult{
				PeriodStart:           detailedTenant.UsageStats.PeriodStart.Format(time.RFC3339),
				PeriodEnd:             detailedTenant.UsageStats.PeriodEnd.Format(time.RFC3339),
				ActiveUsers:           detailedTenant.UsageStats.ActiveUsers,
				TotalEntities:         detailedTenant.UsageStats.TotalEntities,
				StorageUsedMb:         detailedTenant.UsageStats.StorageUsedMB,
				APICalls:              detailedTenant.UsageStats.APICalls,
				TransactionsProcessed: detailedTenant.UsageStats.TransactionsProcessed,
			}
		}

		goaResult.Data[i] = goaTenant
	}

	h.metrics.IncrementCounter("tenant_list_admin_total", metrics.Fields{})

	return goaResult, nil
}

func (h *TenantManagementHandler) GetAuditLog(ctx context.Context, p *tenant_management.GetAuditLogPayload) (*tenant_management.GetAuditLogResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "tenant_management.handler.GetAuditLog")
	defer span.End()

	timer := h.metrics.Timer("tenant_audit_log_duration", metrics.Fields{})
	defer timer.Stop()

	// Parse tenant ID
	tenantID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, tenant_management.MakeBadRequest(fmt.Errorf("invalid tenant ID format"))
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("page", int(uintValue(p.Page))),
		attribute.Int("page_size", int(uintValue(p.PageSize))),
	)

	// Build service request
	serviceReq := tenant.AuditLogRequest{
		TenantID: tenantID,
		Page:     uintValue(p.Page),
		PageSize: uintValue(p.PageSize),
	}

	// Apply filters
	if p.ActionFilter != nil {
		serviceReq.ActionFilter = *p.ActionFilter
	}
	if p.DateFrom != nil {
		if parsedTime, err := time.Parse(time.RFC3339, *p.DateFrom); err == nil {
			serviceReq.DateFrom = &parsedTime
		}
	}
	if p.DateTo != nil {
		if parsedTime, err := time.Parse(time.RFC3339, *p.DateTo); err == nil {
			serviceReq.DateTo = &parsedTime
		}
	}

	// Call service
	result, err := h.provisioningService.GetTenantAuditLog(ctx, serviceReq)
	if err != nil {
		span.RecordError(err)
		h.metrics.IncrementCounter("tenant_audit_log_errors", metrics.Fields{
			"error_type": classifyError(err),
		})
		return nil, h.convertServiceError(err)
	}

	// Convert to GOA response
	goaResult := &tenant_management.GetAuditLogResult{
		Data: make([]*tenant_management.TenantAuditLogEntry, len(result.Data)),
		Pagination: &tenant_management.PaginationMeta{
			CurrentPage: result.Pagination.CurrentPage,
			PageSize:    result.Pagination.PageSize,
			TotalItems:  result.Pagination.TotalItems,
			TotalPages:  result.Pagination.TotalPages,
			HasNext:     result.Pagination.HasNext,
			HasPrev:     result.Pagination.HasPrev,
		},
	}

	// Convert audit log entries
	for i, entry := range result.Data {
		goaEntry := &tenant_management.TenantAuditLogEntry{
			ID:          entry.ID.String(),
			TenantID:    entry.TenantID.String(),
			Action:      entry.Action,
			ActorID:     entry.ActorID.String(),
			Description: entry.Description,
			Timestamp:   entry.Timestamp.Format(time.RFC3339),
			Metadata:    entry.Metadata,
		}

		if entry.ActorName != "" {
			goaEntry.ActorName = stringPtr(entry.ActorName)
		}
		if entry.IPAddress != "" {
			goaEntry.IPAddress = stringPtr(entry.IPAddress)
		}

		goaResult.Data[i] = goaEntry
	}

	h.metrics.IncrementCounter("tenant_audit_log_total", metrics.Fields{})

	return goaResult, nil
}