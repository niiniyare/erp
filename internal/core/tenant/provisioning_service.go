package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/identity"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ProvisioningService extends the basic tenant service with comprehensive provisioning capabilities
type ProvisioningService interface {
	Service

	// Enhanced provisioning methods
	ProvisionTenantComplete(ctx context.Context, req ComprehensiveProvisionRequest) (*ComprehensiveProvisionResult, error)
	SuspendTenant(ctx context.Context, req SuspendTenantRequest) (*TenantActionResult, error)
	ReactivateTenant(ctx context.Context, req ReactivateTenantRequest) (*TenantActionResult, error)
	ArchiveTenant(ctx context.Context, req ArchiveTenantRequest) (*TenantActionResult, error)

	// Configuration management
	UpdateTenantConfiguration(ctx context.Context, req UpdateTenantConfigurationRequest) (*TenantConfigurationResult, error)
	GetTenantConfiguration(ctx context.Context, tenantID uuid.UUID) (*TenantConfigurationResult, error)

	// Usage analytics
	GetTenantUsageAnalytics(ctx context.Context, req UsageAnalyticsRequest) (*TenantUsageAnalyticsResult, error)

	// Bulk operations
	BulkTenantOperation(ctx context.Context, req BulkTenantOperationRequest) (*BulkOperationResult, error)

	// Audit logging
	GetTenantAuditLog(ctx context.Context, req AuditLogRequest) (*AuditLogResult, error)
	LogTenantAction(ctx context.Context, req LogTenantActionRequest) error

	// Advanced listing with filters
	ListTenantsAdvanced(ctx context.Context, req AdvancedListRequest) (*AdvancedListResult, error)
}

// provisioningService implements ProvisioningService with full tenant lifecycle management
type provisioningService struct {
	Service            // Embed the basic tenant service
	identityService    identity.Service
	auditLogger        AuditLogger
	notificationSender NotificationSender
}

// NewProvisioningService creates a new comprehensive provisioning service
func NewProvisioningService(
	basicService Service,
	identityService identity.Service,
	auditLogger AuditLogger,
	notificationSender NotificationSender,
) ProvisioningService {
	return &provisioningService{
		Service:            basicService,
		identityService:    identityService,
		auditLogger:        auditLogger,
		notificationSender: notificationSender,
	}
}

// ProvisionTenantComplete implements comprehensive tenant provisioning
func (s *provisioningService) ProvisionTenantComplete(ctx context.Context, req ComprehensiveProvisionRequest) (*ComprehensiveProvisionResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.ProvisionTenantComplete")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.name", req.Name),
		attribute.String("tenant.subdomain", req.Subdomain),
		attribute.String("plan.type", req.PlanType),
	)

	logger.InfoContext(ctx, "Starting comprehensive tenant provisioning", logger.Fields{
		"tenant_name": req.Name,
		"subdomain":   req.Subdomain,
		"plan_type":   req.PlanType,
		"operation":   "provisioning.ProvisionTenantComplete",
	})

	result := &ComprehensiveProvisionResult{
		SetupStatus: TenantSetupStatus{
			OverallStatus:  "partial",
			CompletedSteps: []string{},
			FailedSteps:    []string{},
			NextSteps:      []string{},
		},
	}

	// Step 1: Provision the basic tenant using existing database function
	basicProvisionReq := ProvisionTenantRequest{
		Name:         req.Name,
		Email:        req.AdminUser.Email,
		Subdomain:    &req.Subdomain,
		Industry:     &req.Industry,
		CompanySize:  &req.CompanySize,
		CurrencyCode: req.InitialSettings.Currency,
		CountryCode:  req.Country,
	}

	provisionedInfo, err := s.Service.ProvisionTenant(ctx, basicProvisionReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Basic tenant provisioning failed")
		result.SetupStatus.FailedSteps = append(result.SetupStatus.FailedSteps, "tenant_creation")
		return result, fmt.Errorf("failed to provision basic tenant: %w", err)
	}

	result.SetupStatus.CompletedSteps = append(result.SetupStatus.CompletedSteps, "tenant_created")
	result.TenantID = provisionedInfo.TenantID

	// Step 2: Create admin user within tenant context
	adminUser, err := s.createAdminUser(ctx, provisionedInfo.TenantID, req.AdminUser)
	if err != nil {
		span.RecordError(err)
		result.SetupStatus.FailedSteps = append(result.SetupStatus.FailedSteps, "admin_user_creation")
		logger.ErrorContext(ctx, "Failed to create admin user", logger.Fields{
			"tenant_id": provisionedInfo.TenantID.String(),
			"error":     err.Error(),
		})
	} else {
		result.SetupStatus.CompletedSteps = append(result.SetupStatus.CompletedSteps, "admin_user_created")
		result.AdminUser = adminUser
	}

	// Step 3: Apply initial configuration
	configResult, err := s.applyInitialConfiguration(ctx, provisionedInfo.TenantID, req)
	if err != nil {
		span.RecordError(err)
		result.SetupStatus.FailedSteps = append(result.SetupStatus.FailedSteps, "configuration_setup")
		logger.ErrorContext(ctx, "Failed to apply initial configuration", logger.Fields{
			"tenant_id": provisionedInfo.TenantID.String(),
			"error":     err.Error(),
		})
	} else {
		result.SetupStatus.CompletedSteps = append(result.SetupStatus.CompletedSteps, "configuration_applied")
		result.Configuration = configResult
	}

	// Step 4: Enable requested modules
	if len(req.EnabledModules) > 0 {
		err := s.enableModules(ctx, provisionedInfo.TenantID, req.EnabledModules)
		if err != nil {
			span.RecordError(err)
			result.SetupStatus.FailedSteps = append(result.SetupStatus.FailedSteps, "module_enablement")
			logger.ErrorContext(ctx, "Failed to enable modules", logger.Fields{
				"tenant_id": provisionedInfo.TenantID.String(),
				"modules":   req.EnabledModules,
				"error":     err.Error(),
			})
		} else {
			result.SetupStatus.CompletedSteps = append(result.SetupStatus.CompletedSteps, "modules_enabled")
		}
	}

	// Step 5: Send welcome communications
	if req.AdminUser.SendWelcomeEmail {
		err := s.sendWelcomeEmail(ctx, provisionedInfo.TenantID, req.AdminUser)
		if err != nil {
			span.RecordError(err)
			result.SetupStatus.FailedSteps = append(result.SetupStatus.FailedSteps, "welcome_email")
			logger.WarnContext(ctx, "Failed to send welcome email", logger.Fields{
				"tenant_id": provisionedInfo.TenantID.String(),
				"email":     req.AdminUser.Email,
				"error":     err.Error(),
			})
		} else {
			result.SetupStatus.CompletedSteps = append(result.SetupStatus.CompletedSteps, "welcome_email_sent")
		}
	}

	// Step 6: Log the provisioning action
	err = s.LogTenantAction(ctx, LogTenantActionRequest{
		TenantID:    provisionedInfo.TenantID,
		Action:      "tenant_provisioned",
		ActorID:     uuid.Nil, // System action
		Description: fmt.Sprintf("Tenant '%s' provisioned with plan '%s'", req.Name, req.PlanType),
		Metadata: map[string]any{
			"plan_type":       req.PlanType,
			"enabled_modules": req.EnabledModules,
			"company_size":    req.CompanySize,
			"industry":        req.Industry,
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log provisioning action", logger.Fields{
			"tenant_id": provisionedInfo.TenantID.String(),
			"error":     err.Error(),
		})
	}

	// Generate access information
	result.AccessInfo = s.generateAccessInfo(req.Subdomain)

	// Determine overall status
	if len(result.SetupStatus.FailedSteps) == 0 {
		result.SetupStatus.OverallStatus = "completed"
		result.SetupStatus.NextSteps = []string{"verify_admin_email", "explore_dashboard", "configure_payment_method"}
	} else if len(result.SetupStatus.CompletedSteps) > 0 {
		result.SetupStatus.OverallStatus = "partial"
		result.SetupStatus.NextSteps = []string{"resolve_setup_issues", "contact_support"}
	} else {
		result.SetupStatus.OverallStatus = "failed"
		result.SetupStatus.NextSteps = []string{"retry_provisioning", "contact_support"}
	}

	span.SetAttributes(
		attribute.String("result.status", result.SetupStatus.OverallStatus),
		attribute.Int("result.completed_steps", len(result.SetupStatus.CompletedSteps)),
		attribute.Int("result.failed_steps", len(result.SetupStatus.FailedSteps)),
	)

	logger.InfoContext(ctx, "Tenant provisioning completed", logger.Fields{
		"tenant_id":       provisionedInfo.TenantID.String(),
		"overall_status":  result.SetupStatus.OverallStatus,
		"completed_steps": len(result.SetupStatus.CompletedSteps),
		"failed_steps":    len(result.SetupStatus.FailedSteps),
		"operation":       "provisioning.ProvisionTenantComplete",
	})

	return result, nil
}

// SuspendTenant implements tenant suspension with audit logging
func (s *provisioningService) SuspendTenant(ctx context.Context, req SuspendTenantRequest) (*TenantActionResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.SuspendTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", req.TenantID.String()))

	// Get current tenant to check status
	tenant, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check if tenant is already suspended
	if tenant.Status == StatusSuspended {
		return &TenantActionResult{
			TenantID:      req.TenantID,
			Action:        "suspend",
			Status:        "failed",
			Message:       "Tenant is already suspended",
			EffectiveDate: time.Now(),
		}, sharedErrors.NewBusinessError("TENANT_ALREADY_SUSPENDED", "Tenant is already suspended")
	}

	// Update tenant status to suspended
	suspendedStatus := StatusSuspended
	updateReq := UpdateTenantRequest{
		Status: &suspendedStatus,
	}

	_, err = s.Service.UpdateTenant(ctx, req.TenantID, updateReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
	}

	// Log the suspension action
	auditLogID := uuid.New()
	err = s.LogTenantAction(ctx, LogTenantActionRequest{
		ID:          auditLogID,
		TenantID:    req.TenantID,
		Action:      "tenant_suspended",
		ActorID:     req.ActorID,
		Description: fmt.Sprintf("Tenant suspended: %s", req.Reason),
		Metadata: map[string]any{
			"reason":          req.Reason,
			"previous_status": string(tenant.Status),
			"notify_users":    req.NotifyUsers,
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log suspension action", logger.Fields{
			"tenant_id": req.TenantID.String(),
			"error":     err.Error(),
		})
	}

	// Send notifications if requested
	if req.NotifyUsers {
		err = s.notifyTenantUsers(ctx, req.TenantID, "tenant_suspended", map[string]any{
			"reason": req.Reason,
		})
		if err != nil {
			logger.WarnContext(ctx, "Failed to notify users of suspension", logger.Fields{
				"tenant_id": req.TenantID.String(),
				"error":     err.Error(),
			})
		}
	}

	result := &TenantActionResult{
		TenantID:      req.TenantID,
		Action:        "suspend",
		Status:        "completed",
		Message:       "Tenant successfully suspended",
		EffectiveDate: time.Now(),
		AuditLogID:    &auditLogID,
	}

	logger.InfoContext(ctx, "Tenant suspended successfully", logger.Fields{
		"tenant_id": req.TenantID.String(),
		"reason":    req.Reason,
		"actor_id":  req.ActorID.String(),
	})

	return result, nil
}

// ReactivateTenant implements tenant reactivation
func (s *provisioningService) ReactivateTenant(ctx context.Context, req ReactivateTenantRequest) (*TenantActionResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.ReactivateTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", req.TenantID.String()))

	// Get current tenant to check status
	tenant, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check if tenant is suspended
	if tenant.Status != StatusSuspended {
		return &TenantActionResult{
			TenantID:      req.TenantID,
			Action:        "reactivate",
			Status:        "failed",
			Message:       "Tenant is not suspended",
			EffectiveDate: time.Now(),
		}, sharedErrors.NewBusinessError("TENANT_NOT_SUSPENDED", "Tenant is not suspended")
	}

	// Update tenant status to active
	activeStatus := StatusActive
	updateReq := UpdateTenantRequest{
		Status: &activeStatus,
	}

	_, err = s.Service.UpdateTenant(ctx, req.TenantID, updateReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to reactivate tenant: %w", err)
	}

	// Log the reactivation action
	auditLogID := uuid.New()
	err = s.LogTenantAction(ctx, LogTenantActionRequest{
		ID:          auditLogID,
		TenantID:    req.TenantID,
		Action:      "tenant_reactivated",
		ActorID:     req.ActorID,
		Description: fmt.Sprintf("Tenant reactivated: %s", req.Reason),
		Metadata: map[string]any{
			"reason":          req.Reason,
			"previous_status": string(tenant.Status),
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log reactivation action", logger.Fields{
			"tenant_id": req.TenantID.String(),
			"error":     err.Error(),
		})
	}

	result := &TenantActionResult{
		TenantID:      req.TenantID,
		Action:        "reactivate",
		Status:        "completed",
		Message:       "Tenant successfully reactivated",
		EffectiveDate: time.Now(),
		AuditLogID:    &auditLogID,
	}

	logger.InfoContext(ctx, "Tenant reactivated successfully", logger.Fields{
		"tenant_id": req.TenantID.String(),
		"reason":    req.Reason,
		"actor_id":  req.ActorID.String(),
	})

	return result, nil
}

// ArchiveTenant implements tenant archiving with data retention policies
func (s *provisioningService) ArchiveTenant(ctx context.Context, req ArchiveTenantRequest) (*TenantActionResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.ArchiveTenant")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", req.TenantID.String()),
		attribute.Bool("immediate_deletion", req.ImmediateDeletion),
		attribute.Int("retention_days", int(req.DataRetentionDays)),
	)

	// Get current tenant
	tenant, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check if tenant is already archived
	if tenant.DeletedAt != nil {
		return &TenantActionResult{
			TenantID:      req.TenantID,
			Action:        "archive",
			Status:        "failed",
			Message:       "Tenant is already archived",
			EffectiveDate: time.Now(),
		}, sharedErrors.NewBusinessError("TENANT_ALREADY_ARCHIVED", "Tenant is already archived")
	}

	// Perform soft delete
	err = s.Service.DeleteTenant(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to archive tenant: %w", err)
	}

	// Schedule data retention cleanup if not immediate deletion
	var scheduledDeletion *time.Time
	if !req.ImmediateDeletion && req.DataRetentionDays > 0 {
		deletionDate := time.Now().AddDate(0, 0, int(req.DataRetentionDays))
		scheduledDeletion = &deletionDate
		// TODO: Schedule background job for permanent deletion
	}

	// Log the archiving action
	auditLogID := uuid.New()
	err = s.LogTenantAction(ctx, LogTenantActionRequest{
		ID:          auditLogID,
		TenantID:    req.TenantID,
		Action:      "tenant_archived",
		ActorID:     req.ActorID,
		Description: fmt.Sprintf("Tenant archived: %s", req.Reason),
		Metadata: map[string]any{
			"reason":              req.Reason,
			"immediate_deletion":  req.ImmediateDeletion,
			"data_retention_days": req.DataRetentionDays,
			"scheduled_deletion":  scheduledDeletion,
			"previous_status":     string(tenant.Status),
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log archiving action", logger.Fields{
			"tenant_id": req.TenantID.String(),
			"error":     err.Error(),
		})
	}

	result := &TenantActionResult{
		TenantID:      req.TenantID,
		Action:        "archive",
		Status:        "completed",
		Message:       "Tenant successfully archived",
		EffectiveDate: time.Now(),
		AuditLogID:    &auditLogID,
	}

	logger.InfoContext(ctx, "Tenant archived successfully", logger.Fields{
		"tenant_id":           req.TenantID.String(),
		"reason":              req.Reason,
		"actor_id":            req.ActorID.String(),
		"immediate_deletion":  req.ImmediateDeletion,
		"data_retention_days": req.DataRetentionDays,
	})

	return result, nil
}

// Helper methods for provisioning workflow

func (s *provisioningService) createAdminUser(ctx context.Context, tenantID uuid.UUID, adminReq AdminUserRequest) (*AdminUserResult, error) {
	// Set tenant context for user creation
	err := s.Service.WithTenantContext(ctx, tenantID, func(tenantCtx context.Context) error {
		// Create user using identity service within tenant context
		// Note: For now, we'll create a basic user record
		createUserReq := identity.CreateUserRequest{
			EntityID:              uuid.New(),     // This should come from entity service
			Username:              adminReq.Email, // Use email as username for now
			Email:                 adminReq.Email,
			Password:              "temp_password", // This should be generated or provided
			UserType:              "admin",
			AccountStatus:         "ACTIVE",
			SessionTimeoutMinutes: 480, // 8 hours
			MfaEnabled:            false,
		}

		_, err := s.identityService.RegisterNewUser(tenantCtx, &createUserReq)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	return &AdminUserResult{
		ID:               uuid.New(), // This should come from the identity service
		Email:            adminReq.Email,
		FirstName:        adminReq.FirstName,
		LastName:         adminReq.LastName,
		Status:           "pending_verification",
		WelcomeEmailSent: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

func (s *provisioningService) applyInitialConfiguration(ctx context.Context, tenantID uuid.UUID, req ComprehensiveProvisionRequest) (*TenantConfigurationResult, error) {
	// Apply initial configuration using the existing configuration system
	// This would integrate with your tenant configuration tables
	return &TenantConfigurationResult{
		Limits: TenantLimits{
			MaxUsers:           req.InitialLimits.MaxUsers,
			MaxStorageMB:       req.InitialLimits.MaxStorageMB,
			MaxAPICallsPerHour: req.InitialLimits.MaxAPICallsPerHour,
		},
		EnabledFeatures: req.EnabledModules,
	}, nil
}

func (s *provisioningService) enableModules(ctx context.Context, tenantID uuid.UUID, modules []string) error {
	// Enable specified modules for the tenant
	// This would integrate with your module management system
	logger.InfoContext(ctx, "Enabling modules for tenant", logger.Fields{
		"tenant_id": tenantID.String(),
		"modules":   modules,
	})
	return nil
}

func (s *provisioningService) sendWelcomeEmail(ctx context.Context, tenantID uuid.UUID, adminUser AdminUserRequest) error {
	// Send welcome email using notification service
	if s.notificationSender != nil {
		return s.notificationSender.SendWelcomeEmail(ctx, tenantID, adminUser.Email, adminUser.FirstName)
	}
	return nil
}

func (s *provisioningService) generateAccessInfo(subdomain string) TenantAccessInfo {
	baseURL := "https://yourdomain.com" // This should come from configuration
	if subdomain != "" {
		baseURL = fmt.Sprintf("https://%s.yourdomain.com", subdomain)
	}

	return TenantAccessInfo{
		TenantURL:        baseURL,
		AdminPortalURL:   fmt.Sprintf("%s/admin", baseURL),
		APIBaseURL:       fmt.Sprintf("%s/api/v1", baseURL),
		DocumentationURL: "https://docs.yourdomain.com",
	}
}

func (s *provisioningService) notifyTenantUsers(ctx context.Context, tenantID uuid.UUID, eventType string, data map[string]any) error {
	if s.notificationSender != nil {
		return s.notificationSender.NotifyTenantUsers(ctx, tenantID, eventType, data)
	}
	return nil
}

func (s *provisioningService) getTracer() tracing.TracingService {
	// This assumes the embedded Service has access to the tracer
	// You may need to adjust this based on your actual service structure
	if basicService, ok := s.Service.(*service); ok {
		return basicService.tracer
	}
	// Return a no-op tracer if not available
	return &noOpTracer{}
}

// NoOpTracer is a fallback tracer implementation
type noOpTracer struct{}

func (n *noOpTracer) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	return ctx, &noOpSpan{}
}

func (n *noOpTracer) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {}
func (n *noOpTracer) SpanFromContext(ctx context.Context) tracing.Span                       { return &noOpSpan{} }
func (n *noOpTracer) InjectHTTPHeaders(ctx context.Context, headers http.Header)             {}
func (n *noOpTracer) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	return ctx
}
func (n *noOpTracer) RecordError(ctx context.Context, err error, opts ...tracing.ErrorOption) {}
func (n *noOpTracer) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue)          {}
func (n *noOpTracer) GetTraceID(ctx context.Context) string                                   { return "" }
func (n *noOpTracer) GetSpanID(ctx context.Context) string                                    { return "" }
func (n *noOpTracer) Shutdown(ctx context.Context) error                                      { return nil }

type noOpSpan struct{}

func (n *noOpSpan) End(opts ...tracing.SpanEndOption)                 {}
func (n *noOpSpan) SetAttributes(attrs ...attribute.KeyValue)         {}
func (n *noOpSpan) RecordError(err error, opts ...trace.EventOption)  {}
func (n *noOpSpan) SetStatus(code codes.Code, description string)     {}
func (n *noOpSpan) AddEvent(name string, attrs ...attribute.KeyValue) {}
func (n *noOpSpan) SetName(name string)                               {}
func (n *noOpSpan) IsRecording() bool                                 { return false }
func (n *noOpSpan) SpanContext() trace.SpanContext                    { return trace.SpanContext{} }

// NotImplemented methods - will be implemented in follow-up

func (s *provisioningService) UpdateTenantConfiguration(ctx context.Context, req UpdateTenantConfigurationRequest) (*TenantConfigurationResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.UpdateTenantConfiguration")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", req.TenantID.String()),
		attribute.String("actor.id", req.ActorID.String()),
	)

	// Validate tenant exists
	tenant, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check if tenant is active
	if tenant.Status != StatusActive {
		return nil, sharedErrors.NewBusinessError("TENANT_NOT_ACTIVE", "Cannot update configuration for inactive tenant")
	}

	// Build configuration update
	configUpdate := &TenantConfigurationResult{
		UpdatedAt: time.Now(),
		UpdatedBy: &req.ActorID,
	}

	// Update limits if provided
	if req.Limits != nil {
		configUpdate.Limits = *req.Limits

		// Validate limits are reasonable
		if err := s.validateTenantLimits(*req.Limits); err != nil {
			span.RecordError(err)
			return nil, err
		}
	} else {
		// Get current limits
		currentConfig, err := s.getCurrentConfiguration(ctx, req.TenantID)
		if err != nil {
			// Use default limits if no current config
			configUpdate.Limits = s.getDefaultLimits()
		} else {
			configUpdate.Limits = currentConfig.Limits
		}
	}

	// Update enabled features if provided
	if req.EnabledFeatures != nil {
		configUpdate.EnabledFeatures = req.EnabledFeatures
	} else {
		// Keep current features
		currentConfig, err := s.getCurrentConfiguration(ctx, req.TenantID)
		if err == nil && currentConfig != nil {
			configUpdate.EnabledFeatures = currentConfig.EnabledFeatures
		}
	}

	// Update security policies if provided
	if req.SecurityPolicies != nil {
		configUpdate.SecurityPolicies = req.SecurityPolicies
	}

	// Update notification preferences if provided
	if req.NotificationPreferences != nil {
		configUpdate.NotificationPreferences = req.NotificationPreferences
	}

	// Store configuration in database
	// In a real implementation, this would use a configuration repository
	err = s.storeConfiguration(ctx, req.TenantID, configUpdate)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to store configuration: %w", err)
	}

	// Log configuration update
	auditLogID := uuid.New()
	err = s.LogTenantAction(ctx, LogTenantActionRequest{
		ID:          auditLogID,
		TenantID:    req.TenantID,
		Action:      "configuration_updated",
		ActorID:     req.ActorID,
		Description: fmt.Sprintf("Tenant configuration updated: %s", req.Reason),
		Metadata: map[string]any{
			"reason":           req.Reason,
			"limits_updated":   req.Limits != nil,
			"features_updated": req.EnabledFeatures != nil,
			"security_updated": req.SecurityPolicies != nil,
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log configuration update", logger.Fields{
			"tenant_id": req.TenantID.String(),
			"error":     err.Error(),
		})
	}

	logger.InfoContext(ctx, "Tenant configuration updated successfully", logger.Fields{
		"tenant_id": req.TenantID.String(),
		"actor_id":  req.ActorID.String(),
		"reason":    req.Reason,
	})

	return configUpdate, nil
}

func (s *provisioningService) GetTenantConfiguration(ctx context.Context, tenantID uuid.UUID) (*TenantConfigurationResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.GetTenantConfiguration")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	// Validate tenant exists
	_, err := s.Service.GetTenantByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Get configuration from storage
	config, err := s.getCurrentConfiguration(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		// Return default configuration if none exists
		logger.InfoContext(ctx, "No configuration found for tenant, returning defaults", logger.Fields{
			"tenant_id": tenantID.String(),
		})

		return &TenantConfigurationResult{
			Limits:          s.getDefaultLimits(),
			EnabledFeatures: []string{"finance"}, // Default enabled features
			UpdatedAt:       time.Now(),
		}, nil
	}

	return config, nil
}

func (s *provisioningService) GetTenantUsageAnalytics(ctx context.Context, req UsageAnalyticsRequest) (*TenantUsageAnalyticsResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.GetTenantUsageAnalytics")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", req.TenantID.String()),
		attribute.String("period", req.Period),
	)

	// Validate tenant exists
	_, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Calculate date range based on period
	startDate, endDate := s.calculatePeriodDates(req.Period)

	// Initialize result
	result := &TenantUsageAnalyticsResult{
		TenantID: req.TenantID,
		Period:   req.Period,
	}

	// Collect metrics based on requested types
	for _, metric := range req.Metrics {
		switch metric {
		case "users":
			userMetrics, err := s.getUserMetrics(ctx, req.TenantID, startDate, endDate)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get user metrics", logger.Fields{
					"tenant_id": req.TenantID.String(),
					"error":     err.Error(),
				})
			} else {
				result.UserMetrics = userMetrics
			}

		case "storage":
			storageMetrics, err := s.getStorageMetrics(ctx, req.TenantID, startDate, endDate)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get storage metrics", logger.Fields{
					"tenant_id": req.TenantID.String(),
					"error":     err.Error(),
				})
			} else {
				result.StorageMetrics = storageMetrics
			}

		case "api_calls":
			apiMetrics, err := s.getAPIMetrics(ctx, req.TenantID, startDate, endDate)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get API metrics", logger.Fields{
					"tenant_id": req.TenantID.String(),
					"error":     err.Error(),
				})
			} else {
				result.APIMetrics = apiMetrics
			}

		case "transactions":
			financialMetrics, err := s.getFinancialMetrics(ctx, req.TenantID, startDate, endDate)
			if err != nil {
				logger.WarnContext(ctx, "Failed to get financial metrics", logger.Fields{
					"tenant_id": req.TenantID.String(),
					"error":     err.Error(),
				})
			} else {
				result.FinancialMetrics = financialMetrics
			}
		}
	}

	// Get feature usage metrics
	featureUsage, err := s.getFeatureUsageMetrics(ctx, req.TenantID, startDate, endDate)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get feature usage metrics", logger.Fields{
			"tenant_id": req.TenantID.String(),
			"error":     err.Error(),
		})
	} else {
		result.FeatureUsage = featureUsage
	}

	logger.InfoContext(ctx, "Retrieved tenant usage analytics", logger.Fields{
		"tenant_id": req.TenantID.String(),
		"period":    req.Period,
		"metrics":   req.Metrics,
	})

	return result, nil
}

func (s *provisioningService) BulkTenantOperation(ctx context.Context, req BulkTenantOperationRequest) (*BulkOperationResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.BulkTenantOperation")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", req.Operation),
		attribute.Int("tenant_count", len(req.TenantIDs)),
		attribute.String("actor.id", req.ActorID.String()),
	)

	if len(req.TenantIDs) == 0 {
		return nil, sharedErrors.NewBusinessError("INVALID_REQUEST", "No tenant IDs provided")
	}

	if len(req.TenantIDs) > 100 {
		return nil, sharedErrors.NewBusinessError("INVALID_REQUEST", "Cannot process more than 100 tenants at once")
	}

	operationID := uuid.New()
	startTime := time.Now()

	result := &BulkOperationResult{
		OperationID: operationID,
		Status:      "in_progress",
		Total:       uint(len(req.TenantIDs)),
		Successful:  0,
		Failed:      0,
		Results:     make([]BulkOperationItemResult, 0, len(req.TenantIDs)),
		StartedAt:   startTime,
	}

	logger.InfoContext(ctx, "Starting bulk tenant operation", logger.Fields{
		"operation_id": operationID.String(),
		"operation":    req.Operation,
		"tenant_count": len(req.TenantIDs),
		"actor_id":     req.ActorID.String(),
	})

	// Process each tenant
	for _, tenantID := range req.TenantIDs {
		itemResult := BulkOperationItemResult{
			TenantID: tenantID,
			Status:   "processing",
		}

		// Execute operation based on type
		switch req.Operation {
		case "suspend":
			err := s.executeBulkSuspend(ctx, tenantID, req)
			if err != nil {
				itemResult.Status = "failed"
				itemResult.Error = err.Error()
				result.Failed++
			} else {
				itemResult.Status = "completed"
				itemResult.Message = "Tenant suspended successfully"
				result.Successful++
			}

		case "reactivate":
			err := s.executeBulkReactivate(ctx, tenantID, req)
			if err != nil {
				itemResult.Status = "failed"
				itemResult.Error = err.Error()
				result.Failed++
			} else {
				itemResult.Status = "completed"
				itemResult.Message = "Tenant reactivated successfully"
				result.Successful++
			}

		case "archive":
			err := s.executeBulkArchive(ctx, tenantID, req)
			if err != nil {
				itemResult.Status = "failed"
				itemResult.Error = err.Error()
				result.Failed++
			} else {
				itemResult.Status = "completed"
				itemResult.Message = "Tenant archived successfully"
				result.Successful++
			}

		case "update_limits":
			err := s.executeBulkUpdateLimits(ctx, tenantID, req)
			if err != nil {
				itemResult.Status = "failed"
				itemResult.Error = err.Error()
				result.Failed++
			} else {
				itemResult.Status = "completed"
				itemResult.Message = "Tenant limits updated successfully"
				result.Successful++
			}

		default:
			itemResult.Status = "failed"
			itemResult.Error = fmt.Sprintf("Unknown operation: %s", req.Operation)
			result.Failed++
		}

		result.Results = append(result.Results, itemResult)
	}

	// Update final status
	completedTime := time.Now()
	result.CompletedAt = &completedTime
	if result.Failed == 0 {
		result.Status = "completed"
	} else if result.Successful == 0 {
		result.Status = "failed"
	} else {
		result.Status = "partial_success"
	}

	// Log bulk operation completion
	auditLogID := uuid.New()
	err := s.LogTenantAction(ctx, LogTenantActionRequest{
		ID:          auditLogID,
		TenantID:    uuid.Nil, // System-level operation
		Action:      fmt.Sprintf("bulk_%s", req.Operation),
		ActorID:     req.ActorID,
		Description: fmt.Sprintf("Bulk operation %s completed: %d successful, %d failed", req.Operation, result.Successful, result.Failed),
		Metadata: map[string]any{
			"operation_id":   operationID.String(),
			"operation_type": req.Operation,
			"total_tenants":  result.Total,
			"successful":     result.Successful,
			"failed":         result.Failed,
			"tenant_ids":     req.TenantIDs,
		},
	})
	if err != nil {
		logger.WarnContext(ctx, "Failed to log bulk operation", logger.Fields{
			"operation_id": operationID.String(),
			"error":        err.Error(),
		})
	}

	logger.InfoContext(ctx, "Bulk tenant operation completed", logger.Fields{
		"operation_id": operationID.String(),
		"operation":    req.Operation,
		"status":       result.Status,
		"successful":   result.Successful,
		"failed":       result.Failed,
		"duration":     completedTime.Sub(startTime).String(),
	})

	return result, nil
}

func (s *provisioningService) GetTenantAuditLog(ctx context.Context, req AuditLogRequest) (*AuditLogResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.GetTenantAuditLog")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", req.TenantID.String()),
		attribute.Int("page", int(req.Page)),
		attribute.Int("page_size", int(req.PageSize)),
	)

	// Validate tenant exists
	_, err := s.Service.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Set defaults for pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Get audit log entries from storage
	entries, totalCount, err := s.getAuditLogEntries(ctx, req.TenantID, req.ActionFilter, req.DateFrom, req.DateTo, offset, req.PageSize)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to retrieve audit log: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.PageSize - 1) / req.PageSize
	pagination := PaginationMeta{
		CurrentPage: req.Page,
		PageSize:    req.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     req.Page < totalPages,
		HasPrev:     req.Page > 1,
	}

	result := &AuditLogResult{
		Data:       entries,
		Pagination: pagination,
	}

	logger.InfoContext(ctx, "Retrieved tenant audit log", logger.Fields{
		"tenant_id":   req.TenantID.String(),
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_items": totalCount,
		"entry_count": len(entries),
	})

	return result, nil
}

func (s *provisioningService) LogTenantAction(ctx context.Context, req LogTenantActionRequest) error {
	if s.auditLogger != nil {
		return s.auditLogger.LogAction(ctx, req)
	}
	return nil
}

func (s *provisioningService) ListTenantsAdvanced(ctx context.Context, req AdvancedListRequest) (*AdvancedListResult, error) {
	ctx, span := s.getTracer().StartSpan(ctx, "tenant.provisioning.ListTenantsAdvanced")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page", int(req.Page)),
		attribute.Int("page_size", int(req.PageSize)),
		attribute.String("search", req.Search),
	)

	// Set defaults for pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Get filtered tenants from storage
	tenants, totalCount, err := s.getFilteredTenants(ctx, req, offset, req.PageSize)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to retrieve tenants: %w", err)
	}

	// Build detailed results
	var detailedResults []DetailedTenantResult
	for _, tenant := range tenants {
		detailed := DetailedTenantResult{
			Tenant: tenant,
		}

		// Add configuration if requested
		config, err := s.GetTenantConfiguration(ctx, tenant.ID)
		if err == nil {
			detailed.Configuration = config
		}

		// Add usage stats
		usageStats, err := s.getTenantUsageStats(ctx, tenant.ID)
		if err == nil {
			detailed.UsageStats = usageStats
		}

		// Add subscription details (mock for now)
		detailed.SubscriptionDetails = &SubscriptionDetails{
			Plan:         "professional",
			Status:       "active",
			BillingCycle: "monthly",
		}

		// Add security settings (mock for now)
		detailed.SecuritySettings = &SecuritySettings{
			SecurityScore:   85,
			EnabledFeatures: []string{"mfa", "encryption", "audit_logs"},
		}

		detailedResults = append(detailedResults, detailed)
	}

	// Generate summary statistics
	summary, err := s.generateTenantSummaryStats(ctx)
	if err != nil {
		logger.WarnContext(ctx, "Failed to generate summary stats", logger.Fields{
			"error": err.Error(),
		})
		// Use empty summary if generation fails
		summary = &TenantSummaryStats{}
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.PageSize - 1) / req.PageSize
	pagination := PaginationMeta{
		CurrentPage: req.Page,
		PageSize:    req.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     req.Page < totalPages,
		HasPrev:     req.Page > 1,
	}

	result := &AdvancedListResult{
		Data:       detailedResults,
		Pagination: pagination,
		Summary:    *summary,
	}

	logger.InfoContext(ctx, "Retrieved advanced tenant list", logger.Fields{
		"page":         req.Page,
		"page_size":    req.PageSize,
		"total_items":  totalCount,
		"result_count": len(detailedResults),
		"filters": map[string]any{
			"status":   req.StatusFilter,
			"plan":     req.PlanFilter,
			"search":   req.Search,
			"archived": req.IncludeArchived,
		},
	})

	return result, nil
}

// Helper methods for configuration management

func (s *provisioningService) validateTenantLimits(limits TenantLimits) error {
	// Validate tenant limits are within reasonable bounds
	if limits.MaxUsers > 10000 {
		return sharedErrors.NewBusinessError("INVALID_LIMITS", "Maximum users cannot exceed 10,000")
	}
	if limits.MaxStorageMB > 1000000 { // 1TB
		return sharedErrors.NewBusinessError("INVALID_LIMITS", "Maximum storage cannot exceed 1TB")
	}
	if limits.MaxAPICallsPerHour > 1000000 {
		return sharedErrors.NewBusinessError("INVALID_LIMITS", "Maximum API calls per hour cannot exceed 1,000,000")
	}
	return nil
}

func (s *provisioningService) getDefaultLimits() TenantLimits {
	return TenantLimits{
		MaxUsers:           100,
		MaxStorageMB:       10240, // 10GB
		MaxAPICallsPerHour: 10000,
	}
}

func (s *provisioningService) getCurrentConfiguration(ctx context.Context, tenantID uuid.UUID) (*TenantConfigurationResult, error) {
	// In a real implementation, this would query the tenant_configurations table
	// For now, return a mock configuration
	return &TenantConfigurationResult{
		Limits:          s.getDefaultLimits(),
		EnabledFeatures: []string{"finance"},
		UpdatedAt:       time.Now(),
	}, nil
}

func (s *provisioningService) storeConfiguration(ctx context.Context, tenantID uuid.UUID, config *TenantConfigurationResult) error {
	// In a real implementation, this would store to tenant_configurations table
	logger.InfoContext(ctx, "Storing tenant configuration", logger.Fields{
		"tenant_id": tenantID.String(),
		"limits":    config.Limits,
		"features":  config.EnabledFeatures,
	})
	return nil
}

// Helper methods for usage analytics

func (s *provisioningService) calculatePeriodDates(period string) (time.Time, time.Time) {
	now := time.Now()
	switch period {
	case "current_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	case "last_month":
		lastMonth := now.AddDate(0, -1, 0)
		start := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, -1)
		return start, end
	case "last_3_months":
		start := now.AddDate(0, -3, 0)
		return start, now
	case "last_year":
		start := now.AddDate(-1, 0, 0)
		return start, now
	default:
		// Default to current month
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	}
}

func (s *provisioningService) getUserMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*UserMetrics, error) {
	// In a real implementation, this would query user activity tables
	return &UserMetrics{
		TotalUsers:         25,
		ActiveUsers:        18,
		NewUsers:           3,
		AvgSessionDuration: 1800, // 30 minutes in seconds
	}, nil
}

func (s *provisioningService) getStorageMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*StorageMetrics, error) {
	// In a real implementation, this would query storage usage tables
	return &StorageMetrics{
		TotalUsedMB:    2048, // 2GB
		DocumentsCount: 156,
		MediaUsedMB:    512,  // 512MB
		GrowthRate:     0.15, // 15% growth
	}, nil
}

func (s *provisioningService) getAPIMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*APIMetrics, error) {
	// In a real implementation, this would query API usage logs
	return &APIMetrics{
		TotalCalls:      50000,
		SuccessfulCalls: 48500,
		ErrorRate:       0.03, // 3% error rate
		AvgResponseTime: 250,  // 250ms
	}, nil
}

func (s *provisioningService) getFinancialMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*FinancialMetrics, error) {
	// In a real implementation, this would query financial transaction tables
	return &FinancialMetrics{
		TransactionsCount: 1250,
		TotalAmount: Money{
			Amount:   "125000.00",
			Currency: "USD",
		},
		AvgTransactionAmount: Money{
			Amount:   "100.00",
			Currency: "USD",
		},
	}, nil
}

func (s *provisioningService) getFeatureUsageMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*FeatureUsageMetrics, error) {
	// In a real implementation, this would query feature usage tables
	return &FeatureUsageMetrics{
		FinanceUsage:   850,
		InventoryUsage: 320,
		HRUsage:        180,
		ReportingUsage: 95,
	}, nil
}

// Helper methods for audit logging

func (s *provisioningService) getAuditLogEntries(ctx context.Context, tenantID uuid.UUID, actionFilter []string, dateFrom, dateTo *time.Time, offset, pageSize uint) ([]TenantAuditLogEntry, uint, error) {
	// In a real implementation, this would query the audit_logs table
	// For now, return mock data
	mockEntries := []TenantAuditLogEntry{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Action:      "tenant_created",
			ActorID:     uuid.New(),
			ActorName:   "System Administrator",
			Description: "Tenant was created successfully",
			Metadata: map[string]any{
				"plan_type": "professional",
				"subdomain": "acme-corp",
			},
			Timestamp: time.Now().Add(-24 * time.Hour),
			IPAddress: "192.168.1.100",
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Action:      "configuration_updated",
			ActorID:     uuid.New(),
			ActorName:   "Admin User",
			Description: "Tenant configuration was updated",
			Metadata: map[string]any{
				"reason":           "Increased user limit",
				"limits_updated":   true,
				"features_updated": false,
			},
			Timestamp: time.Now().Add(-12 * time.Hour),
			IPAddress: "192.168.1.101",
		},
	}

	// Apply filters in a real implementation
	filteredEntries := mockEntries
	if len(actionFilter) > 0 {
		// Filter by action types
		var filtered []TenantAuditLogEntry
		for _, entry := range mockEntries {
			for _, filter := range actionFilter {
				if entry.Action == filter {
					filtered = append(filtered, entry)
					break
				}
			}
		}
		filteredEntries = filtered
	}

	// Apply date filters in a real implementation
	if dateFrom != nil || dateTo != nil {
		// Apply date filtering logic
	}

	// Apply pagination
	totalCount := uint(len(filteredEntries))
	start := offset
	end := offset + pageSize
	if start > uint(len(filteredEntries)) {
		start = uint(len(filteredEntries))
	}
	if end > uint(len(filteredEntries)) {
		end = uint(len(filteredEntries))
	}

	paginatedEntries := filteredEntries[start:end]

	return paginatedEntries, totalCount, nil
}

// Helper methods for bulk operations

func (s *provisioningService) executeBulkSuspend(ctx context.Context, tenantID uuid.UUID, req BulkTenantOperationRequest) error {
	reason := "Bulk suspension operation"
	if reasonParam, ok := req.Parameters["reason"].(string); ok {
		reason = reasonParam
	}

	suspendReq := SuspendTenantRequest{
		TenantID:    tenantID,
		Reason:      reason,
		ActorID:     req.ActorID,
		NotifyUsers: false, // Don't notify users during bulk operations
	}

	_, err := s.SuspendTenant(ctx, suspendReq)
	return err
}

func (s *provisioningService) executeBulkReactivate(ctx context.Context, tenantID uuid.UUID, req BulkTenantOperationRequest) error {
	reason := "Bulk reactivation operation"
	if reasonParam, ok := req.Parameters["reason"].(string); ok {
		reason = reasonParam
	}

	reactivateReq := ReactivateTenantRequest{
		TenantID: tenantID,
		Reason:   reason,
		ActorID:  req.ActorID,
	}

	_, err := s.ReactivateTenant(ctx, reactivateReq)
	return err
}

func (s *provisioningService) executeBulkArchive(ctx context.Context, tenantID uuid.UUID, req BulkTenantOperationRequest) error {
	reason := "Bulk archiving operation"
	retentionDays := uint(90) // Default 90 days
	immediateDeletion := false

	if reasonParam, ok := req.Parameters["reason"].(string); ok {
		reason = reasonParam
	}
	if retentionParam, ok := req.Parameters["retention_days"].(float64); ok {
		retentionDays = uint(retentionParam)
	}
	if deletionParam, ok := req.Parameters["immediate_deletion"].(bool); ok {
		immediateDeletion = deletionParam
	}

	archiveReq := ArchiveTenantRequest{
		TenantID:          tenantID,
		Reason:            reason,
		ActorID:           req.ActorID,
		DataRetentionDays: retentionDays,
		ImmediateDeletion: immediateDeletion,
	}

	_, err := s.ArchiveTenant(ctx, archiveReq)
	return err
}

func (s *provisioningService) executeBulkUpdateLimits(ctx context.Context, tenantID uuid.UUID, req BulkTenantOperationRequest) error {
	reason := "Bulk limits update operation"
	if reasonParam, ok := req.Parameters["reason"].(string); ok {
		reason = reasonParam
	}

	// Extract limits from parameters
	var limits *TenantLimits
	if limitsParam, ok := req.Parameters["limits"].(map[string]any); ok {
		limits = &TenantLimits{}
		if maxUsers, ok := limitsParam["max_users"].(float64); ok {
			limits.MaxUsers = uint(maxUsers)
		}
		if maxStorage, ok := limitsParam["max_storage_mb"].(float64); ok {
			limits.MaxStorageMB = uint64(maxStorage)
		}
		if maxAPI, ok := limitsParam["max_api_calls_per_hour"].(float64); ok {
			limits.MaxAPICallsPerHour = uint(maxAPI)
		}
	}

	if limits == nil {
		return fmt.Errorf("no limits provided for update operation")
	}

	updateReq := UpdateTenantConfigurationRequest{
		TenantID: tenantID,
		Limits:   limits,
		Reason:   reason,
		ActorID:  req.ActorID,
	}

	_, err := s.UpdateTenantConfiguration(ctx, updateReq)
	return err
}

// Helper methods for advanced listing

func (s *provisioningService) getFilteredTenants(ctx context.Context, req AdvancedListRequest, offset, pageSize uint) ([]Tenant, uint, error) {
	// In a real implementation, this would build and execute SQL queries with filters
	// For now, return mock data
	acmeSubdomain := "acme"
	techSubdomain := "techstart"
	globalSubdomain := "global"

	mockTenants := []Tenant{
		{
			ID:        uuid.New(),
			Name:      "Acme Corporation",
			Slug:      "acme-corp",
			Subdomain: &acmeSubdomain,
			Status:    StatusActive,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Name:      "Tech Startup Inc",
			Slug:      "tech-startup",
			Subdomain: &techSubdomain,
			Status:    StatusActive,
			CreatedAt: time.Now().Add(-15 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Name:      "Global Services Ltd",
			Slug:      "global-services",
			Subdomain: &globalSubdomain,
			Status:    StatusSuspended,
			CreatedAt: time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-5 * time.Hour),
		},
	}

	// Apply filters in a real implementation
	filteredTenants := mockTenants

	// Apply status filter
	if len(req.StatusFilter) > 0 {
		var filtered []Tenant
		for _, tenant := range mockTenants {
			for _, status := range req.StatusFilter {
				if string(tenant.Status) == status {
					filtered = append(filtered, tenant)
					break
				}
			}
		}
		filteredTenants = filtered
	}

	// Apply search filter
	if req.Search != "" {
		var filtered []Tenant
		for _, tenant := range filteredTenants {
			subdomainStr := ""
			if tenant.Subdomain != nil {
				subdomainStr = *tenant.Subdomain
			}
			if contains(tenant.Name, req.Search) || contains(subdomainStr, req.Search) {
				filtered = append(filtered, tenant)
			}
		}
		filteredTenants = filtered
	}

	// Apply date filters
	if req.CreatedAfter != nil {
		var filtered []Tenant
		for _, tenant := range filteredTenants {
			if tenant.CreatedAt.After(*req.CreatedAfter) {
				filtered = append(filtered, tenant)
			}
		}
		filteredTenants = filtered
	}

	if req.CreatedBefore != nil {
		var filtered []Tenant
		for _, tenant := range filteredTenants {
			if tenant.CreatedAt.Before(*req.CreatedBefore) {
				filtered = append(filtered, tenant)
			}
		}
		filteredTenants = filtered
	}

	// Apply pagination
	totalCount := uint(len(filteredTenants))
	start := offset
	end := offset + pageSize
	if start > uint(len(filteredTenants)) {
		start = uint(len(filteredTenants))
	}
	if end > uint(len(filteredTenants)) {
		end = uint(len(filteredTenants))
	}

	paginatedTenants := filteredTenants[start:end]

	return paginatedTenants, totalCount, nil
}

func (s *provisioningService) getTenantUsageStats(ctx context.Context, tenantID uuid.UUID) (*TenantUsageStatsResult, error) {
	// In a real implementation, this would query usage statistics tables
	return &TenantUsageStatsResult{
		PeriodStart:           time.Now().AddDate(0, -1, 0),
		PeriodEnd:             time.Now(),
		ActiveUsers:           18,
		TotalEntities:         156,
		StorageUsedMB:         2048,
		APICalls:              50000,
		TransactionsProcessed: 1250,
	}, nil
}

func (s *provisioningService) generateTenantSummaryStats(ctx context.Context) (*TenantSummaryStats, error) {
	// In a real implementation, this would aggregate statistics from the database
	return &TenantSummaryStats{
		TotalTenants:     125,
		ActiveTenants:    98,
		SuspendedTenants: 15,
		PendingTenants:   8,
		ArchivedTenants:  4,
		TotalRevenue: &Money{
			Amount:   "1250000.00",
			Currency: "USD",
		},
		AvgUsersPerTenant: 24.5,
	}, nil
}

// Utility function for string search
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		strings.Contains(strings.ToLower(str), strings.ToLower(substr)))
}
