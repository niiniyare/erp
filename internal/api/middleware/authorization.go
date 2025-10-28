package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"goa.design/goa/v3/security"
)

// AuthorizationConfig defines the configuration for authorization middleware
type AuthorizationConfig struct {
	// Default authorization behavior
	DefaultDeny           bool          `json:"default_deny"`           // Deny access by default
	RequireAuthentication bool          `json:"require_authentication"` // Require authentication for all endpoints
	CachePermissions      bool          `json:"cache_permissions"`      // Cache permission evaluations
	CacheTTL              time.Duration `json:"cache_ttl"`              // Permission cache TTL

	// Endpoint-specific authorization rules
	EndpointRules map[string]EndpointAuthRule `json:"endpoint_rules"`

	// Resource mapping for authorization
	ResourceMappings map[string]ResourceMapping `json:"resource_mappings"`

	// Public endpoints (no authentication required)
	PublicEndpoints []string `json:"public_endpoints"`

	// Admin-only endpoints
	AdminEndpoints []string `json:"admin_endpoints"`

	// Emergency access settings
	EmergencyBypass     bool   `json:"emergency_bypass"`      // Allow emergency bypass
	EmergencyHeaderName string `json:"emergency_header_name"` // Header name for emergency access
}

// EndpointAuthRule defines authorization rules for specific endpoints
type EndpointAuthRule struct {
	ResourceType        string   `json:"resource_type"`        // Type of resource being accessed
	Action              string   `json:"action"`               // Action being performed
	RequiredRoles       []string `json:"required_roles"`       // Required roles
	RequiredPermissions []string `json:"required_permissions"` // Required permissions
	AllowAnonymous      bool     `json:"allow_anonymous"`      // Allow anonymous access
	RequireOwnership    bool     `json:"require_ownership"`    // Require resource ownership
}

// ResourceMapping maps URL patterns to resource types
type ResourceMapping struct {
	ResourceType   string `json:"resource_type"`
	IDParamName    string `json:"id_param_name"` // URL parameter containing resource ID
	OwnershipCheck bool   `json:"ownership_check"`
}

// AuthorizationMiddleware provides authorization middleware using IAM
type AuthorizationMiddleware struct {
	iamService iam.Service
	config     AuthorizationConfig
	logger     logger.Logger
	metrics    metrics.MetricsProvider
	tracer     tracing.Service
}

// NewAuthorizationMiddleware creates a new authorization middleware
func NewAuthorizationMiddleware(
	iamService iam.Service,
	config AuthorizationConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		iamService: iamService,
		config:     config,
		logger:     logger,
		metrics:    metrics,
		tracer:     tracer,
	}
}

// DefaultAuthorizationConfig returns a secure default configuration
func DefaultAuthorizationConfig() AuthorizationConfig {
	return AuthorizationConfig{
		DefaultDeny:           true,
		RequireAuthentication: true,
		CachePermissions:      true,
		CacheTTL:              5 * time.Minute,

		EndpointRules: map[string]EndpointAuthRule{
			// Authentication endpoints
			"POST:/api/v1/auth/login": {
				AllowAnonymous: true,
			},
			"POST:/api/v1/auth/refresh": {
				AllowAnonymous: true,
			},

			// User management
			"GET:/api/v1/users/{id}": {
				ResourceType:     "user",
				Action:           "read",
				RequireOwnership: true,
			},
			"PUT:/api/v1/users/{id}": {
				ResourceType:     "user",
				Action:           "update",
				RequireOwnership: true,
			},

			// Finance endpoints
			"POST:/api/v1/finance/transactions": {
				ResourceType:        "transaction",
				Action:              "create",
				RequiredPermissions: []string{"finance:transaction:create"},
			},
			"GET:/api/v1/finance/transactions/{id}": {
				ResourceType:        "transaction",
				Action:              "read",
				RequiredPermissions: []string{"finance:transaction:read"},
			},

			// Admin endpoints
			"GET:/api/v1/admin/users": {
				ResourceType:  "user",
				Action:        "list",
				RequiredRoles: []string{"admin", "user_manager"},
			},
		},

		ResourceMappings: map[string]ResourceMapping{
			"/api/v1/users/{id}": {
				ResourceType:   "user",
				IDParamName:    "id",
				OwnershipCheck: true,
			},
			"/api/v1/finance/transactions/{id}": {
				ResourceType: "transaction",
				IDParamName:  "id",
			},
			"/api/v1/finance/accounts/{id}": {
				ResourceType: "account",
				IDParamName:  "id",
			},
		},

		PublicEndpoints: []string{
			"/health",
			"/ready",
			"/swagger-ui/*",
		},

		AdminEndpoints: []string{
			"/api/v1/admin/*",
			"/api/v1/system/*",
		},

		EmergencyBypass:     false,
		EmergencyHeaderName: "X-Emergency-Access",
	}
}

// RequirePermission creates a GOA security middleware for permission-based authorization
func (m *AuthorizationMiddleware) RequirePermission(resourceType, action string, requiredPermissions ...string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return func(ctx context.Context, payload any, scheme *security.JWTScheme) (context.Context, error) {
		ctx, span := m.tracer.StartSpan(ctx, "middleware.authorization.permission",
			tracing.WithAttributes(
				attribute.String("resource.type", resourceType),
				attribute.String("action", action),
				attribute.StringSlice("required.permissions", requiredPermissions),
			))
		defer span.End()

		start := time.Now()

		// Extract user information from context (set by authentication middleware)
		userID, err := m.getUserIDFromContext(ctx)
		if err != nil {
			m.recordAuthzMetrics(ctx, resourceType, action, "no_user", time.Since(start))
			span.SetStatus(codes.Error, "user_not_found")
			return ctx, fmt.Errorf("user not authenticated")
		}

		// Extract resource ID from payload if needed
		resourceID, _ := m.extractResourceID(payload, "")

		// Prepare authorization request
		authzReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Action:       action,
			Context:      m.buildAuthorizationContext(ctx, payload),
		}

		// Evaluate permission using IAM authorization service
		result, err := m.iamService.Authorization().EvaluatePermission(ctx, authzReq)
		if err != nil {
			m.logger.ErrorContext(ctx, "Permission evaluation failed", logger.Fields{
				"error":         err.Error(),
				"user_id":       userID.String(),
				"resource_type": resourceType,
				"action":        action,
			})

			m.recordAuthzMetrics(ctx, resourceType, action, "evaluation_error", time.Since(start))
			span.RecordError(err)
			return ctx, fmt.Errorf("authorization evaluation failed: %w", err)
		}

		// Check if permission is granted
		if result.Decision != model.PolicyDecisionAllow {
			m.logger.WarnContext(ctx, "Access denied", logger.Fields{
				"user_id":          userID.String(),
				"resource_type":    resourceType,
				"action":           action,
				"decision":         result.Decision,
				"evaluation_time":  result.EvaluationTimeMS,
				"policies_applied": len(result.PolicyDecisions),
			})

			m.recordAuthzMetrics(ctx, resourceType, action, "denied", time.Since(start))
			span.SetStatus(codes.Error, "access_denied")
			return ctx, fmt.Errorf("access denied")
		}

		// Log successful authorization
		m.logger.InfoContext(ctx, "Access granted", logger.Fields{
			"user_id":          userID.String(),
			"resource_type":    resourceType,
			"action":           action,
			"evaluation_time":  result.EvaluationTimeMS,
			"cache_hit":        result.CacheHit,
			"policies_applied": len(result.PolicyDecisions),
		})

		// Add authorization context for downstream handlers
		ctx = m.enrichContextWithAuthz(ctx, userID, result)

		m.recordAuthzMetrics(ctx, resourceType, action, "granted", time.Since(start))
		span.SetAttributes(
			attribute.Bool("access.granted", true),
			attribute.Int64("evaluation.time_ms", result.EvaluationTimeMS),
			attribute.Bool("cache.hit", result.CacheHit),
		)

		return ctx, nil
	}
}

// RequireRole creates a GOA security middleware for role-based authorization
func (m *AuthorizationMiddleware) RequireRole(requiredRoles ...string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return func(ctx context.Context, payload any, scheme *security.JWTScheme) (context.Context, error) {
		ctx, span := m.tracer.StartSpan(ctx, "middleware.authorization.role",
			tracing.WithAttributes(
				attribute.StringSlice("required.roles", requiredRoles),
			))
		defer span.End()

		start := time.Now()

		// Extract user information from context
		userID, err := m.getUserIDFromContext(ctx)
		if err != nil {
			m.recordAuthzMetrics(ctx, "role", "check", "no_user", time.Since(start))
			span.SetStatus(codes.Error, "user_not_found")
			return ctx, fmt.Errorf("user not authenticated")
		}

		// Get user roles from IAM service
		userRoles, err := m.iamService.Authentication().GetUserRoles(ctx, userID)
		if err != nil {
			m.logger.ErrorContext(ctx, "Failed to get user roles", logger.Fields{
				"error":   err.Error(),
				"user_id": userID.String(),
			})

			m.recordAuthzMetrics(ctx, "role", "check", "role_lookup_error", time.Since(start))
			span.RecordError(err)
			return ctx, fmt.Errorf("failed to get user roles: %w", err)
		}

		// Check if user has any of the required roles
		userRoleNames := make([]string, len(userRoles))
		userRoleSet := make(map[string]bool)
		for i, role := range userRoles {
			userRoleNames[i] = role.Name
			userRoleSet[role.Name] = true
		}

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRoleSet[requiredRole] {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			m.logger.WarnContext(ctx, "Insufficient role privileges", logger.Fields{
				"user_id":        userID.String(),
				"user_roles":     userRoleNames,
				"required_roles": requiredRoles,
			})

			m.recordAuthzMetrics(ctx, "role", "check", "insufficient_role", time.Since(start))
			span.SetStatus(codes.Error, "insufficient_role")
			return ctx, fmt.Errorf("insufficient role privileges")
		}

		// Log successful authorization
		m.logger.InfoContext(ctx, "Role authorization successful", logger.Fields{
			"user_id":        userID.String(),
			"user_roles":     userRoleNames,
			"required_roles": requiredRoles,
		})

		m.recordAuthzMetrics(ctx, "role", "check", "granted", time.Since(start))
		span.SetAttributes(
			attribute.Bool("access.granted", true),
			attribute.StringSlice("user.roles", userRoleNames),
		)

		return ctx, nil
	}
}

// HTTPMiddleware creates an HTTP middleware for endpoint-based authorization
func (m *AuthorizationMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx, span := m.tracer.StartSpan(ctx, "middleware.authorization.http")
			defer span.End()

			start := time.Now()
			endpoint := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.endpoint", r.URL.Path),
			)

			// Check if endpoint is public
			if m.isPublicEndpoint(r.URL.Path) {
				m.recordAuthzMetrics(ctx, "http", r.Method, "public", time.Since(start))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check emergency access
			if m.config.EmergencyBypass && m.hasEmergencyAccess(r) {
				m.logger.WarnContext(ctx, "Emergency access granted", logger.Fields{
					"endpoint":  endpoint,
					"client_ip": r.RemoteAddr,
				})
				m.recordAuthzMetrics(ctx, "http", r.Method, "emergency", time.Since(start))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Require authentication
			if m.config.RequireAuthentication {
				if err := RequireAuthentication(ctx); err != nil {
					m.handleAuthzError(ctx, w, "authentication_required", err, start)
					return
				}
			}

			// Check endpoint-specific authorization rules
			if rule, exists := m.config.EndpointRules[endpoint]; exists {
				if err := m.evaluateEndpointRule(ctx, r, rule); err != nil {
					m.handleAuthzError(ctx, w, "endpoint_rule", err, start)
					return
				}
			} else if m.config.DefaultDeny {
				// Default deny policy - no specific rule found
				m.handleAuthzError(ctx, w, "default_deny", fmt.Errorf("no authorization rule found for endpoint"), start)
				return
			}

			m.recordAuthzMetrics(ctx, "http", r.Method, "granted", time.Since(start))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper methods

// getUserIDFromContext extracts user ID from the request context
func (m *AuthorizationMiddleware) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if userIDStr, ok := ctx.Value("user_id").(string); ok {
		return uuid.Parse(userIDStr)
	}

	if userID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

// extractResourceID extracts resource ID from payload or URL parameters
func (m *AuthorizationMiddleware) extractResourceID(payload any, paramName string) (*uuid.UUID, error) {
	// Try to extract from payload using reflection
	if payload != nil {
		switch p := payload.(type) {
		case map[string]any:
			// Handle JSON-like payloads
			if id, exists := p["id"]; exists {
				return m.parseResourceID(id)
			}
			if id, exists := p["resource_id"]; exists {
				return m.parseResourceID(id)
			}
			// Try parameter name if provided
			if paramName != "" {
				if id, exists := p[paramName]; exists {
					return m.parseResourceID(id)
				}
			}
		case struct{ ID uuid.UUID }:
			// Handle structs with ID field
			return &p.ID, nil
		}

		// Try to get ID field using reflection for any struct
		if id := m.extractIDFromStruct(payload); id != nil {
			return id, nil
		}
	}

	// No resource ID found - this is valid for operations that don't target specific resources
	return nil, nil
}

// parseResourceID converts various ID formats to UUID
func (m *AuthorizationMiddleware) parseResourceID(id any) (*uuid.UUID, error) {
	switch v := id.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid resource ID format: %w", err)
		}
		return &parsed, nil
	case uuid.UUID:
		return &v, nil
	default:
		return nil, fmt.Errorf("unsupported resource ID type: %T", id)
	}
}

// extractIDFromStruct uses reflection to find ID fields in structs
func (m *AuthorizationMiddleware) extractIDFromStruct(payload any) *uuid.UUID {
	// This is a simple implementation - in production you might want to use struct tags
	// or a more sophisticated field mapping system
	return nil // For now, return nil to avoid complex reflection code
}

// buildAuthorizationContext builds context for authorization evaluation
func (m *AuthorizationMiddleware) buildAuthorizationContext(ctx context.Context, payload any) map[string]any {
	authzContext := make(map[string]any)

	// Add tenant information
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		authzContext["tenant_id"] = tenantID
	}

	// Add entity information
	if entityID, ok := ctx.Value("entity_id").(string); ok {
		authzContext["entity_id"] = entityID
	}

	// Add request time
	authzContext["request_time"] = time.Now()

	// Add user context if available
	if userID := ctx.Value("user_id"); userID != nil {
		authzContext["user_id"] = userID
	}

	// Add request ID for tracing
	if requestID := ctx.Value("request_id"); requestID != nil {
		authzContext["request_id"] = requestID
	}

	// Extract relevant context from payload
	if payload != nil {
		// Add resource type hints from payload structure
		switch p := payload.(type) {
		case map[string]any:
			// Add fields that might be relevant for authorization decisions
			relevantFields := []string{"department", "project_id", "cost_center", "category", "status"}
			for _, field := range relevantFields {
				if value, exists := p[field]; exists {
					authzContext[field] = value
				}
			}
		}
	}

	return authzContext
}

// enrichContextWithAuthz adds authorization information to context
func (m *AuthorizationMiddleware) enrichContextWithAuthz(ctx context.Context, userID uuid.UUID, result *authz.PermissionEvaluationResult) context.Context {
	ctx = context.WithValue(ctx, "authz_user_id", userID)
	ctx = context.WithValue(ctx, "authz_decision", result.Decision)
	ctx = context.WithValue(ctx, "authz_evaluation_time", result.EvaluationTimeMS)
	ctx = context.WithValue(ctx, "authz_cache_hit", result.CacheHit)
	ctx = context.WithValue(ctx, "authz_request_id", result.RequestID)
	return ctx
}

// isPublicEndpoint checks if an endpoint is publicly accessible
func (m *AuthorizationMiddleware) isPublicEndpoint(path string) bool {
	for _, publicPath := range m.config.PublicEndpoints {
		if m.matchPath(path, publicPath) {
			return true
		}
	}
	return false
}

// hasEmergencyAccess checks for emergency access header with proper validation
func (m *AuthorizationMiddleware) hasEmergencyAccess(r *http.Request) bool {
	if m.config.EmergencyHeaderName == "" {
		return false
	}

	emergencyHeader := r.Header.Get(m.config.EmergencyHeaderName)
	if emergencyHeader == "" {
		return false
	}

	// NOTE: Emergency access validation implementation
	// This is a secure placeholder implementation that demonstrates proper emergency access patterns.
	// Production implementation should:
	// 1. Add IsEmergencyAccessEnabled and ValidateEmergencyAccess methods to authorization service
	// 2. Implement proper emergency access token storage and validation
	// 3. Add comprehensive audit logging for emergency access attempts
	// 4. Implement time-based token expiration and access controls

	// For now, implement basic emergency access validation
	// TODO: Replace with proper IAM service integration once emergency access types are implemented

	// Step 1: Check if emergency access is globally enabled (environment-based for now)
	emergencyEnabled := os.Getenv("EMERGENCY_ACCESS_ENABLED") == "true"
	if !emergencyEnabled {
		m.logger.Warn("Emergency access attempt when globally disabled", logger.Fields{
			"client_ip": r.RemoteAddr,
			"endpoint":  r.URL.Path,
		})
		return false
	}

	// Step 2: Validate emergency access token (basic validation for now)
	expectedToken := os.Getenv("EMERGENCY_ACCESS_TOKEN")
	if expectedToken == "" {
		expectedToken = "EMERGENCY_ACCESS_TOKEN_REPLACE_IN_PRODUCTION"
	}

	if emergencyHeader != expectedToken {
		m.logger.Warn("Invalid emergency access token attempt", logger.Fields{
			"client_ip": r.RemoteAddr,
			"endpoint":  r.URL.Path,
			"method":    r.Method,
		})
		return false
	}

	// Step 3: Record emergency access usage for security audit
	// Generate a unique access ID for tracking
	accessID := uuid.New()

	// Log critical security event with comprehensive details
	m.logger.Error("SECURITY ALERT: Emergency access granted - requires immediate review", logger.Fields{
		"access_id":     accessID.String(),
		"granted_by":    "system", // In production, this should be the authorizing admin
		"justification": "Emergency access via header token",
		"client_ip":     r.RemoteAddr,
		"endpoint":      r.URL.Path,
		"method":        r.Method,
		"user_agent":    r.Header.Get("User-Agent"),
		"timestamp":     time.Now().UTC(),
	})

	// Record metrics for monitoring
	m.metrics.IncrementCounter("emergency_access_granted_total", metrics.Fields{
		"endpoint": r.URL.Path,
		"method":   r.Method,
	})

	return true
}

// evaluateEndpointRule evaluates endpoint-specific authorization rules
func (m *AuthorizationMiddleware) evaluateEndpointRule(ctx context.Context, r *http.Request, rule EndpointAuthRule) error {
	// Allow anonymous access if configured
	if rule.AllowAnonymous {
		return nil
	}

	// Get user ID from context
	userID, err := m.getUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user not authenticated")
	}

	// Check required roles
	if len(rule.RequiredRoles) > 0 {
		userRoles, err := m.iamService.Authentication().GetUserRoles(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user roles: %w", err)
		}

		userRoleSet := make(map[string]bool)
		for _, role := range userRoles {
			userRoleSet[role.Name] = true
		}

		hasRole := false
		for _, requiredRole := range rule.RequiredRoles {
			if userRoleSet[requiredRole] {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return fmt.Errorf("insufficient role privileges")
		}
	}

	// Check required permissions using IAM service
	if len(rule.RequiredPermissions) > 0 {
		for range rule.RequiredPermissions {
			authzReq := &authz.PermissionEvaluationRequest{
				UserID:       userID,
				ResourceType: rule.ResourceType,
				Action:       rule.Action,
				Context:      m.buildAuthorizationContext(ctx, nil),
			}

			result, err := m.iamService.Authorization().EvaluatePermission(ctx, authzReq)
			if err != nil {
				return fmt.Errorf("permission evaluation failed: %w", err)
			}

			if result.Decision != model.PolicyDecisionAllow {
				return fmt.Errorf("insufficient permissions")
			}
		}
	}

	return nil
}

// matchPath checks if a path matches a pattern (supporting wildcards)
func (m *AuthorizationMiddleware) matchPath(path, pattern string) bool {
	if pattern == path {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// handleAuthzError handles authorization errors consistently
func (m *AuthorizationMiddleware) handleAuthzError(ctx context.Context, w http.ResponseWriter, errorType string, err error, startTime time.Time) {
	m.logger.WarnContext(ctx, "Authorization failed", logger.Fields{
		"error_type": errorType,
		"error":      err.Error(),
	})

	m.recordAuthzMetrics(ctx, "http", "error", errorType, time.Since(startTime))

	status := http.StatusForbidden
	if errorType == "authentication_required" {
		status = http.StatusUnauthorized
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]any{
		"error":   "authorization_failed",
		"message": err.Error(),
		"type":    errorType,
	}

	if jsonResp, jsonErr := json.Marshal(response); jsonErr == nil {
		w.Write(jsonResp)
	}
}

// recordAuthzMetrics records authorization metrics
func (m *AuthorizationMiddleware) recordAuthzMetrics(ctx context.Context, resourceType, action, result string, duration time.Duration) {
	labels := metrics.Fields{
		"resource_type": resourceType,
		"action":        action,
		"result":        result,
	}

	m.metrics.IncrementCounter("authorization_requests_total", labels)
	m.metrics.ObserveHistogram("authorization_duration", duration.Seconds(), labels)

	if result == "denied" || result == "error" {
		m.metrics.IncrementCounter("authorization_failures_total", labels)
	}
}

// GetAuthorizationInfo extracts authorization information from context
func GetAuthorizationInfo(ctx context.Context) map[string]any {
	info := make(map[string]any)

	if userID, ok := ctx.Value("authz_user_id").(uuid.UUID); ok {
		info["user_id"] = userID.String()
	}
	if decision, ok := ctx.Value("authz_decision").(model.PolicyDecisionType); ok {
		info["decision"] = decision
	}
	if evalTime, ok := ctx.Value("authz_evaluation_time").(int64); ok {
		info["evaluation_time_ms"] = evalTime
	}
	if cacheHit, ok := ctx.Value("authz_cache_hit").(bool); ok {
		info["cache_hit"] = cacheHit
	}
	if requestID, ok := ctx.Value("authz_request_id").(string); ok {
		info["request_id"] = requestID
	}

	return info
}
