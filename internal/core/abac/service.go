package abac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Service defines the ABAC service interface
type Service interface {
	// Policy Evaluation
	EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error)
	BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) (*BulkPermissionEvaluationResult, error)
	TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error)

	// User Permissions
	GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*UserEffectivePermissions, error)
	CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*RoleHierarchy, error)

	// Cache Management
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
	InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error
	GetCacheStatistics(ctx context.Context) (*CacheStatistics, error)
}

// service implements the ABAC service
type service struct {
	// Activities
	policyEvaluationActivities    *activities.PolicyEvaluationActivities
	attributeCollectionActivities *activities.AttributeCollectionActivities
	cacheActivities               *activities.CacheActivities

	// Repositories
	policyRepo           repository.PolicyRepository
	attributeRepo        repository.AttributeRepository
	policyEvaluationRepo repository.PolicyEvaluationRepository

	// External services
	identityService identity.Service
	tenantService   tenant.Service

	// Infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewService creates a new ABAC service instance
func NewService(
	policyRepo repository.PolicyRepository,
	attributeRepo repository.AttributeRepository,
	policyEvaluationRepo repository.PolicyEvaluationRepository,
	identityService identity.Service,
	tenantService tenant.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) Service {
	// Create activities
	policyEvaluationActivities := activities.NewPolicyEvaluationActivities(
		policyRepo, policyEvaluationRepo, logger, metrics, tracer)

	attributeCollectionActivities := activities.NewAttributeCollectionActivities(
		attributeRepo, identityService, tenantService, logger, metrics, tracer)

	cacheActivities := activities.NewCacheActivities(
		policyEvaluationRepo, attributeRepo, logger, metrics, tracer)

	return &service{
		policyEvaluationActivities:    policyEvaluationActivities,
		attributeCollectionActivities: attributeCollectionActivities,
		cacheActivities:               cacheActivities,
		policyRepo:                    policyRepo,
		attributeRepo:                 attributeRepo,
		policyEvaluationRepo:          policyEvaluationRepo,
		identityService:               identityService,
		tenantService:                 tenantService,
		logger:                        logger,
		metrics:                       metrics,
		tracer:                        tracer,
	}
}

// PermissionEvaluationRequest represents a permission evaluation request
type PermissionEvaluationRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Action       string                 `json:"action" validate:"required"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
}

// PermissionEvaluationResult represents the result of permission evaluation
type PermissionEvaluationResult struct {
	Decision         types.PolicyDecisionType `json:"decision"`
	PolicyDecisions  []*models.PolicyDecision `json:"policy_decisions"`
	EvaluationTimeMS int64                    `json:"evaluation_time_ms"`
	CacheHit         bool                     `json:"cache_hit"`
	RequestID        string                   `json:"request_id"`
	Timestamp        time.Time                `json:"timestamp"`
}

// EvaluatePermission evaluates a single permission request
func (s *service) EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.EvaluatePermission",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
		))
	defer span.End()

	startTime := time.Now()
	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	s.logger.InfoContext(ctx, "Starting permission evaluation",
		logger.Fields{
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"request_id":    requestID,
		})

	// Validate tenant context
	currentTenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for ABAC evaluation").WithErr(err)
	}

	// Collect user attributes
	userAttrInput := &activities.CollectUserAttributesInput{
		UserID:    req.UserID,
		EntityID:  req.EntityID,
		RequestID: requestID,
	}

	userAttrs, err := s.attributeCollectionActivities.CollectUserAttributes(ctx, userAttrInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "USER_ATTRIBUTE_COLLECTION_FAILED", "Failed to collect user attributes").WithErr(err)
	}

	// Collect resource attributes
	resourceAttrInput := &activities.CollectResourceAttributesInput{
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		EntityID:     req.EntityID,
		RequestID:    requestID,
	}

	resourceAttrs, err := s.attributeCollectionActivities.CollectResourceAttributes(ctx, resourceAttrInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "RESOURCE_ATTRIBUTE_COLLECTION_FAILED", "Failed to collect resource attributes").WithErr(err)
	}

	// Build environment context
	envContextInput := &activities.BuildEnvironmentContextInput{
		RequestTime: time.Now(),
		RequestID:   requestID,
	}

	// Extract context from request or context
	if ipAddress := s.extractIPFromContext(ctx); ipAddress != "" {
		envContextInput.IPAddress = ipAddress
	}
	if userAgent := s.extractUserAgentFromContext(ctx); userAgent != "" {
		envContextInput.UserAgent = userAgent
	}

	envContext, err := s.attributeCollectionActivities.BuildEnvironmentContext(ctx, envContextInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "ENVIRONMENT_CONTEXT_FAILED", "Failed to build environment context").WithErr(err)
	}

	// Combine all attributes
	allAttributes := make(map[string]interface{})

	// Add user attributes
	for k, v := range userAttrs.Attributes {
		allAttributes[k] = v
	}

	// Add resource attributes
	for k, v := range resourceAttrs.Attributes {
		allAttributes[k] = v
	}

	// Add environment context
	for k, v := range envContext.Context {
		allAttributes[k] = v
	}

	// Add request context
	for k, v := range req.Context {
		allAttributes[k] = v
	}

	// Add tenant context
	allAttributes["tenant.id"] = currentTenant.ID.String()
	allAttributes["tenant.name"] = currentTenant.Name
	allAttributes["tenant.status"] = currentTenant.Status

	// Evaluate policies
	evalInput := &activities.EvaluatePoliciesActivityInput{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
		Attributes:   allAttributes,
		RequestID:    requestID,
	}

	evalResult, err := s.policyEvaluationActivities.EvaluatePolicies(ctx, evalInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_FAILED", "Failed to evaluate policies").WithErr(err)
	}

	totalDuration := time.Since(startTime)
	span.SetAttributes(
		tracing.StringAttribute("decision", string(evalResult.Decision)),
		tracing.BoolAttribute("cache_hit", evalResult.CacheHit),
		tracing.IntAttribute("policies_evaluated", len(evalResult.PolicyDecisions)),
	)

	// Record metrics
	s.recordEvaluationMetrics(ctx, req, evalResult, totalDuration)

	// Audit log
	s.auditPermissionEvaluation(ctx, req, evalResult, totalDuration)

	result := &PermissionEvaluationResult{
		Decision:         evalResult.Decision,
		PolicyDecisions:  evalResult.PolicyDecisions,
		EvaluationTimeMS: totalDuration.Milliseconds(),
		CacheHit:         evalResult.CacheHit,
		RequestID:        requestID,
		Timestamp:        time.Now(),
	}

	s.logger.InfoContext(ctx, "Permission evaluation completed",
		logger.Fields{
			"decision":           result.Decision,
			"policies_evaluated": len(result.PolicyDecisions),
			"evaluation_time_ms": result.EvaluationTimeMS,
			"cache_hit":          result.CacheHit,
			"request_id":         requestID,
		})

	return result, nil
}

// BulkPermissionEvaluationRequest represents a bulk permission evaluation request
type BulkPermissionEvaluationRequest struct {
	Requests  []*PermissionEvaluationRequest `json:"requests" validate:"required,min=1,max=100"`
	RequestID string                         `json:"request_id,omitempty"`
}

// BulkPermissionEvaluationResult represents the result of bulk permission evaluation
type BulkPermissionEvaluationResult struct {
	Results         []*PermissionEvaluationResult `json:"results"`
	TotalRequests   int                           `json:"total_requests"`
	SuccessfulCount int                           `json:"successful_count"`
	FailedCount     int                           `json:"failed_count"`
	TotalTimeMS     int64                         `json:"total_time_ms"`
	AverageTimeMS   float64                       `json:"average_time_ms"`
	RequestID       string                        `json:"request_id"`
	Timestamp       time.Time                     `json:"timestamp"`
}

// BulkEvaluatePermissions evaluates multiple permission requests
func (s *service) BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) (*BulkPermissionEvaluationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.BulkEvaluatePermissions",
		tracing.WithAttributes(
			tracing.IntAttribute("request_count", len(req.Requests)),
		))
	defer span.End()

	startTime := time.Now()
	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	s.logger.InfoContext(ctx, "Starting bulk permission evaluation",
		logger.Fields{
			"request_count": len(req.Requests),
			"request_id":    requestID,
		})

	results := make([]*PermissionEvaluationResult, len(req.Requests))
	successCount := 0
	failedCount := 0

	// Process requests concurrently (simplified implementation)
	for i, request := range req.Requests {
		request.RequestID = fmt.Sprintf("%s-%d", requestID, i)

		result, err := s.EvaluatePermission(ctx, request)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to evaluate permission in bulk request",
				logger.Fields{
					"index":      i,
					"user_id":    request.UserID,
					"resource":   request.ResourceType,
					"action":     request.Action,
					"error":      err.Error(),
					"request_id": request.RequestID,
				})
			failedCount++

			// Create error result
			results[i] = &PermissionEvaluationResult{
				Decision:         types.PolicyDecisionDeny,
				PolicyDecisions:  []*models.PolicyDecision{},
				EvaluationTimeMS: 0,
				CacheHit:         false,
				RequestID:        request.RequestID,
				Timestamp:        time.Now(),
			}
		} else {
			results[i] = result
			successCount++
		}
	}

	totalDuration := time.Since(startTime)
	averageTime := float64(totalDuration.Milliseconds()) / float64(len(req.Requests))

	span.SetAttributes(
		tracing.IntAttribute("successful_count", successCount),
		tracing.IntAttribute("failed_count", failedCount),
	)

	// Record bulk metrics
	s.metrics.IncrementCounter("abac_bulk_evaluations_total",
		metrics.Fields{"request_count": fmt.Sprintf("%d", len(req.Requests))})
	s.metrics.ObserveHistogram("abac_bulk_evaluation_duration_seconds", totalDuration.Seconds(),
		metrics.Fields{"request_count": fmt.Sprintf("%d", len(req.Requests))})

	result := &BulkPermissionEvaluationResult{
		Results:         results,
		TotalRequests:   len(req.Requests),
		SuccessfulCount: successCount,
		FailedCount:     failedCount,
		TotalTimeMS:     totalDuration.Milliseconds(),
		AverageTimeMS:   averageTime,
		RequestID:       requestID,
		Timestamp:       time.Now(),
	}

	s.logger.InfoContext(ctx, "Bulk permission evaluation completed",
		logger.Fields{
			"total_requests":   result.TotalRequests,
			"successful_count": result.SuccessfulCount,
			"failed_count":     result.FailedCount,
			"average_time_ms":  result.AverageTimeMS,
			"request_id":       requestID,
		})

	return result, nil
}

// PolicyTestRequest represents a policy test request
type PolicyTestRequest struct {
	PolicyID    uuid.UUID              `json:"policy_id" validate:"required"`
	TestContext map[string]interface{} `json:"test_context" validate:"required"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// PolicyTestResult represents the result of policy testing
type PolicyTestResult struct {
	PolicyID         uuid.UUID              `json:"policy_id"`
	PolicyName       string                 `json:"policy_name"`
	EvaluationResult *models.PolicyDecision `json:"evaluation_result"`
	EvaluationTimeMS int64                  `json:"evaluation_time_ms"`
	RequestID        string                 `json:"request_id"`
	Timestamp        time.Time              `json:"timestamp"`
}

// TestPolicy tests a specific policy against provided context
func (s *service) TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.TestPolicy",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_id", req.PolicyID.String()),
		))
	defer span.End()

	startTime := time.Now()
	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	s.logger.InfoContext(ctx, "Starting policy test",
		logger.Fields{
			"policy_id":  req.PolicyID,
			"request_id": requestID,
		})

	// Get policy
	policy, err := s.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_test_policy", "policy_not_found")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found for testing").WithErr(err)
	}

	// This is a simplified policy test implementation
	// In a full implementation, you would use the policy evaluation activities
	decision := &models.PolicyDecision{
		PolicyID:   policy.ID,
		PolicyName: policy.Name,
		Effect:     policy.Effect,
		Reason:     "Policy test evaluation",
	}

	totalDuration := time.Since(startTime)

	result := &PolicyTestResult{
		PolicyID:         policy.ID,
		PolicyName:       policy.Name,
		EvaluationResult: decision,
		EvaluationTimeMS: totalDuration.Milliseconds(),
		RequestID:        requestID,
		Timestamp:        time.Now(),
	}

	s.metrics.IncrementSuccessCount("abac_test_policy")
	s.metrics.ObserveHistogram("abac_test_policy_duration_seconds", totalDuration.Seconds(),
		metrics.Fields{"policy_id": policy.ID.String()})

	s.logger.InfoContext(ctx, "Policy test completed",
		logger.Fields{
			"policy_id":          result.PolicyID,
			"policy_name":        result.PolicyName,
			"evaluation_time_ms": result.EvaluationTimeMS,
			"request_id":         requestID,
		})

	return result, nil
}

// UserEffectivePermissions represents a user's effective permissions
type UserEffectivePermissions struct {
	UserID      uuid.UUID           `json:"user_id"`
	EntityID    *uuid.UUID          `json:"entity_id,omitempty"`
	Permissions map[string][]string `json:"permissions"` // resource_type -> actions
	Roles       []string            `json:"roles"`
	Timestamp   time.Time           `json:"timestamp"`
}

// GetUserEffectivePermissions gets the effective permissions for a user
func (s *service) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*UserEffectivePermissions, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.GetUserEffectivePermissions",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", userID.String()),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Getting user effective permissions",
		logger.Fields{
			"user_id":   userID,
			"entity_id": entityID,
		})

	// Get user roles from identity service
	userRoles, err := s.identityService.GetUserRoles(ctx, userID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_get_effective_permissions", "get_user_roles_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "GET_USER_ROLES_FAILED", "Failed to retrieve user roles").WithErr(err)
	}

	roles := make([]string, len(userRoles))
	for i, role := range userRoles {
		roles[i] = role.Name
	}

	// This is a placeholder implementation for permissions.
	// In a full implementation, you would evaluate policies based on user roles and other attributes.
	permissions := make(map[string][]string)

	result := &UserEffectivePermissions{
		UserID:      userID,
		EntityID:    entityID,
		Permissions: permissions,
		Roles:       roles,
		Timestamp:   time.Now(),
	}

	s.metrics.IncrementSuccessCount("abac_get_effective_permissions")
	s.metrics.RecordGauge("abac_user_effective_roles_count", float64(len(roles)),
		metrics.Fields{"user_id": userID.String()})

	s.logger.InfoContext(ctx, "User effective permissions retrieved",
		logger.Fields{
			"user_id":           userID,
			"role_count":        len(roles),
			"permissions_count": len(permissions),
		})

	return result, nil
}

// RoleHierarchy represents a user's role hierarchy
type RoleHierarchy struct {
	UserID    uuid.UUID           `json:"user_id"`
	EntityID  *uuid.UUID          `json:"entity_id,omitempty"`
	Roles     []string            `json:"roles"`
	Hierarchy map[string][]string `json:"hierarchy"` // role -> inherited roles
	Timestamp time.Time           `json:"timestamp"`
}

// CalculateRoleHierarchy calculates the role hierarchy for a user
func (s *service) CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*RoleHierarchy, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.CalculateRoleHierarchy",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", userID.String()),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Calculating role hierarchy",
		logger.Fields{
			"user_id":   userID,
			"entity_id": entityID,
		})

	// Get user roles from identity service
	userRoles, err := s.identityService.GetUserRoles(ctx, userID)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_calculate_role_hierarchy", "get_user_roles_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "GET_USER_ROLES_FAILED", "Failed to retrieve user roles for hierarchy calculation").WithErr(err)
	}

	roles := make([]string, len(userRoles))
	for i, role := range userRoles {
		roles[i] = role.Name
	}

	// This is a placeholder implementation for hierarchy calculation.
	// In a full implementation, you would traverse role relationships (e.g., parent roles).
	hierarchy := make(map[string][]string)
	for _, role := range roles {
		hierarchy[role] = []string{} // For now, no inherited roles
	}

	result := &RoleHierarchy{
		UserID:    userID,
		EntityID:  entityID,
		Roles:     roles,
		Hierarchy: hierarchy,
		Timestamp: time.Now(),
	}

	s.metrics.IncrementSuccessCount("abac_calculate_role_hierarchy")
	s.metrics.RecordGauge("abac_user_role_hierarchy_depth", float64(1), // Placeholder depth
		metrics.Fields{"user_id": userID.String()})

	s.logger.InfoContext(ctx, "Role hierarchy calculated",
		logger.Fields{
			"user_id":           userID,
			"role_count":        len(roles),
			"hierarchy_entries": len(hierarchy),
		})

	return result, nil
}

// InvalidateUserCache invalidates cache entries for a specific user
func (s *service) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.InvalidateUserCache",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", userID.String()),
		))
	defer span.End()

	input := &activities.InvalidatePolicyCacheInput{
		UserID:    &userID,
		RequestID: uuid.New().String(),
	}

	_, err := s.cacheActivities.InvalidatePolicyCache(ctx, input)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_invalidate_user_cache", "activity_failed")
		return errors.NewBusinessErrorWithContext(ctx, "CACHE_INVALIDATION_FAILED", "Failed to invalidate user cache").WithErr(err)
	}

	s.metrics.IncrementSuccessCount("abac_invalidate_user_cache")
	s.logger.InfoContext(ctx, "User cache invalidated", logger.Fields{"user_id": userID})

	return nil
}

// InvalidatePolicyCache invalidates cache entries for specific policies
func (s *service) InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.InvalidatePolicyCache",
		tracing.WithAttributes(
			tracing.IntAttribute("policy_count", len(policyIDs)),
		))
	defer span.End()

	input := &activities.InvalidatePolicyCacheInput{
		PolicyIDs: policyIDs,
		RequestID: uuid.New().String(),
	}

	_, err := s.cacheActivities.InvalidatePolicyCache(ctx, input)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_invalidate_policy_cache", "activity_failed")
		return errors.NewBusinessErrorWithContext(ctx, "CACHE_INVALIDATION_FAILED", "Failed to invalidate policy cache").WithErr(err)
	}

	s.metrics.IncrementSuccessCount("abac_invalidate_policy_cache")
	s.logger.InfoContext(ctx, "Policy cache invalidated", logger.Fields{"policy_ids_count": len(policyIDs)})

	return nil
}

// CacheStatistics represents cache performance statistics
type CacheStatistics struct {
	PolicyEvaluationStats *repository.EvaluationCacheStats `json:"policy_evaluation_stats"`
	AttributeStats        *repository.AttributeStats       `json:"attribute_stats"`
	GeneratedAt           time.Time                        `json:"generated_at"`
}

// GetCacheStatistics gets cache performance statistics
func (s *service) GetCacheStatistics(ctx context.Context) (*CacheStatistics, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.service.GetCacheStatistics")
	defer span.End()

	input := &activities.GetCacheStatsInput{
		RequestID: uuid.New().String(),
	}

	stats, err := s.cacheActivities.GetCacheStats(ctx, input)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementErrorCount("abac_get_cache_statistics", "activity_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "GET_CACHE_STATS_FAILED", "Failed to retrieve cache statistics").WithErr(err)
	}

	result := &CacheStatistics{
		PolicyEvaluationStats: stats.PolicyEvaluationStats,
		AttributeStats:        stats.AttributeStats,
		GeneratedAt:           stats.GeneratedAt,
	}

	s.metrics.IncrementSuccessCount("abac_get_cache_statistics")
	s.logger.InfoContext(ctx, "Cache statistics retrieved",
		logger.Fields{
			"total_cached_evaluations": stats.PolicyEvaluationStats.TotalCachedEvaluations,
			"cache_hit_rate":           stats.PolicyEvaluationStats.CacheHitRate,
		})

	return result, nil
}

// Helper methods

// extractIPFromContext extracts IP address from context
func (s *service) extractIPFromContext(ctx context.Context) string {
	// This would extract from Gin context or other HTTP context
	// Implementation depends on your HTTP framework
	return ""
}

// extractUserAgentFromContext extracts user agent from context
func (s *service) extractUserAgentFromContext(ctx context.Context) string {
	// This would extract from Gin context or other HTTP context
	// Implementation depends on your HTTP framework
	return ""
}

// recordEvaluationMetrics records metrics for permission evaluation
func (s *service) recordEvaluationMetrics(ctx context.Context, req *PermissionEvaluationRequest, result *activities.EvaluatePoliciesActivityOutput, duration time.Duration) {
	labels := metrics.Fields{
		"decision":      string(result.Decision),
		"resource_type": req.ResourceType,
		"action":        req.Action,
		"cache_hit":     fmt.Sprintf("%t", result.CacheHit),
	}

	s.metrics.IncrementCounter("abac_permission_evaluations_total", labels)
	s.metrics.ObserveHistogram("abac_permission_evaluation_duration_seconds", duration.Seconds(), labels)
	s.metrics.RecordGauge("abac_permission_policies_evaluated", float64(len(result.PolicyDecisions)), labels)
}

// auditPermissionEvaluation creates an audit log entry
func (s *service) auditPermissionEvaluation(ctx context.Context, req *PermissionEvaluationRequest, result *activities.EvaluatePoliciesActivityOutput, duration time.Duration) {
	s.logger.InfoContext(ctx, "ABAC permission evaluation audit",
		logger.Fields{
			"event_type":         "permission_evaluation",
			"user_id":            req.UserID,
			"resource_type":      req.ResourceType,
			"resource_id":        req.ResourceID,
			"action":             req.Action,
			"entity_id":          req.EntityID,
			"decision":           result.Decision,
			"policies_evaluated": len(result.PolicyDecisions),
			"evaluation_time_ms": duration.Milliseconds(),
			"cache_hit":          result.CacheHit,
			"request_id":         req.RequestID,
		})
}
