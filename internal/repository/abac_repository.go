package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/niiniyare/erp/db/sqlc"
	// "github.com/niiniyare/erp/internal/core/abac" // TODO: Fix import path
)

// SQLCPolicyRepository implements the PolicyRepository interface using sqlc.
type SQLCPolicyRepository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// NewSQLCPolicyRepository creates a new SQLCPolicyRepository.
func NewSQLCPolicyRepository(db *sql.DB) abac.PolicyRepository {
	return &SQLCPolicyRepository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// CreatePolicy creates a new ABAC policy.
func (r *SQLCPolicyRepository) CreatePolicy(ctx context.Context, arg sqlc.CreatePolicyParams) (sqlc.Policy, error) {
	return r.queries.CreatePolicy(ctx, arg)
}

// GetPolicy retrieves a policy by its ID.
func (r *SQLCPolicyRepository) GetPolicy(ctx context.Context, id uuid.UUID) (sqlc.Policy, error) {
	return r.queries.GetPolicy(ctx, id)
}

// GetPolicyByName retrieves a policy by its name.
func (r *SQLCPolicyRepository) GetPolicyByName(ctx context.Context, name string) (sqlc.Policy, error) {
	return r.queries.GetPolicyByName(ctx, name)
}

// UpdatePolicy updates an existing policy.
func (r *SQLCPolicyRepository) UpdatePolicy(ctx context.Context, arg sqlc.UpdatePolicyParams) (sqlc.Policy, error) {
	return r.queries.UpdatePolicy(ctx, arg)
}

// SoftDeletePolicy soft deletes a policy by its ID.
func (r *SQLCPolicyRepository) SoftDeletePolicy(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeletePolicy(ctx, id)
}

// HardDeletePolicy hard deletes a policy by its ID.
func (r *SQLCPolicyRepository) HardDeletePolicy(ctx context.Context, id uuid.UUID) error {
	return r.queries.HardDeletePolicy(ctx, id)
}

// ListPolicies lists all policies.
func (r *SQLCPolicyRepository) ListPolicies(ctx context.Context) ([]sqlc.Policy, error) {
	return r.queries.ListPolicies(ctx)
}

// ListActivePolicies lists all active policies.
func (r *SQLCPolicyRepository) ListActivePolicies(ctx context.Context) ([]sqlc.Policy, error) {
	return r.queries.ListActivePolicies(ctx)
}

// ListPoliciesByCategory lists policies by category.
func (r *SQLCPolicyRepository) ListPoliciesByCategory(ctx context.Context, category string) ([]sqlc.Policy, error) {
	return r.queries.ListPoliciesByCategory(ctx, category)
}

// ListPoliciesByEffect lists policies by effect.
func (r *SQLCPolicyRepository) ListPoliciesByEffect(ctx context.Context, effect string) ([]sqlc.Policy, error) {
	return r.queries.ListPoliciesByEffect(ctx, effect)
}

// SearchPolicies searches policies by name or description.
func (r *SQLCPolicyRepository) SearchPolicies(ctx context.Context, arg sqlc.SearchPoliciesParams) ([]sqlc.Policy, error) {
	return r.queries.SearchPolicies(ctx, arg)
}

// GetApplicablePolicies retrieves policies applicable to a given resource and action.
func (r *SQLCPolicyRepository) GetApplicablePolicies(ctx context.Context, resourceName string, actionName string) ([]sqlc.Policy, error) {
	return r.queries.GetApplicablePolicies(ctx, sqlc.GetApplicablePoliciesParams{
		ResourceName: pgtype.Text{String: resourceName, Valid: true},
		ActionName:   pgtype.Text{String: actionName, Valid: true},
	})
}

// SQLCAttributeDefinitionRepository implements the AttributeDefinitionRepository interface using sqlc.
type SQLCAttributeDefinitionRepository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// NewSQLCAttributeDefinitionRepository creates a new SQLCAttributeDefinitionRepository.
func NewSQLCAttributeDefinitionRepository(db *sql.DB) abac.AttributeDefinitionRepository {
	return &SQLCAttributeDefinitionRepository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// CreateAttributeDefinition creates a new attribute definition.
func (r *SQLCAttributeDefinitionRepository) CreateAttributeDefinition(ctx context.Context, arg sqlc.CreateAttributeDefinitionParams) (sqlc.AttributeDefinition, error) {
	return r.queries.CreateAttributeDefinition(ctx, arg)
}

// GetAttributeDefinition retrieves an attribute definition by its ID.
func (r *SQLCAttributeDefinitionRepository) GetAttributeDefinition(ctx context.Context, id uuid.UUID) (sqlc.AttributeDefinition, error) {
	return r.queries.GetAttributeDefinition(ctx, id)
}

// GetAttributeDefinitionByName retrieves an attribute definition by its name.
func (r *SQLCAttributeDefinitionRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (sqlc.AttributeDefinition, error) {
	return r.queries.GetAttributeDefinitionByName(ctx, name)
}

// UpdateAttributeDefinition updates an existing attribute definition.
func (r *SQLCAttributeDefinitionRepository) UpdateAttributeDefinition(ctx context.Context, arg sqlc.UpdateAttributeDefinitionParams) (sqlc.AttributeDefinition, error) {
	return r.queries.UpdateAttributeDefinition(ctx, arg)
}

// DeleteAttributeDefinition deletes an attribute definition by its ID.
func (r *SQLCAttributeDefinitionRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteAttributeDefinition(ctx, id)
}

// ListAttributeDefinitions lists all attribute definitions.
func (r *SQLCAttributeDefinitionRepository) ListAttributeDefinitions(ctx context.Context) ([]sqlc.AttributeDefinition, error) {
	return r.queries.ListAttributeDefinitions(ctx)
}

// ListAttributeDefinitionsByCategory lists attribute definitions by category.
func (r *SQLCAttributeDefinitionRepository) ListAttributeDefinitionsByCategory(ctx context.Context, category string) ([]sqlc.AttributeDefinition, error) {
	return r.queries.ListAttributeDefinitionsByCategory(ctx, category)
}

// SQLCPolicyEvaluationRepository implements the PolicyEvaluationRepository interface using sqlc.
type SQLCPolicyEvaluationRepository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// NewSQLCPolicyEvaluationRepository creates a new SQLCPolicyEvaluationRepository.
func NewSQLCPolicyEvaluationRepository(db *sql.DB) abac.PolicyEvaluationRepository {
	return &SQLCPolicyEvaluationRepository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// CreatePolicyEvaluation creates a new policy evaluation cache entry.
func (r *SQLCPolicyEvaluationRepository) CreatePolicyEvaluation(ctx context.Context, arg sqlc.CreatePolicyEvaluationParams) (sqlc.PolicyEvaluation, error) {
	return r.queries.CreatePolicyEvaluation(ctx, arg)
}

// GetPolicyEvaluation retrieves a policy evaluation by its ID.
func (r *SQLCPolicyEvaluationRepository) GetPolicyEvaluation(ctx context.Context, id uuid.UUID) (sqlc.PolicyEvaluation, error) {
	return r.queries.GetPolicyEvaluation(ctx, id)
}

// GetPolicyEvaluationByContextHash retrieves a policy evaluation by context hash.
func (r *SQLCPolicyEvaluationRepository) GetPolicyEvaluationByContextHash(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID, actionID uuid.UUID, contextHash string) (sqlc.PolicyEvaluation, error) {
	return r.queries.GetPolicyEvaluationByContextHash(ctx, sqlc.GetPolicyEvaluationByContextHashParams{
		UserID:      userID,
		ResourceID:  resourceID,
		ActionID:    actionID,
		ContextHash: contextHash,
	})
}

// DeletePolicyEvaluation deletes a policy evaluation by its ID.
func (r *SQLCPolicyEvaluationRepository) DeletePolicyEvaluation(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeletePolicyEvaluation(ctx, id)
}

// DeleteExpiredPolicyEvaluations deletes all expired policy evaluation cache entries.
func (r *SQLCPolicyEvaluationRepository) DeleteExpiredPolicyEvaluations(ctx context.Context) error {
	return r.queries.DeleteExpiredPolicyEvaluations(ctx)
}

// ListPolicyEvaluationsForUser lists policy evaluations for a specific user.
func (r *SQLCPolicyEvaluationRepository) ListPolicyEvaluationsForUser(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]sqlc.PolicyEvaluation, error) {
	return r.queries.ListPolicyEvaluationsForUser(ctx, sqlc.ListPolicyEvaluationsForUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}
