package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/types"
)

// ─── CORE ABAC DOMAIN MODELS ─────────────────────────────────────────────

// AttributeSource represents an external source for attribute collection
type AttributeSource struct {
	ID        uuid.UUID      `json:"id"`
	TenantID  uuid.UUID      `json:"tenant_id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"` // "ldap", "rest_api", "database", etc.
	Config    map[string]any `json:"config"`
	IsActive  bool           `json:"is_active"`
	Priority  int32          `json:"priority"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty"`
}

// AttributeDefinition defines the structure and constraints of an attribute
type AttributeDefinition struct {
	ID                 uuid.UUID               `json:"id"`
	TenantID           uuid.UUID               `json:"tenant_id"`
	Name               string                  `json:"name"`
	DisplayName        *string                 `json:"display_name,omitempty"`
	Description        *string                 `json:"description,omitempty"`
	DataType           types.AttributeDataType `json:"data_type"`
	Category           types.AttributeCategory `json:"category"`
	IsRequired         bool                    `json:"is_required"`
	IsSensitive        bool                    `json:"is_sensitive"`
	DefaultValue       *string                 `json:"default_value,omitempty"`
	AllowedValues      []string                `json:"allowed_values,omitempty"`
	ValidationRules    map[string]any          `json:"validation_rules,omitempty"`
	EncryptionRequired bool                    `json:"encryption_required"`
	IsActive           bool                    `json:"is_active"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	DeletedAt          *time.Time              `json:"deleted_at,omitempty"`
}

// Validate validates the attribute definition using shared error handling
func (ad *AttributeDefinition) Validate() error {
	if ad.Name == "" {
		return errors.NewBusinessError("ATTRIBUTE_NAME_REQUIRED", "Attribute name is required").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Provide a valid attribute name")
	}

	if !ad.DataType.IsValid() {
		return errors.NewBusinessError("INVALID_ATTRIBUTE_DATA_TYPE", "Invalid attribute data type").
			WithCategory(errors.CategoryValidation).
			WithDetail("data_type", ad.DataType).
			WithSuggestion("Use a valid data type: STRING, NUMBER, BOOLEAN, DATE, JSON, ARRAY, ENUM")
	}

	if !ad.Category.IsValid() {
		return errors.NewBusinessError("INVALID_ATTRIBUTE_CATEGORY", "Invalid attribute category").
			WithCategory(errors.CategoryValidation).
			WithDetail("category", ad.Category).
			WithSuggestion("Use a valid category: USER, RESOURCE, ENVIRONMENT, ACTION, ENTITY, SESSION")
	}

	// Validate enum values if data type is ENUM
	if ad.DataType == types.AttributeDataTypeEnum {
		if len(ad.AllowedValues) == 0 {
			return errors.NewBusinessError("ENUM_VALUES_REQUIRED", "Enum data type requires allowed values").
				WithCategory(errors.CategoryValidation).
				WithSuggestion("Provide at least one allowed value for enum attributes")
		}
	}

	return nil
}

// LogCreation logs attribute definition creation with structured logging
func (ad *AttributeDefinition) LogCreation(ctx context.Context, log logger.Logger) {
	log.InfoContext(ctx, "ABAC attribute definition created",
		logger.Fields{
			"attribute_id":   ad.ID,
			"attribute_name": ad.Name,
			"data_type":      ad.DataType,
			"category":       ad.Category,
			"is_required":    ad.IsRequired,
			"is_sensitive":   ad.IsSensitive,
			"tenant_id":      ad.TenantID,
		})
}

// Policy represents an ABAC policy with enhanced integration
type Policy struct {
	ID                 uuid.UUID                      `json:"id"`
	TenantID           uuid.UUID                      `json:"tenant_id"`
	Name               string                         `json:"name"`
	DisplayName        *string                        `json:"display_name,omitempty"`
	Description        *string                        `json:"description,omitempty"`
	PolicyType         types.PolicyType               `json:"policy_type"`
	Effect             types.PolicyEffect             `json:"effect"`
	Priority           int32                          `json:"priority"`
	Category           types.PolicyCategory           `json:"category"`
	Target             map[string]any                 `json:"target"`
	Rule               map[string]any                 `json:"rule"`
	Obligations        map[string]any                 `json:"obligations,omitempty"`
	Advice             map[string]any                 `json:"advice,omitempty"`
	CombiningAlgorithm types.PolicyCombiningAlgorithm `json:"combining_algorithm"`
	IsActive           bool                           `json:"is_active"`
	Version            int32                          `json:"version"`
	PreviousVersionID  *uuid.UUID                     `json:"previous_version_id,omitempty"`
	ExpiresAt          *time.Time                     `json:"expires_at,omitempty"`
	ApprovedBy         *uuid.UUID                     `json:"approved_by,omitempty"`
	ApprovedAt         *time.Time                     `json:"approved_at,omitempty"`
	CreatedBy          uuid.UUID                      `json:"created_by"`
	CreatedAt          time.Time                      `json:"created_at"`
	UpdatedAt          time.Time                      `json:"updated_at"`
	DeletedAt          *time.Time                     `json:"deleted_at,omitempty"`
}

// Validate validates the policy using shared error handling
func (p *Policy) Validate() error {
	if p.Name == "" {
		return errors.NewBusinessError("POLICY_NAME_REQUIRED", "Policy name is required").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Provide a valid policy name")
	}

	if !p.Effect.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_EFFECT", "Invalid policy effect").
			WithCategory(errors.CategoryValidation).
			WithDetail("effect", p.Effect).
			WithSuggestion("Use ALLOW, DENY, or NOT_APPLICABLE as policy effect")
	}

	if !p.PolicyType.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_TYPE", "Invalid policy type").
			WithCategory(errors.CategoryValidation).
			WithDetail("policy_type", p.PolicyType).
			WithSuggestion("Use a valid policy type: ACCESS, DELEGATION, OBLIGATION, REFRAIN, CONDITION")
	}

	if !p.Category.IsValid() {
		return errors.NewBusinessError("INVALID_POLICY_CATEGORY", "Invalid policy category").
			WithCategory(errors.CategoryValidation).
			WithDetail("category", p.Category).
			WithSuggestion("Use a valid policy category: SECURITY, COMPLIANCE, BUSINESS, OPERATIONAL, SYSTEM")
	}

	if p.Priority < 0 {
		return errors.NewBusinessError("INVALID_POLICY_PRIORITY", "Policy priority must be non-negative").
			WithCategory(errors.CategoryValidation).
			WithDetail("priority", p.Priority).
			WithSuggestion("Use a non-negative priority value")
	}

	if len(p.Target) == 0 {
		return errors.NewBusinessError("POLICY_TARGET_REQUIRED", "Policy target is required").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Define policy target conditions")
	}

	if len(p.Rule) == 0 {
		return errors.NewBusinessError("POLICY_RULE_REQUIRED", "Policy rule is required").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Define policy rule conditions")
	}

	if !p.CombiningAlgorithm.IsValid() {
		return errors.NewBusinessError("INVALID_COMBINING_ALGORITHM", "Invalid policy combining algorithm").
			WithCategory(errors.CategoryValidation).
			WithDetail("combining_algorithm", p.CombiningAlgorithm).
			WithSuggestion("Use a valid combining algorithm")
	}

	return nil
}

// IsExpired checks if the policy has expired
func (p *Policy) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// LogEvaluation logs policy evaluation with enhanced context
func (p *Policy) LogEvaluation(ctx context.Context, log logger.Logger, decision types.PolicyDecisionType, evaluationTime time.Duration) {
	log.InfoContext(ctx, "ABAC policy evaluated",
		logger.Fields{
			"policy_id":          p.ID,
			"policy_name":        p.Name,
			"policy_effect":      p.Effect,
			"policy_type":        p.PolicyType,
			"decision":           decision,
			"evaluation_time_ms": evaluationTime.Milliseconds(),
			"policy_priority":    p.Priority,
			"tenant_id":          p.TenantID,
		})
}

// RecordMetrics records policy evaluation metrics
func (p *Policy) RecordMetrics(ctx context.Context, m metrics.MetricsProvider, decision types.PolicyDecisionType, evaluationTime time.Duration) {

	// Policy evaluation counter
	policyCounter := m.Counter(
		"abac_policy_evaluations_total",
		"Total number of ABAC policy evaluations",
		"policy_id", "policy_name", "policy_type", "effect", "decision", "tenant_id",
	)

	policyCounter.Inc(metrics.Fields{
		"policy_id":   p.ID.String(),
		"policy_name": p.Name,
		"policy_type": string(p.PolicyType),
		"effect":      string(p.Effect),
		"decision":    string(decision),
		"tenant_id":   p.TenantID.String(),
	})

	// Evaluation duration histogram
	durationHist := m.Histogram(
		"abac_policy_evaluation_duration_seconds",
		"Time spent evaluating ABAC policies",
		metrics.StandardHTTPDurationBuckets(),
		"policy_type", "policy_effect", "decision",
	)

	durationHist.Observe(evaluationTime.Seconds(), metrics.Fields{
		"policy_type":   string(p.PolicyType),
		"policy_effect": string(p.Effect),
		"decision":      string(decision),
	})
}

// PolicyEvaluationRequest represents a request to evaluate policies
type PolicyEvaluationRequest struct {
	UserID              uuid.UUID       `json:"user_id" validate:"required"`
	ResourceType        string          `json:"resource_type" validate:"required"`
	ResourceID          *uuid.UUID      `json:"resource_id,omitempty"`
	Action              string          `json:"action" validate:"required"`
	EntityID            *uuid.UUID      `json:"entity_id,omitempty"`
	UserAttributes      map[string]any  `json:"user_attributes,omitempty"`
	ResourceAttributes  map[string]any  `json:"resource_attributes,omitempty"`
	EnvironmentContext  map[string]any  `json:"environment_context,omitempty"`
	SessionAttributes   map[string]any  `json:"session_attributes,omitempty"`
	RequestID           string          `json:"request_id"`
	TenantID            uuid.UUID       `json:"tenant_id"`
	TenantFeatures      map[string]bool `json:"tenant_features,omitempty"`
	TenantSecurityLevel string          `json:"tenant_security_level,omitempty"`
}

// GetTraceAttributes returns OpenTelemetry attributes for tracing
func (per *PolicyEvaluationRequest) GetTraceAttributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("abac.user_id", per.UserID.String()),
		attribute.String("abac.resource_type", per.ResourceType),
		attribute.String("abac.action", per.Action),
		attribute.String("abac.request_id", per.RequestID),
		attribute.String("abac.tenant_id", per.TenantID.String()),
	}

	if per.ResourceID != nil {
		attrs = append(attrs, attribute.String("abac.resource_id", per.ResourceID.String()))
	}

	if per.EntityID != nil {
		attrs = append(attrs, attribute.String("abac.entity_id", per.EntityID.String()))
	}

	attrs = append(attrs, attribute.Int("abac.user_attributes_count", len(per.UserAttributes)))
	attrs = append(attrs, attribute.Int("abac.resource_attributes_count", len(per.ResourceAttributes)))
	attrs = append(attrs, attribute.Int("abac.environment_context_count", len(per.EnvironmentContext)))

	return attrs
}

// Validate validates the evaluation request
func (per *PolicyEvaluationRequest) Validate() error {
	if per.UserID == uuid.Nil {
		return errors.NewBusinessError("USER_ID_REQUIRED", "User ID is required for policy evaluation").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Provide a valid user ID")
	}

	if per.ResourceType == "" {
		return errors.NewBusinessError("RESOURCE_TYPE_REQUIRED", "Resource type is required for policy evaluation").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Specify the type of resource being accessed")
	}

	if per.Action == "" {
		return errors.NewBusinessError("ACTION_REQUIRED", "Action is required for policy evaluation").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Specify the action being performed")
	}

	if per.RequestID == "" {
		return errors.NewBusinessError("REQUEST_ID_REQUIRED", "Request ID is required for tracing").
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Provide a unique request ID for tracing purposes")
	}

	return nil
}

// PolicyEvaluationResult represents the result of policy evaluation
type PolicyEvaluationResult struct {
	RequestID          string                   `json:"request_id"`
	Decision           types.PolicyDecisionType `json:"decision"`
	Status             types.EvaluationStatus   `json:"status"`
	ApplicablePolicies []uuid.UUID              `json:"applicable_policies"`
	PolicyDecisions    []PolicyDecisionInfo     `json:"policy_decisions"`
	EvaluationTimeMS   int64                    `json:"evaluation_time_ms"`
	CacheHit           bool                     `json:"cache_hit"`
	Obligations        []PolicyObligation       `json:"obligations,omitempty"`
	Advice             []PolicyAdvice           `json:"advice,omitempty"`
	EvaluationContext  map[string]any           `json:"evaluation_context,omitempty"`
	EvaluatedAt        time.Time                `json:"evaluated_at"`
	CachedAt           time.Time                `json:"cached_at,omitempty"`
	ExpiresAt          time.Time                `json:"expires_at,omitempty"`
	TenantID           uuid.UUID                `json:"tenant_id"`
	Error              *string                  `json:"error,omitempty"`
}

// PolicyDecisionInfo provides detailed information about a policy decision
type PolicyDecisionInfo struct {
	PolicyID       uuid.UUID                `json:"policy_id"`
	PolicyName     string                   `json:"policy_name"`
	PolicyType     types.PolicyType         `json:"policy_type"`
	Decision       types.PolicyDecisionType `json:"decision"`
	Effect         types.PolicyEffect       `json:"effect"`
	Priority       int32                    `json:"priority"`
	Reason         string                   `json:"reason,omitempty"`
	MatchedRules   []string                 `json:"matched_rules,omitempty"`
	EvaluationTime int64                    `json:"evaluation_time_ms"`
	TargetMatched  bool                     `json:"target_matched"`
	RuleEvaluated  bool                     `json:"rule_evaluated"`
}

// PolicyObligation represents an action that must be taken when a policy is applied
type PolicyObligation struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Fulfilled   bool           `json:"fulfilled"`
}

// PolicyAdvice represents advisory information from policy evaluation
type PolicyAdvice struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// LogResult logs the evaluation result with enhanced context
func (per *PolicyEvaluationResult) LogResult(ctx context.Context, log logger.Logger) {
	log.InfoContext(ctx, "ABAC policy evaluation completed",
		logger.Fields{
			"request_id":          per.RequestID,
			"decision":            per.Decision,
			"status":              per.Status,
			"evaluation_time_ms":  per.EvaluationTimeMS,
			"cache_hit":           per.CacheHit,
			"applicable_policies": len(per.ApplicablePolicies),
			"policy_decisions":    len(per.PolicyDecisions),
			"obligations":         len(per.Obligations),
			"advice":              len(per.Advice),
			"tenant_id":           per.TenantID,
			"error":               per.Error,
		})
}

// GetTraceAttributes returns OpenTelemetry attributes for result tracing
func (per *PolicyEvaluationResult) GetTraceAttributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("abac.result.request_id", per.RequestID),
		attribute.String("abac.result.decision", string(per.Decision)),
		attribute.String("abac.result.status", string(per.Status)),
		attribute.Int64("abac.result.evaluation_time_ms", per.EvaluationTimeMS),
		attribute.Bool("abac.result.cache_hit", per.CacheHit),
		attribute.Int("abac.result.applicable_policies_count", len(per.ApplicablePolicies)),
		attribute.Int("abac.result.policy_decisions_count", len(per.PolicyDecisions)),
		attribute.Int("abac.result.obligations_count", len(per.Obligations)),
		attribute.Int("abac.result.advice_count", len(per.Advice)),
		attribute.String("abac.result.tenant_id", per.TenantID.String()),
	}

	if per.Error != nil {
		attrs = append(attrs, attribute.String("abac.result.error", *per.Error))
	}

	return attrs
}

// RecordMetrics records evaluation result metrics
func (per *PolicyEvaluationResult) RecordMetrics(ctx context.Context, m metrics.MetricsProvider) {
	// Evaluation result counter
	resultCounter := m.Counter(
		"abac_evaluations_total",
		"Total number of ABAC evaluations",
		"decision", "status", "cache_hit", "tenant_id",
	)

	resultCounter.Inc(metrics.Fields{
		"decision":  string(per.Decision),
		"status":    string(per.Status),
		"cache_hit": fmt.Sprintf("%t", per.CacheHit),
		"tenant_id": per.TenantID.String(),
	})

	// Evaluation duration histogram
	durationHist := m.Histogram(
		"abac_evaluation_duration_seconds",
		"Time spent on ABAC evaluations",
		metrics.StandardHTTPDurationBuckets(),
		"decision", "status", "cache_hit",
	)

	durationHist.Observe(float64(per.EvaluationTimeMS)/1000.0, metrics.Fields{
		"decision":  string(per.Decision),
		"status":    string(per.Status),
		"cache_hit": fmt.Sprintf("%t", per.CacheHit),
	})

	// Policy count histogram
	policyCountHist := m.Histogram(
		"abac_applicable_policies_count",
		"Number of applicable policies per evaluation",
		[]float64{0, 1, 2, 5, 10, 20, 50, 100},
		"decision", "status",
	)

	policyCountHist.Observe(float64(len(per.ApplicablePolicies)), metrics.Fields{
		"decision": string(per.Decision),
		"status":   string(per.Status),
	})
}

// AttributeValue represents a typed attribute value with metadata
type AttributeValue struct {
	DefinitionID uuid.UUID               `json:"definition_id"`
	Name         string                  `json:"name"`
	Value        any                     `json:"value"`
	DataType     types.AttributeDataType `json:"data_type"`
	Category     types.AttributeCategory `json:"category"`
	Source       types.AttributeSource   `json:"source"`
	Confidence   float64                 `json:"confidence"`
	Timestamp    time.Time               `json:"timestamp"`
	ExpiresAt    *time.Time              `json:"expires_at,omitempty"`
	IsSensitive  bool                    `json:"is_sensitive"`
}

// IsExpired checks if the attribute value has expired
func (av *AttributeValue) IsExpired() bool {
	if av.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*av.ExpiresAt)
}

// GetSecureValue returns the value, masking sensitive data if needed
func (av *AttributeValue) GetSecureValue() any {
	if av.IsSensitive {
		switch v := av.Value.(type) {
		case string:
			if len(v) > 4 {
				return v[:2] + "***" + v[len(v)-2:]
			}
			return "***"
		default:
			return "***"
		}
	}
	return av.Value
}

// MarshalJSON implements custom JSON marshaling for sensitive data
func (av *AttributeValue) MarshalJSON() ([]byte, error) {
	type Alias AttributeValue
	return json.Marshal(&struct {
		Value any `json:"value"`
		*Alias
	}{
		Value: av.GetSecureValue(),
		Alias: (*Alias)(av),
	})
}

// AttributeContext represents a collection of attributes for evaluation
type AttributeContext struct {
	UserAttributes        map[string]*AttributeValue `json:"user_attributes"`
	ResourceAttributes    map[string]*AttributeValue `json:"resource_attributes"`
	EnvironmentAttributes map[string]*AttributeValue `json:"environment_attributes"`
	SessionAttributes     map[string]*AttributeValue `json:"session_attributes"`
	EntityAttributes      map[string]*AttributeValue `json:"entity_attributes"`
	ActionAttributes      map[string]*AttributeValue `json:"action_attributes"`
	CollectedAt           time.Time                  `json:"collected_at"`
	TenantID              uuid.UUID                  `json:"tenant_id"`
}

// GetAllAttributes returns all attributes in a flat map
func (ac *AttributeContext) GetAllAttributes() map[string]*AttributeValue {
	allAttrs := make(map[string]*AttributeValue)

	for k, v := range ac.UserAttributes {
		allAttrs[k] = v
	}
	for k, v := range ac.ResourceAttributes {
		allAttrs[k] = v
	}
	for k, v := range ac.EnvironmentAttributes {
		allAttrs[k] = v
	}
	for k, v := range ac.SessionAttributes {
		allAttrs[k] = v
	}
	for k, v := range ac.EntityAttributes {
		allAttrs[k] = v
	}
	for k, v := range ac.ActionAttributes {
		allAttrs[k] = v
	}

	return allAttrs
}

// GetExpiredAttributes returns attributes that have expired
func (ac *AttributeContext) GetExpiredAttributes() map[string]*AttributeValue {
	expired := make(map[string]*AttributeValue)

	for k, v := range ac.GetAllAttributes() {
		if v.IsExpired() {
			expired[k] = v
		}
	}

	return expired
}

// LogCollection logs attribute collection with metrics
func (ac *AttributeContext) LogCollection(ctx context.Context, log logger.Logger) {
	totalAttrs := len(ac.UserAttributes) + len(ac.ResourceAttributes) +
		len(ac.EnvironmentAttributes) + len(ac.SessionAttributes) +
		len(ac.EntityAttributes) + len(ac.ActionAttributes)

	expiredCount := len(ac.GetExpiredAttributes())

	log.InfoContext(ctx, "ABAC attribute context collected",
		logger.Fields{
			"total_attributes":       totalAttrs,
			"user_attributes":        len(ac.UserAttributes),
			"resource_attributes":    len(ac.ResourceAttributes),
			"environment_attributes": len(ac.EnvironmentAttributes),
			"session_attributes":     len(ac.SessionAttributes),
			"entity_attributes":      len(ac.EntityAttributes),
			"action_attributes":      len(ac.ActionAttributes),
			"expired_attributes":     expiredCount,
			"collected_at":           ac.CollectedAt,
			"tenant_id":              ac.TenantID,
		})
}

// ─── HELPER FUNCTIONS ─────────────────────────────────────────────

// NewAttributeValue creates a new attribute value with proper defaults
func NewAttributeValue(definitionID uuid.UUID, name string, value any, dataType types.AttributeDataType, category types.AttributeCategory, source types.AttributeSource) *AttributeValue {
	return &AttributeValue{
		DefinitionID: definitionID,
		Name:         name,
		Value:        value,
		DataType:     dataType,
		Category:     category,
		Source:       source,
		Confidence:   1.0,
		Timestamp:    time.Now(),
		IsSensitive:  false,
	}
}

// NewAttributeContext creates a new attribute context with initialized maps
func NewAttributeContext(tenantID uuid.UUID) *AttributeContext {
	return &AttributeContext{
		UserAttributes:        make(map[string]*AttributeValue),
		ResourceAttributes:    make(map[string]*AttributeValue),
		EnvironmentAttributes: make(map[string]*AttributeValue),
		SessionAttributes:     make(map[string]*AttributeValue),
		EntityAttributes:      make(map[string]*AttributeValue),
		ActionAttributes:      make(map[string]*AttributeValue),
		CollectedAt:           time.Now(),
		TenantID:              tenantID,
	}
}

// PolicyEvaluation represents a stored policy evaluation record
type PolicyEvaluation struct {
	ID                 uuid.UUID                `json:"id"`
	TenantID           uuid.UUID                `json:"tenant_id"`
	UserID             uuid.UUID                `json:"user_id"`
	ResourceType       string                   `json:"resource_type"`
	ResourceID         *uuid.UUID               `json:"resource_id,omitempty"`
	Action             string                   `json:"action"`
	EntityID           *uuid.UUID               `json:"entity_id,omitempty"`
	ContextHash        string                   `json:"context_hash"`
	Decision           types.PolicyDecisionType `json:"decision"`
	ApplicablePolicies []uuid.UUID              `json:"applicable_policies"`
	PolicyDecisions    []*PolicyDecision        `json:"policy_decisions"`
	EvaluationTimeMS   *int64                   `json:"evaluation_time_ms,omitempty"`
	EvaluatedAt        time.Time                `json:"evaluated_at"`
	ExpiresAt          time.Time                `json:"expires_at"`
}

type PolicyDecision struct {
	PolicyID      uuid.UUID                `json:"policy_id"`
	Decision      types.PolicyDecisionType `json:"decision"`
	Reason        string                   `json:"reason,omitempty"`
	MatchedRule   string                   `json:"matched_rule,omitempty"`
	EvaluationMS  int64                    `json:"evaluation_ms,omitempty"`
	TargetMatched bool                     `json:"target_matched"`
}

// PolicyRule represents a policy rule for evaluation
type PolicyRule struct {
	ID         string         `json:"id"`
	Name       string         `json:"name,omitempty"`
	Expression string         `json:"expression"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Conditions []string       `json:"conditions,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// PolicyTarget represents a policy target for evaluation
type PolicyTarget struct {
	Resources     []string       `json:"resources,omitempty"`
	Actions       []string       `json:"actions,omitempty"`
	Subjects      []string       `json:"subjects,omitempty"`
	Environment   map[string]any `json:"environment,omitempty"`
	ResourceTypes []string       `json:"resource_types,omitempty"`
}

// PolicyCondition represents a policy condition for evaluation
type PolicyCondition struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Expression string         `json:"expression"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Operator   string         `json:"operator,omitempty"`
	Values     []any          `json:"values,omitempty"`
}

// AttributeMatch represents an attribute match condition for evaluation
type AttributeMatch struct {
	AttributeName string `json:"attribute_name"`
	MatchType     string `json:"match_type"` // equals, contains, regex, in
	Value         any    `json:"value"`
	CaseSensitive bool   `json:"case_sensitive"`
}
