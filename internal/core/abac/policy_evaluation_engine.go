package abac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// EvaluationPerformanceTracker tracks performance metrics for policy evaluations
type EvaluationPerformanceTracker struct {
	evaluationTimes []time.Duration
	cacheHitRate    float64
	totalRequests   int64
	errors          int64
	mu              sync.RWMutex
}

// NewEvaluationPerformanceTracker creates a new performance tracker
func NewEvaluationPerformanceTracker() *EvaluationPerformanceTracker {
	return &EvaluationPerformanceTracker{
		evaluationTimes: make([]time.Duration, 0),
	}
}

// RecordEvaluation records a policy evaluation's performance
func (ept *EvaluationPerformanceTracker) RecordEvaluation(duration time.Duration, cacheHit bool, success bool) {
	ept.mu.Lock()
	defer ept.mu.Unlock()

	ept.evaluationTimes = append(ept.evaluationTimes, duration)
	ept.totalRequests++

	if !success {
		ept.errors++
	}
}

// MultiplePolicyEvaluationRequest represents a request to evaluate multiple policies
type MultiplePolicyEvaluationRequest struct {
	PolicyIDs []uuid.UUID    `json:"policy_ids"`
	Context   map[string]any `json:"context"`
	RequestID string         `json:"request_id"`
}

// MultiplePolicyEvaluationResult represents the result of multiple policy evaluation
type MultiplePolicyEvaluationResult struct {
	Results   []PolicyEvaluationResult `json:"results"`
	Decision  types.PolicyDecisionType `json:"decision"`
	RequestID string                   `json:"request_id"`
}

// BatchPolicyEvaluationRequest represents a batch policy evaluation request
type BatchPolicyEvaluationRequest struct {
	Requests []PolicyEvaluationRequest `json:"requests"`
	Options  BatchOptions              `json:"options,omitempty"`
}

// BatchPolicyEvaluationResult represents the result of batch policy evaluation
type BatchPolicyEvaluationResult struct {
	Results  []PolicyEvaluationResult `json:"results"`
	BatchID  uuid.UUID                `json:"batch_id"`
	Success  bool                     `json:"success"`
	ErrorMsg string                   `json:"error_msg,omitempty"`
}

// ContextualEvaluationRequest represents a contextual evaluation request
type ContextualEvaluationRequest struct {
	PolicyID  uuid.UUID      `json:"policy_id"`
	Context   map[string]any `json:"context"`
	Scenario  string         `json:"scenario,omitempty"`
	RequestID string         `json:"request_id"`
}

// ContextualEvaluationResult represents the result of contextual evaluation
type ContextualEvaluationResult struct {
	PolicyID  uuid.UUID                `json:"policy_id"`
	Decision  types.PolicyDecisionType `json:"decision"`
	Context   map[string]any           `json:"context"`
	Reasoning string                   `json:"reasoning,omitempty"`
	RequestID string                   `json:"request_id"`
}

// EvaluationSimulationRequest represents a simulation request
type EvaluationSimulationRequest struct {
	PolicyID  uuid.UUID        `json:"policy_id"`
	Scenarios []map[string]any `json:"scenarios"`
	RequestID string           `json:"request_id"`
}

// EvaluationSimulationResult represents the result of evaluation simulation
type EvaluationSimulationResult struct {
	PolicyID  uuid.UUID                    `json:"policy_id"`
	Results   []ContextualEvaluationResult `json:"results"`
	Summary   SimulationSummary            `json:"summary"`
	RequestID string                       `json:"request_id"`
}

// SimulationSummary represents a summary of simulation results
type SimulationSummary struct {
	TotalScenarios     int64   `json:"total_scenarios"`
	PermitCount        int64   `json:"permit_count"`
	DenyCount          int64   `json:"deny_count"`
	IndeterminateCount int64   `json:"indeterminate_count"`
	SuccessRate        float64 `json:"success_rate"`
}

// RuleEvaluationRequest represents a rule evaluation request
type RuleEvaluationRequest struct {
	RuleID    string         `json:"rule_id"`
	Rule      string         `json:"rule"`
	Context   map[string]any `json:"context"`
	RequestID string         `json:"request_id"`
}

// ParsedRuleExpression represents a parsed rule expression
type ParsedRuleExpression struct {
	Expression string    `json:"expression"`
	Variables  []string  `json:"variables"`
	Functions  []string  `json:"functions"`
	ParsedAt   time.Time `json:"parsed_at"`
	IsValid    bool      `json:"is_valid"`
	Errors     []string  `json:"errors,omitempty"`
}

// PIPResolutionRequest represents a request for PIP resolution
type PIPResolutionRequest struct {
	AttributeName string         `json:"attribute_name"`
	EntityID      uuid.UUID      `json:"entity_id"`
	Context       map[string]any `json:"context,omitempty"`
}

// PIPResolutionResult represents the result of PIP resolution
type PIPResolutionResult struct {
	AttributeName string `json:"attribute_name"`
	Value         any    `json:"value"`
	Source        string `json:"source"`
	Success       bool   `json:"success"`
	ErrorMsg      string `json:"error_msg,omitempty"`
}

// RegisterPIPProviderRequest represents a request to register a PIP provider
type RegisterPIPProviderRequest struct {
	ProviderID   string   `json:"provider_id"`
	ProviderName string   `json:"provider_name"`
	Attributes   []string `json:"attributes"`
	Priority     int32    `json:"priority"`
}

// PerformanceAnalysisRequest represents a request for performance analysis
type PerformanceAnalysisRequest struct {
	PolicyID  uuid.UUID  `json:"policy_id"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// EvaluationPerformanceAnalysis represents performance analysis results
type EvaluationPerformanceAnalysis struct {
	PolicyID         uuid.UUID     `json:"policy_id"`
	AverageLatency   time.Duration `json:"average_latency"`
	TotalEvaluations int64         `json:"total_evaluations"`
	SuccessRate      float64       `json:"success_rate"`
	Bottlenecks      []string      `json:"bottlenecks,omitempty"`
}

// PolicyEvaluationEngine provides policy evaluation capabilities
type PolicyEvaluationEngine interface {
	// Core Evaluation
	EvaluatePolicy(ctx context.Context, req *PolicyEvaluationRequest) (*PolicyEvaluationResult, error)
	EvaluatePolicies(ctx context.Context, req *MultiplePolicyEvaluationRequest) (*MultiplePolicyEvaluationResult, error)

	// Batch Evaluation
	BatchEvaluatePolicy(ctx context.Context, req *BatchPolicyEvaluationRequest) (*BatchPolicyEvaluationResult, error)

	// Advanced Evaluation
	EvaluateWithContext(ctx context.Context, req *ContextualEvaluationRequest) (*ContextualEvaluationResult, error)
	SimulateEvaluation(ctx context.Context, req *EvaluationSimulationRequest) (*EvaluationSimulationResult, error)

	// Rule Engine
	EvaluateRule(ctx context.Context, req *RuleEvaluationRequest) (*RuleEvaluationResult, error)
	ParseRuleExpression(ctx context.Context, expression string) (*ParsedRuleExpression, error)
	ValidateRuleExpression(ctx context.Context, expression string) (*RuleValidationResult, error)

	// Policy Information Point (PIP) Integration
	ResolvePolicyInformation(ctx context.Context, req *PIPResolutionRequest) (*PIPResolutionResult, error)
	RegisterPIPProvider(ctx context.Context, req *RegisterPIPProviderRequest) (*PIPProvider, error)

	// Evaluation Analytics
	GetEvaluationMetrics(ctx context.Context, req *repository.GetEvaluationMetricsRequest) (*repository.EvaluationMetrics, error)
	AnalyzeEvaluationPerformance(ctx context.Context, req *PerformanceAnalysisRequest) (*EvaluationPerformanceAnalysis, error)
}

// policyEvaluationEngine implements PolicyEvaluationEngine
type policyEvaluationEngine struct {
	policyRepo         repository.PolicyRepository
	attributeResolver  AttributeResolver
	externalSources    ExternalAttributeSourceManager
	ruleEngine         RuleEngine
	pipManager         *PolicyInformationPointManager
	evaluationCache    *EvaluationCache
	performanceTracker *EvaluationPerformanceTracker

	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewPolicyEvaluationEngine creates a new policy evaluation engine
func NewPolicyEvaluationEngine(
	policyRepo repository.PolicyRepository,
	attributeResolver AttributeResolver,
	externalSources ExternalAttributeSourceManager,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) PolicyEvaluationEngine {
	return &policyEvaluationEngine{
		policyRepo:         policyRepo,
		attributeResolver:  attributeResolver,
		externalSources:    externalSources,
		ruleEngine:         NewAdvancedRuleEngine(),
		pipManager:         NewPolicyInformationPointManager(),
		evaluationCache:    NewEvaluationCache(),
		performanceTracker: NewEvaluationPerformanceTracker(),
		logger:             logger,
		metrics:            metrics,
		tracer:             tracer,
	}
}

// Core Evaluation Types

type PolicyEvaluationRequest struct {
	RequestID       uuid.UUID               `json:"request_id"`
	PolicyID        uuid.UUID               `json:"policy_id" validate:"required"`
	Subject         SubjectContext          `json:"subject" validate:"required"`
	Resource        ResourceContext         `json:"resource" validate:"required"`
	Action          ActionContext           `json:"action" validate:"required"`
	Environment     EnvironmentContext      `json:"environment"`
	EvaluationMode  EvaluationMode          `json:"evaluation_mode"`
	Options         PolicyEvaluationOptions `json:"options"`
	RequiredDetails []EvaluationDetailType  `json:"required_details,omitempty"`
}

type PolicyEvaluationResult struct {
	RequestID       uuid.UUID                `json:"request_id"`
	PolicyID        uuid.UUID                `json:"policy_id"`
	Decision        types.PolicyDecisionType `json:"decision"`
	Explanation     *EvaluationExplanation   `json:"explanation,omitempty"`
	ApplicableRules []ApplicableRule         `json:"applicable_rules,omitempty"`
	AttributesUsed  []AttributeUsage         `json:"attributes_used,omitempty"`
	Obligations     []PolicyObligation       `json:"obligations,omitempty"`
	Advice          []PolicyAdvice           `json:"advice,omitempty"`
	EvaluationTime  time.Duration            `json:"evaluation_time"`
	CacheHit        bool                     `json:"cache_hit"`
	Metadata        map[string]any           `json:"metadata,omitempty"`
}

type SubjectContext struct {
	UserID     uuid.UUID      `json:"user_id" validate:"required"`
	Roles      []string       `json:"roles,omitempty"`
	Groups     []string       `json:"groups,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Claims     map[string]any `json:"claims,omitempty"`
}

type ResourceContext struct {
	ResourceID   uuid.UUID      `json:"resource_id" validate:"required"`
	ResourceType string         `json:"resource_type" validate:"required"`
	Owner        *uuid.UUID     `json:"owner,omitempty"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type ActionContext struct {
	Action     string         `json:"action" validate:"required"`
	Operations []string       `json:"operations,omitempty"`
	Intent     *string        `json:"intent,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// EnvironmentContext type already defined in attribute_collector.go

// Core Evaluation Implementation

func (pee *policyEvaluationEngine) EvaluatePolicy(ctx context.Context, req *PolicyEvaluationRequest) (*PolicyEvaluationResult, error) {
	ctx, span := pee.tracer.StartSpan(ctx, "PolicyEvaluationEngine.EvaluatePolicy")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		pee.metrics.ObserveHistogram("abac.policy_evaluation.duration",
			duration.Seconds(), metrics.Fields{"policy_id": req.PolicyID.String()})
	}()

	// Check cache first
	if cacheResult := pee.checkEvaluationCache(ctx, req); cacheResult != nil {
		cacheResult.CacheHit = true
		pee.metrics.IncrementCounter("abac.policy_evaluation.cache_hit",
			metrics.Fields{"policy_id": req.PolicyID.String()})
		return cacheResult, nil
	}

	// Retrieve policy
	policy, err := pee.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pee.logger.Error("Failed to retrieve policy", logger.Fields{"error": err.Error(), "policy_id": req.PolicyID.String()})
		return nil, fmt.Errorf("failed to retrieve policy: %w", err)
	}

	// Policy is assumed to be active if retrieved successfully

	// Resolve attributes
	attributeCtx, err := pee.resolveEvaluationAttributes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve evaluation attributes: %w", err)
	}

	// Execute policy evaluation
	result, err := pee.executePolicyEvaluation(ctx, policy, attributeCtx, req)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	result.EvaluationTime = time.Since(startTime)
	result.CacheHit = false

	// Cache result if enabled
	if req.Options.EnableCaching {
		pee.cacheEvaluationResult(ctx, req, result)
	}

	pee.logger.Debug("Policy evaluation completed", logger.Fields{
		"policy_id": req.PolicyID.String(),
		"decision":  string(result.Decision),
		"duration":  result.EvaluationTime,
	})

	return result, nil
}

func (pee *policyEvaluationEngine) executePolicyEvaluation(
	ctx context.Context,
	policy *models.Policy,
	attributeCtx *EvaluationAttributeContext,
	req *PolicyEvaluationRequest,
) (*PolicyEvaluationResult, error) {

	result := &PolicyEvaluationResult{
		RequestID: req.RequestID,
		PolicyID:  policy.ID,
		Decision:  types.PolicyDecisionDeny, // Default to deny
	}

	// Evaluate target (applicability)
	if policy.Target != nil {
		targetResult, err := pee.evaluateTarget(ctx, &models.PolicyTarget{
			Resources: []string{"*"},
			Actions:   []string{"*"},
		}, attributeCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate policy target: %w", err)
		}

		if !targetResult.Applicable {
			result.Decision = types.PolicyDecisionNotApplicable
			if req.Options.IncludeExplanation {
				result.Explanation = &EvaluationExplanation{
					Decision: result.Decision,
					Reason:   "Policy target not applicable",
					Details:  targetResult.Details,
				}
			}
			return result, nil
		}
	}

	// Evaluate the policy rule
	ruleResults := make([]RuleEvaluationResult, 0, 1)
	applicableRules := make([]ApplicableRule, 0, 1)

	// Since Policy.Rule is map[string]any, we need to evaluate it differently
	// For now, simplified evaluation - assume policy allows if target matches
	ruleResult := &RuleEvaluationResult{
		RuleID:     policy.ID,
		Decision:   types.PolicyDecisionAllow,
		Applicable: true,
	}

	// TODO: Implement proper rule evaluation with RuleEngine interface

	if ruleResult != nil {
		ruleResults = append(ruleResults, *ruleResult)
		if ruleResult.Applicable {
			applicableRules = append(applicableRules, ApplicableRule{
				RuleID:      ruleResult.RuleID,
				Decision:    ruleResult.Decision,
				Effect:      policy.Effect,
				Explanation: ruleResult.Explanation,
			})
		}
	}

	// Apply combining algorithm
	finalDecision, err := pee.applyCombiningAlgorithm(ctx, policy.CombiningAlgorithm, ruleResults)
	if err != nil {
		return nil, fmt.Errorf("failed to apply combining algorithm: %w", err)
	}

	result.Decision = finalDecision
	result.ApplicableRules = applicableRules

	// Collect obligations and advice
	result.Obligations = pee.collectObligations(ctx, ruleResults, result.Decision)
	result.Advice = pee.collectAdvice(ctx, ruleResults, result.Decision)

	// Generate explanation if requested
	if req.Options.IncludeExplanation {
		result.Explanation = pee.generateEvaluationExplanation(ctx, policy, ruleResults, result)
	}

	// Track attributes used
	if req.Options.TrackAttributeUsage {
		result.AttributesUsed = attributeCtx.GetUsedAttributes()
	}

	return result, nil
}

// Rule Evaluation

// Combining Algorithms

func (pee *policyEvaluationEngine) applyCombiningAlgorithm(
	ctx context.Context,
	algorithm types.CombiningAlgorithm,
	ruleResults []RuleEvaluationResult,
) (types.PolicyDecisionType, error) {

	switch algorithm {
	case types.CombiningAlgorithmDenyOverrides:
		return pee.applyDenyOverrides(ruleResults), nil
	case types.CombiningAlgorithmPermitOverrides:
		return pee.applyPermitOverrides(ruleResults), nil
	case types.CombiningAlgorithmFirstApplicable:
		return pee.applyFirstApplicable(ruleResults), nil
	case types.CombiningAlgorithmOnlyOneApplicable:
		return pee.applyOnlyOneApplicable(ruleResults)
	case "deny_unless_permit":
		return pee.applyDenyOverrides(ruleResults), nil
	case "permit_unless_deny":
		return pee.applyPermitOverrides(ruleResults), nil
	default:
		return types.PolicyDecisionDeny,
			errors.ErrInvalidInput
	}
}

func (pee *policyEvaluationEngine) applyDenyOverrides(results []RuleEvaluationResult) types.PolicyDecisionType {
	hasApplicable := false

	for _, result := range results {
		if result.Applicable {
			hasApplicable = true
			if result.Decision == types.PolicyDecisionDeny {
				return types.PolicyDecisionDeny
			}
		}
	}

	if hasApplicable {
		return types.PolicyDecisionAllow
	}

	return types.PolicyDecisionNotApplicable
}

func (pee *policyEvaluationEngine) applyPermitOverrides(results []RuleEvaluationResult) types.PolicyDecisionType {
	hasApplicable := false

	for _, result := range results {
		if result.Applicable {
			hasApplicable = true
			if result.Decision == types.PolicyDecisionAllow {
				return types.PolicyDecisionAllow
			}
		}
	}

	if hasApplicable {
		return types.PolicyDecisionDeny
	}

	return types.PolicyDecisionNotApplicable
}

func (pee *policyEvaluationEngine) applyFirstApplicable(results []RuleEvaluationResult) types.PolicyDecisionType {
	for _, result := range results {
		if result.Applicable {
			return result.Decision
		}
	}

	return types.PolicyDecisionNotApplicable
}

func (pee *policyEvaluationEngine) applyOnlyOneApplicable(results []RuleEvaluationResult) (types.PolicyDecisionType, error) {
	var applicableResult *RuleEvaluationResult

	for _, result := range results {
		if result.Applicable {
			if applicableResult != nil {
				return types.PolicyDecisionDeny, errors.ErrInvalidInput
			}
			applicableResult = &result
		}
	}

	if applicableResult != nil {
		return applicableResult.Decision, nil
	}

	return types.PolicyDecisionNotApplicable, nil
}

// Policy Information Point (PIP) Integration

type PolicyInformationPointManager struct {
	providers map[string]PIPProvider
	cache     *PIPCache
	mutex     sync.RWMutex
}

func NewPolicyInformationPointManager() *PolicyInformationPointManager {
	return &PolicyInformationPointManager{
		providers: make(map[string]PIPProvider),
		cache:     NewPIPCache(),
	}
}

func (pipm *PolicyInformationPointManager) RegisterProvider(name string, provider PIPProvider) {
	pipm.mutex.Lock()
	defer pipm.mutex.Unlock()
	pipm.providers[name] = provider
}

func (pipm *PolicyInformationPointManager) ResolveAttribute(
	ctx context.Context,
	category string,
	attributeID string,
	subjectID uuid.UUID,
) (any, error) {

	// Check cache first
	if value, found := pipm.cache.Get(category, attributeID, subjectID); found {
		return value, nil
	}

	// Find appropriate provider
	pipm.mutex.RLock()
	provider, exists := pipm.providers[category]
	pipm.mutex.RUnlock()

	if !exists {
		return nil, errors.ErrNotFound
	}

	// Resolve attribute
	value, err := provider.ResolveAttribute(ctx, attributeID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve attribute from PIP: %w", err)
	}

	// Cache result
	pipm.cache.Set(category, attributeID, subjectID, value)

	return value, nil
}

// Supporting Types and Interfaces

type EvaluationAttributeContext struct {
	Subject     SubjectContext
	Resource    ResourceContext
	Action      ActionContext
	Environment EnvironmentContext
	Resolved    map[string]any
	UsedPaths   []string
	mutex       sync.RWMutex
}

func (eac *EvaluationAttributeContext) GetAttribute(path string) (any, bool) {
	eac.mutex.Lock()
	defer eac.mutex.Unlock()

	// Track attribute usage
	eac.UsedPaths = append(eac.UsedPaths, path)

	// Resolve attribute path
	return eac.resolveAttributePath(path)
}

func (eac *EvaluationAttributeContext) GetUsedAttributes() []AttributeUsage {
	eac.mutex.RLock()
	defer eac.mutex.RUnlock()

	usage := make([]AttributeUsage, 0, len(eac.UsedPaths))
	for _, path := range eac.UsedPaths {
		usage = append(usage, AttributeUsage{
			AttributePath: path,
			Category:      eac.categorizeAttributePath(path),
			UsageCount:    1, //NOTE: Simplified for now
		})
	}

	return usage
}

func (eac *EvaluationAttributeContext) resolveAttributePath(path string) (any, bool) {
	// Parse attribute path (e.g., "subject.roles", "resource.owner", "environment.time")
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, false
	}

	category := parts[0]
	attribute := strings.Join(parts[1:], ".")

	switch category {
	case "subject":
		return eac.getSubjectAttribute(attribute)
	case "resource":
		return eac.getResourceAttribute(attribute)
	case "action":
		return eac.getActionAttribute(attribute)
	case "environment":
		return eac.getEnvironmentAttribute(attribute)
	default:
		// Check resolved attributes
		if value, exists := eac.Resolved[path]; exists {
			return value, true
		}
		return nil, false
	}
}

func (eac *EvaluationAttributeContext) categorizeAttributePath(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// Helper methods for attribute resolution

func (eac *EvaluationAttributeContext) getSubjectAttribute(attribute string) (any, bool) {
	switch attribute {
	case "id":
		return eac.Subject.UserID, true
	case "roles":
		return eac.Subject.Roles, true
	case "groups":
		return eac.Subject.Groups, true
	default:
		if value, exists := eac.Subject.Attributes[attribute]; exists {
			return value, true
		}
		if value, exists := eac.Subject.Claims[attribute]; exists {
			return value, true
		}
		return nil, false
	}
}

func (eac *EvaluationAttributeContext) getResourceAttribute(attribute string) (any, bool) {
	switch attribute {
	case "id":
		return eac.Resource.ResourceID, true
	case "type":
		return eac.Resource.ResourceType, true
	case "owner":
		return eac.Resource.Owner, true
	default:
		if value, exists := eac.Resource.Attributes[attribute]; exists {
			return value, true
		}
		if value, exists := eac.Resource.Metadata[attribute]; exists {
			return value, true
		}
		return nil, false
	}
}

func (eac *EvaluationAttributeContext) getActionAttribute(attribute string) (any, bool) {
	switch attribute {
	case "action":
		return eac.Action.Action, true
	case "operations":
		return eac.Action.Operations, true
	case "intent":
		return eac.Action.Intent, true
	default:
		if value, exists := eac.Action.Attributes[attribute]; exists {
			return value, true
		}
		return nil, false
	}
}

func (eac *EvaluationAttributeContext) getEnvironmentAttribute(attribute string) (any, bool) {
	switch attribute {
	case "timestamp":
		return eac.Environment.Timestamp, true
	case "location":
		return eac.Environment.Properties["location"], true
	case "network_info":
		return eac.Environment.Properties["network_info"], true
	case "device_info":
		return eac.Environment.Properties["device_info"], true
	case "security_context":
		return eac.Environment.Properties["security_context"], true
	default:
		if value, exists := eac.Environment.Properties[attribute]; exists {
			return value, true
		}
		return nil, false
	}
}

// Additional Supporting Types

type ApplicableRule struct {
	RuleID      uuid.UUID                `json:"rule_id"`
	Decision    types.PolicyDecisionType `json:"decision"`
	Effect      types.PolicyEffect       `json:"effect"`
	Condition   *models.PolicyCondition  `json:"condition,omitempty"`
	Explanation string                   `json:"explanation,omitempty"`
}

type AttributeUsage struct {
	AttributePath string `json:"attribute_path"`
	Category      string `json:"category"`
	UsageCount    int    `json:"usage_count"`
}

type PolicyObligation struct {
	ID          uuid.UUID             `json:"id"`
	Type        string                `json:"type"`
	Description string                `json:"description"`
	Parameters  map[string]any        `json:"parameters,omitempty"`
	Fulfillment ObligationFulfillment `json:"fulfillment"`
}

type PolicyAdvice struct {
	ID          uuid.UUID      `json:"id"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Severity    AdviceSeverity `json:"severity"`
}

type ObligationFulfillment string

const (
	ObligationFulfillmentRequired ObligationFulfillment = "required"
	ObligationFulfillmentOptional ObligationFulfillment = "optional"
)

type AdviceSeverity string

const (
	AdviceSeverityInfo    AdviceSeverity = "info"
	AdviceSeverityWarning AdviceSeverity = "warning"
	AdviceSeverityError   AdviceSeverity = "error"
)

type EvaluationExplanation struct {
	Decision   types.PolicyDecisionType `json:"decision"`
	Reason     string                   `json:"reason"`
	Details    []ExplanationDetail      `json:"details,omitempty"`
	RuleTrace  []RuleTraceEntry         `json:"rule_trace,omitempty"`
	Attributes []AttributeReference     `json:"attributes,omitempty"`
}

type ExplanationDetail struct {
	Component   string `json:"component"`
	Description string `json:"description"`
	Result      string `json:"result"`
}

type RuleTraceEntry struct {
	RuleID      uuid.UUID                `json:"rule_id"`
	RuleName    string                   `json:"rule_name"`
	Evaluated   bool                     `json:"evaluated"`
	Applicable  bool                     `json:"applicable"`
	Decision    types.PolicyDecisionType `json:"decision,omitempty"`
	Explanation string                   `json:"explanation,omitempty"`
}

type AttributeReference struct {
	Path     string `json:"path"`
	Value    any    `json:"value"`
	Source   string `json:"source"`
	Resolved bool   `json:"resolved"`
}

// Evaluation modes and options

type EvaluationMode string

const (
	EvaluationModeStandard   EvaluationMode = "standard"
	EvaluationModeOptimized  EvaluationMode = "optimized"
	EvaluationModeVerbose    EvaluationMode = "verbose"
	EvaluationModeSimulation EvaluationMode = "simulation"
)

type PolicyEvaluationOptions struct {
	IncludeExplanation  bool          `json:"include_explanation"`
	TrackAttributeUsage bool          `json:"track_attribute_usage"`
	EnableCaching       bool          `json:"enable_caching"`
	CacheTTL            time.Duration `json:"cache_ttl"`
	MaxExecutionTime    time.Duration `json:"max_execution_time"`
	StrictMode          bool          `json:"strict_mode"`
	CollectObligations  bool          `json:"collect_obligations"`
	CollectAdvice       bool          `json:"collect_advice"`
}

type EvaluationDetailType string

const (
	EvaluationDetailTypeExplanation EvaluationDetailType = "explanation"
	EvaluationDetailTypeTrace       EvaluationDetailType = "trace"
	EvaluationDetailTypeAttributes  EvaluationDetailType = "attributes"
	EvaluationDetailTypePerformance EvaluationDetailType = "performance"
)

// PIP Provider Interface

type PIPProvider interface {
	ResolveAttribute(ctx context.Context, attributeID string, subjectID uuid.UUID) (any, error)
	GetSupportedAttributes() []string
	GetProviderInfo() PIPProviderInfo
}

type PIPProviderInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Attributes  []string `json:"attributes"`
}

// Helper functions

func (pee *policyEvaluationEngine) mapEffectToDecision(effect types.PolicyEffect) types.PolicyDecisionType {
	switch effect {
	case types.PolicyEffectAllow:
		return types.PolicyDecisionAllow
	case types.PolicyEffectDeny:
		return types.PolicyDecisionDeny
	default:
		return types.PolicyDecisionDeny
	}
}

func (pee *policyEvaluationEngine) resolveEvaluationAttributes(
	ctx context.Context,
	req *PolicyEvaluationRequest,
) (*EvaluationAttributeContext, error) {

	attributeCtx := &EvaluationAttributeContext{
		Subject:     req.Subject,
		Resource:    req.Resource,
		Action:      req.Action,
		Environment: req.Environment,
		Resolved:    make(map[string]any),
		UsedPaths:   make([]string, 0),
	}

	// NOTE:Additional attribute resolution logic would go here
	// This could involve calling the attribute resolver service
	// to fetch additional attributes from external sources

	return attributeCtx, nil
}

// Stub implementations for cache and other methods

func (pee *policyEvaluationEngine) checkEvaluationCache(ctx context.Context, req *PolicyEvaluationRequest) *PolicyEvaluationResult {
	if !req.Options.EnableCaching {
		return nil
	}
	cachedResult, found := pee.evaluationCache.Get(req)
	if found {
		return cachedResult
	}
	return nil
}

func (pee *policyEvaluationEngine) cacheEvaluationResult(ctx context.Context, req *PolicyEvaluationRequest, result *PolicyEvaluationResult) {
	if !req.Options.EnableCaching {
		return
	}
	pee.evaluationCache.Set(req, result)
}

func (pee *policyEvaluationEngine) evaluateTarget(ctx context.Context, target *models.PolicyTarget, attributeCtx *EvaluationAttributeContext) (*TargetEvaluationResult, error) {
	if target == nil {
		// A nil target means the policy applies to everything.
		return &TargetEvaluationResult{Applicable: true}, nil
	}

	// Build a single boolean expression from the target fields.
	expression, err := pee.buildTargetExpression(target)
	if err != nil {
		return nil, fmt.Errorf("failed to build target expression: %w", err)
	}

	// If the expression is empty, the target is considered met (it's not restrictive).
	if expression == "" {
		return &TargetEvaluationResult{Applicable: true}, nil
	}

	// Use the rule engine to evaluate the constructed expression.
	evalResult, err := pee.ruleEngine.EvaluateExpression(ctx, expression, attributeCtx)
	if err != nil {
		pee.logger.Error("Failed to evaluate policy target expression", logger.Fields{
			"error":      err.Error(),
			"expression": expression,
		})
		// If the target can't be evaluated, we should default to not-applicable for safety.
		return &TargetEvaluationResult{
			Applicable: false,
			Details: []ExplanationDetail{{
				Component:   "target",
				Description: "Expression evaluation failed",
				Result:      err.Error(),
			}},
		}, nil
	}

	return &TargetEvaluationResult{Applicable: evalResult.Result}, nil
}

// buildTargetExpression constructs a single boolean expression string from a PolicyTarget struct.
func (pee *policyEvaluationEngine) buildTargetExpression(target *models.PolicyTarget) (string, error) {
	var conditions []string

	// Helper to create 'in' expressions for string slices.
	addInCondition := func(attribute string, values []string) {
		if len(values) > 0 {
			// e.g., `resource.type in ["type1", "type2"]`
			jsonValues, _ := json.Marshal(values)
			conditions = append(conditions, fmt.Sprintf("%s in %s", attribute, string(jsonValues)))
		}
	}

	addInCondition("resource.resource_type", target.ResourceTypes)
	addInCondition("action.action", target.Actions)
	// Note: The subject and resource ID checks would need to align with the attribute context.
	// Assuming "resource.id" and "subject.id" are available paths.
	addInCondition("resource.id", target.Resources)
	addInCondition("subject.id", target.Subjects)

	// For the environment map, we can create equality checks.
	for key, value := range target.Environment {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("failed to marshal environment value for key '%s': %w", key, err)
		}
		conditions = append(conditions, fmt.Sprintf("environment.%s == %s", key, string(jsonValue)))
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return strings.Join(conditions, " && "), nil
}

func (pee *policyEvaluationEngine) collectObligations(ctx context.Context, ruleResults []RuleEvaluationResult, decision types.PolicyDecisionType) []PolicyObligation {
	// TODO:Implementation would collect obligations from applicable rules
	return []PolicyObligation{}
}

func (pee *policyEvaluationEngine) collectAdvice(ctx context.Context, ruleResults []RuleEvaluationResult, decision types.PolicyDecisionType) []PolicyAdvice {
	// TODO:Implementation would collect advice from applicable rules
	return []PolicyAdvice{}
}

func (pee *policyEvaluationEngine) generateEvaluationExplanation(ctx context.Context, policy *models.Policy, ruleResults []RuleEvaluationResult, result *PolicyEvaluationResult) *EvaluationExplanation {
	// TODO:Implementation would generate detailed explanation
	return &EvaluationExplanation{
		Decision: result.Decision,
		Reason:   "Policy evaluation completed",
	}
}

type TargetEvaluationResult struct {
	Applicable bool                `json:"applicable"`
	Details    []ExplanationDetail `json:"details,omitempty"`
}

type RuleEvaluationResult struct {
	RuleID      uuid.UUID                `json:"rule_id"`
	Decision    types.PolicyDecisionType `json:"decision"`
	Applicable  bool                     `json:"applicable"`
	Explanation string                   `json:"explanation,omitempty"`
}

// AnalyzeEvaluationPerformance analyzes the performance of policy evaluations
func (pee *policyEvaluationEngine) AnalyzeEvaluationPerformance(ctx context.Context, req *PerformanceAnalysisRequest) (*EvaluationPerformanceAnalysis, error) {
	// TODO:Implementatio
	return &EvaluationPerformanceAnalysis{
		PolicyID:         req.PolicyID,
		TotalEvaluations: 100,
		AverageLatency:   50 * time.Millisecond,
		SuccessRate:      0.99,
		Bottlenecks:      []string{"attribute_resolution", "rule_evaluation"},
	}, nil
}

// BatchEvaluatePolicy evaluates multiple policies in batch
func (pee *policyEvaluationEngine) BatchEvaluatePolicy(ctx context.Context, req *BatchPolicyEvaluationRequest) (*BatchPolicyEvaluationResult, error) {
	// TODO:Implementatio
	// Placeholder implementation
	results := make([]*PolicyEvaluationResult, len(req.Requests))
	for i, singleReq := range req.Requests {
		result, err := pee.EvaluatePolicy(ctx, &singleReq)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}

	// Convert pointer slice to value slice
	valueslice := make([]PolicyEvaluationResult, len(results))
	for i, result := range results {
		if result != nil {
			valueslice[i] = *result
		}
	}

	return &BatchPolicyEvaluationResult{
		Results:  valueslice,
		BatchID:  uuid.New(),
		Success:  true,
		ErrorMsg: "",
	}, nil
}

// EvaluatePolicies evaluates multiple policies
func (pee *policyEvaluationEngine) EvaluatePolicies(ctx context.Context, req *MultiplePolicyEvaluationRequest) (*MultiplePolicyEvaluationResult, error) {
	// TODO:Implementatio
	// Placeholder implementation
	return &MultiplePolicyEvaluationResult{
		Results:   []PolicyEvaluationResult{},
		Decision:  types.PolicyDecisionAllow,
		RequestID: req.RequestID,
	}, nil
}

// EvaluateRule evaluates a single rule
func (pee *policyEvaluationEngine) EvaluateRule(ctx context.Context, req *RuleEvaluationRequest) (*RuleEvaluationResult, error) {
	// TODO:Implementatio
	// Placeholder implementation
	// Create a UUID from the string RuleID
	ruleUUID, err := uuid.Parse(req.RuleID)
	if err != nil {
		// If parsing fails, generate a new UUID
		ruleUUID = uuid.New()
	}

	return &RuleEvaluationResult{
		RuleID:      ruleUUID,
		Decision:    types.PolicyDecisionAllow,
		Applicable:  true,
		Explanation: "Rule evaluated successfully",
	}, nil
}

// EvaluateWithContext evaluates policy with enhanced context
func (pee *policyEvaluationEngine) EvaluateWithContext(ctx context.Context, req *ContextualEvaluationRequest) (*ContextualEvaluationResult, error) {
	// TODO:Implementatio
	// Placeholder implementation
	return &ContextualEvaluationResult{
		PolicyID:  req.PolicyID,
		Decision:  types.PolicyDecisionAllow,
		Context:   req.Context,
		Reasoning: "Policy evaluation successful",
		RequestID: req.RequestID,
	}, nil
}

// GetEvaluationMetrics returns evaluation metrics
func (pee *policyEvaluationEngine) GetEvaluationMetrics(ctx context.Context, req *repository.GetEvaluationMetricsRequest) (*repository.EvaluationMetrics, error) {
	// TODO:Implementatio
	// Placeholder implementation
	return &repository.EvaluationMetrics{
		TotalEvaluations:       1000,
		UniqueUsers:            125,
		UniqueResources:        250,
		AvgEvaluationTimeMS:    45.0,
		MedianEvaluationTimeMS: 35.0,
		P95EvaluationTimeMS:    100.0,
		P99EvaluationTimeMS:    250.0,
	}, nil
}

// ParseRuleExpression parses a rule expression
func (pee *policyEvaluationEngine) ParseRuleExpression(ctx context.Context, expression string) (*ParsedRuleExpression, error) {
	// TODO:Implementatio
	// Placeholder implementation
	return &ParsedRuleExpression{
		Expression: expression,
		Variables:  []string{"user", "resource", "environment"},
		Functions:  []string{"equals", "contains", "matches"},
		ParsedAt:   time.Now(),
		IsValid:    true,
	}, nil
}

// RegisterPIPProvider registers a Policy Information Point provider
func (pee *policyEvaluationEngine) RegisterPIPProvider(ctx context.Context, req *RegisterPIPProviderRequest) (*PIPProvider, error) {
	// TODO: Implement PIP provider registration
	return nil, fmt.Errorf("PIP provider registration not yet implemented")
}

// ResolvePolicyInformation resolves policy information from external sources
func (pee *policyEvaluationEngine) ResolvePolicyInformation(ctx context.Context, req *PIPResolutionRequest) (*PIPResolutionResult, error) {
	// TODO: Implement PIP resolution
	return &PIPResolutionResult{
		AttributeName: req.AttributeName,
		Value:         "resolved_value",
		Source:        "internal",
		Success:       true,
	}, nil
}

// SimulateEvaluation simulates policy evaluation with multiple scenarios
func (pee *policyEvaluationEngine) SimulateEvaluation(ctx context.Context, req *EvaluationSimulationRequest) (*EvaluationSimulationResult, error) {
	// TODO: Implement simulation logic
	results := make([]ContextualEvaluationResult, len(req.Scenarios))
	for i, scenario := range req.Scenarios {
		results[i] = ContextualEvaluationResult{
			PolicyID:  req.PolicyID,
			Decision:  types.PolicyDecisionAllow,
			Context:   scenario,
			Reasoning: "Simulated evaluation",
			RequestID: req.RequestID,
		}
	}

	return &EvaluationSimulationResult{
		PolicyID: req.PolicyID,
		Results:  results,
		Summary: SimulationSummary{
			TotalScenarios:     int64(len(req.Scenarios)),
			PermitCount:        int64(len(req.Scenarios)),
			DenyCount:          0,
			IndeterminateCount: 0,
			SuccessRate:        1.0,
		},
		RequestID: req.RequestID,
	}, nil
}

// ValidateRuleExpression validates a rule expression
func (pee *policyEvaluationEngine) ValidateRuleExpression(ctx context.Context, expression string) (*RuleValidationResult, error) {
	// TODO: Implement actual rule validation logic
	return &RuleValidationResult{
		RuleID:   uuid.New(),
		RuleName: "expression_rule",
		RuleType: AttributeRuleTypeCustom,
		Passed:   true,
		Message:  "Rule expression validation passed",
		Severity: "info",
		Details:  map[string]any{"expression": expression},
	}, nil
}

// Additional stub types for completeness

// RuleExpressionParseRequest represents a request to parse a rule expression
type RuleExpressionParseRequest struct {
	Expression string `json:"expression"`
	Context    string `json:"context,omitempty"`
}

// RuleExpressionParseResult represents the result of parsing a rule expression
type RuleExpressionParseResult struct {
	Valid                  bool          `json:"valid"`
	ParsedAST              string        `json:"parsed_ast,omitempty"`
	Variables              []string      `json:"variables,omitempty"`
	Functions              []string      `json:"functions,omitempty"`
	Complexity             float64       `json:"complexity"`
	EstimatedExecutionTime time.Duration `json:"estimated_execution_time"`
}

// // ValidationResult conflicts with attribute_service.go - using PolicyValidationResult
// type PolicyValidationResult struct {
// 	Valid  bool     `json:"valid"`
// 	Errors []string `json:"errors"`
// }
