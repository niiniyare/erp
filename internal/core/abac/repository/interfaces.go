package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeDefinitionRepository defines the interface for attribute definition persistence
type AttributeDefinitionRepository interface {
	// Core CRUD operations
	CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error)
	GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*models.AttributeDefinition, error)
	GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error)
	UpdateAttributeDefinition(ctx context.Context, id uuid.UUID, req *UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error)
	DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error

	// List and search operations
	ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error)
	GetAttributeDefinitionsByCategory(ctx context.Context, category types.AttributeCategory) ([]*models.AttributeDefinition, error)
	GetRequiredAttributeDefinitions(ctx context.Context) ([]*models.AttributeDefinition, error)

	// Bulk operations
	GetAttributeDefinitionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.AttributeDefinition, error)
}

// PolicyRepository defines the interface for policy persistence
type PolicyRepository interface {
	// Core CRUD operations
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*models.Policy, error)
	GetPolicyByID(ctx context.Context, id uuid.UUID) (*models.Policy, error)
	GetPolicyByName(ctx context.Context, name string) (*models.Policy, error)
	UpdatePolicy(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*models.Policy, error)
	DeletePolicy(ctx context.Context, id uuid.UUID) error

	// List and search operations
	ListPolicies(ctx context.Context, req *ListPoliciesRequest) ([]*models.Policy, error)
	GetPoliciesByCategory(ctx context.Context, category types.PolicyCategory) ([]*models.Policy, error)
	GetPoliciesByType(ctx context.Context, policyType types.PolicyType) ([]*models.Policy, error)
	GetActivePolicies(ctx context.Context) ([]*models.Policy, error)

	// Evaluation operations
	GetPoliciesForEvaluation(ctx context.Context, req *GetPoliciesForEvaluationRequest) ([]*models.Policy, error)
	GetApplicablePolicies(ctx context.Context, resourceType, action string) ([]*models.Policy, error)

	// Bulk operations
	GetPoliciesByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Policy, error)

	// Version management
	GetPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*models.Policy, error)
	GetLatestPolicyVersion(ctx context.Context, policyName string) (*models.Policy, error)

	// Conflict analysis
	GetConflictingPolicies(ctx context.Context, policy *models.Policy) ([]*models.Policy, error)
}

// PolicyEvaluationRepository defines the interface for policy evaluation caching
type PolicyEvaluationRepository interface {
	// Cache operations
	CacheEvaluationResult(ctx context.Context, req *CacheEvaluationResultRequest) error
	GetCachedEvaluationResult(ctx context.Context, req *GetCachedEvaluationResultRequest) (*models.PolicyEvaluationResult, error)
	InvalidateEvaluationCache(ctx context.Context, req *InvalidateEvaluationCacheRequest) error

	// Bulk cache operations
	GetCachedEvaluationResults(ctx context.Context, requests []*GetCachedEvaluationResultRequest) ([]*models.PolicyEvaluationResult, error)
	InvalidateEvaluationCacheByPolicyID(ctx context.Context, policyID uuid.UUID) error
	InvalidateEvaluationCacheByUserID(ctx context.Context, userID uuid.UUID) error

	// Cache maintenance
	CleanupExpiredEvaluations(ctx context.Context) error
	GetEvaluationCacheStats(ctx context.Context) (*EvaluationCacheStats, error)
}

// AttributeRepository defines the interface for attribute value persistence
type AttributeRepository interface {
	// Attribute value operations
	StoreAttributeValue(ctx context.Context, req *StoreAttributeValueRequest) (*models.AttributeValue, error)
	GetAttributeValue(ctx context.Context, definitionID uuid.UUID, entityID uuid.UUID) (*models.AttributeValue, error)
	GetAttributeValuesByEntity(ctx context.Context, entityID uuid.UUID, category types.AttributeCategory) ([]*models.AttributeValue, error)
	UpdateAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID, value any) (*models.AttributeValue, error)
	DeleteAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID) error

	// Bulk operations
	StoreAttributeValues(ctx context.Context, values []*StoreAttributeValueRequest) ([]*models.AttributeValue, error)
	GetAttributeValuesByDefinitions(ctx context.Context, definitionIDs []uuid.UUID, entityID uuid.UUID) ([]*models.AttributeValue, error)

	// Context operations
	BuildAttributeContext(ctx context.Context, req *BuildAttributeContextRequest) (*models.AttributeContext, error)
	GetUserAttributeContext(ctx context.Context, userID uuid.UUID) (*models.AttributeContext, error)
	GetResourceAttributeContext(ctx context.Context, resourceType string, resourceID uuid.UUID) (*models.AttributeContext, error)

	// Maintenance
	CleanupExpiredAttributes(ctx context.Context) error
	GetAttributeStats(ctx context.Context) (*AttributeStats, error)
}

// ─── REQUEST/RESPONSE MODELS ─────────────────────────────────────────────

// CreateAttributeDefinitionRequest represents attribute definition creation request
type CreateAttributeDefinitionRequest struct {
	Name            string                  `json:"name" validate:"required,min=1,max=100"`
	DisplayName     *string                 `json:"display_name,omitempty"`
	Description     *string                 `json:"description,omitempty"`
	DataType        types.AttributeDataType `json:"data_type" validate:"required"`
	Category        types.AttributeCategory `json:"category" validate:"required"`
	IsRequired      bool                    `json:"is_required"`
	IsSensitive     bool                    `json:"is_sensitive"`
	DefaultValue    *string                 `json:"default_value,omitempty"`
	AllowedValues   []string                `json:"allowed_values,omitempty"`
	ValidationRules map[string]any          `json:"validation_rules,omitempty"`
}

// UpdateAttributeDefinitionRequest represents attribute definition update request
type UpdateAttributeDefinitionRequest struct {
	DisplayName     *string        `json:"display_name,omitempty"`
	Description     *string        `json:"description,omitempty"`
	IsRequired      *bool          `json:"is_required,omitempty"`
	IsSensitive     *bool          `json:"is_sensitive,omitempty"`
	DefaultValue    *string        `json:"default_value,omitempty"`
	AllowedValues   []string       `json:"allowed_values,omitempty"`
	ValidationRules map[string]any `json:"validation_rules,omitempty"`
	IsActive        *bool          `json:"is_active,omitempty"`
}

// ListAttributeDefinitionsRequest represents attribute definition list request
type ListAttributeDefinitionsRequest struct {
	Category    *types.AttributeCategory `json:"category,omitempty"`
	DataType    *types.AttributeDataType `json:"data_type,omitempty"`
	IsRequired  *bool                    `json:"is_required,omitempty"`
	IsSensitive *bool                    `json:"is_sensitive,omitempty"`
	IsActive    *bool                    `json:"is_active,omitempty"`
	Limit       int                      `json:"limit"`
	Offset      int                      `json:"offset"`
}

// CreatePolicyRequest represents policy creation request
type CreatePolicyRequest struct {
	Name               string                         `json:"name" validate:"required,min=1,max=100"`
	DisplayName        *string                        `json:"display_name,omitempty"`
	Description        *string                        `json:"description,omitempty"`
	PolicyType         types.PolicyType               `json:"policy_type" validate:"required"`
	Effect             types.PolicyEffect             `json:"effect" validate:"required"`
	Priority           int32                          `json:"priority"`
	Category           types.PolicyCategory           `json:"category" validate:"required"`
	Target             map[string]any                 `json:"target" validate:"required"`
	Rule               map[string]any                 `json:"rule" validate:"required"`
	Obligations        map[string]any                 `json:"obligations,omitempty"`
	Advice             map[string]any                 `json:"advice,omitempty"`
	CombiningAlgorithm types.PolicyCombiningAlgorithm `json:"combining_algorithm"`
	ExpiresAt          *time.Time                     `json:"expires_at,omitempty"`
	CreatedBy          uuid.UUID                      `json:"created_by" validate:"required"`
}

// UpdatePolicyRequest represents policy update request
type UpdatePolicyRequest struct {
	DisplayName        *string                         `json:"display_name,omitempty"`
	Description        *string                         `json:"description,omitempty"`
	Priority           *int32                          `json:"priority,omitempty"`
	Target             map[string]any                  `json:"target,omitempty"`
	Rule               map[string]any                  `json:"rule,omitempty"`
	Obligations        map[string]any                  `json:"obligations,omitempty"`
	Advice             map[string]any                  `json:"advice,omitempty"`
	CombiningAlgorithm *types.PolicyCombiningAlgorithm `json:"combining_algorithm,omitempty"`
	IsActive           *bool                           `json:"is_active,omitempty"`
	ExpiresAt          *time.Time                      `json:"expires_at,omitempty"`
}

// ListPoliciesRequest represents policy list request
type ListPoliciesRequest struct {
	PolicyType *types.PolicyType     `json:"policy_type,omitempty"`
	Effect     *types.PolicyEffect   `json:"effect,omitempty"`
	Category   *types.PolicyCategory `json:"category,omitempty"`
	IsActive   *bool                 `json:"is_active,omitempty"`
	CreatedBy  *uuid.UUID            `json:"created_by,omitempty"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
}

// GetPoliciesForEvaluationRequest represents request for policies applicable to evaluation
type GetPoliciesForEvaluationRequest struct {
	ResourceType string         `json:"resource_type" validate:"required"`
	Action       string         `json:"action" validate:"required"`
	UserType     *string        `json:"user_type,omitempty"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// CacheEvaluationResultRequest represents cache evaluation result request
type CacheEvaluationResultRequest struct {
	UserID             uuid.UUID                      `json:"user_id" validate:"required"`
	ResourceType       string                         `json:"resource_type" validate:"required"`
	ResourceID         *uuid.UUID                     `json:"resource_id,omitempty"`
	Action             string                         `json:"action" validate:"required"`
	ContextHash        string                         `json:"context_hash" validate:"required"`
	Decision           types.PolicyDecisionType       `json:"decision" validate:"required"`
	ApplicablePolicies []uuid.UUID                    `json:"applicable_policies"`
	EvaluationTimeMS   int64                          `json:"evaluation_time_ms"`
	ExpiresAt          time.Time                      `json:"expires_at"`
	Result             *models.PolicyEvaluationResult `json:"result"`
}

// GetCachedEvaluationResultRequest represents get cached evaluation result request
type GetCachedEvaluationResultRequest struct {
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ResourceType string     `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action" validate:"required"`
	ContextHash  string     `json:"context_hash" validate:"required"`
}

// InvalidateEvaluationCacheRequest represents invalidate evaluation cache request
type InvalidateEvaluationCacheRequest struct {
	UserID        *uuid.UUID  `json:"user_id,omitempty"`
	ResourceType  *string     `json:"resource_type,omitempty"`
	ResourceID    *uuid.UUID  `json:"resource_id,omitempty"`
	Action        *string     `json:"action,omitempty"`
	PolicyIDs     []uuid.UUID `json:"policy_ids,omitempty"`
	InvalidateAll bool        `json:"invalidate_all"`
}

// StoreAttributeValueRequest represents store attribute value request
type StoreAttributeValueRequest struct {
	DefinitionID uuid.UUID             `json:"definition_id" validate:"required"`
	EntityID     uuid.UUID             `json:"entity_id" validate:"required"`
	EntityType   string                `json:"entity_type" validate:"required"`
	Value        any                   `json:"value" validate:"required"`
	Source       types.AttributeSource `json:"source" validate:"required"`
	Confidence   float64               `json:"confidence"`
	ExpiresAt    *time.Time            `json:"expires_at,omitempty"`
}

// BuildAttributeContextRequest represents build attribute context request
type BuildAttributeContextRequest struct {
	UserID          uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType    string         `json:"resource_type" validate:"required"`
	ResourceID      *uuid.UUID     `json:"resource_id,omitempty"`
	Action          string         `json:"action" validate:"required"`
	EntityID        *uuid.UUID     `json:"entity_id,omitempty"`
	SessionData     map[string]any `json:"session_data,omitempty"`
	EnvironmentData map[string]any `json:"environment_data,omitempty"`
}

// ─── STATISTICS MODELS ─────────────────────────────────────────────

// EvaluationCacheStats represents evaluation cache statistics
type EvaluationCacheStats struct {
	TotalCachedEvaluations int64            `json:"total_cached_evaluations"`
	CacheHitRate           float64          `json:"cache_hit_rate"`
	CacheMissRate          float64          `json:"cache_miss_rate"`
	ExpiredEvaluations     int64            `json:"expired_evaluations"`
	AverageEvaluationTime  time.Duration    `json:"average_evaluation_time"`
	EvaluationsByDecision  map[string]int64 `json:"evaluations_by_decision"`
	EvaluationsByResource  map[string]int64 `json:"evaluations_by_resource"`
}

// AttributeStats represents attribute statistics
type AttributeStats struct {
	TotalAttributes      int64            `json:"total_attributes"`
	AttributesByCategory map[string]int64 `json:"attributes_by_category"`
	AttributesByDataType map[string]int64 `json:"attributes_by_data_type"`
	AttributesBySource   map[string]int64 `json:"attributes_by_source"`
	ExpiredAttributes    int64            `json:"expired_attributes"`
	SensitiveAttributes  int64            `json:"sensitive_attributes"`
	RequiredAttributes   int64            `json:"required_attributes"`
}
