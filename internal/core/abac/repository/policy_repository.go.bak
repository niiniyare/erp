package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// policyRepository implements PolicyRepository interface
type policyRepository struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.TracingService
	metrics metrics.Provider
	logger  logger.Logger
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(
	store db.Store,
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.Provider,
	logger logger.Logger,
) PolicyRepository {
	return &policyRepository{
		store:   store,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
		logger:  logger,
	}
}

// CreatePolicy creates a new ABAC policy
func (r *policyRepository) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*models.Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.CreatePolicy",
		tracing.WithAttributes(
			attribute.String("policy.name", req.Name),
			attribute.String("policy.type", string(req.PolicyType)),
			attribute.String("policy.effect", string(req.Effect)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Creating ABAC policy",
		logger.Fields{
			"policy_name": req.Name,
			"policy_type": req.PolicyType,
			"effect":      req.Effect,
			"priority":    req.Priority,
			"created_by":  req.CreatedBy,
		})

	// Validate request
	if err := req.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Policy validation failed")
		r.recordPolicyMetrics(ctx, "create", "validation_error", 0)
		return nil, err
	}

	// Convert to SQLC params
	params, err := r.toPolicyCreateParams(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert policy params")
		r.recordPolicyMetrics(ctx, "create", "conversion_error", 0)
		return nil, errors.NewBusinessError("POLICY_CONVERSION_FAILED", "Failed to convert policy parameters").
			WithDetail("error", err.Error())
	}

	startTime := time.Now()

	// Create policy using SQLC (tenant_id handled by current_tenant_id())
	sqlcPolicy, err := r.store.CreatePolicy(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create policy in database")
		r.recordPolicyMetrics(ctx, "create", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_CREATE_FAILED", "Failed to create policy").
			WithDetail("error", err.Error())
	}

	// Convert back to domain model
	policy, err := r.fromSQLCPolicy(sqlcPolicy)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert policy from database")
		return nil, err
	}

	// Invalidate relevant caches
	if err := r.invalidatePolicyCaches(ctx, policy); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate policy caches",
			logger.Fields{"policy_id": policy.ID, "error": err.Error()})
	}

	// Record metrics
	r.recordPolicyMetrics(ctx, "create", "success", time.Since(startTime))

	// Log successful creation
	policy.LogEvaluation(ctx, r.logger, types.PolicyDecisionAllow, time.Since(startTime))

	r.logger.InfoContext(ctx, "ABAC policy created successfully",
		logger.Fields{
			"policy_id":   policy.ID,
			"policy_name": policy.Name,
			"tenant_id":   policy.TenantID,
		})

	return policy, nil
}

// GetPolicyByID retrieves a policy by ID with caching
func (r *policyRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.GetPolicyByID",
		tracing.WithAttributes(attribute.String("policy.id", id.String())))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("policy:%s", id)
	var cachedPolicy models.Policy
	if err := r.cache.Get(ctx, cacheKey, &cachedPolicy); err == nil {
		r.recordPolicyMetrics(ctx, "get", "cache_hit", 0)
		return &cachedPolicy, nil
	}

	startTime := time.Now()

	// Get from database (tenant filtering handled by SQLC query using current_tenant_id())
	sqlcPolicy, err := r.store.GetPolicyByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordPolicyMetrics(ctx, "get", "not_found", time.Since(startTime))
			return nil, errors.ErrPolicyNotFound
		}
		span.SetStatus(codes.Error, "Failed to get policy from database")
		r.recordPolicyMetrics(ctx, "get", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_GET_FAILED", "Failed to retrieve policy").
			WithDetail("policy_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Convert to domain model
	policy, err := r.fromSQLCPolicy(sqlcPolicy)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Cache the result
	cacheTTL := 15 * time.Minute
	if err := r.cache.Set(ctx, cacheKey, policy, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache policy",
			logger.Fields{"policy_id": id, "error": err.Error()})
	}

	r.recordPolicyMetrics(ctx, "get", "success", time.Since(startTime))
	return policy, nil
}

// GetPolicyByName retrieves a policy by name
func (r *policyRepository) GetPolicyByName(ctx context.Context, name string) (*models.Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.GetPolicyByName",
		tracing.WithAttributes(attribute.String("policy.name", name)))
	defer span.End()

	startTime := time.Now()

	// Get from database (tenant filtering handled automatically)
	sqlcPolicy, err := r.store.GetPolicyByName(ctx, name)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordPolicyMetrics(ctx, "get_by_name", "not_found", time.Since(startTime))
			return nil, errors.ErrPolicyNotFound
		}
		span.SetStatus(codes.Error, "Failed to get policy by name")
		r.recordPolicyMetrics(ctx, "get_by_name", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_GET_FAILED", "Failed to retrieve policy by name").
			WithDetail("policy_name", name).
			WithDetail("error", err.Error())
	}

	policy, err := r.fromSQLCPolicy(sqlcPolicy)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	r.recordPolicyMetrics(ctx, "get_by_name", "success", time.Since(startTime))
	return policy, nil
}

// GetPoliciesForEvaluation retrieves policies applicable for evaluation
func (r *policyRepository) GetPoliciesForEvaluation(ctx context.Context, req *GetPoliciesForEvaluationRequest) ([]*models.Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.GetPoliciesForEvaluation",
		tracing.WithAttributes(
			attribute.String("resource_type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("policies_for_eval:%s:%s", req.ResourceType, req.Action)
	var cachedPolicies []*models.Policy
	if err := r.cache.Get(ctx, cacheKey, &cachedPolicies); err == nil {
		r.recordPolicyMetrics(ctx, "get_for_evaluation", "cache_hit", 0)
		return cachedPolicies, nil
	}

	startTime := time.Now()

	// Get from database - this query would need to be added to SQLC
	// For now, use a placeholder approach
	sqlcPolicies, err := r.store.GetPoliciesForEvaluation(ctx, db.GetPoliciesForEvaluationParams{
		ResourceType: req.ResourceType,
		Action:       req.Action,
		// tenant_id filtering handled by current_tenant_id() in the SQLC query
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get policies for evaluation")
		r.recordPolicyMetrics(ctx, "get_for_evaluation", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_EVALUATION_GET_FAILED", "Failed to get policies for evaluation").
			WithDetail("resource_type", req.ResourceType).
			WithDetail("action", req.Action).
			WithDetail("error", err.Error())
	}

	// Convert to domain models
	policies := make([]*models.Policy, len(sqlcPolicies))
	for i, sqlcPolicy := range sqlcPolicies {
		policy, err := r.fromSQLCPolicy(&sqlcPolicy)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		policies[i] = policy
	}

	// Cache the results
	cacheTTL := 5 * time.Minute // Shorter TTL for evaluation policies
	if err := r.cache.Set(ctx, cacheKey, policies, cacheTTL); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache evaluation policies",
			logger.Fields{"cache_key": cacheKey, "error": err.Error()})
	}

	r.recordPolicyMetrics(ctx, "get_for_evaluation", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "Retrieved policies for evaluation",
		logger.Fields{
			"resource_type":      req.ResourceType,
			"action":             req.Action,
			"policy_count":       len(policies),
			"evaluation_time_ms": time.Since(startTime).Milliseconds(),
		})

	return policies, nil
}

// UpdatePolicy updates an existing policy
func (r *policyRepository) UpdatePolicy(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*models.Policy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.UpdatePolicy",
		tracing.WithAttributes(attribute.String("policy.id", id.String())))
	defer span.End()

	r.logger.InfoContext(ctx, "Updating ABAC policy",
		logger.Fields{
			"policy_id": id,
		})

	startTime := time.Now()

	// Convert to SQLC params
	params, err := r.toPolicyUpdateParams(id, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert policy update params")
		r.recordPolicyMetrics(ctx, "update", "conversion_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_UPDATE_CONVERSION_FAILED", "Failed to convert policy update parameters").
			WithDetail("error", err.Error())
	}

	// Update policy
	sqlcPolicy, err := r.store.UpdatePolicy(ctx, *params)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordPolicyMetrics(ctx, "update", "not_found", time.Since(startTime))
			return nil, errors.ErrPolicyNotFound
		}
		span.SetStatus(codes.Error, "Failed to update policy in database")
		r.recordPolicyMetrics(ctx, "update", "db_error", time.Since(startTime))
		return nil, errors.NewBusinessError("POLICY_UPDATE_FAILED", "Failed to update policy").
			WithDetail("policy_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Convert back to domain model
	policy, err := r.fromSQLCPolicy(sqlcPolicy)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Invalidate caches
	if err := r.invalidatePolicyCaches(ctx, policy); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate policy caches",
			logger.Fields{"policy_id": policy.ID, "error": err.Error()})
	}

	r.recordPolicyMetrics(ctx, "update", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "ABAC policy updated successfully",
		logger.Fields{
			"policy_id":   policy.ID,
			"policy_name": policy.Name,
		})

	return policy, nil
}

// DeletePolicy soft-deletes a policy
func (r *policyRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "policyRepository.DeletePolicy",
		tracing.WithAttributes(attribute.String("policy.id", id.String())))
	defer span.End()

	r.logger.InfoContext(ctx, "Deleting ABAC policy", logger.Fields{"policy_id": id})

	startTime := time.Now()

	// Soft delete policy
	err := r.store.SoftDeletePolicy(ctx, id)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			r.recordPolicyMetrics(ctx, "delete", "not_found", time.Since(startTime))
			return errors.ErrPolicyNotFound
		}
		span.SetStatus(codes.Error, "Failed to delete policy")
		r.recordPolicyMetrics(ctx, "delete", "db_error", time.Since(startTime))
		return errors.NewBusinessError("POLICY_DELETE_FAILED", "Failed to delete policy").
			WithDetail("policy_id", id.String()).
			WithDetail("error", err.Error())
	}

	// Invalidate all related caches
	cachePatterns := []string{
		fmt.Sprintf("policy:%s", id),
		"policies_for_eval:*",
		"policies:*",
	}

	for _, pattern := range cachePatterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			r.logger.WarnContext(ctx, "Failed to invalidate cache pattern",
				logger.Fields{"pattern": pattern, "error": err.Error()})
		}
	}

	r.recordPolicyMetrics(ctx, "delete", "success", time.Since(startTime))

	r.logger.InfoContext(ctx, "ABAC policy deleted successfully", logger.Fields{"policy_id": id})
	return nil
}

// Implement placeholder methods for other interface methods
func (r *policyRepository) ListPolicies(ctx context.Context, req *ListPoliciesRequest) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetPoliciesByCategory(ctx context.Context, category types.PolicyCategory) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetPoliciesByType(ctx context.Context, policyType types.PolicyType) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetActivePolicies(ctx context.Context) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetApplicablePolicies(ctx context.Context, resourceType, action string) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetPoliciesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

func (r *policyRepository) GetLatestPolicyVersion(ctx context.Context, policyName string) (*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return nil, errors.ErrPolicyNotFound
}

func (r *policyRepository) GetConflictingPolicies(ctx context.Context, policy *models.Policy) ([]*models.Policy, error) {
	// TODO: Implement with proper SQLC query
	return []*models.Policy{}, nil
}

// ─── HELPER METHODS ─────────────────────────────────────────────

// Validate validates CreatePolicyRequest
func (req *CreatePolicyRequest) Validate() error {
	if req.Name == "" {
		return errors.NewBusinessError("POLICY_NAME_REQUIRED", "Policy name is required")
	}
	if !req.PolicyType.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_TYPE", "Invalid policy type")
	}
	if !req.Effect.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_EFFECT", "Invalid policy effect")
	}
	if !req.Category.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_CATEGORY", "Invalid policy category")
	}
	if len(req.Target) == 0 {
		return errors.NewBusinessError("POLICY_TARGET_REQUIRED", "Policy target is required")
	}
	if len(req.Rule) == 0 {
		return errors.NewBusinessError("POLICY_RULE_REQUIRED", "Policy rule is required")
	}
	return nil
}

// toPolicyCreateParams converts CreatePolicyRequest to SQLC params
func (r *policyRepository) toPolicyCreateParams(req *CreatePolicyRequest) (*db.CreatePolicyParams, error) {
	targetJSON, err := json.Marshal(req.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target: %w", err)
	}

	ruleJSON, err := json.Marshal(req.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rule: %w", err)
	}

	var obligationsJSON, adviceJSON []byte
	if req.Obligations != nil {
		obligationsJSON, err = json.Marshal(req.Obligations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal obligations: %w", err)
		}
	}
	if req.Advice != nil {
		adviceJSON, err = json.Marshal(req.Advice)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal advice: %w", err)
		}
	}

	var expiresAt sql.NullTime
	if req.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *req.ExpiresAt, Valid: true}
	}

	return &db.CreatePolicyParams{
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		PolicyType:         string(req.PolicyType),
		Effect:             string(req.Effect),
		Priority:           req.Priority,
		Category:           string(req.Category),
		Target:             targetJSON,
		Rule:               ruleJSON,
		Obligations:        obligationsJSON,
		Advice:             adviceJSON,
		CombiningAlgorithm: string(req.CombiningAlgorithm),
		ExpiresAt:          expiresAt,
		CreatedBy:          req.CreatedBy,
		// tenant_id is set automatically by current_tenant_id() in SQLC query
	}, nil
}

// toPolicyUpdateParams converts UpdatePolicyRequest to SQLC params
func (r *policyRepository) toPolicyUpdateParams(id uuid.UUID, req *UpdatePolicyRequest) (*db.UpdatePolicyParams, error) {
	params := &db.UpdatePolicyParams{
		ID:          id,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
	}

	if req.Target != nil {
		targetJSON, err := json.Marshal(req.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target: %w", err)
		}
		params.Target = targetJSON
	}

	if req.Rule != nil {
		ruleJSON, err := json.Marshal(req.Rule)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rule: %w", err)
		}
		params.Rule = ruleJSON
	}

	if req.Obligations != nil {
		obligationsJSON, err := json.Marshal(req.Obligations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal obligations: %w", err)
		}
		params.Obligations = obligationsJSON
	}

	if req.Advice != nil {
		adviceJSON, err := json.Marshal(req.Advice)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal advice: %w", err)
		}
		params.Advice = adviceJSON
	}

	if req.CombiningAlgorithm != nil {
		algorithm := string(*req.CombiningAlgorithm)
		params.CombiningAlgorithm = &algorithm
	}

	if req.ExpiresAt != nil {
		params.ExpiresAt = sql.NullTime{Time: *req.ExpiresAt, Valid: true}
	}

	return params, nil
}

// fromSQLCPolicy converts SQLC policy to domain model
func (r *policyRepository) fromSQLCPolicy(sqlcPolicy *db.Policy) (*models.Policy, error) {
	var target, rule, obligations, advice map[string]any

	if err := json.Unmarshal(sqlcPolicy.Target, &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal target: %w", err)
	}

	if err := json.Unmarshal(sqlcPolicy.Rule, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
	}

	if len(sqlcPolicy.Obligations) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Obligations, &obligations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal obligations: %w", err)
		}
	}

	if len(sqlcPolicy.Advice) > 0 {
		if err := json.Unmarshal(sqlcPolicy.Advice, &advice); err != nil {
			return nil, fmt.Errorf("failed to unmarshal advice: %w", err)
		}
	}

	var expiresAt *time.Time
	if sqlcPolicy.ExpiresAt.Valid {
		expiresAt = &sqlcPolicy.ExpiresAt.Time
	}

	var approvedAt *time.Time
	if sqlcPolicy.ApprovedAt.Valid {
		approvedAt = &sqlcPolicy.ApprovedAt.Time
	}

	var deletedAt *time.Time
	if sqlcPolicy.DeletedAt.Valid {
		deletedAt = &sqlcPolicy.DeletedAt.Time
	}

	return &models.Policy{
		ID:                 sqlcPolicy.ID,
		TenantID:           sqlcPolicy.TenantID,
		Name:               sqlcPolicy.Name,
		DisplayName:        sqlcPolicy.DisplayName,
		Description:        sqlcPolicy.Description,
		PolicyType:         types.PolicyType(sqlcPolicy.PolicyType),
		Effect:             types.PolicyEffect(sqlcPolicy.Effect),
		Priority:           sqlcPolicy.Priority,
		Category:           types.PolicyCategory(sqlcPolicy.Category),
		Target:             target,
		Rule:               rule,
		Obligations:        obligations,
		Advice:             advice,
		CombiningAlgorithm: types.PolicyCombiningAlgorithm(sqlcPolicy.CombiningAlgorithm),
		IsActive:           sqlcPolicy.IsActive,
		Version:            sqlcPolicy.Version,
		PreviousVersionID:  sqlcPolicy.PreviousVersionID,
		ExpiresAt:          expiresAt,
		ApprovedBy:         sqlcPolicy.ApprovedBy,
		ApprovedAt:         approvedAt,
		CreatedBy:          sqlcPolicy.CreatedBy,
		CreatedAt:          sqlcPolicy.CreatedAt,
		UpdatedAt:          sqlcPolicy.UpdatedAt,
		DeletedAt:          deletedAt,
	}, nil
}

// invalidatePolicyCaches invalidates all caches related to a policy
func (r *policyRepository) invalidatePolicyCaches(ctx context.Context, policy *models.Policy) error {
	cachePatterns := []string{
		fmt.Sprintf("policy:%s", policy.ID),
		"policies_for_eval:*",
		fmt.Sprintf("policies:category:%s", policy.Category),
		fmt.Sprintf("policies:type:%s", policy.PolicyType),
	}

	for _, pattern := range cachePatterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to delete cache pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// recordPolicyMetrics records policy operation metrics
func (r *policyRepository) recordPolicyMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	counter := r.metrics.Counter(
		"abac_policy_repository_operations_total",
		"Total number of policy repository operations",
		"operation", "status",
	)

	counter.Inc(ctx, metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	if duration > 0 {
		histogram := r.metrics.Histogram(
			"abac_policy_repository_operation_duration_seconds",
			"Duration of policy repository operations",
			metrics.StandardHTTPDurationBuckets(),
			"operation", "status",
		)

		histogram.Observe(ctx, duration.Seconds(), metrics.Fields{
			"operation": operation,
			"status":    status,
		})
	}
}
