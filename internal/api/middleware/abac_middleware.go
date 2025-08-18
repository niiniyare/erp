package middleware

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
	"go.opentelemetry.io/otel/attribute"
	"goa.design/goa/v3/security"
)

// ABACMiddleware provides ABAC-based authorization middleware
type ABACMiddleware struct {
	abacService         abac.Service
	permissionEvaluator *featureflag.AdminPermissionEvaluator
	logger              logger.Logger
	metrics             *metrics.MetricsService
	tracing             tracing.TracingService
}

// NewABACMiddleware creates a new ABAC middleware instance
func NewABACMiddleware(
	abacService abac.Service,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
) *ABACMiddleware {
	return &ABACMiddleware{
		abacService:         abacService,
		permissionEvaluator: featureflag.NewAdminPermissionEvaluator(abacService),
		logger:              logger,
		metrics:             metrics,
		tracing:             tracing,
	}
}

// ABACAuthorizationConfig holds configuration for ABAC authorization
type ABACAuthorizationConfig struct {
	ResourceType string
	Action       string
	RequiredRole string
	IsHighRisk   bool
}

// RequireBulkOperationPermission creates middleware for bulk operations
func (m *ABACMiddleware) RequireBulkOperationPermission(action string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return func(ctx context.Context, payload any, scheme *security.JWTScheme) (context.Context, error) {
		ctx, span := m.tracing.StartSpan(ctx, "abac.middleware.bulk_operation",
			tracing.WithAttributes(
				attribute.String("action", action),
				attribute.String("resource_type", featureflag.ResourceTypeFeatureFlagBulk),
			))
		defer span.End()

		// Extract user info from JWT token (this would be implemented based on your JWT handling)
		userID, tenantID, err := m.extractUserFromJWT(ctx, scheme)
		if err != nil {
			m.recordAuthorizationMetrics(ctx, action, "jwt_extraction_failed", false)
			return ctx, fmt.Errorf("invalid JWT token: %w", err)
		}

		// Extract bulk operation details from payload
		bulkSize, reason := m.extractBulkOperationDetails(payload)

		// Evaluate permission
		result, err := m.permissionEvaluator.EvaluateBulkOperationPermission(
			ctx, userID, action, tenantID, bulkSize, reason)
		if err != nil {
			m.logger.Error("ABAC evaluation failed for bulk operation", logger.Fields{
				"error":     err.Error(),
				"user_id":   userID,
				"tenant_id": tenantID,
				"action":    action,
				"bulk_size": bulkSize,
			})
			m.recordAuthorizationMetrics(ctx, action, "evaluation_error", false)
			return ctx, fmt.Errorf("authorization evaluation failed: %w", err)
		}

		if result.Decision != types.PolicyDecisionAllow {
			m.logger.Warn("ABAC denied bulk operation", logger.Fields{
				"user_id":    userID,
				"tenant_id":  tenantID,
				"action":     action,
				"bulk_size":  bulkSize,
				"decision":   result.Decision,
				"policies":   len(result.PolicyDecisions),
				"request_id": result.RequestID,
			})
			m.recordAuthorizationMetrics(ctx, action, "permission_denied", false)
			return ctx, fmt.Errorf("insufficient permissions for bulk operation")
		}

		// Log successful authorization
		m.logger.Info("ABAC authorized bulk operation", logger.Fields{
			"user_id":            userID,
			"tenant_id":          tenantID,
			"action":             action,
			"bulk_size":          bulkSize,
			"evaluation_time_ms": result.EvaluationTimeMS,
			"cache_hit":          result.CacheHit,
			"request_id":         result.RequestID,
		})

		// Record successful authorization
		m.recordAuthorizationMetrics(ctx, action, "authorized", true)

		// Add authorization context for downstream handlers
		ctx = m.addAuthorizationContext(ctx, userID, tenantID, result)

		return ctx, nil
	}
}

// RequireSystemOperationPermission creates middleware for system operations
func (m *ABACMiddleware) RequireSystemOperationPermission(action string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return func(ctx context.Context, payload any, scheme *security.JWTScheme) (context.Context, error) {
		ctx, span := m.tracing.StartSpan(ctx, "abac.middleware.system_operation",
			tracing.WithAttributes(
				attribute.String("action", action),
				attribute.String("resource_type", featureflag.ResourceTypeFeatureFlagSystem),
			))
		defer span.End()

		// Extract user info from JWT token
		userID, tenantID, err := m.extractUserFromJWT(ctx, scheme)
		if err != nil {
			m.recordAuthorizationMetrics(ctx, action, "jwt_extraction_failed", false)
			return ctx, fmt.Errorf("invalid JWT token: %w", err)
		}

		// Evaluate permission
		result, err := m.permissionEvaluator.EvaluateSystemOperationPermission(
			ctx, userID, action, tenantID)
		if err != nil {
			m.logger.Error("ABAC evaluation failed for system operation", logger.Fields{
				"error":     err.Error(),
				"user_id":   userID,
				"tenant_id": tenantID,
				"action":    action,
			})
			m.recordAuthorizationMetrics(ctx, action, "evaluation_error", false)
			return ctx, fmt.Errorf("authorization evaluation failed: %w", err)
		}

		if result.Decision != types.PolicyDecisionAllow {
			m.logger.Warn("ABAC denied system operation", logger.Fields{
				"user_id":    userID,
				"tenant_id":  tenantID,
				"action":     action,
				"decision":   result.Decision,
				"policies":   len(result.PolicyDecisions),
				"request_id": result.RequestID,
			})
			m.recordAuthorizationMetrics(ctx, action, "permission_denied", false)
			return ctx, fmt.Errorf("insufficient permissions for system operation")
		}

		// Log successful authorization
		m.logger.Info("ABAC authorized system operation", logger.Fields{
			"user_id":            userID,
			"tenant_id":          tenantID,
			"action":             action,
			"evaluation_time_ms": result.EvaluationTimeMS,
			"cache_hit":          result.CacheHit,
			"request_id":         result.RequestID,
		})

		// Record successful authorization
		m.recordAuthorizationMetrics(ctx, action, "authorized", true)

		// Add authorization context
		ctx = m.addAuthorizationContext(ctx, userID, tenantID, result)

		return ctx, nil
	}
}

// RequireEmergencyOperationPermission creates middleware for emergency operations
func (m *ABACMiddleware) RequireEmergencyOperationPermission() func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return func(ctx context.Context, payload any, scheme *security.JWTScheme) (context.Context, error) {
		ctx, span := m.tracing.StartSpan(ctx, "abac.middleware.emergency_operation",
			tracing.WithAttributes(
				attribute.String("action", featureflag.ActionEmergencyControl),
				attribute.String("resource_type", featureflag.ResourceTypeFeatureFlagSystem),
				attribute.Bool("emergency", true),
			))
		defer span.End()

		// Extract user info from JWT token
		userID, tenantID, err := m.extractUserFromJWT(ctx, scheme)
		if err != nil {
			m.recordAuthorizationMetrics(ctx, featureflag.ActionEmergencyControl, "jwt_extraction_failed", false)
			return ctx, fmt.Errorf("invalid JWT token: %w", err)
		}

		// Extract reason from payload
		reason := m.extractReasonFromPayload(payload)
		if reason == "" {
			return ctx, fmt.Errorf("emergency operations require a reason")
		}

		// Evaluate emergency permission
		result, err := m.permissionEvaluator.EvaluateEmergencyOperationPermission(
			ctx, userID, tenantID, reason)
		if err != nil {
			m.logger.Error("ABAC evaluation failed for emergency operation", logger.Fields{
				"error":     err.Error(),
				"user_id":   userID,
				"tenant_id": tenantID,
				"reason":    reason,
			})
			m.recordAuthorizationMetrics(ctx, featureflag.ActionEmergencyControl, "evaluation_error", false)
			return ctx, fmt.Errorf("authorization evaluation failed: %w", err)
		}

		if result.Decision != types.PolicyDecisionAllow {
			m.logger.Warn("ABAC denied emergency operation", logger.Fields{
				"user_id":    userID,
				"tenant_id":  tenantID,
				"reason":     reason,
				"decision":   result.Decision,
				"policies":   len(result.PolicyDecisions),
				"request_id": result.RequestID,
			})
			m.recordAuthorizationMetrics(ctx, featureflag.ActionEmergencyControl, "permission_denied", false)
			return ctx, fmt.Errorf("insufficient permissions for emergency operation")
		}

		// Log emergency operation authorization (high priority)
		m.logger.Warn("EMERGENCY: ABAC authorized emergency operation", logger.Fields{
			"user_id":            userID,
			"tenant_id":          tenantID,
			"reason":             reason,
			"evaluation_time_ms": result.EvaluationTimeMS,
			"request_id":         result.RequestID,
			"emergency":          true,
		})

		// Record emergency authorization
		m.recordAuthorizationMetrics(ctx, featureflag.ActionEmergencyControl, "authorized", true)

		// Add emergency context
		ctx = m.addEmergencyContext(ctx, userID, tenantID, reason, result)

		return ctx, nil
	}
}

// Helper methods

// extractUserFromJWT extracts user and tenant information from JWT token
func (m *ABACMiddleware) extractUserFromJWT(ctx context.Context, scheme *security.JWTScheme) (uuid.UUID, uuid.UUID, error) {
	// NOTE:
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Parse and validate the JWT token
	// 2. Extract user ID and tenant ID from claims
	// 3. Validate token signature and expiry

	// For now, return mock values - replace with actual JWT parsing
	userID := uuid.New()   // Extract from JWT claims
	tenantID := uuid.New() // Extract from JWT claims or context

	return userID, tenantID, nil
}

// extractBulkOperationDetails extracts bulk operation details from payload
func (m *ABACMiddleware) extractBulkOperationDetails(payload any) (int, string) {
	// NOTE:
	// This would extract from the actual Goa payload structure
	// Implementation depends on the specific payload types

	// Example extraction (replace with actual implementation):
	bulkSize := 10              // Extract from payload.FlagNames length
	reason := "admin_operation" // Extract from payload.Reason

	return bulkSize, reason
}

// extractReasonFromPayload extracts reason from payload
func (m *ABACMiddleware) extractReasonFromPayload(payload any) string {
	// Extract reason from the emergency operation payload
	return "emergency_reason" // Replace with actual extraction
}

// addAuthorizationContext adds authorization information to context
func (m *ABACMiddleware) addAuthorizationContext(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	result *abac.PermissionEvaluationResult,
) context.Context {
	ctx = context.WithValue(ctx, "authorized_user_id", userID)
	ctx = context.WithValue(ctx, "authorized_tenant_id", tenantID)
	ctx = context.WithValue(ctx, "abac_request_id", result.RequestID)
	ctx = context.WithValue(ctx, "abac_evaluation_time", result.EvaluationTimeMS)
	ctx = context.WithValue(ctx, "abac_cache_hit", result.CacheHit)
	return ctx
}

// addEmergencyContext adds emergency operation context
func (m *ABACMiddleware) addEmergencyContext(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	reason string,
	result *abac.PermissionEvaluationResult,
) context.Context {
	ctx = m.addAuthorizationContext(ctx, userID, tenantID, result)
	ctx = context.WithValue(ctx, "emergency_operation", true)
	ctx = context.WithValue(ctx, "emergency_reason", reason)
	return ctx
}

// recordAuthorizationMetrics records metrics for authorization events
func (m *ABACMiddleware) recordAuthorizationMetrics(ctx context.Context, action, outcome string, authorized bool) {
	labels := metrics.Fields{
		"action":   action,
		"outcome":  outcome,
		"resource": featureflag.ResourceTypeFeatureFlagBulk,
	}

	m.metrics.IncrementCounter("abac_authorization_attempts_total", labels)

	if authorized {
		m.metrics.IncrementCounter("abac_authorization_success_total", labels)
	} else {
		m.metrics.IncrementCounter("abac_authorization_failures_total", labels)
	}
}

// AuthorizationInfo holds authorization context information
type AuthorizationInfo struct {
	UserID           uuid.UUID
	TenantID         uuid.UUID
	RequestID        string
	EvaluationTimeMS int64
	CacheHit         bool
	Emergency        bool
	Reason           string
}

// GetAuthorizationInfo extracts authorization info from context
func GetAuthorizationInfo(ctx context.Context) *AuthorizationInfo {
	info := &AuthorizationInfo{}

	if userID, ok := ctx.Value("authorized_user_id").(uuid.UUID); ok {
		info.UserID = userID
	}
	if tenantID, ok := ctx.Value("authorized_tenant_id").(uuid.UUID); ok {
		info.TenantID = tenantID
	}
	if requestID, ok := ctx.Value("abac_request_id").(string); ok {
		info.RequestID = requestID
	}
	if evalTime, ok := ctx.Value("abac_evaluation_time").(int64); ok {
		info.EvaluationTimeMS = evalTime
	}
	if cacheHit, ok := ctx.Value("abac_cache_hit").(bool); ok {
		info.CacheHit = cacheHit
	}
	if emergency, ok := ctx.Value("emergency_operation").(bool); ok {
		info.Emergency = emergency
	}
	if reason, ok := ctx.Value("emergency_reason").(string); ok {
		info.Reason = reason
	}

	return info
}

// ValidateMinimumRole validates that user has minimum required role
func (m *ABACMiddleware) ValidateMinimumRole(ctx context.Context, userID uuid.UUID, requiredRole string) error {
	// This would integrate with the identity service to check user roles
	// For now, return nil (implement based on your identity system)
	return nil
}

// LogAuthorizationDecision logs the authorization decision for audit purposes
func (m *ABACMiddleware) LogAuthorizationDecision(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	result *abac.PermissionEvaluationResult,
	payload any,
) {
	m.logger.Info("ABAC authorization decision logged", logger.Fields{
		"event_type":         "authorization_decision",
		"user_id":            userID,
		"action":             action,
		"decision":           result.Decision,
		"policies_applied":   len(result.PolicyDecisions),
		"evaluation_time_ms": result.EvaluationTimeMS,
		"cache_hit":          result.CacheHit,
		"request_id":         result.RequestID,
		"timestamp":          result.Timestamp,
	})
}
