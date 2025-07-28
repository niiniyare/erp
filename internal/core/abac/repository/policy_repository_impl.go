package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/niiniyare/erp/db/sqlc"
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
	store   sqlc.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewPolicyRepository creates a new policy repository implementation
func NewPolicyRepository(
	store sqlc.Store,
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
			tracing.StringAttribute("policy_name", req.Name),
			tracing.StringAttribute("policy_type", string(req.PolicyType)),
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

	// Use tenant-aware transaction for policy creation
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		params := sqlc.CreatePolicyParams{
			Name:        req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			PolicyType:  string(req.PolicyType),
			Effect:      string(req.Effect),
			Priority:    req.Priority,
			Category:    string(req.Category),
			Target:      req.Target,
			Rule:        req.Rule,
			Obligations: req.Obligations,
			Advice:      req.Advice,
			CreatedBy:   req.CreatedBy,
		}

		// Use transaction store queries (tenant context is already set)
		sqlcPolicy, err := txStore.CreatePolicy(ctx, params)
		if err != nil {
			return err
		}

		// Convert SQLC model to domain model
		policy = r.convertSQLCPolicyToModel(sqlcPolicy)
		return nil
	})

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "create_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_CREATE_FAILED", "Failed to create policy").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_create")

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
			tracing.StringAttribute("policy_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context for cache key prefixing
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy retrieval")
	}

	// Try cache first (Redis will automatically prefix with tenant_id)
	cacheKey := fmt.Sprintf("policy:%s", id.String())
	var policy *models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policy); err == nil {
		r.metrics.IncrementCounter("policy_repository_cache_hit")
		return policy, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicy, err := r.store.GetPolicyByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy").WithErr(err)
	}

	// Convert to domain model
	policy = r.convertSQLCPolicyToModel(sqlcPolicy)

	// Cache the result (15 minute TTL)
	if err := r.cache.Set(ctx, cacheKey, policy, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache policy", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("policy_repository_get")
	return policy, nil
}

// GetPolicyByName retrieves a policy by name
func (r *policyRepository) GetPolicyByName(ctx context.Context, name string) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyByName",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_name", name),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("policy_name:%s", name)
	var policy *models.Policy
	if err := r.cache.Get(ctx, cacheKey, &policy); err == nil {
		r.metrics.IncrementCounter("policy_repository_name_cache_hit")
		return policy, nil
	}

	// Cache miss - get from database
	sqlcPolicy, err := r.store.GetPolicyByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_by_name_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy by name").WithErr(err)
	}

	// Convert to domain model
	policy = r.convertSQLCPolicyToModel(sqlcPolicy)

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, policy, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache policy by name", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_by_name")
	return policy, nil
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_by_name_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy by name").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_by_name")
	return r.convertSQLCPolicyToModel(policy), nil
}

// UpdatePolicy updates an existing policy
func (r *policyRepository) UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.UpdatePolicy",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_id", req.ID.String()),
		))
	defer span.End()

	params := sqlc.UpdatePolicyParams{
		ID:          req.ID,
		EntityID:    req.EntityID,
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
		IsActive:    req.IsActive,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	var policy sqlc.Policy
	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		var err error
		policy, err = txStore.UpdatePolicy(ctx, params)
		return err
	})

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "update_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_UPDATE_FAILED", "Failed to update policy").WithErr(err)
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

	r.metrics.IncrementSuccessCount("policy_repository_update")
	return r.convertSQLCPolicyToModel(policy), nil
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "update_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_UPDATE_FAILED", "Failed to update policy").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_update")
	return r.convertSQLCPolicyToModel(policy), nil
}

// DeletePolicy soft deletes a policy
func (r *policyRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.DeletePolicy",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_id", id.String()),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		return txStore.SoftDeletePolicy(ctx, id)
	})

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "delete_failed")
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_DELETE_FAILED", "Failed to delete policy").WithErr(err)
	}

	// Invalidate cache entries for this policy
	cacheKey := fmt.Sprintf("policy:%s", id.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		r.logger.WarnContext(ctx, "Failed to invalidate deleted policy cache", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("policy_repository_delete")

	r.logger.InfoContext(ctx, "Policy deleted successfully",
		logger.Fields{"policy_id": id})

	return nil
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "delete_failed")
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_DELETE_FAILED", "Failed to delete policy").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_delete")

	r.logger.InfoContext(ctx, "Policy deleted successfully",
		logger.Fields{"policy_id": id})

	return nil
}

// ListPolicies lists policies with filtering and pagination
func (r *policyRepository) ListPolicies(ctx context.Context, req *ListPoliciesRequest) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.ListPolicies")
	defer span.End()

	var policies []sqlc.Policy
	var err error

	if req.Search != "" || req.Category != "" || req.IsActive != nil {
		// Use filtered search
		// Extract tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
		}

		count, err := r.store.CountPolicies(ctx, sqlc.CountPoliciesParams{
			Column1: &req.Search,
			Column2: &req.Category,
			Column3: req.IsActive,
		})
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_COUNT_FAILED", "Failed to count policies").WithErr(err)
		}

		// Set the total count for pagination
		req.TotalCount = int(count)

		// Get filtered policies (this would need a custom query)
		// For now, fallback to basic list
		policies, err = r.store.ListPolicies(ctx)
	} else {
		// Extract tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
		}

		policies, err = r.store.ListPolicies(ctx)
	}

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "list_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_LIST_FAILED", "Failed to list policies").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_list")

	result := make([]*models.Policy, len(policies))
	for i, policy := range policies {
		result[i] = r.convertSQLCPolicyToModel(policy)
	}

	return result, nil
}

// GetPoliciesForEvaluation gets policies applicable for evaluation with caching
func (r *policyRepository) GetPoliciesForEvaluation(ctx context.Context, req *GetPoliciesForEvaluationRequest) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesForEvaluation",
		tracing.WithAttributes(
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy evaluation")
	}

	// Create cache key for evaluation policies (Redis automatically prefixes with tenant_id)
	cacheKey := fmt.Sprintf("policies:eval:%s:%s", req.ResourceType, req.Action)
	var policies []*models.Policy

	// Try cache first
	if err := r.cache.Get(ctx, cacheKey, &policies); err == nil {
		r.metrics.IncrementCounter("policy_repository_eval_cache_hit")
		return policies, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicies, err := r.store.GetPoliciesForEvaluation(ctx, sqlc.GetPoliciesForEvaluationParams{
		Column1: req.EntityID,
		Column2: req.ResourceType,
		Column3: req.Action,
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_for_evaluation_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_GET_FAILED", "Failed to get policies for evaluation").WithErr(err)
	}

	// Convert to domain models
	policies = make([]*models.Policy, len(sqlcPolicies))
	for i, policy := range sqlcPolicies {
		policies[i] = r.convertSQLCPolicyToModel(policy)
	}

	// Cache the result (10 minute TTL for evaluation policies)
	if err := r.cache.Set(ctx, cacheKey, policies, 10*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to cache evaluation policies", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_for_evaluation")
	return policies, nil
}

// GetPoliciesByIDs gets policies by their IDs using bulk cache operations
func (r *policyRepository) GetPoliciesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesByIDs",
		tracing.WithAttributes(
			tracing.IntAttribute("policy_count", len(ids)),
		))
	defer span.End()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context is required for policy retrieval")
	}

	// Build cache keys for bulk operations
	cacheKeys := make([]string, len(ids))
	for i, id := range ids {
		cacheKeys[i] = fmt.Sprintf("policy:%s", id.String())
	}

	// Try bulk cache get first
	var cachedPolicies []*models.Policy
	if err := r.cache.MGet(ctx, cacheKeys, &cachedPolicies); err == nil && len(cachedPolicies) == len(ids) {
		r.metrics.IncrementCounter("policy_repository_bulk_cache_hit")
		return cachedPolicies, nil
	}

	// Cache miss - get from database (RLS automatically filters by tenant)
	sqlcPolicies, err := r.store.GetPoliciesByIDs(ctx, ids)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_by_ids_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICIES_GET_FAILED", "Failed to get policies by IDs").WithErr(err)
	}

	// Convert to domain models
	policies := make([]*models.Policy, len(sqlcPolicies))
	cacheData := make(map[string]interface{})

	for i, sqlcPolicy := range sqlcPolicies {
		policy := r.convertSQLCPolicyToModel(sqlcPolicy)
		policies[i] = policy

		// Prepare for bulk cache set
		cacheKey := fmt.Sprintf("policy:%s", policy.ID.String())
		cacheData[cacheKey] = policy
	}

	// Bulk cache the results (15 minute TTL)
	if err := r.cache.MSet(ctx, cacheData, 15*time.Minute); err != nil {
		r.logger.WarnContext(ctx, "Failed to bulk cache policies", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_by_ids")
	return policies, nil
}

// Helper method to convert SQLC Policy to domain model
func (r *policyRepository) convertSQLCPolicyToModel(policy sqlc.Policy) *models.Policy {
	var entityID *uuid.UUID
	if policy.EntityID.Valid {
		entityID = &policy.EntityID.UUID
	}

	var displayName *string
	if policy.DisplayName.Valid {
		displayName = &policy.DisplayName.String
	}

	var description *string
	if policy.Description.Valid {
		description = &policy.Description.String
	}

	var createdBy *uuid.UUID
	if policy.CreatedBy.Valid {
		createdBy = &policy.CreatedBy.UUID
	}

	return &models.Policy{
		ID:          policy.ID,
		TenantID:    policy.TenantID,
		EntityID:    entityID,
		Name:        policy.Name,
		DisplayName: displayName,
		Description: description,
		PolicyType:  types.PolicyType(policy.PolicyType),
		Effect:      types.PolicyEffect(policy.Effect),
		Priority:    policy.Priority,
		Category:    types.PolicyCategory(policy.Category),
		Target:      policy.Target,
		Rule:        policy.Rule,
		Obligations: policy.Obligations,
		Advice:      policy.Advice,
		IsActive:    policy.IsActive,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt.Time,
		CreatedBy:   createdBy,
	}
}
