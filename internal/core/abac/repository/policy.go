package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// policyRepository implements PolicyRepository using SQLC and Store
type policyRepository struct {
	store   db.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewPolicyRepository creates a new policy repository implementation
func NewPolicyRepository(
	store db.Store,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) PolicyRepository {
	return &policyRepository{
		store:   store,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreatePolicy creates a new policy using tenant-aware transaction
func (r *policyRepository) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CreatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_name", req.Name),
			attribute.String("policy_type", string(req.PolicyType)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Creating policy",
		logger.Fields{
			"name":        req.Name,
			"policy_type": req.PolicyType,
			"effect":      req.Effect,
		})

	// Extract tenant ID from context (injected by middleware)
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy creation")
	}

	var policy *models.Policy

	// Serialize JSON fields
	targetBytes, err := json.Marshal(req.Target)
	if err != nil {
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_TARGET_MARSHAL_FAILED", "Failed to marshal policy target").WithDetail("error", err.Error())
	}

	ruleBytes, err := json.Marshal(req.Rule)
	if err != nil {
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_RULE_MARSHAL_FAILED", "Failed to marshal policy rule").WithDetail("error", err.Error())
	}

	var obligationsBytes []byte
	if req.Obligations != nil {
		obligationsBytes, err = json.Marshal(req.Obligations)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_OBLIGATIONS_MARSHAL_FAILED", "Failed to marshal policy obligations").WithDetail("error", err.Error())
		}
	}

	var adviceBytes []byte
	if req.Advice != nil {
		adviceBytes, err = json.Marshal(req.Advice)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_ADVICE_MARSHAL_FAILED", "Failed to marshal policy advice").WithDetail("error", err.Error())
		}
	}

	// Use tenant-aware transaction for policy creation
	err = r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		policyTypeStr := string(req.PolicyType)
		effectStr := string(req.Effect)
		categoryStr := string(req.Category)
		var description string
		if req.Description != nil {
			description = *req.Description
		}
		params := db.CreatePolicyParams{
			Name:        req.Name,
			DisplayName: req.DisplayName,
			Description: description,
			PolicyType:  &policyTypeStr,
			Effect:      &effectStr,
			Priority:    &req.Priority,
			Category:    &categoryStr,
			Target:      targetBytes,
			Rule:        ruleBytes,
			Obligations: obligationsBytes,
			Advice:      adviceBytes,
			CreatedBy:   &req.CreatedBy,
		}

		// Use transaction store queries (tenant context is already set)
		sqlcPolicy, err := txStore.CreatePolicy(ctx, params)
		if err != nil {
			return err
		}

		// Convert SQLC model to domain model
		policy = r.convertSQLCPolicyToModel(*sqlcPolicy)
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_create_failed", metrics.Fields{"operation": "create"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_CREATE_FAILED", "Failed to create policy").WithDetail("error", err.Error())
	}

	r.metrics.IncrementCounter("policy_repository_create_success", metrics.Fields{"operation": "create"})

	r.logger.InfoContext(ctx, "Policy created successfully",
		logger.Fields{
			"policy_id": policy.ID,
			"name":      policy.Name,
		})

	return policy, nil
}

// GetPolicyByID retrieves a policy by ID with cache integration
func (r *policyRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyByID",
		tracing.WithAttributes(
			attribute.String("policy_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context for cache key prefixing
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy retrieval")
	}

	// Try cache first with tenant-specific key
	cacheKey := fmt.Sprintf("policy:%s:%s", tenantID.String(), id.String())
	var policy *models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policy); err == nil {
		r.metrics.IncrementCounter("policy_repository_cache_hit", metrics.Fields{"cache": "hit"})
		return policy, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicy, err := r.store.GetPolicy(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found", metrics.Fields{"result": "not_found"})
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_get_failed", metrics.Fields{"operation": "get"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy").WithDetail("error", err.Error())
	}

	// Convert to domain model
	policy = r.convertSQLCPolicyToModel(*sqlcPolicy)

	// Cache the result (15 minute TTL)
	if err := r.cache.Set(ctx, cacheKey, policy, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache policy", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_get_success", metrics.Fields{"operation": "get"})
	return policy, nil
}

// GetPolicyByName retrieves a policy by name
func (r *policyRepository) GetPolicyByName(ctx context.Context, name string) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyByName",
		tracing.WithAttributes(
			attribute.String("policy_name", name),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first with tenant-specific key
	cacheKey := fmt.Sprintf("policy_name:%s:%s", tenantID.String(), name)
	var policy *models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policy); err == nil {
		r.metrics.IncrementCounter("policy_repository_name_cache_hit", metrics.Fields{"cache": "hit"})
		return policy, nil
	}

	// Cache miss - get from database
	sqlcPolicy, err := r.store.GetPolicyByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found", metrics.Fields{"result": "not_found"})
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_get_by_name_failed", metrics.Fields{"operation": "get_by_name"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy by name").WithDetail("error", err.Error())
	}

	// Convert to domain model
	policy = r.convertSQLCPolicyToModel(*sqlcPolicy)

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, policy, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache policy by name", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_get_by_name_success", metrics.Fields{"operation": "get_by_name"})
	return policy, nil
}

// UpdatePolicy updates an existing policy
func (r *policyRepository) UpdatePolicy(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.UpdatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", id.String()),
		))
	defer span.End()

	// Serialize JSON fields if they exist
	var targetBytes, ruleBytes, obligationsBytes, adviceBytes []byte
	var err error

	if req.Target != nil {
		targetBytes, err = json.Marshal(req.Target)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_TARGET_MARSHAL_FAILED", "Failed to marshal policy target").WithDetail("error", err.Error())
		}
	}

	if req.Rule != nil {
		ruleBytes, err = json.Marshal(req.Rule)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_RULE_MARSHAL_FAILED", "Failed to marshal policy rule").WithDetail("error", err.Error())
		}
	}

	if req.Obligations != nil {
		obligationsBytes, err = json.Marshal(req.Obligations)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_OBLIGATIONS_MARSHAL_FAILED", "Failed to marshal policy obligations").WithDetail("error", err.Error())
		}
	}

	if req.Advice != nil {
		adviceBytes, err = json.Marshal(req.Advice)
		if err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_ADVICE_MARSHAL_FAILED", "Failed to marshal policy advice").WithDetail("error", err.Error())
		}
	}

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	params := db.UpdatePolicyParams{
		ID:          id,
		DisplayName: req.DisplayName,
		Description: description,
		Priority:    req.Priority,
		Target:      targetBytes,
		Rule:        ruleBytes,
		Obligations: obligationsBytes,
		Advice:      adviceBytes,
		IsActive:    req.IsActive,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	var policy *db.Policy
	// Use tenant-aware transaction
	err = r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		var err error
		policy, err = txStore.UpdatePolicy(ctx, params)
		return err
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_update_failed", metrics.Fields{"operation": "update"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_UPDATE_FAILED", "Failed to update policy").WithDetail("error", err.Error())
	}

	// Invalidate cache entries for this policy
	cacheKeys := []string{
		fmt.Sprintf("policy:%s", policy.ID.String()),
		fmt.Sprintf("policy_name:%s", policy.Name),
	}
	for _, key := range cacheKeys {
		if err := r.cache.Delete(ctx, key); err != nil {
			r.logger.WarnContext(ctx, "Failed to invalidate policy cache", logger.Fields{"key": key, "error": err.Error()})
		}
	}

	r.metrics.IncrementCounter("policy_repository_update_success", metrics.Fields{"operation": "update"})
	return r.convertSQLCPolicyToModel(*policy), nil
}

// DeletePolicy soft deletes a policy
func (r *policyRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.DeletePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		return txStore.SoftDeletePolicy(ctx, id)
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_delete_failed", metrics.Fields{"operation": "delete"})
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_DELETE_FAILED", "Failed to delete policy").WithDetail("error", err.Error())
	}

	// Invalidate cache entries for this policy
	cacheKey := fmt.Sprintf("policy:%s", id.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate deleted policy cache", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_delete_success", metrics.Fields{"operation": "delete"})

	r.logger.InfoContext(ctx, "Policy deleted successfully",
		logger.Fields{"policy_id": id})

	return nil
}

// ListPolicies lists policies with filtering and pagination
func (r *policyRepository) ListPolicies(ctx context.Context, req *ListPoliciesRequest) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.ListPolicies")
	defer span.End()

	var policies []*db.Policy
	var err error

	// Get policies (RLS automatically filters by tenant)
	policies, err = r.store.ListPolicies(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_list_failed", metrics.Fields{"operation": "list"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_LIST_FAILED", "Failed to list policies").WithDetail("error", err.Error())
	}

	r.metrics.IncrementCounter("policy_repository_list_success", metrics.Fields{"operation": "list"})

	result := make([]*models.Policy, len(policies))
	for i, policy := range policies {
		result[i] = r.convertSQLCPolicyToModel(*policy)
	}

	return result, nil
}

// GetPoliciesForEvaluation gets policies applicable for evaluation with caching
func (r *policyRepository) GetPoliciesForEvaluation(ctx context.Context, req *GetPoliciesForEvaluationRequest) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesForEvaluation",
		tracing.WithAttributes(
			attribute.String("resource_type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	// Create cache key for evaluation policies
	cacheKey := fmt.Sprintf("policies:eval:%s:%s", req.ResourceType, req.Action)
	var policies []*models.Policy

	// Try cache first
	if err := r.cache.Get(ctx, cacheKey, &policies); err == nil {
		r.metrics.IncrementCounter("policy_repository_eval_cache_hit", metrics.Fields{"cache": "hit"})
		return policies, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	var entityID uuid.UUID
	if req.EntityID != nil {
		entityID = *req.EntityID
	}
	sqlcPolicies, err := r.store.GetPoliciesForEvaluation(ctx, db.GetPoliciesForEvaluationParams{
		EntityID:     entityID,
		ResourceType: []byte(req.ResourceType),
		Action:       []byte(req.Action),
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_get_for_evaluation_failed", metrics.Fields{"operation": "get_for_evaluation"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_GET_FAILED", "Failed to get policies for evaluation").WithDetail("error", err.Error())
	}

	// Convert to domain models
	policies = make([]*models.Policy, len(sqlcPolicies))
	for i, policyRow := range sqlcPolicies {
		// Convert GetPoliciesForEvaluationRow to Policy
		policy := &db.Policy{
			ID:          policyRow.PolicyID,
			EntityID:    policyRow.EntityID,
			Name:        policyRow.Name,
			Effect:      policyRow.Effect,
			Priority:    policyRow.Priority,
			Category:    policyRow.Category,
			Target:      policyRow.Target,
			Rule:        policyRow.Rule,
			Obligations: policyRow.Obligations,
			IsActive:    policyRow.IsActive,
			CreatedAt:   policyRow.CreatedAt,
		}
		policies[i] = r.convertSQLCPolicyToModel(*policy)
	}

	// Cache the result (10 minute TTL for evaluation policies)
	if err := r.cache.Set(ctx, cacheKey, policies, 10*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache evaluation policies", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_get_for_evaluation_success", metrics.Fields{"operation": "get_for_evaluation"})
	return policies, nil
}

// GetPoliciesByIDs gets policies by their IDs using bulk cache operations
func (r *policyRepository) GetPoliciesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesByIDs",
		tracing.WithAttributes(
			attribute.Int("policy_count", len(ids)),
		))
	defer span.End()

	// For now, skip bulk cache operations and go directly to database
	// TODO: Implement bulk cache operations when cache service supports MGet/MSet

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicies, err := r.store.GetPoliciesByIDs(ctx, ids)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_get_by_ids_failed", metrics.Fields{"operation": "get_by_ids"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICIES_GET_FAILED", "Failed to get policies by IDs").WithDetail("error", err.Error())
	}

	// Convert to domain models
	policies := make([]*models.Policy, len(sqlcPolicies))
	for i, sqlcPolicy := range sqlcPolicies {
		policies[i] = r.convertSQLCPolicyToModel(*sqlcPolicy)
	}

	r.metrics.IncrementCounter("policy_repository_get_by_ids_success", metrics.Fields{"operation": "get_by_ids"})
	return policies, nil
}

// Helper method to convert SQLC Policy to domain model
func (r *policyRepository) convertSQLCPolicyToModel(policy db.Policy) *models.Policy {
	var displayName *string
	if policy.DisplayName != nil {
		displayName = policy.DisplayName
	}

	var description *string
	// Assuming description is a regular string field, not nullable
	// If it's nullable, we'd need to check if it's a pointer or sql.NullString

	var createdBy *uuid.UUID
	if policy.CreatedBy != nil {
		createdBy = policy.CreatedBy
	}

	// Convert JSON byte arrays to map[string]any
	var target, rule, obligations, advice map[string]any
	if len(policy.Target) > 0 {
		json.Unmarshal(policy.Target, &target)
	}
	if len(policy.Rule) > 0 {
		json.Unmarshal(policy.Rule, &rule)
	}
	if len(policy.Obligations) > 0 {
		json.Unmarshal(policy.Obligations, &obligations)
	}
	if len(policy.Advice) > 0 {
		json.Unmarshal(policy.Advice, &advice)
	}

	var createdByVal uuid.UUID
	if createdBy != nil {
		createdByVal = *createdBy
	}

	return &models.Policy{
		ID:          policy.ID,
		TenantID:    policy.TenantID,
		Name:        policy.Name,
		DisplayName: displayName,
		Description: description,
		PolicyType:  types.PolicyType(*policy.PolicyType),
		Effect:      types.PolicyEffect(*policy.Effect),
		Priority:    *policy.Priority,
		Category:    types.PolicyCategory(*policy.Category),
		Target:      target,
		Rule:        rule,
		Obligations: obligations,
		Advice:      advice,
		IsActive:    *policy.IsActive,
		CreatedAt:   policy.CreatedAt.Time,
		UpdatedAt:   policy.UpdatedAt.Time,
		CreatedBy:   createdByVal,
	}
}

// GetActivePolicies retrieves all active policies using cache with tenant isolation
func (r *policyRepository) GetActivePolicies(ctx context.Context) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetActivePolicies")
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving active policies")

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy retrieval")
	}

	// Try cache first with tenant-specific key
	cacheKey := fmt.Sprintf("policies:active:%s", tenantID.String())
	var policies []*models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policies); err == nil {
		r.metrics.IncrementCounter("policy_repository_active_cache_hit", metrics.Fields{"cache": "hit"})
		return policies, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicies, err := r.store.ListActivePolicies(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_get_active_failed", metrics.Fields{"operation": "get_active"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_ACTIVE_GET_FAILED", "Failed to get active policies").WithDetail("error", err.Error())
	}

	// Convert to domain models
	policies = make([]*models.Policy, len(sqlcPolicies))
	for i, policy := range sqlcPolicies {
		policies[i] = r.convertSQLCPolicyToModel(*policy)
	}

	// Cache the result (15 minute TTL for active policies)
	if err := r.cache.Set(ctx, cacheKey, policies, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache active policies", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_get_active_success", metrics.Fields{"operation": "get_active"})
	return policies, nil
}

// GetApplicablePolicies gets policies applicable to a specific resource type and action
func (r *policyRepository) GetApplicablePolicies(ctx context.Context, resourceType, action string) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetApplicablePolicies",
		tracing.WithAttributes(
			attribute.String("resource_type", resourceType),
			attribute.String("action", action),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving applicable policies",
		logger.Fields{
			"resource_type": resourceType,
			"action":        action,
		})

	// Try cache first with resource-specific key
	cacheKey := fmt.Sprintf("policies:applicable:%s:%s", resourceType, action)
	var policies []*models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policies); err == nil {
		r.metrics.IncrementCounter("policy_repository_applicable_cache_hit", metrics.Fields{"cache": "hit"})
		return policies, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicies, err := r.store.GetApplicablePolicies(ctx, db.GetApplicablePoliciesParams{
		Column1: resourceType,
		Column2: action,
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("policy_repository_applicable_failed", metrics.Fields{"operation": "get_applicable"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_APPLICABLE_GET_FAILED", "Failed to get applicable policies").WithDetail("error", err.Error())
	}

	// Convert to domain models
	policies = make([]*models.Policy, len(sqlcPolicies))
	for i, policy := range sqlcPolicies {
		policies[i] = r.convertSQLCPolicyToModel(*policy)
	}

	// Cache the result (10 minute TTL for applicable policies)
	if err := r.cache.Set(ctx, cacheKey, policies, 10*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache applicable policies", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("policy_repository_get_applicable_success", metrics.Fields{"operation": "get_applicable"})
	return policies, nil
}

// GetConflictingPolicies finds policies that might conflict with the given policy
func (r *policyRepository) GetConflictingPolicies(ctx context.Context, policy *models.Policy) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetConflictingPolicies",
		tracing.WithAttributes(
			attribute.String("policy_id", policy.ID.String()),
			attribute.String("policy_type", string(policy.PolicyType)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Finding conflicting policies",
		logger.Fields{
			"policy_id":   policy.ID.String(),
			"policy_type": policy.PolicyType,
		})

	// For now, return empty list as this would require complex logic
	// to determine policy conflicts based on targets, rules, etc.
	// This is a placeholder implementation
	return []*models.Policy{}, nil
}

// GetPoliciesByCategory retrieves policies by category
func (r *policyRepository) GetPoliciesByCategory(ctx context.Context, category types.PolicyCategory) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesByCategory",
		tracing.WithAttributes(
			attribute.String("category", string(category)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving policies by category",
		logger.Fields{"category": category})

	// Placeholder implementation - would need SQLC query
	return []*models.Policy{}, nil
}

// GetPoliciesByType retrieves policies by type
func (r *policyRepository) GetPoliciesByType(ctx context.Context, policyType types.PolicyType) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesByType",
		tracing.WithAttributes(
			attribute.String("policy_type", string(policyType)),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving policies by type",
		logger.Fields{"policy_type": policyType})

	// Placeholder implementation - would need SQLC query
	return []*models.Policy{}, nil
}

// GetPolicyVersions retrieves all versions of a policy
func (r *policyRepository) GetPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyVersions",
		tracing.WithAttributes(
			attribute.String("policy_id", policyID.String()),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving policy versions",
		logger.Fields{"policy_id": policyID.String()})

	// Placeholder implementation - would need SQLC query for versioning
	return []*models.Policy{}, nil
}

// GetLatestPolicyVersion retrieves the latest version of a policy by name
func (r *policyRepository) GetLatestPolicyVersion(ctx context.Context, policyName string) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetLatestPolicyVersion",
		tracing.WithAttributes(
			attribute.String("policy_name", policyName),
		))
	defer span.End()

	r.logger.InfoContext(ctx, "Retrieving latest policy version",
		logger.Fields{"policy_name": policyName})

	// For now, delegate to GetPolicyByName as a placeholder
	return r.GetPolicyByName(ctx, policyName)
}
