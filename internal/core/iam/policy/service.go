package policy

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam/model"
)

// Service defines the policy service interface
// Handles policy evaluation, caching, and policy lifecycle management
type Service interface {
	// Policy Management
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*model.Policy, error)
	GetPolicy(ctx context.Context, policyID uuid.UUID) (*model.Policy, error)
	UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*model.Policy, error)
	DeletePolicy(ctx context.Context, policyID uuid.UUID) error
	ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*ListPoliciesResult, error)

	// Policy Evaluation
	EvaluatePolicy(ctx context.Context, req *PolicyEvaluationRequest) (*PolicyEvaluationResult, error)
	BulkEvaluatePolicies(ctx context.Context, req *BulkPolicyEvaluationRequest) (*BulkPolicyEvaluationResult, error)
	TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error)

	// Policy Templates
	CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*model.PolicyTemplate, error)
	GetPolicyTemplate(ctx context.Context, templateID uuid.UUID) (*model.PolicyTemplate, error)
	ListPolicyTemplates(ctx context.Context) ([]*model.PolicyTemplate, error)
	InstantiatePolicyFromTemplate(ctx context.Context, req *InstantiatePolicyRequest) (*model.Policy, error)

	// Policy Versions
	CreatePolicyVersion(ctx context.Context, req *CreatePolicyVersionRequest) (*model.PolicyVersion, error)
	GetPolicyVersion(ctx context.Context, policyID uuid.UUID, version int) (*model.PolicyVersion, error)
	ListPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*model.PolicyVersion, error)
	PromotePolicyVersion(ctx context.Context, policyID uuid.UUID, version int) error

	// Policy Attributes
	CreateAttribute(ctx context.Context, req *CreateAttributeRequest) (*model.Attribute, error)
	GetAttribute(ctx context.Context, attributeID uuid.UUID) (*model.Attribute, error)
	UpdateAttribute(ctx context.Context, req *UpdateAttributeRequest) (*model.Attribute, error)
	DeleteAttribute(ctx context.Context, attributeID uuid.UUID) error
	ListAttributes(ctx context.Context, req *ListAttributesRequest) (*ListAttributesResult, error)

	// Policy Combining
	CombinePolicyDecisions(ctx context.Context, decisions []*model.PolicyDecision, algorithm string) (*CombinedPolicyDecision, error)

	// Policy Analysis
	AnalyzePolicyConflicts(ctx context.Context, policyIDs []uuid.UUID) (*PolicyConflictAnalysis, error)
	GetPolicyImpactAnalysis(ctx context.Context, policyID uuid.UUID) (*PolicyImpactAnalysis, error)

	// Policy Caching
	InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error
	WarmupPolicyCache(ctx context.Context, policyIDs []uuid.UUID) error
	GetPolicyCacheStats(ctx context.Context) (*PolicyCacheStats, error)
}

// Policy Management types
type CreatePolicyRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description"`
	Version     int                    `json:"version" validate:"min=1"`
	Effect      model.PolicyEffect     `json:"effect" validate:"required,oneof=allow deny"`
	Target      *model.PolicyTarget    `json:"target" validate:"required"`
	Condition   *model.PolicyCondition `json:"condition,omitempty"`
	Rules       []*model.PolicyRule    `json:"rules,omitempty"`
	Priority    int                    `json:"priority" validate:"min=0"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

type UpdatePolicyRequest struct {
	PolicyID    uuid.UUID              `json:"policy_id" validate:"required"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Effect      *model.PolicyEffect    `json:"effect,omitempty"`
	Target      *model.PolicyTarget    `json:"target,omitempty"`
	Condition   *model.PolicyCondition `json:"condition,omitempty"`
	Rules       []*model.PolicyRule    `json:"rules,omitempty"`
	Priority    *int                   `json:"priority,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

type ListPoliciesRequest struct {
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	ResourceType *string    `json:"resource_type,omitempty"`
	Enabled      *bool      `json:"enabled,omitempty"`
	Limit        int        `json:"limit" validate:"min=1,max=100"`
	Offset       int        `json:"offset" validate:"min=0"`
}

type ListPoliciesResult struct {
	Policies []*model.Policy `json:"policies"`
	Total    int             `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
	HasMore  bool            `json:"has_more"`
}

// Policy Evaluation types
type PolicyEvaluationRequest struct {
	PolicyID    uuid.UUID                `json:"policy_id" validate:"required"`
	Subject     *model.PolicySubject     `json:"subject" validate:"required"`
	Resource    *model.PolicyResource    `json:"resource" validate:"required"`
	Action      string                   `json:"action" validate:"required"`
	Environment *model.PolicyEnvironment `json:"environment,omitempty"`
	Context     map[string]any           `json:"context,omitempty"`
}

type PolicyEvaluationResult struct {
	PolicyID    uuid.UUID                 `json:"policy_id"`
	Decision    model.PolicyDecisionType  `json:"decision"`
	Effect      model.PolicyEffect        `json:"effect"`
	Reason      string                    `json:"reason"`
	Obligations []*model.PolicyObligation `json:"obligations,omitempty"`
	Advice      []*model.PolicyAdvice     `json:"advice,omitempty"`
	Attributes  map[string]any            `json:"attributes,omitempty"`
	EvaluatedAt time.Time                 `json:"evaluated_at"`
}

type BulkPolicyEvaluationRequest struct {
	Requests  []*PolicyEvaluationRequest `json:"requests" validate:"required,min=1,max=50"`
	RequestID string                     `json:"request_id,omitempty"`
}

type BulkPolicyEvaluationResult struct {
	Results   []*PolicyEvaluationResult `json:"results"`
	RequestID string                    `json:"request_id"`
	Timestamp time.Time                 `json:"timestamp"`
}

type PolicyTestRequest struct {
	PolicyID  uuid.UUID               `json:"policy_id" validate:"required"`
	TestCases []*model.PolicyTestCase `json:"test_cases" validate:"required,min=1"`
	RequestID string                  `json:"request_id,omitempty"`
}

type PolicyTestResult struct {
	PolicyID    uuid.UUID                 `json:"policy_id"`
	PolicyName  string                    `json:"policy_name"`
	TestResults []*model.PolicyTestResult `json:"test_results"`
	PassedCount int                       `json:"passed_count"`
	FailedCount int                       `json:"failed_count"`
	RequestID   string                    `json:"request_id"`
	Timestamp   time.Time                 `json:"timestamp"`
}

// Policy Template types
type CreatePolicyTemplateRequest struct {
	Name        string                     `json:"name" validate:"required"`
	Description string                     `json:"description"`
	Category    string                     `json:"category" validate:"required"`
	Template    *model.PolicyTemplateSpec  `json:"template" validate:"required"`
	Parameters  []*model.TemplateParameter `json:"parameters,omitempty"`
	Metadata    map[string]any             `json:"metadata,omitempty"`
}

type InstantiatePolicyRequest struct {
	TemplateID  uuid.UUID      `json:"template_id" validate:"required"`
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// Policy Version types
type CreatePolicyVersionRequest struct {
	PolicyID  uuid.UUID              `json:"policy_id" validate:"required"`
	Changes   string                 `json:"changes" validate:"required"`
	Target    *model.PolicyTarget    `json:"target,omitempty"`
	Condition *model.PolicyCondition `json:"condition,omitempty"`
	Rules     []*model.PolicyRule    `json:"rules,omitempty"`
	Metadata  map[string]any         `json:"metadata,omitempty"`
}

// Attribute Management types
type CreateAttributeRequest struct {
	Name         string                      `json:"name" validate:"required"`
	Type         string                      `json:"type" validate:"required,oneof=string number boolean datetime array object"`
	Category     string                      `json:"category" validate:"required,oneof=subject resource action environment"`
	Description  string                      `json:"description"`
	Required     bool                        `json:"required"`
	Multivalued  bool                        `json:"multivalued"`
	DefaultValue any                         `json:"default_value,omitempty"`
	Constraints  *model.AttributeConstraints `json:"constraints,omitempty"`
	Metadata     map[string]any              `json:"metadata,omitempty"`
}

type UpdateAttributeRequest struct {
	AttributeID  uuid.UUID                   `json:"attribute_id" validate:"required"`
	Name         *string                     `json:"name,omitempty"`
	Type         *string                     `json:"type,omitempty"`
	Category     *string                     `json:"category,omitempty"`
	Description  *string                     `json:"description,omitempty"`
	Required     *bool                       `json:"required,omitempty"`
	Multivalued  *bool                       `json:"multivalued,omitempty"`
	DefaultValue any                         `json:"default_value,omitempty"`
	Constraints  *model.AttributeConstraints `json:"constraints,omitempty"`
	Metadata     map[string]any              `json:"metadata,omitempty"`
}

type ListAttributesRequest struct {
	Category *string `json:"category,omitempty"`
	Type     *string `json:"type,omitempty"`
	Limit    int     `json:"limit" validate:"min=1,max=100"`
	Offset   int     `json:"offset" validate:"min=0"`
}

type ListAttributesResult struct {
	Attributes []*model.Attribute `json:"attributes"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// Policy Combining types
type CombinedPolicyDecision struct {
	Decision    model.PolicyDecisionType  `json:"decision"`
	Algorithm   string                    `json:"algorithm"`
	Decisions   []*model.PolicyDecision   `json:"decisions"`
	Obligations []*model.PolicyObligation `json:"obligations,omitempty"`
	Advice      []*model.PolicyAdvice     `json:"advice,omitempty"`
	Timestamp   time.Time                 `json:"timestamp"`
}

// Policy Analysis types
type PolicyConflictAnalysis struct {
	Conflicts   []*model.PolicyConflict   `json:"conflicts"`
	Warnings    []*model.PolicyWarning    `json:"warnings"`
	Suggestions []*model.PolicySuggestion `json:"suggestions"`
	AnalyzedAt  time.Time                 `json:"analyzed_at"`
}

type PolicyImpactAnalysis struct {
	PolicyID        uuid.UUID      `json:"policy_id"`
	AffectedUsers   int            `json:"affected_users"`
	AffectedRoles   int            `json:"affected_roles"`
	Impact          string         `json:"impact"`
	Recommendations []string       `json:"recommendations"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	AnalyzedAt      time.Time      `json:"analyzed_at"`
}

// Policy Caching types
type PolicyCacheStats struct {
	HitRate       float64   `json:"hit_rate"`
	MissRate      float64   `json:"miss_rate"`
	TotalRequests int64     `json:"total_requests"`
	CacheHits     int64     `json:"cache_hits"`
	CacheMisses   int64     `json:"cache_misses"`
	EvictionCount int64     `json:"eviction_count"`
	CacheSize     int64     `json:"cache_size"`
	LastUpdated   time.Time `json:"last_updated"`
}
