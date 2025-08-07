package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	adminfeatureflag "github.com/niiniyare/erp/internal/api/gen/admin_featureflag"
	corefeatureflag "github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"goa.design/goa/v3/security"
)

// AdminFeatureFlagService implements the Goa-generated admin_featureflag.Service interface
type AdminFeatureFlagService struct {
	adminService corefeatureflag.AdminService
	logger       logger.Logger
	metrics      *metrics.MetricsService
	tracing      tracing.TracingService
}

// NewAdminFeatureFlagService creates a new Goa admin feature flag service implementation
func NewAdminFeatureFlagService(
	adminService corefeatureflag.AdminService,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
) adminfeatureflag.Service {
	return &AdminFeatureFlagService{
		adminService: adminService,
		logger:       logger,
		metrics:      metrics,
		tracing:      tracing,
	}
}

// BulkEnable implements admin_featureflag.Service.
func (s *AdminFeatureFlagService) BulkEnable(ctx context.Context, p *adminfeatureflag.BulkEnablePayload) (res *adminfeatureflag.BulkEnableResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin_featureflag.service.bulk_enable")
	defer span.End()

	s.logger.Info("Starting bulk enable operation", logger.Fields{
		"tenant_id":  p.TenantID,
		"flag_count": len(p.FlagNames),
		"reason":     p.Reason,
	})

	// Convert Goa payload to domain request
	request := &corefeatureflag.BulkEnableFlagsRequest{
		FlagNames: p.FlagNames,
		Reason:    p.Reason,
	}

	// Execute bulk enable operation
	result, err := s.adminService.BulkEnableFlags(ctx, request)
	if err != nil {
		s.logger.Error("Failed to bulk enable feature flags", logger.Fields{
			"error":      err.Error(),
			"tenant_id":  p.TenantID,
			"flag_count": len(p.FlagNames),
			"reason":     p.Reason,
		})
		return nil, s.mapError(err)
	}

	s.logger.Info("Bulk enable operation completed", logger.Fields{
		"tenant_id":        p.TenantID,
		"total_requested":  result.TotalRequested,
		"successful":       result.Successful,
		"failed":           result.Failed,
		"success_rate":     result.Summary.SuccessRate,
		"execution_time":   result.ExecutionTime.String(),
	})

	// Record metrics
	if s.metrics != nil {
		labels := metrics.Fields{
			"operation": "bulk_enable",
			"tenant_id": p.TenantID,
		}
		s.metrics.IncrementCounter("admin_bulk_operations_total", labels)
		s.metrics.ObserveHistogram("admin_bulk_operation_duration", float64(result.ExecutionTime.Milliseconds()), labels)
		s.metrics.SetGauge("admin_bulk_operation_success_rate", result.Summary.SuccessRate, labels)
	}

	// Convert domain result to Goa result
	return &adminfeatureflag.BulkEnableResult{
		TotalRequested: result.TotalRequested,
		Successful:     result.Successful,
		Failed:         result.Failed,
		ExecutedAt:     result.ExecutedAt.Format(time.RFC3339),
	}, nil
}

// BulkDisable implements admin_featureflag.Service.
func (s *AdminFeatureFlagService) BulkDisable(ctx context.Context, p *adminfeatureflag.BulkDisablePayload) (res *adminfeatureflag.BulkDisableResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin_featureflag.service.bulk_disable")
	defer span.End()

	s.logger.Info("Starting bulk disable operation", logger.Fields{
		"tenant_id":  p.TenantID,
		"flag_count": len(p.FlagNames),
		"reason":     p.Reason,
	})

	// Convert Goa payload to domain request
	request := &corefeatureflag.BulkDisableFlagsRequest{
		FlagNames: p.FlagNames,
		Reason:    p.Reason,
	}

	// Execute bulk disable operation
	result, err := s.adminService.BulkDisableFlags(ctx, request)
	if err != nil {
		s.logger.Error("Failed to bulk disable feature flags", logger.Fields{
			"error":      err.Error(),
			"tenant_id":  p.TenantID,
			"flag_count": len(p.FlagNames),
			"reason":     p.Reason,
		})
		return nil, s.mapError(err)
	}

	s.logger.Info("Bulk disable operation completed", logger.Fields{
		"tenant_id":        p.TenantID,
		"total_requested":  result.TotalRequested,
		"successful":       result.Successful,
		"failed":           result.Failed,
		"success_rate":     result.Summary.SuccessRate,
		"execution_time":   result.ExecutionTime.String(),
	})

	// Record metrics
	if s.metrics != nil {
		labels := metrics.Fields{
			"operation": "bulk_disable",
			"tenant_id": p.TenantID,
		}
		s.metrics.IncrementCounter("admin_bulk_operations_total", labels)
		s.metrics.ObserveHistogram("admin_bulk_operation_duration", float64(result.ExecutionTime.Milliseconds()), labels)
		s.metrics.SetGauge("admin_bulk_operation_success_rate", result.Summary.SuccessRate, labels)
	}

	// Convert domain result to Goa result
	return &adminfeatureflag.BulkDisableResult{
		TotalRequested: result.TotalRequested,
		Successful:     result.Successful,
		Failed:         result.Failed,
		ExecutedAt:     result.ExecutedAt.Format(time.RFC3339),
	}, nil
}

// SystemHealth implements admin_featureflag.Service.
func (s *AdminFeatureFlagService) SystemHealth(ctx context.Context, p *adminfeatureflag.SystemHealthPayload) (res *adminfeatureflag.SystemHealthResult, err error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin_featureflag.service.system_health")
	defer span.End()

	s.logger.Debug("Getting system health", logger.Fields{
		"tenant_id": p.TenantID,
	})

	// Execute system health check
	result, err := s.adminService.GetSystemHealth(ctx)
	if err != nil {
		s.logger.Error("Failed to get system health", logger.Fields{
			"error":     err.Error(),
			"tenant_id": p.TenantID,
		})
		return nil, s.mapError(err)
	}

	s.logger.Info("System health check completed", logger.Fields{
		"tenant_id":      p.TenantID,
		"overall_status": result.Status,
		"overall_score":  result.OverallScore,
		"db_status":      result.DatabaseStatus,
		"cache_status":   result.CacheStatus,
	})

	// Record health metrics
	if s.metrics != nil {
		labels := metrics.Fields{
			"tenant_id": p.TenantID,
		}
		s.metrics.SetGauge("admin_system_health_score", float64(result.OverallScore), labels)
		
		// Record component health
		for component, health := range result.ComponentHealth {
			componentLabels := metrics.Fields{
				"tenant_id": p.TenantID,
				"component": component,
			}
			
			var statusValue float64
			switch health.Status {
			case "healthy":
				statusValue = 1.0
			case "degraded":
				statusValue = 0.5
			case "unhealthy":
				statusValue = 0.0
			default:
				statusValue = 0.0
			}
			
			s.metrics.SetGauge("admin_component_health_status", statusValue, componentLabels)
			s.metrics.ObserveHistogram("admin_component_response_time", float64(health.ResponseTime.Milliseconds()), componentLabels)
		}
	}

	// Convert domain result to Goa result
	return &adminfeatureflag.SystemHealthResult{
		Status:         result.Status,
		Timestamp:      result.Timestamp.Format(time.RFC3339),
		DatabaseStatus: result.DatabaseStatus,
		CacheStatus:    result.CacheStatus,
		OverallScore:   result.OverallScore,
	}, nil
}

// JWTAuth implements the authorization logic for the JWT security scheme.
func (s *AdminFeatureFlagService) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	// This is a placeholder implementation. In a real application, you would:
	// 1. Validate the JWT token
	// 2. Extract user information and permissions
	// 3. Check if the user has admin privileges for feature flag management
	// 4. Add the validated user context to the context
	
	s.logger.Debug("JWT authentication for admin feature flag service", logger.Fields{
		"has_token": token != "",
	})

	// For now, we'll pass through the token validation to the underlying service
	// The actual JWT validation should be handled by middleware or a dedicated auth service
	
	if token == "" {
		return nil, adminfeatureflag.MakeUnauthorized(fmt.Errorf("missing authentication token"))
	}

	// Add token to context for downstream services to validate
	ctx = context.WithValue(ctx, "jwt_token", token)
	
	return ctx, nil
}

// Helper methods for error mapping

func (s *AdminFeatureFlagService) mapError(err error) error {
	// Map domain errors to Goa errors
	switch {
	case err == corefeatureflag.ErrFeatureFlagNotFound:
		return adminfeatureflag.MakeBadRequest(fmt.Errorf("feature flag not found: %w", err))
	case err == corefeatureflag.ErrFeatureFlagAlreadyExists:
		return adminfeatureflag.MakeBadRequest(fmt.Errorf("feature flag already exists: %w", err))
	case err == corefeatureflag.ErrInvalidTenantContext:
		return adminfeatureflag.MakeUnauthorized(fmt.Errorf("invalid tenant context: %w", err))
	default:
		// Check for validation errors
		if IsValidationError(err) {
			return adminfeatureflag.MakeBadRequest(fmt.Errorf("validation error: %w", err))
		}
		
		// Check for permission errors
		if IsPermissionError(err) {
			return adminfeatureflag.MakeUnauthorized(fmt.Errorf("permission denied: %w", err))
		}
		
		// Log unexpected errors for debugging
		s.logger.Error("Unmapped error in admin feature flag service", logger.Fields{
			"error":      err.Error(),
			"error_type": fmt.Sprintf("%T", err),
		})
		
		// Return internal error for unexpected errors
		return adminfeatureflag.MakeInternalError(fmt.Errorf("internal server error: %w", err))
	}
}

// Helper functions for error classification

func IsValidationError(err error) bool {
	// Check if error is a validation error
	// This could check for specific error types or error message patterns
	errorMsg := err.Error()
	validationKeywords := []string{
		"validation",
		"invalid",
		"required",
		"format",
		"limit",
		"range",
	}
	
	for _, keyword := range validationKeywords {
		if contains(errorMsg, keyword) {
			return true
		}
	}
	
	return false
}

func IsPermissionError(err error) bool {
	// Check if error is a permission/authorization error
	errorMsg := err.Error()
	permissionKeywords := []string{
		"unauthorized",
		"permission",
		"access denied",
		"forbidden",
		"not allowed",
	}
	
	for _, keyword := range permissionKeywords {
		if contains(errorMsg, keyword) {
			return true
		}
	}
	
	return false
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) && 
		   (len(substr) == 0 || 
		    text == substr || 
		    (len(text) > len(substr) && 
		     (text[:len(substr)] == substr || 
		      text[len(text)-len(substr):] == text || 
		      containsHelper(text, substr))))
}

func containsHelper(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}