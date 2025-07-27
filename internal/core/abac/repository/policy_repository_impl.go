package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// policyRepository implements PolicyRepository using SQLC
type policyRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewPolicyRepository creates a new policy repository implementation
func NewPolicyRepository(
	db sqlc.DBTX,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) PolicyRepository {
	return &policyRepository{
		db:      db,
		queries: sqlc.New(db),
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CreatePolicy creates a new policy
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

	params := sqlc.CreatePolicyParams{
		EntityID:    req.EntityID,
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
		IsActive:    req.IsActive,
		CreatedBy:   req.CreatedBy,
	}

	policy, err := r.queries.CreatePolicy(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "create_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_CREATE_FAILED", "Failed to create policy").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_create")

	result := r.convertSQLCPolicyToModel(policy)

	r.logger.InfoContext(ctx, "Policy created successfully",
		logger.Fields{
			"policy_id": result.ID,
			"name":      result.Name,
		})

	return result, nil
}

// GetPolicyByID retrieves a policy by ID
func (r *policyRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyByID",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_id", id.String()),
		))
	defer span.End()

	policy, err := r.queries.GetPolicy(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_repository_not_found")
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_GET_FAILED", "Failed to get policy").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_get")
	return r.convertSQLCPolicyToModel(policy), nil
}

// GetPolicyByName retrieves a policy by name
func (r *policyRepository) GetPolicyByName(ctx context.Context, name string) (*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPolicyByName",
		tracing.WithAttributes(
			tracing.StringAttribute("policy_name", name),
		))
	defer span.End()

	policy, err := r.queries.GetPolicyByName(ctx, name)
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

	policy, err := r.queries.UpdatePolicy(ctx, params)
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

	err := r.queries.SoftDeletePolicy(ctx, id)
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
		count, err := r.queries.CountPolicies(ctx, sqlc.CountPoliciesParams{
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
		policies, err = r.queries.ListPolicies(ctx)
	} else {
		policies, err = r.queries.ListPolicies(ctx)
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

// GetPoliciesForEvaluation gets policies applicable for evaluation
func (r *policyRepository) GetPoliciesForEvaluation(ctx context.Context, req *GetPoliciesForEvaluationRequest) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesForEvaluation",
		tracing.WithAttributes(
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
		))
	defer span.End()

	policies, err := r.queries.GetPoliciesForEvaluation(ctx, sqlc.GetPoliciesForEvaluationParams{
		Column1: req.EntityID,
		Column2: req.ResourceType,
		Column3: req.Action,
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_for_evaluation_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_GET_FAILED", "Failed to get policies for evaluation").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_for_evaluation")

	result := make([]*models.Policy, len(policies))
	for i, policy := range policies {
		result[i] = r.convertSQLCPolicyToModel(policy)
	}

	return result, nil
}

// GetPoliciesByIDs gets policies by their IDs
func (r *policyRepository) GetPoliciesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Policy, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetPoliciesByIDs",
		tracing.WithAttributes(
			tracing.IntAttribute("policy_count", len(ids)),
		))
	defer span.End()

	policies, err := r.queries.GetPoliciesByIDs(ctx, ids)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_repository", "get_by_ids_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICIES_GET_FAILED", "Failed to get policies by IDs").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_repository_get_by_ids")

	result := make([]*models.Policy, len(policies))
	for i, policy := range policies {
		result[i] = r.convertSQLCPolicyToModel(policy)
	}

	return result, nil
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
