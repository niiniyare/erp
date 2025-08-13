package policy

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
	CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*PolicyTemplate, error)
	GetPolicyTemplate(ctx context.Context, templateID uuid.UUID) (*PolicyTemplate, error)
	ListPolicyTemplates(ctx context.Context) ([]*PolicyTemplate, error)
	InstantiatePolicyFromTemplate(ctx context.Context, req *InstantiatePolicyRequest) (*Policy, error)
	
	// Policy Versions
	CreatePolicyVersion(ctx context.Context, req *CreatePolicyVersionRequest) (*PolicyVersion, error)
	GetPolicyVersion(ctx context.Context, policyID uuid.UUID, version int) (*PolicyVersion, error)
	ListPolicyVersions(ctx context.Context, policyID uuid.UUID) ([]*PolicyVersion, error)
	PromotePolicyVersion(ctx context.Context, policyID uuid.UUID, version int) error
	
	// Policy Attributes
	CreateAttribute(ctx context.Context, req *CreateAttributeRequest) (*Attribute, error)
	GetAttribute(ctx context.Context, attributeID uuid.UUID) (*Attribute, error)
	UpdateAttribute(ctx context.Context, req *UpdateAttributeRequest) (*Attribute, error)
	DeleteAttribute(ctx context.Context, attributeID uuid.UUID) error
	ListAttributes(ctx context.Context, req *ListAttributesRequest) (*ListAttributesResult, error)
	
	// Policy Combining
	CombinePolicyDecisions(ctx context.Context, decisions []*PolicyDecision, algorithm string) (*CombinedPolicyDecision, error)
	
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
	Effect      types.PolicyEffect     `json:"effect" validate:"required,oneof=allow deny"`
	Target      *PolicyTarget          `json:"target" validate:"required"`
	Condition   *PolicyCondition       `json:"condition,omitempty"`
	Rules       []*PolicyRule          `json:"rules,omitempty"`
	Priority    int                    `json:"priority" validate:"min=0"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdatePolicyRequest struct {
	PolicyID    uuid.UUID              `json:"policy_id" validate:"required"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Effect      *types.PolicyEffect    `json:"effect,omitempty"`
	Target      *PolicyTarget          `json:"target,omitempty"`
	Condition   *PolicyCondition       `json:"condition,omitempty"`
	Rules       []*PolicyRule          `json:"rules,omitempty"`
	Priority    *int                   `json:"priority,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ListPoliciesRequest struct {
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	ResourceType *string    `json:"resource_type,omitempty"`
	Enabled      *bool      `json:"enabled,omitempty"`
	Limit        int        `json:"limit" validate:"min=1,max=100"`
	Offset       int        `json:"offset" validate:"min=0"`
}

type ListPoliciesResult struct {
	Policies []*Policy `json:"policies"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
	HasMore  bool      `json:"has_more"`
}

// Policy Evaluation types
type PolicyEvaluationRequest struct {
	PolicyID     uuid.UUID              `json:"policy_id" validate:"required"`
	Subject      *PolicySubject         `json:"subject" validate:"required"`
	Resource     *PolicyResource        `json:"resource" validate:"required"`
	Action       string                 `json:"action" validate:"required"`
	Environment  *PolicyEnvironment     `json:"environment,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type PolicyEvaluationResult struct {
	PolicyID     uuid.UUID                `json:"policy_id"`
	Decision     types.PolicyDecisionType `json:"decision"`
	Effect       types.PolicyEffect       `json:"effect"`
	Reason       string                   `json:"reason"`
	Obligations  []*PolicyObligation      `json:"obligations,omitempty"`
	Advice       []*PolicyAdvice          `json:"advice,omitempty"`
	Attributes   map[string]interface{}   `json:"attributes,omitempty"`
	EvaluatedAt  time.Time                `json:"evaluated_at"`
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
	PolicyID     uuid.UUID              `json:"policy_id" validate:"required"`
	TestCases    []*PolicyTestCase      `json:"test_cases" validate:"required,min=1"`
	RequestID    string                 `json:"request_id,omitempty"`
}

type PolicyTestResult struct {
	PolicyID     uuid.UUID           `json:"policy_id"`
	PolicyName   string              `json:"policy_name"`
	TestResults  []*PolicyTestResult `json:"test_results"`
	PassedCount  int                 `json:"passed_count"`
	FailedCount  int                 `json:"failed_count"`
	RequestID    string              `json:"request_id"`
	Timestamp    time.Time           `json:"timestamp"`
}

// Policy Template types
type CreatePolicyTemplateRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category" validate:"required"`
	Template    *PolicyTemplateSpec    `json:"template" validate:"required"`
	Parameters  []*TemplateParameter   `json:"parameters,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type InstantiatePolicyRequest struct {
	TemplateID   uuid.UUID              `json:"template_id" validate:"required"`
	Name         string                 `json:"name" validate:"required"`
	Description  string                 `json:"description"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// Policy Version types
type CreatePolicyVersionRequest struct {
	PolicyID    uuid.UUID              `json:"policy_id" validate:"required"`
	Changes     string                 `json:"changes" validate:"required"`
	Target      *PolicyTarget          `json:"target,omitempty"`
	Condition   *PolicyCondition       `json:"condition,omitempty"`
	Rules       []*PolicyRule          `json:"rules,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Attribute Management types
type CreateAttributeRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Type        string                 `json:"type" validate:"required,oneof=string number boolean datetime array object"`
	Category    string                 `json:"category" validate:"required,oneof=subject resource action environment"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Multivalued bool                   `json:"multivalued"`
	DefaultValue interface{}           `json:"default_value,omitempty"`
	Constraints *AttributeConstraints  `json:"constraints,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateAttributeRequest struct {
	AttributeID  uuid.UUID              `json:"attribute_id" validate:"required"`
	Name         *string                `json:"name,omitempty"`
	Type         *string                `json:"type,omitempty"`
	Category     *string                `json:"category,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Required     *bool                  `json:"required,omitempty"`
	Multivalued  *bool                  `json:"multivalued,omitempty"`
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Constraints  *AttributeConstraints  `json:"constraints,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ListAttributesRequest struct {
	Category *string `json:"category,omitempty"`
	Type     *string `json:"type,omitempty"`
	Limit    int     `json:"limit" validate:"min=1,max=100"`
	Offset   int     `json:"offset" validate:"min=0"`
}

type ListAttributesResult struct {
	Attributes []*Attribute `json:"attributes"`
	Total      int          `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
	HasMore    bool         `json:"has_more"`
}

// Policy Combining types
type CombinedPolicyDecision struct {
	Decision    types.PolicyDecisionType `json:"decision"`
	Algorithm   string                   `json:"algorithm"`
	Decisions   []*PolicyDecision        `json:"decisions"`
	Obligations []*PolicyObligation      `json:"obligations,omitempty"`
	Advice      []*PolicyAdvice          `json:"advice,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
}

// Policy Analysis types
type PolicyConflictAnalysis struct {
	Conflicts    []*PolicyConflict `json:"conflicts"`
	Warnings     []*PolicyWarning  `json:"warnings"`
	Suggestions  []*PolicySuggestion `json:"suggestions"`
	AnalyzedAt   time.Time         `json:"analyzed_at"`
}

type PolicyImpactAnalysis struct {
	PolicyID      uuid.UUID              `json:"policy_id"`
	AffectedUsers int                    `json:"affected_users"`
	AffectedRoles int                    `json:"affected_roles"`
	Impact        string                 `json:"impact"`
	Recommendations []string             `json:"recommendations"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	AnalyzedAt    time.Time              `json:"analyzed_at"`
}

// Policy Caching types
type PolicyCacheStats struct {
	HitRate         float64   `json:"hit_rate"`
	MissRate        float64   `json:"miss_rate"`
	TotalRequests   int64     `json:"total_requests"`
	CacheHits       int64     `json:"cache_hits"`
	CacheMisses     int64     `json:"cache_misses"`
	EvictionCount   int64     `json:"eviction_count"`
	CacheSize       int64     `json:"cache_size"`
	LastUpdated     time.Time `json:"last_updated"`
}