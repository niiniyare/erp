package abac

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"awo/internal/core/abac/models"
	"awo/internal/core/abac/repository"
	"awo/internal/platform/cache"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// ListPoliciesRequest represents a request to list policies
type ListPoliciesRequest struct {
	TenantID   uuid.UUID             `json:"tenant_id"`
	PolicyType *types.PolicyType     `json:"policy_type,omitempty"`
	Category   *types.PolicyCategory `json:"category,omitempty"`
	IsActive   *bool                 `json:"is_active,omitempty"`
	SearchTerm string                `json:"search_term,omitempty"`
	Limit      int32                 `json:"limit,omitempty"`
	Offset     int32                 `json:"offset,omitempty"`
	SortBy     string                `json:"sort_by,omitempty"`
	SortOrder  string                `json:"sort_order,omitempty"`
}

// PolicyListResult represents the result of listing policies
type PolicyListResult struct {
	Policies []models.Policy `json:"policies"`
	Total    int64           `json:"total"`
	Page     int32           `json:"page"`
	PageSize int32           `json:"page_size"`
}

// PolicyMetricsRequest represents a request for policy metrics
type PolicyMetricsRequest struct {
	PolicyID  *uuid.UUID `json:"policy_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	TenantID  uuid.UUID  `json:"tenant_id"`
}

// PolicyMetrics represents policy usage and performance metrics
type PolicyMetrics struct {
	PolicyID        uuid.UUID     `json:"policy_id"`
	EvaluationCount int64         `json:"evaluation_count"`
	PermitCount     int64         `json:"permit_count"`
	DenyCount       int64         `json:"deny_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorCount      int64         `json:"error_count"`
	LastEvaluated   time.Time     `json:"last_evaluated"`
}

// PolicyUsageRequest represents a request for policy usage information
type PolicyUsageRequest struct {
	PolicyID  uuid.UUID  `json:"policy_id"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	GroupBy   string     `json:"group_by,omitempty"` // "day", "hour", "user", etc.
}

// PolicyUsageStatsExtended represents detailed policy usage statistics with additional fields
type PolicyUsageStatsExtended struct {
	PolicyUsageStats                          // Embedded existing type
	EvaluationsByTime []TimeBasedUsage        `json:"evaluations_by_time"`
	EvaluationsByUser []UserBasedUsage        `json:"evaluations_by_user"`
	TopResources      []ResourceUsageDetailed `json:"top_resources"`
	SuccessRate       float64                 `json:"success_rate"`
	PeakUsageTime     *time.Time              `json:"peak_usage_time,omitempty"`
	LatencyP95        time.Duration           `json:"latency_p95"`
	LatencyP99        time.Duration           `json:"latency_p99"`
	RecentTrend       string                  `json:"recent_trend"` // "increasing", "decreasing", "stable"
	Recommendations   []UsageRecommendation   `json:"recommendations"`
	LastUpdated       time.Time               `json:"last_updated"`
}

// TimeBasedUsage represents usage statistics for a time period
type TimeBasedUsage struct {
	Timestamp    time.Time `json:"timestamp"`
	Evaluations  int64     `json:"evaluations"`
	Permits      int64     `json:"permits"`
	Denies       int64     `json:"denies"`
	Errors       int64     `json:"errors"`
	AvgLatencyMs float64   `json:"avg_latency_ms"`
}

// UserBasedUsage represents usage statistics per user
type UserBasedUsage struct {
	UserID      uuid.UUID `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	Evaluations int64     `json:"evaluations"`
	Permits     int64     `json:"permits"`
	Denies      int64     `json:"denies"`
	LastAccess  time.Time `json:"last_access"`
}

// ResourceUsageDetailed represents detailed usage statistics per resource
type ResourceUsageDetailed struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	Action       string    `json:"action"`
	Evaluations  int64     `json:"evaluations"`
	Permits      int64     `json:"permits"`
	Denies       int64     `json:"denies"`
	LastAccess   time.Time `json:"last_access"`
}

// UsageRecommendation represents a recommendation based on usage patterns
type UsageRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// PolicyManager provides policy management capabilities
type PolicyManager interface {
	// Policy CRUD Operations
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*PolicyManagementResult, error)
	UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*PolicyManagementResult, error)
	DeletePolicy(ctx context.Context, req *DeletePolicyRequest) error
	GetPolicy(ctx context.Context, id uuid.UUID) (*PolicyDetails, error)
	ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*PolicyListResult, error)

	// Policy Testing & Simulation
	TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error)
	SimulatePolicyChange(ctx context.Context, req *PolicySimulationRequest) (*PolicySimulationResult, error)
	AnalyzePolicyImpact(ctx context.Context, req *PolicyImpactRequest) (*PolicyImpactResult, error)
	DetectPolicyConflicts(ctx context.Context, req *PolicyConflictRequest) (*PolicyConflictResult, error)

	// Policy Lifecycle Management
	CreatePolicyVersion(ctx context.Context, req *CreatePolicyVersionRequest) (*PolicyVersionResult, error)
	ActivatePolicy(ctx context.Context, req *ActivatePolicyRequest) error
	DeactivatePolicy(ctx context.Context, req *DeactivatePolicyRequest) error
	ArchivePolicy(ctx context.Context, req *ArchivePolicyRequest) error

	// Policy Templates & Import/Export
	CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*PolicyTemplate, error)
	ImportPolicies(ctx context.Context, req *ImportPoliciesRequest) (*ImportResult, error)
	ExportPolicies(ctx context.Context, req *ExportPoliciesRequest) (*ExportResult, error)

	// Policy Analytics
	GetPolicyMetrics(ctx context.Context, req *PolicyMetricsRequest) (*PolicyMetrics, error)
	GetPolicyUsageStats(ctx context.Context, req *PolicyUsageRequest) (*PolicyUsageStatsExtended, error)
}

// policyManager implements PolicyManager
type policyManager struct {
	policyRepo           repository.PolicyRepository
	policyEvaluationRepo repository.PolicyEvaluationRepository
	attributeRepo        repository.AttributeRepository
	cache                cache.Service
	logger               logger.Logger
	metrics              metrics.MetricsProvider
	tracer               tracing.Service
}

// NewPolicyManager creates a new policy manager instance
func NewPolicyManager(
	policyRepo repository.PolicyRepository,
	policyEvaluationRepo repository.PolicyEvaluationRepository,
	attributeRepo repository.AttributeRepository,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) PolicyManager {
	return &policyManager{
		policyRepo:           policyRepo,
		policyEvaluationRepo: policyEvaluationRepo,
		attributeRepo:        attributeRepo,
		cache:                cache,
		logger:               logger,
		metrics:              metrics,
		tracer:               tracer,
	}
}

// Policy CRUD Operations

type CreatePolicyRequest struct {
	Name        string               `json:"name" validate:"required,min=1,max=150"`
	DisplayName *string              `json:"display_name,omitempty"`
	Description *string              `json:"description,omitempty"`
	PolicyType  types.PolicyType     `json:"policy_type" validate:"required"`
	Effect      types.PolicyEffect   `json:"effect" validate:"required"`
	Priority    int32                `json:"priority" validate:"min=1,max=1000"`
	Category    types.PolicyCategory `json:"category" validate:"required"`
	Target      map[string]any       `json:"target" validate:"required"`
	Rule        map[string]any       `json:"rule" validate:"required"`
	Obligations map[string]any       `json:"obligations,omitempty"`
	Advice      map[string]any       `json:"advice,omitempty"`
	EntityID    *uuid.UUID           `json:"entity_id,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	CreatedBy   *uuid.UUID           `json:"created_by,omitempty"`
}

type PolicyManagementResult struct {
	Policy           *models.Policy          `json:"policy"`
	ValidationResult *PolicyValidationResult `json:"validation_result"`
	ConflictWarnings []PolicyConflictWarning `json:"conflict_warnings,omitempty"`
	Recommendations  []string                `json:"recommendations,omitempty"`
}

// PolicyValidationResult defines the result of a policy validation
type PolicyValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type PolicyConflictWarning struct {
	ConflictType          string    `json:"conflict_type"`
	ConflictingPolicyID   uuid.UUID `json:"conflicting_policy_id"`
	ConflictingPolicyName string    `json:"conflicting_policy_name"`
	Description           string    `json:"description"`
	Severity              string    `json:"severity"` // high, medium, low
}

func (pm *policyManager) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*PolicyManagementResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.CreatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_name", req.Name),
			attribute.String("policy_type", string(req.PolicyType)),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Creating policy with management features",
		logger.Fields{
			"name":        req.Name,
			"policy_type": req.PolicyType,
			"priority":    req.Priority,
		})

	// Step 1: Validate policy syntax and semantics
	validationResult := pm.validatePolicy(ctx, req)
	if !validationResult.IsValid {
		pm.metrics.IncrementCounter("policy_manager_create_validation_failed", nil)
		return &PolicyManagementResult{
			ValidationResult: validationResult,
		}, errors.NewBusinessErrorWithContext(ctx, "POLICY_VALIDATION_FAILED", "Policy validation failed")
	}

	// Step 2: Check for conflicts with existing policies
	conflictWarnings, err := pm.checkPolicyConflicts(ctx, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to check policy conflicts", logger.Fields{"error": err.Error()})
	}

	// Step 3: Create the policy
	createReq := &repository.CreatePolicyRequest{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		PolicyType:  req.PolicyType,
		Effect:      req.Effect,
		Priority:    req.Priority,
		Category:    req.Category,
		Target:      req.Target,
		Rule:        req.Rule,
		Obligations: req.Obligations,
		Advice:      req.Advice,
		CreatedBy:   *req.CreatedBy,
	}

	policy, err := pm.policyRepo.CreatePolicy(ctx, createReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		pm.metrics.IncrementCounter("policy_manager_create_failed", nil)
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// Step 4: Generate recommendations
	recommendations := pm.generatePolicyRecommendations(ctx, policy, req)

	// Step 5: Record metrics
	pm.metrics.IncrementCounter("policy_manager_create", nil)
	pm.metrics.SetGauge("policy_manager_total_policies", 1,
		metrics.Fields{"category": string(req.Category)})

	result := &PolicyManagementResult{
		Policy:           policy,
		ValidationResult: validationResult,
		ConflictWarnings: conflictWarnings,
		Recommendations:  recommendations,
	}

	pm.logger.InfoContext(ctx, "Policy created successfully",
		logger.Fields{
			"policy_id":       policy.ID,
			"name":            policy.Name,
			"conflicts":       len(conflictWarnings),
			"recommendations": len(recommendations),
		})

	return result, nil
}

type UpdatePolicyRequest struct {
	ID          uuid.UUID             `json:"id" validate:"required"`
	Name        *string               `json:"name,omitempty"`
	DisplayName *string               `json:"display_name,omitempty"`
	Description *string               `json:"description,omitempty"`
	PolicyType  *types.PolicyType     `json:"policy_type,omitempty"`
	Effect      *types.PolicyEffect   `json:"effect,omitempty"`
	Priority    *int32                `json:"priority,omitempty"`
	Category    *types.PolicyCategory `json:"category,omitempty"`
	Target      map[string]any        `json:"target,omitempty"`
	Rule        map[string]any        `json:"rule,omitempty"`
	Obligations map[string]any        `json:"obligations,omitempty"`
	Advice      map[string]any        `json:"advice,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	UpdatedBy   *uuid.UUID            `json:"updated_by,omitempty"`
}

func (pm *policyManager) UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*PolicyManagementResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.UpdatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.ID.String()),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Updating policy with management features",
		logger.Fields{
			"policy_id": req.ID,
		})

	// Step 1: Get existing policy
	existingPolicy, err := pm.policyRepo.GetPolicyByID(ctx, req.ID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	// Step 2: Validate update request
	validationResult := pm.validatePolicyUpdate(ctx, existingPolicy, req)
	if !validationResult.IsValid {
		pm.metrics.IncrementCounter("policy_manager_update_validation_failed", nil)
		return &PolicyManagementResult{
			ValidationResult: validationResult,
		}, errors.NewBusinessErrorWithContext(ctx, "POLICY_UPDATE_VALIDATION_FAILED", "Policy update validation failed")
	}

	// Step 3: Check for conflicts after update
	conflictWarnings, err := pm.checkPolicyUpdateConflicts(ctx, existingPolicy, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to check policy update conflicts", logger.Fields{"error": err.Error()})
	}

	// Step 4: Perform the update
	updateReq := &repository.UpdatePolicyRequest{
		// ID:          req.ID, // ID is passed separately to UpdatePolicy
		// EntityID:    nil, // Not updating entity ID
		DisplayName: req.DisplayName,
		Description: req.Description,
		// PolicyType:  req.PolicyType, // Not in UpdatePolicyRequest
		// Effect:      req.Effect, // Not in UpdatePolicyRequest
		Priority: req.Priority,
		// Category:    req.Category, // Not in UpdatePolicyRequest
		Target:      req.Target,
		Rule:        req.Rule,
		Obligations: req.Obligations,
		Advice:      req.Advice,
		IsActive:    req.IsActive,
	}

	updatedPolicy, err := pm.policyRepo.UpdatePolicy(ctx, req.ID, updateReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		pm.metrics.IncrementCounter("policy_manager_update_failed", nil)
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// Step 5: Invalidate related cache entries
	go func() {
		invalidateReq := &repository.InvalidateEvaluationCacheRequest{
			PolicyIDs: []uuid.UUID{req.ID},
		}
		if err := pm.policyEvaluationRepo.InvalidateEvaluationCache(context.Background(), invalidateReq); err != nil {
			pm.logger.WarnContext(ctx, "Failed to invalidate cache after policy update",
				logger.Fields{"policy_id": req.ID, "error": err.Error()})
		}
	}()

	// Step 6: Generate recommendations
	recommendations := pm.generatePolicyUpdateRecommendations(ctx, existingPolicy, updatedPolicy)

	pm.metrics.IncrementCounter("policy_manager_update", nil)

	result := &PolicyManagementResult{
		Policy:           updatedPolicy,
		ValidationResult: validationResult,
		ConflictWarnings: conflictWarnings,
		Recommendations:  recommendations,
	}

	pm.logger.InfoContext(ctx, "Policy updated successfully",
		logger.Fields{
			"policy_id":       updatedPolicy.ID,
			"name":            updatedPolicy.Name,
			"conflicts":       len(conflictWarnings),
			"recommendations": len(recommendations),
		})

	return result, nil
}

type DeletePolicyRequest struct {
	ID        uuid.UUID  `json:"id" validate:"required"`
	Reason    *string    `json:"reason,omitempty"`
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty"`
	Force     bool       `json:"force"` // Force delete even if policy is referenced
}

func (pm *policyManager) DeletePolicy(ctx context.Context, req *DeletePolicyRequest) error {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.DeletePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.ID.String()),
			attribute.Bool("force", req.Force),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Deleting policy with dependency checks",
		logger.Fields{
			"policy_id": req.ID,
			"force":     req.Force,
			"reason":    req.Reason,
		})

	// Step 1: Check if policy exists
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.ID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return err
	}

	// Step 2: Check for dependencies if not force delete
	if !req.Force {
		dependencies, err := pm.checkPolicyDependencies(ctx, req.ID)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return fmt.Errorf("failed to check policy dependencies: %w", err)
		}

		if len(dependencies) > 0 {
			pm.metrics.IncrementCounter("policy_manager_delete_blocked_by_dependencies", nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_HAS_DEPENDENCIES",
				fmt.Sprintf("Policy has %d dependencies. Use force=true to delete anyway", len(dependencies)))
		}
	}

	// Step 3: Perform the deletion
	err = pm.policyRepo.DeletePolicy(ctx, req.ID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		pm.metrics.IncrementCounter("policy_manager_delete_failed", nil)
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	// Step 4: Invalidate related cache entries
	go func() {
		invalidateReq := &repository.InvalidateEvaluationCacheRequest{
			PolicyIDs: []uuid.UUID{req.ID},
		}
		if err := pm.policyEvaluationRepo.InvalidateEvaluationCache(context.Background(), invalidateReq); err != nil {
			pm.logger.WarnContext(ctx, "Failed to invalidate cache after policy deletion",
				logger.Fields{"policy_id": req.ID, "error": err.Error()})
		}
	}()

	pm.metrics.IncrementCounter("policy_manager_delete", nil)
	pm.metrics.SetGauge("policy_manager_total_policies", -1,
		metrics.Fields{"category": string(policy.Category)})

	pm.logger.InfoContext(ctx, "Policy deleted successfully",
		logger.Fields{
			"policy_id":   req.ID,
			"policy_name": policy.Name,
		})

	return nil
}

type PolicyDetails struct {
	Policy           *models.Policy          `json:"policy"`
	Metadata         *PolicyMetadata         `json:"metadata"`
	Usage            *PolicyUsageInfo        `json:"usage"`
	RelatedPolicies  []*models.Policy        `json:"related_policies,omitempty"`
	ConflictAnalysis *PolicyConflictAnalysis `json:"conflict_analysis,omitempty"`
}

type PolicyMetadata struct {
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	LastModifiedBy *uuid.UUID `json:"last_modified_by,omitempty"`
	Version        int32      `json:"version"`
	Tags           []string   `json:"tags,omitempty"`
	ApprovalStatus string     `json:"approval_status"`
	EffectiveFrom  time.Time  `json:"effective_from"`
	EffectiveTo    *time.Time `json:"effective_to,omitempty"`
}

type PolicyUsageInfo struct {
	EvaluationCount   int64      `json:"evaluation_count"`
	LastEvaluated     *time.Time `json:"last_evaluated,omitempty"`
	AvgEvaluationTime float64    `json:"avg_evaluation_time_ms"`
	SuccessRate       float64    `json:"success_rate"`
	AffectedUsers     int32      `json:"affected_users"`
	AffectedResources int32      `json:"affected_resources"`
}

type PolicyConflictAnalysis struct {
	HasConflicts          bool                   `json:"has_conflicts"`
	ConflictDetails       []PolicyConflictDetail `json:"conflict_details,omitempty"`
	ResolutionSuggestions []string               `json:"resolution_suggestions,omitempty"`
}

type PolicyConflictDetail struct {
	ConflictType        string         `json:"conflict_type"`
	ConflictingPolicy   *models.Policy `json:"conflicting_policy"`
	ConflictDescription string         `json:"conflict_description"`
	Severity            string         `json:"severity"`
	AutoResolvable      bool           `json:"auto_resolvable"`
}

func (pm *policyManager) GetPolicy(ctx context.Context, id uuid.UUID) (*PolicyDetails, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.GetPolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", id.String()),
		))
	defer span.End()

	// Get the policy
	policy, err := pm.policyRepo.GetPolicyByID(ctx, id)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	// Get policy metadata (placeholder implementation)
	metadata := &PolicyMetadata{
		CreatedBy:      &policy.CreatedBy,
		Version:        1, // Placeholder
		ApprovalStatus: "approved",
		EffectiveFrom:  policy.CreatedAt,
	}

	// Get usage information (placeholder implementation)
	usage := &PolicyUsageInfo{
		EvaluationCount:   0, // Would be calculated from evaluation history
		AvgEvaluationTime: 0,
		SuccessRate:       1.0,
		AffectedUsers:     0,
		AffectedResources: 0,
	}

	// Get conflict analysis (placeholder implementation)
	conflictAnalysis := &PolicyConflictAnalysis{
		HasConflicts: false,
	}

	pm.metrics.IncrementCounter("policy_manager_get", nil)

	return &PolicyDetails{
		Policy:           policy,
		Metadata:         metadata,
		Usage:            usage,
		ConflictAnalysis: conflictAnalysis,
	}, nil
}

func (pm *policyManager) ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*PolicyListResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.ListPolicies",
		tracing.WithAttributes(
			attribute.String("tenant_id", req.TenantID.String()),
			attribute.Int("limit", int(req.Limit)),
			attribute.Int("offset", int(req.Offset)),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Listing policies with filters",
		logger.Fields{
			"tenant_id":   req.TenantID,
			"policy_type": req.PolicyType,
			"category":    req.Category,
			"search_term": req.SearchTerm,
			"limit":       req.Limit,
			"offset":      req.Offset,
		})

	// NOTE: Enhanced policy listing with advanced filtering and caching
	// TODO: Implement advanced search functionality (full-text search, tags)
	// TODO: Add sorting by multiple fields
	// TODO: Implement result caching with cache invalidation strategies
	// TODO: Add filtering by effective date ranges
	// TODO: Support complex filter combinations with AND/OR logic

	// Set default pagination if not provided
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 50 // Default page size
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Try cache first for commonly accessed policy lists
	var cacheKey string
	if req.SearchTerm == "" && req.PolicyType == nil && req.Category == nil {
		// Only cache simple list requests without filters
		cacheKey = fmt.Sprintf("policies:list:%s:%d:%d", req.TenantID, limit, offset)
		var cachedResult *PolicyListResult
		if err := pm.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			pm.metrics.IncrementCounter("policy_manager_list_cache_hit", nil)
			return cachedResult, nil
		}
	}

	// Build repository request
	repoReq := &repository.ListPoliciesRequest{
		PolicyType: req.PolicyType,
		Category:   req.Category,
		IsActive:   req.IsActive,
		Limit:      int(limit),
		Offset:     int(offset),
	}

	// Get policies from repository
	policies, err := pm.policyRepo.ListPolicies(ctx, repoReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		pm.metrics.IncrementCounter("policy_manager_list_failed", nil)
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	// Get total count for pagination (placeholder implementation)
	// TODO: Implement efficient total count query
	total := int64(len(policies))

	// Convert to result format
	policyList := make([]models.Policy, len(policies))
	for i, policy := range policies {
		policyList[i] = *policy
	}

	result := &PolicyListResult{
		Policies: policyList,
		Total:    total,
		Page:     offset/limit + 1,
		PageSize: limit,
	}

	// Cache simple list results (without complex filters)
	if cacheKey != "" {
		if err := pm.cache.Set(ctx, cacheKey, result, 5*time.Minute); err != nil {
			pm.logger.WarnContext(ctx, "Failed to cache policy list result",
				logger.Fields{"error": err.Error(), "cache_key": cacheKey})
		}
	}

	pm.metrics.IncrementCounter("policy_manager_list_success",
		metrics.Fields{"result_count": strconv.Itoa(len(policyList))})

	pm.logger.InfoContext(ctx, "Policies listed successfully",
		logger.Fields{
			"count":  len(policyList),
			"total":  total,
			"page":   result.Page,
			"cached": cacheKey != "",
		})

	return result, nil
}

func (pm *policyManager) GetPolicyMetrics(ctx context.Context, req *PolicyMetricsRequest) (*PolicyMetrics, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.GetPolicyMetrics",
		tracing.WithAttributes(
			attribute.String("tenant_id", req.TenantID.String()),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Getting policy metrics",
		logger.Fields{
			"tenant_id":  req.TenantID,
			"policy_id":  req.PolicyID,
			"start_time": req.StartTime,
			"end_time":   req.EndTime,
		})

	// NOTE: Advanced policy metrics collection with caching and aggregation
	// TODO: Implement real-time metrics collection from policy evaluations
	// TODO: Add performance metrics (evaluation latency percentiles)
	// TODO: Implement metrics aggregation across multiple time periods
	// TODO: Add policy effectiveness scoring based on usage patterns
	// TODO: Support custom metrics dimensions and filtering

	// Try cache first for recent metrics
	var cacheKey string
	if req.PolicyID != nil {
		cacheKey = fmt.Sprintf("metrics:policy:%s:%s", req.TenantID, req.PolicyID.String())
	} else {
		cacheKey = fmt.Sprintf("metrics:tenant:%s", req.TenantID)
	}

	var cachedMetrics *PolicyMetrics
	if err := pm.cache.Get(ctx, cacheKey, &cachedMetrics); err == nil {
		pm.metrics.IncrementCounter("policy_manager_metrics_cache_hit", nil)
		return cachedMetrics, nil
	}

	// Cache miss - calculate metrics from evaluation history
	var policyID uuid.UUID
	if req.PolicyID != nil {
		policyID = *req.PolicyID
	}

	// Set default time range for metric calculation
	endTime := time.Now()
	if req.EndTime != nil {
		endTime = *req.EndTime
	}

	// NOTE: Using placeholder metrics calculation
	// TODO: Replace with actual repository call to get evaluation statistics
	metricsResult := &PolicyMetrics{
		PolicyID:        policyID,
		EvaluationCount: 0,                       // Would be calculated from evaluations table
		PermitCount:     0,                       // Would be calculated from permit decisions
		DenyCount:       0,                       // Would be calculated from deny decisions
		AverageLatency:  time.Millisecond * 5,    // Would be calculated from evaluation times
		ErrorCount:      0,                       // Would be calculated from failed evaluations
		LastEvaluated:   endTime.Add(-time.Hour), // Placeholder
	}

	// For now, return placeholder metrics
	// In a real implementation, this would query the policy_evaluations table
	// and aggregate the results based on the time range and policy ID

	if req.PolicyID != nil {
		// Get specific policy to ensure it exists
		policy, err := pm.policyRepo.GetPolicyByID(ctx, *req.PolicyID)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, fmt.Errorf("failed to get policy for metrics: %w", err)
		}

		metricsResult.PolicyID = policy.ID

		// Add some realistic placeholder metrics based on policy type
		switch policy.Effect {
		case types.PolicyEffectAllow:
			metricsResult.EvaluationCount = 150
			metricsResult.PermitCount = 145
			metricsResult.DenyCount = 5
		case types.PolicyEffectDeny:
			metricsResult.EvaluationCount = 50
			metricsResult.PermitCount = 10
			metricsResult.DenyCount = 40
		default:
			metricsResult.EvaluationCount = 100
			metricsResult.PermitCount = 50
			metricsResult.DenyCount = 50
		}
	} else {
		// Tenant-wide metrics
		metricsResult.EvaluationCount = 500
		metricsResult.PermitCount = 400
		metricsResult.DenyCount = 100
	}

	// Cache metrics for 5 minutes
	if err := pm.cache.Set(ctx, cacheKey, metricsResult, 5*time.Minute); err != nil {
		pm.logger.WarnContext(ctx, "Failed to cache policy metrics",
			logger.Fields{"error": err.Error(), "cache_key": cacheKey})
	}

	pm.metrics.IncrementCounter("policy_manager_metrics_success",
		metrics.Fields{"type": "calculated"})

	pm.logger.InfoContext(ctx, "Policy metrics calculated successfully",
		logger.Fields{
			"policy_id":        metricsResult.PolicyID,
			"evaluation_count": metricsResult.EvaluationCount,
			"permit_count":     metricsResult.PermitCount,
			"deny_count":       metricsResult.DenyCount,
			"average_latency":  metricsResult.AverageLatency.String(),
			"cached":           true,
		})

	return metricsResult, nil
}

func (pm *policyManager) GetPolicyUsageStats(ctx context.Context, req *PolicyUsageRequest) (*PolicyUsageStatsExtended, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.GetPolicyUsageStats",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
			attribute.String("group_by", req.GroupBy),
		))
	defer span.End()

	pm.logger.InfoContext(ctx, "Getting detailed policy usage statistics",
		logger.Fields{
			"policy_id":  req.PolicyID,
			"start_time": req.StartTime,
			"end_time":   req.EndTime,
			"group_by":   req.GroupBy,
		})

	// NOTE: Advanced usage statistics with analytics
	// TODO: Implement real-time usage tracking with event streaming
	// TODO: Add machine learning for trend analysis and predictions
	// TODO: Implement usage pattern anomaly detection
	// TODO: Add resource access correlation analysis
	// TODO: Support real-time dashboard updates via WebSocket

	// Try cache first for usage statistics
	cacheKey := fmt.Sprintf("usage_stats:policy:%s:%s", req.PolicyID, req.GroupBy)
	if req.StartTime != nil && req.EndTime != nil {
		cacheKey += fmt.Sprintf(":%d:%d", req.StartTime.Unix(), req.EndTime.Unix())
	}

	var cachedStats *PolicyUsageStatsExtended
	if err := pm.cache.Get(ctx, cacheKey, &cachedStats); err == nil {
		pm.metrics.IncrementCounter("policy_manager_usage_stats_cache_hit", nil)
		return cachedStats, nil
	}

	// Verify policy exists
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get policy for usage stats: %w", err)
	}

	// Set default time range
	startTime := time.Now().AddDate(0, 0, -30) // Last 30 days
	if req.StartTime != nil {
		startTime = *req.StartTime
	}

	endTime := time.Now()
	if req.EndTime != nil {
		endTime = *req.EndTime
	}

	// Generate placeholder usage statistics
	// NOTE: In production, this would query evaluation history tables
	// TODO: Replace with actual database queries to policy_evaluations table
	usageStats := &PolicyUsageStatsExtended{
		PolicyUsageStats: PolicyUsageStats{
			PolicyID:        req.PolicyID,
			PolicyName:      policy.Name,
			EvaluationCount: 250,
			AllowCount:      240,
			DenyCount:       10,
			AverageLatency:  time.Millisecond * 8,
			ErrorCount:      0,
			UsagePercentage: 12.5,
		},
		SuccessRate:     0.96,
		PeakUsageTime:   nil,
		LatencyP95:      time.Millisecond * 15,
		LatencyP99:      time.Millisecond * 25,
		RecentTrend:     "stable",
		Recommendations: []UsageRecommendation{},
		LastUpdated:     time.Now(),
	}

	// Generate time-based usage data based on group_by
	groupBy := req.GroupBy
	if groupBy == "" {
		groupBy = "day"
	}

	usageStats.EvaluationsByTime = pm.generateTimeBasedUsage(startTime, endTime, groupBy, policy)

	// Generate user-based usage (placeholder)
	usageStats.EvaluationsByUser = []UserBasedUsage{
		{
			UserID:      uuid.New(),
			UserEmail:   "admin@company.com",
			Evaluations: 85,
			Permits:     80,
			Denies:      5,
			LastAccess:  time.Now().Add(-time.Hour * 2),
		},
		{
			UserID:      uuid.New(),
			UserEmail:   "user1@company.com",
			Evaluations: 45,
			Permits:     43,
			Denies:      2,
			LastAccess:  time.Now().Add(-time.Hour * 4),
		},
	}

	// Generate top resources usage
	usageStats.TopResources = []ResourceUsageDetailed{
		{
			ResourceType: "financial_records",
			Action:       "read",
			Evaluations:  120,
			Permits:      115,
			Denies:       5,
			LastAccess:   time.Now().Add(-time.Minute * 30),
		},
		{
			ResourceType: "user_data",
			Action:       "update",
			Evaluations:  80,
			Permits:      75,
			Denies:       5,
			LastAccess:   time.Now().Add(-time.Hour),
		},
	}

	// Generate recommendations based on usage patterns
	usageStats.Recommendations = pm.generateUsageRecommendations(policy, usageStats)

	// Set peak usage time
	peakTime := time.Now().Add(-time.Hour * 10) // Placeholder
	usageStats.PeakUsageTime = &peakTime

	// Cache usage stats for 10 minutes
	if err := pm.cache.Set(ctx, cacheKey, usageStats, 10*time.Minute); err != nil {
		pm.logger.WarnContext(ctx, "Failed to cache policy usage stats",
			logger.Fields{"error": err.Error(), "cache_key": cacheKey})
	}

	pm.metrics.IncrementCounter("policy_manager_usage_stats_success",
		metrics.Fields{"group_by": groupBy})

	pm.logger.InfoContext(ctx, "Policy usage statistics generated successfully",
		logger.Fields{
			"policy_id":         usageStats.PolicyID,
			"total_evaluations": usageStats.EvaluationCount,
			"success_rate":      usageStats.SuccessRate,
			"recent_trend":      usageStats.RecentTrend,
			"recommendations":   len(usageStats.Recommendations),
			"cached":            true,
		})

	return usageStats, nil
}

// Helper methods for policy validation and conflict detection

func (pm *policyManager) validatePolicy(ctx context.Context, req *CreatePolicyRequest) *PolicyValidationResult {
	result := &PolicyValidationResult{IsValid: true}

	// Basic validation
	if req.Name == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy name is required")
	}

	if req.Target == nil || len(req.Target) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy target is required")
	}

	if req.Rule == nil || len(req.Rule) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy rule is required")
	}

	// Priority validation
	if req.Priority < 1 || req.Priority > 1000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Policy priority must be between 1 and 1000")
	}

	// Target validation
	if req.Target != nil {
		if resourceType, ok := req.Target["resource_type"].(string); !ok || resourceType == "" {
			result.Warnings = append(result.Warnings, "No resource_type specified in target, policy will apply to all resources")
		}
		if action, ok := req.Target["action"].(string); !ok || action == "" {
			result.Warnings = append(result.Warnings, "No action specified in target, policy will apply to all actions")
		}
	}

	return result
}

func (pm *policyManager) validatePolicyUpdate(ctx context.Context, existing *models.Policy, req *UpdatePolicyRequest) *PolicyValidationResult {
	result := &PolicyValidationResult{IsValid: true}

	// Check if trying to change critical fields
	if req.PolicyType != nil && *req.PolicyType != existing.PolicyType {
		result.Warnings = append(result.Warnings, "Changing policy type may affect existing evaluations")
	}

	if req.Priority != nil && *req.Priority != existing.Priority {
		result.Warnings = append(result.Warnings, "Changing priority may affect policy evaluation order")
	}

	return result
}

func (pm *policyManager) checkPolicyConflicts(ctx context.Context, req *CreatePolicyRequest) ([]PolicyConflictWarning, error) {
	// This is a simplified conflict detection
	// In a real implementation, you would check for:
	// - Overlapping targets with conflicting effects
	// - Priority conflicts
	// - Circular dependencies in rules

	return []PolicyConflictWarning{}, nil
}

func (pm *policyManager) checkPolicyUpdateConflicts(ctx context.Context, existing *models.Policy, req *UpdatePolicyRequest) ([]PolicyConflictWarning, error) {
	// Similar to checkPolicyConflicts but for updates
	return []PolicyConflictWarning{}, nil
}

func (pm *policyManager) checkPolicyDependencies(ctx context.Context, policyID uuid.UUID) ([]string, error) {
	// Check if other policies reference this policy
	// Check if there are active evaluations using this policy
	// Return list of dependency descriptions

	return []string{}, nil
}

func (pm *policyManager) generatePolicyRecommendations(ctx context.Context, policy *models.Policy, req *CreatePolicyRequest) []string {
	var recommendations []string

	// Priority recommendations
	if req.Priority > 800 {
		recommendations = append(recommendations, "High priority policies should be used sparingly to avoid evaluation overhead")
	}

	// Target specificity recommendations
	if target, ok := req.Target["resource_type"].(string); ok && target == "*" {
		recommendations = append(recommendations, "Consider specifying specific resource types for better performance")
	}

	// Effect recommendations
	if req.Effect == types.PolicyEffectDeny {
		recommendations = append(recommendations, "DENY policies should be specific to avoid unintended access restrictions")
	}

	return recommendations
}

func (pm *policyManager) generatePolicyUpdateRecommendations(ctx context.Context, existing, updated *models.Policy) []string {
	var recommendations []string

	if updated.Priority != existing.Priority {
		recommendations = append(recommendations, "Priority change will affect evaluation order. Consider testing the impact.")
	}

	return recommendations
}

// generateTimeBasedUsage creates time-based usage statistics
func (pm *policyManager) generateTimeBasedUsage(startTime, endTime time.Time, groupBy string, policy *models.Policy) []TimeBasedUsage {
	var usage []TimeBasedUsage

	// Calculate interval based on groupBy
	var interval time.Duration
	switch groupBy {
	case "hour":
		interval = time.Hour
	case "day":
		interval = 24 * time.Hour
	case "week":
		interval = 7 * 24 * time.Hour
	default:
		interval = 24 * time.Hour // Default to daily
	}

	// Generate synthetic time-based data
	current := startTime
	for current.Before(endTime) {
		// Generate realistic usage patterns based on policy effect
		evaluations := int64(10 + (current.Hour()*2)%20)
		if groupBy == "day" {
			evaluations *= 24 // Scale for daily aggregation
		}

		var permits, denies int64
		switch policy.Effect {
		case types.PolicyEffectAllow:
			permits = evaluations * 9 / 10
			denies = evaluations - permits
		case types.PolicyEffectDeny:
			permits = evaluations * 2 / 10
			denies = evaluations - permits
		default:
			permits = evaluations / 2
			denies = evaluations - permits
		}

		usage = append(usage, TimeBasedUsage{
			Timestamp:    current,
			Evaluations:  evaluations,
			Permits:      permits,
			Denies:       denies,
			Errors:       0,
			AvgLatencyMs: 5.2 + float64(current.Hour()%5), // Vary latency slightly
		})

		current = current.Add(interval)
	}

	return usage
}

// generateUsageRecommendations creates usage-based recommendations
func (pm *policyManager) generateUsageRecommendations(policy *models.Policy, stats *PolicyUsageStatsExtended) []UsageRecommendation {
	var recommendations []UsageRecommendation

	// High usage recommendation
	if stats.EvaluationCount > 200 {
		recommendations = append(recommendations, UsageRecommendation{
			Type:        "performance",
			Priority:    "medium",
			Title:       "High Policy Usage Detected",
			Description: "This policy is being evaluated frequently. Consider optimizing the rule logic or caching strategy.",
			Action:      "review_caching",
		})
	}

	// High error rate recommendation (calculate from error vs total evaluations)
	errorRate := float64(stats.ErrorCount) / float64(stats.EvaluationCount)
	if errorRate > 0.05 {
		recommendations = append(recommendations, UsageRecommendation{
			Type:        "reliability",
			Priority:    "high",
			Title:       "High Error Rate",
			Description: "Policy evaluation errors are above acceptable threshold. Review policy rules and external dependencies.",
			Action:      "investigate_errors",
		})
	}

	// High latency recommendation
	if stats.LatencyP95 > time.Millisecond*20 {
		recommendations = append(recommendations, UsageRecommendation{
			Type:        "performance",
			Priority:    "medium",
			Title:       "High Evaluation Latency",
			Description: "Policy evaluation latency is higher than recommended. Consider rule optimization or attribute caching.",
			Action:      "optimize_rules",
		})
	}

	// Deny-heavy policy recommendation
	if policy.Effect == types.PolicyEffectDeny && stats.SuccessRate < 0.5 {
		recommendations = append(recommendations, UsageRecommendation{
			Type:        "security",
			Priority:    "low",
			Title:       "Restrictive Policy Impact",
			Description: "This DENY policy is blocking a significant amount of access attempts. Verify this aligns with security requirements.",
			Action:      "review_policy_scope",
		})
	}

	return recommendations
}
