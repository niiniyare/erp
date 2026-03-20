package repository

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"
	"awo/internal/core/abac/models"
	"awo/internal/shared/types"
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

	// History and metrics
	GetUserEvaluationHistory(ctx context.Context, req *GetUserEvaluationHistoryRequest) ([]*models.PolicyEvaluation, error)
	GetEvaluationMetrics(ctx context.Context, req *GetEvaluationMetricsRequest) (*EvaluationMetrics, error)
}

// AttributeRepository defines the interface for attribute value persistence
type AttributeRepository interface {
	// Attribute definition operations
	CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error)
	GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*models.AttributeDefinition, error)
	GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error)
	UpdateAttributeDefinition(ctx context.Context, req *UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error)
	DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error
	ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error)

	// Attribute value operations
	CreateAttributeValue(ctx context.Context, req *CreateAttributeValueRequest) (*models.AttributeValue, error)
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

// AttributeSourceRepository defines the interface for external attribute source management
type AttributeSourceRepository interface {
	// Source CRUD operations
	CreateAttributeSource(ctx context.Context, req *CreateAttributeSourceRequest) (*models.AttributeSource, error)
	GetAttributeSourceByID(ctx context.Context, id uuid.UUID) (*models.AttributeSource, error)
	UpdateAttributeSource(ctx context.Context, id uuid.UUID, req *UpdateAttributeSourceRequest) (*models.AttributeSource, error)
	DeleteAttributeSource(ctx context.Context, id uuid.UUID) error

	// Source listing and filtering
	ListAttributeSources(ctx context.Context, req *ListAttributeSourcesRequest) ([]*models.AttributeSource, error)
	GetAttributeSourcesByType(ctx context.Context, sourceType string) ([]*models.AttributeSource, error)
	GetActiveAttributeSources(ctx context.Context) ([]*models.AttributeSource, error)

	// Source health and metrics
	UpdateSourceHealth(ctx context.Context, sourceID uuid.UUID, health *SourceHealthStatus) error
	GetSourceMetrics(ctx context.Context, sourceID uuid.UUID) (*SourceMetrics, error)
}

// ─── REQUEST/RESPONSE MODELS ─────────────────────────────────────────────

// CreateAttributeSourceRequest represents attribute source creation request
type CreateAttributeSourceRequest struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Config   map[string]any `json:"config"`
	TenantID uuid.UUID      `json:"tenant_id"`
	IsActive bool           `json:"is_active"`
}

// UpdateAttributeSourceRequest represents attribute source update request
type UpdateAttributeSourceRequest struct {
	Name     *string        `json:"name,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
}

// ListAttributeSourcesRequest represents attribute source listing request
type ListAttributeSourcesRequest struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	SourceType string    `json:"source_type,omitempty"`
	Active     *bool     `json:"active,omitempty"`
	Limit      int32     `json:"limit,omitempty"`
	Offset     int32     `json:"offset,omitempty"`
}

// SourceHealthStatus represents the health status of an attribute source
type SourceHealthStatus struct {
	SourceID     uuid.UUID `json:"source_id"`
	IsHealthy    bool      `json:"is_healthy"`
	LastChecked  time.Time `json:"last_checked"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// SourceMetrics represents metrics for an attribute source
type SourceMetrics struct {
	SourceID        uuid.UUID     `json:"source_id"`
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
}

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	// Audit log operations
	CreateAuditLog(ctx context.Context, req *CreateAuditLogRequest) error
	GetAuditLogs(ctx context.Context, req *GetAuditLogsRequest) ([]*AuditLogEntry, error)
	GetAuditLogsByEntity(ctx context.Context, entityID uuid.UUID, entityType string) ([]*AuditLogEntry, error)
	SearchAuditLogs(ctx context.Context, req *SearchAuditLogsRequest) ([]*AuditLogEntry, error)

	// Cleanup operations
	CleanupOldAuditLogs(ctx context.Context, retentionDays int32) error
}

// CreateAuditLogRequest represents audit log creation request
type CreateAuditLogRequest struct {
	EventType  string         `json:"event_type"`
	EntityID   uuid.UUID      `json:"entity_id"`
	EntityType string         `json:"entity_type"`
	UserID     uuid.UUID      `json:"user_id"`
	Action     string         `json:"action"`
	Details    map[string]any `json:"details,omitempty"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
}

// GetAuditLogsRequest represents audit logs retrieval request
type GetAuditLogsRequest struct {
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	EventType  string     `json:"event_type,omitempty"`
	EntityType string     `json:"entity_type,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Limit      int32      `json:"limit,omitempty"`
	Offset     int32      `json:"offset,omitempty"`
}

// SearchAuditLogsRequest represents audit logs search request
type SearchAuditLogsRequest struct {
	Query     string     `json:"query"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int32      `json:"limit,omitempty"`
	Offset    int32      `json:"offset,omitempty"`
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID         uuid.UUID      `json:"id"`
	EventType  string         `json:"event_type"`
	EntityID   uuid.UUID      `json:"entity_id"`
	EntityType string         `json:"entity_type"`
	UserID     uuid.UUID      `json:"user_id"`
	Action     string         `json:"action"`
	Details    map[string]any `json:"details,omitempty"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// CreateAttributeDefinitionRequest represents attribute definition creation request
type CreateAttributeDefinitionRequest struct {
	Name               string                  `json:"name" validate:"required,min=1,max=100"`
	DisplayName        *string                 `json:"display_name,omitempty"`
	Description        *string                 `json:"description,omitempty"`
	DataType           types.AttributeDataType `json:"data_type" validate:"required"`
	Category           types.AttributeCategory `json:"category" validate:"required"`
	IsRequired         bool                    `json:"is_required"`
	IsSensitive        bool                    `json:"is_sensitive"`
	DefaultValue       *string                 `json:"default_value,omitempty"`
	AllowedValues      []string                `json:"allowed_values,omitempty"`
	ValidationRules    map[string]any          `json:"validation_rules,omitempty"`
	EncryptionRequired bool                    `json:"encryption_required"`
	IsActive           bool                    `json:"is_active"`
}

// UpdateAttributeDefinitionRequest represents attribute definition update request
type UpdateAttributeDefinitionRequest struct {
	DisplayName        *string        `json:"display_name,omitempty"`
	Description        *string        `json:"description,omitempty"`
	IsRequired         *bool          `json:"is_required,omitempty"`
	IsSensitive        *bool          `json:"is_sensitive,omitempty"`
	DefaultValue       *string        `json:"default_value,omitempty"`
	AllowedValues      []string       `json:"allowed_values,omitempty"`
	ValidationRules    map[string]any `json:"validation_rules,omitempty"`
	IsActive           *bool          `json:"is_active,omitempty"`
	EncryptionRequired *bool          `json:"encryption_required,omitempty"`
}

// ListAttributeDefinitionsRequest represents attribute definition list request
type ListAttributeDefinitionsRequest struct {
	Category    *types.AttributeCategory `json:"category,omitempty"`
	Search      string                   `json:"search"`
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

// CreateAttributeValueRequest represents create attribute value request
type CreateAttributeValueRequest struct {
	DefinitionID   uuid.UUID  `json:"definition_id" validate:"required"`
	EntityID       uuid.UUID  `json:"entity_id" validate:"required"`
	Value          string     `json:"value" validate:"required"`
	EncryptedValue []byte     `json:"encrypted_value,omitempty"`
	IsEncrypted    bool       `json:"is_encrypted"`
	Version        int32      `json:"version"`
	EffectiveFrom  time.Time  `json:"effective_from" validate:"required"`
	EffectiveTo    *time.Time `json:"effective_to,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by" validate:"required"`
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
	TotalDefinitions       int   `json:"total_definitions"`
	TotalValues            int   `json:"total_values"`
	EntitiesWithAttributes int   `json:"entities_with_attributes"`
	UserAttributes         int64 `json:"user_attributes"`
	ResourceAttributes     int64 `json:"resource_attributes"`
	EnvironmentAttributes  int64 `json:"environment_attributes"`
}

// ─── HISTORY AND METRICS REQUEST TYPES ────────────────────────────────────

// GetUserEvaluationHistoryRequest represents a request to get user evaluation history
type GetUserEvaluationHistoryRequest struct {
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	ResourceType *string   `json:"resource_type,omitempty"`
	Action       *string   `json:"action,omitempty"`
	Limit        int       `json:"limit" validate:"min=1,max=1000"`
	Offset       int       `json:"offset" validate:"min=0"`
}

// GetEvaluationMetricsRequest represents a request to get evaluation metrics
type GetEvaluationMetricsRequest struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// ─── RESULT TYPES ──────────────────────────────────────────────────────────

// CachedEvaluationResult represents a cached evaluation result
type CachedEvaluationResult struct {
	Decision        types.PolicyDecisionType `json:"decision"`
	PolicyDecisions []*models.PolicyDecision `json:"policy_decisions"`
	CachedAt        time.Time                `json:"cached_at"`
	ExpiresAt       time.Time                `json:"expires_at"`
}

// EvaluationMetrics represents evaluation performance metrics
type EvaluationMetrics struct {
	TotalEvaluations       int     `json:"total_evaluations"`
	UniqueUsers            int     `json:"unique_users"`
	UniqueResources        int     `json:"unique_resources"`
	AvgEvaluationTimeMS    float64 `json:"avg_evaluation_time_ms"`
	MedianEvaluationTimeMS float64 `json:"median_evaluation_time_ms"`
	P95EvaluationTimeMS    float64 `json:"p95_evaluation_time_ms"`
	P99EvaluationTimeMS    float64 `json:"p99_evaluation_time_ms"`
}
