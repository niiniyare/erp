package abac

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/niiniyare/erp/db/sqlc"
)

// PolicyRepository defines the interface for interacting with ABAC policies.
type PolicyRepository interface {
	CreatePolicy(ctx context.Context, arg sqlc.CreatePolicyParams) (sqlc.Policy, error)
	GetPolicy(ctx context.Context, id uuid.UUID) (sqlc.Policy, error)
	GetPolicyByName(ctx context.Context, name string) (sqlc.Policy, error)
	UpdatePolicy(ctx context.Context, arg sqlc.UpdatePolicyParams) (sqlc.Policy, error)
	SoftDeletePolicy(ctx context.Context, id uuid.UUID) error
	HardDeletePolicy(ctx context.Context, id uuid.UUID) error
	ListPolicies(ctx context.Context) ([]sqlc.Policy, error)
	ListActivePolicies(ctx context.Context) ([]sqlc.Policy, error)
	ListPoliciesByCategory(ctx context.Context, category string) ([]sqlc.Policy, error)
	ListPoliciesByEffect(ctx context.Context, effect string) ([]sqlc.Policy, error)
	SearchPolicies(ctx context.Context, arg sqlc.SearchPoliciesParams) ([]sqlc.Policy, error)
	GetApplicablePolicies(ctx context.Context, resourceName string, actionName string) ([]sqlc.Policy, error)
}

// AttributeDefinitionRepository defines the interface for interacting with ABAC attribute definitions.
type AttributeDefinitionRepository interface {
	CreateAttributeDefinition(ctx context.Context, arg sqlc.CreateAttributeDefinitionParams) (sqlc.AttributeDefinition, error)
	GetAttributeDefinition(ctx context.Context, id uuid.UUID) (sqlc.AttributeDefinition, error)
	GetAttributeDefinitionByName(ctx context.Context, name string) (sqlc.AttributeDefinition, error)
	UpdateAttributeDefinition(ctx context.Context, arg sqlc.UpdateAttributeDefinitionParams) (sqlc.AttributeDefinition, error)
	DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error
	ListAttributeDefinitions(ctx context.Context) ([]sqlc.AttributeDefinition, error)
	ListAttributeDefinitionsByCategory(ctx context.Context, category string) ([]sqlc.AttributeDefinition, error)
}

// PolicyEvaluationRepository defines the interface for interacting with ABAC policy evaluation cache.
type PolicyEvaluationRepository interface {
	CreatePolicyEvaluation(ctx context.Context, arg sqlc.CreatePolicyEvaluationParams) (sqlc.PolicyEvaluation, error)
	GetPolicyEvaluation(ctx context.Context, id uuid.UUID) (sqlc.PolicyEvaluation, error)
	GetPolicyEvaluationByContextHash(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID, actionID uuid.UUID, contextHash string) (sqlc.PolicyEvaluation, error)
	DeletePolicyEvaluation(ctx context.Context, id uuid.UUID) error
	DeleteExpiredPolicyEvaluations(ctx context.Context) error
	ListPolicyEvaluationsForUser(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]sqlc.PolicyEvaluation, error)
}
