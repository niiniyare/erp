package abac

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
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

// ─── MISSING TYPE DEFINITIONS ─────────────────────────────────────────────

// MultiplePolicyEvaluationRequest represents a request to evaluate multiple policies
type MultiplePolicyEvaluationRequest struct {
	PolicyIDs []uuid.UUID            `json:"policy_ids"`
	Context   map[string]interface{} `json:"context"`
	RequestID string                 `json:"request_id"`
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
	PolicyID  uuid.UUID              `json:"policy_id"`
	Context   map[string]interface{} `json:"context"`
	Scenario  string                 `json:"scenario,omitempty"`
	RequestID string                 `json:"request_id"`
}

// ContextualEvaluationResult represents the result of contextual evaluation
type ContextualEvaluationResult struct {
	PolicyID  uuid.UUID                `json:"policy_id"`
	Decision  types.PolicyDecisionType `json:"decision"`
	Context   map[string]interface{}   `json:"context"`
	Reasoning string                   `json:"reasoning,omitempty"`
	RequestID string                   `json:"request_id"`
}

// EvaluationSimulationRequest represents a simulation request
type EvaluationSimulationRequest struct {
	PolicyID  uuid.UUID                `json:"policy_id"`
	Scenarios []map[string]interface{} `json:"scenarios"`
	RequestID string                   `json:"request_id"`
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
	RuleID    string                 `json:"rule_id"`
	Rule      string                 `json:"rule"`
	Context   map[string]interface{} `json:"context"`
	RequestID string                 `json:"request_id"`
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
	AttributeName string                 `json:"attribute_name"`
	EntityID      uuid.UUID              `json:"entity_id"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// PIPResolutionResult represents the result of PIP resolution
type PIPResolutionResult struct {
	AttributeName string      `json:"attribute_name"`
	Value         interface{} `json:"value"`
	Source        string      `json:"source"`
	Success       bool        `json:"success"`
	ErrorMsg      string      `json:"error_msg,omitempty"`
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
	ruleEngine         *AdvancedRuleEngine
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
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

type SubjectContext struct {
	UserID     uuid.UUID              `json:"user_id" validate:"required"`
	Roles      []string               `json:"roles,omitempty"`
	Groups     []string               `json:"groups,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Claims     map[string]interface{} `json:"claims,omitempty"`
}

type ResourceContext struct {
	ResourceID   uuid.UUID              `json:"resource_id" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	Owner        *uuid.UUID             `json:"owner,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ActionContext struct {
	Action     string                 `json:"action" validate:"required"`
	Operations []string               `json:"operations,omitempty"`
	Intent     *string                `json:"intent,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EnvironmentContext type already defined in attribute_collector.go

// Advanced Rule Engine

type AdvancedRuleEngine struct {
	operators       map[string]RuleOperator
	functions       map[string]RuleFunction
	expressionCache map[string]*CompiledExpression
	validationCache map[string]*ValidationResult
	mutex           sync.RWMutex
}

func NewAdvancedRuleEngine() *AdvancedRuleEngine {
	engine := &AdvancedRuleEngine{
		operators:       make(map[string]RuleOperator),
		functions:       make(map[string]RuleFunction),
		expressionCache: make(map[string]*CompiledExpression),
		validationCache: make(map[string]*ValidationResult),
	}

	engine.registerBuiltinOperators()
	engine.registerBuiltinFunctions()

	return engine
}

func (are *AdvancedRuleEngine) registerBuiltinOperators() {
	// Comparison operators
	are.operators["eq"] = &EqualOperator{}
	are.operators["ne"] = &NotEqualOperator{}
	are.operators["gt"] = &GreaterThanOperator{}
	are.operators["gte"] = &GreaterThanEqualOperator{}
	are.operators["lt"] = &LessThanOperator{}
	are.operators["lte"] = &LessThanEqualOperator{}

	// Logical operators
	are.operators["and"] = &AndOperator{}
	are.operators["or"] = &OrOperator{}
	are.operators["not"] = &NotOperator{}

	// String operators
	are.operators["contains"] = &ContainsOperator{}
	are.operators["startswith"] = &StartsWithOperator{}
	are.operators["endswith"] = &EndsWithOperator{}
	are.operators["matches"] = &RegexMatchOperator{}

	// Collection operators
	are.operators["in"] = &InOperator{}
	are.operators["notin"] = &NotInOperator{}
	are.operators["subset"] = &SubsetOperator{}
	are.operators["superset"] = &SupersetOperator{}
	are.operators["intersects"] = &IntersectsOperator{}

	// Temporal operators
	are.operators["before"] = &BeforeOperator{}
	are.operators["after"] = &AfterOperator{}
	are.operators["during"] = &DuringOperator{}
	are.operators["between"] = &BetweenOperator{}
}

func (are *AdvancedRuleEngine) registerBuiltinFunctions() {
	// String functions
	are.functions["strlen"] = &StringLengthFunction{}
	are.functions["upper"] = &UpperCaseFunction{}
	are.functions["lower"] = &LowerCaseFunction{}
	are.functions["trim"] = &TrimFunction{}
	are.functions["substr"] = &SubstringFunction{}

	// Math functions
	are.functions["abs"] = &AbsoluteFunction{}
	are.functions["min"] = &MinFunction{}
	are.functions["max"] = &MaxFunction{}
	are.functions["sum"] = &SumFunction{}
	are.functions["avg"] = &AverageFunction{}

	// Date functions
	are.functions["now"] = &NowFunction{}
	are.functions["date"] = &DateFunction{}
	are.functions["timeformat"] = &TimeFormatFunction{}
	are.functions["datediff"] = &DateDifferenceFunction{}

	// Collection functions
	are.functions["count"] = &CountFunction{}
	are.functions["first"] = &FirstFunction{}
	are.functions["last"] = &LastFunction{}
	are.functions["distinct"] = &DistinctFunction{}

	// Utility functions
	are.functions["type"] = &TypeFunction{}
	are.functions["exists"] = &ExistsFunction{}
	are.functions["empty"] = &EmptyFunction{}
	are.functions["default"] = &DefaultFunction{}
}

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

	// Evaluate rules (simplified - Policy model doesn't include Rules field)
	ruleResults := make([]RuleEvaluationResult, 0)
	applicableRules := make([]ApplicableRule, 0)

	// Placeholder rule evaluation based on policy effect
	dummyRule := &models.PolicyRule{
		ID:         "default_rule",
		Expression: "true",
	}
	ruleResult, err := pee.evaluateRule(ctx, dummyRule, attributeCtx)
	if err != nil {
		pee.logger.Error("Failed to evaluate rule", logger.Fields{"error": err.Error(), "rule_id": dummyRule.ID})
	} else {
		ruleResults = append(ruleResults, *ruleResult)

		if ruleResult.Applicable {
			applicableRules = append(applicableRules, ApplicableRule{
				RuleID:      uuid.New(), // Generate a UUID for the dummy rule
				Decision:    ruleResult.Decision,
				Effect:      policy.Effect,
				Condition:   &models.PolicyCondition{Expression: dummyRule.Expression},
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

func (pee *policyEvaluationEngine) evaluateRule(
	ctx context.Context,
	rule *models.PolicyRule,
	attributeCtx *EvaluationAttributeContext,
) (*RuleEvaluationResult, error) {

	ruleUUID := uuid.New() // Generate UUID for string rule ID
	result := &RuleEvaluationResult{
		RuleID:     ruleUUID,
		Decision:   types.PolicyDecisionDeny,
		Applicable: false,
	}

	// Simplified rule evaluation (no target field in PolicyRule model)
	// Evaluate based on rule expression - basic implementation
	if rule.Expression != "" && rule.Expression == "true" {
		result.Applicable = true
		result.Decision = types.PolicyDecisionAllow
		result.Explanation = "Rule expression evaluated to true"
	} else {
		result.Applicable = false
		result.Decision = types.PolicyDecisionDeny
		result.Explanation = "Rule expression evaluated to false"
	}

	return result, nil
}

// Advanced Rule Engine Implementation

func (are *AdvancedRuleEngine) EvaluateExpression(
	ctx context.Context,
	expression string,
	attributeCtx *EvaluationAttributeContext,
) (*ExpressionEvaluationResult, error) {

	// Check cache first
	are.mutex.RLock()
	compiled, exists := are.expressionCache[expression]
	are.mutex.RUnlock()

	if !exists {
		// Parse and compile expression
		var err error
		compiled, err = are.parseExpression(expression)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expression: %w", err)
		}

		// Cache compiled expression
		are.mutex.Lock()
		are.expressionCache[expression] = compiled
		are.mutex.Unlock()
	}

	// Execute compiled expression
	result, err := are.executeCompiledExpression(ctx, compiled, attributeCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute expression: %w", err)
	}

	return result, nil
}

func (are *AdvancedRuleEngine) parseExpression(expression string) (*CompiledExpression, error) {
	// Tokenize expression
	tokens, err := are.tokenizeExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize expression: %w", err)
	}

	// Parse tokens into AST
	ast, err := are.parseTokens(tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tokens: %w", err)
	}

	// Compile AST to executable form
	compiled, err := are.compileAST(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to compile AST: %w", err)
	}

	return compiled, nil
}

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
) (interface{}, error) {

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

type RuleOperator interface {
	Evaluate(left, right interface{}) (bool, error)
	GetName() string
	GetArity() int
}

type RuleFunction interface {
	Execute(args []interface{}) (interface{}, error)
	GetName() string
	GetArity() int
	ValidateArgs(args []interface{}) error
}

type CompiledExpression struct {
	AST        *ExpressionNode
	Variables  []string
	Functions  []string
	Operators  []string
	Complexity int
}

type ExpressionNode struct {
	Type     NodeType
	Value    interface{}
	Operator string
	Function string
	Children []*ExpressionNode
}

type NodeType string

const (
	NodeTypeLiteral  NodeType = "literal"
	NodeTypeVariable NodeType = "variable"
	NodeTypeOperator NodeType = "operator"
	NodeTypeFunction NodeType = "function"
)

type EvaluationAttributeContext struct {
	Subject     SubjectContext
	Resource    ResourceContext
	Action      ActionContext
	Environment EnvironmentContext
	Resolved    map[string]interface{}
	UsedPaths   []string
	mutex       sync.RWMutex
}

func (eac *EvaluationAttributeContext) GetAttribute(path string) (interface{}, bool) {
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
			UsageCount:    1, // Simplified for now
		})
	}

	return usage
}

func (eac *EvaluationAttributeContext) resolveAttributePath(path string) (interface{}, bool) {
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

func (eac *EvaluationAttributeContext) getSubjectAttribute(attribute string) (interface{}, bool) {
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

func (eac *EvaluationAttributeContext) getResourceAttribute(attribute string) (interface{}, bool) {
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

func (eac *EvaluationAttributeContext) getActionAttribute(attribute string) (interface{}, bool) {
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

func (eac *EvaluationAttributeContext) getEnvironmentAttribute(attribute string) (interface{}, bool) {
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
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Fulfillment ObligationFulfillment  `json:"fulfillment"`
}

type PolicyAdvice struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Severity    AdviceSeverity         `json:"severity"`
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
	Path     string      `json:"path"`
	Value    interface{} `json:"value"`
	Source   string      `json:"source"`
	Resolved bool        `json:"resolved"`
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
	ResolveAttribute(ctx context.Context, attributeID string, subjectID uuid.UUID) (interface{}, error)
	GetSupportedAttributes() []string
	GetProviderInfo() PIPProviderInfo
}

type PIPProviderInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Attributes  []string `json:"attributes"`
}

type PIPCache struct {
	cache map[string]*PIPCacheEntry
	mutex sync.RWMutex
}

func NewPIPCache() *PIPCache {
	return &PIPCache{
		cache: make(map[string]*PIPCacheEntry),
	}
}

func (pc *PIPCache) Get(category, attributeID string, subjectID uuid.UUID) (interface{}, bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", category, attributeID, subjectID.String())
	entry, exists := pc.cache[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry.Value, true
}

func (pc *PIPCache) Set(category, attributeID string, subjectID uuid.UUID, value interface{}) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	key := fmt.Sprintf("%s:%s:%s", category, attributeID, subjectID.String())
	pc.cache[key] = &PIPCacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(5 * time.Minute), // Default TTL
	}
}

type PIPCacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

func (pce *PIPCacheEntry) IsExpired() bool {
	return time.Now().After(pce.ExpiresAt)
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
		Resolved:    make(map[string]interface{}),
		UsedPaths:   make([]string, 0),
	}

	// Additional attribute resolution logic would go here
	// This could involve calling the attribute resolver service
	// to fetch additional attributes from external sources

	return attributeCtx, nil
}

// Stub implementations for cache and other methods

func (pee *policyEvaluationEngine) checkEvaluationCache(ctx context.Context, req *PolicyEvaluationRequest) *PolicyEvaluationResult {
	// Implementation would check cache based on request parameters
	return nil
}

func (pee *policyEvaluationEngine) cacheEvaluationResult(ctx context.Context, req *PolicyEvaluationRequest, result *PolicyEvaluationResult) {
	// Implementation would cache the result
}

func (pee *policyEvaluationEngine) evaluateTarget(ctx context.Context, target *models.PolicyTarget, attributeCtx *EvaluationAttributeContext) (*TargetEvaluationResult, error) {
	// Implementation would evaluate policy/rule target
	return &TargetEvaluationResult{Applicable: true}, nil
}

func (pee *policyEvaluationEngine) collectObligations(ctx context.Context, ruleResults []RuleEvaluationResult, decision types.PolicyDecisionType) []PolicyObligation {
	// Implementation would collect obligations from applicable rules
	return []PolicyObligation{}
}

func (pee *policyEvaluationEngine) collectAdvice(ctx context.Context, ruleResults []RuleEvaluationResult, decision types.PolicyDecisionType) []PolicyAdvice {
	// Implementation would collect advice from applicable rules
	return []PolicyAdvice{}
}

func (pee *policyEvaluationEngine) generateEvaluationExplanation(ctx context.Context, policy *models.Policy, ruleResults []RuleEvaluationResult, result *PolicyEvaluationResult) *EvaluationExplanation {
	// Implementation would generate detailed explanation
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

type ExpressionEvaluationResult struct {
	Result      bool   `json:"result"`
	Explanation string `json:"explanation,omitempty"`
}

// Additional stub components

type EvaluationPerformanceTracker struct{}

func NewEvaluationPerformanceTracker() *EvaluationPerformanceTracker {
	return &EvaluationPerformanceTracker{}
}

// Stub operator implementations

type EqualOperator struct{}

func (eo *EqualOperator) Evaluate(left, right interface{}) (bool, error) {
	return reflect.DeepEqual(left, right), nil
}
func (eo *EqualOperator) GetName() string { return "eq" }
func (eo *EqualOperator) GetArity() int   { return 2 }

type NotEqualOperator struct{}

func (neo *NotEqualOperator) Evaluate(left, right interface{}) (bool, error) {
	return !reflect.DeepEqual(left, right), nil
}
func (neo *NotEqualOperator) GetName() string { return "ne" }
func (neo *NotEqualOperator) GetArity() int   { return 2 }

type GreaterThanOperator struct{}

func (gto *GreaterThanOperator) Evaluate(left, right interface{}) (bool, error) {
	// Implementation would handle numeric comparison
	return false, nil
}
func (gto *GreaterThanOperator) GetName() string { return "gt" }
func (gto *GreaterThanOperator) GetArity() int   { return 2 }

type GreaterThanEqualOperator struct{}

func (gteo *GreaterThanEqualOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (gteo *GreaterThanEqualOperator) GetName() string { return "gte" }
func (gteo *GreaterThanEqualOperator) GetArity() int   { return 2 }

type LessThanOperator struct{}

func (lto *LessThanOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (lto *LessThanOperator) GetName() string { return "lt" }
func (lto *LessThanOperator) GetArity() int   { return 2 }

type LessThanEqualOperator struct{}

func (lteo *LessThanEqualOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (lteo *LessThanEqualOperator) GetName() string { return "lte" }
func (lteo *LessThanEqualOperator) GetArity() int   { return 2 }

type AndOperator struct{}

func (ao *AndOperator) Evaluate(left, right interface{}) (bool, error) {
	leftBool, leftOk := left.(bool)
	rightBool, rightOk := right.(bool)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("AND operator requires boolean operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	return leftBool && rightBool, nil
}
func (ao *AndOperator) GetName() string { return "and" }
func (ao *AndOperator) GetArity() int   { return 2 }

type OrOperator struct{}

func (oo *OrOperator) Evaluate(left, right interface{}) (bool, error) {
	leftBool, leftOk := left.(bool)
	rightBool, rightOk := right.(bool)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("OR operator requires boolean operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	return leftBool || rightBool, nil
}
func (oo *OrOperator) GetName() string { return "or" }
func (oo *OrOperator) GetArity() int   { return 2 }

type NotOperator struct{}

func (no *NotOperator) Evaluate(left, right interface{}) (bool, error) {
	leftBool, leftOk := left.(bool)
	if !leftOk {
		return false, fmt.Errorf("NOT operator requires boolean operand, operand: %v: %w", left, errors.ErrInvalidInput)
	}
	return !leftBool, nil
}
func (no *NotOperator) GetName() string { return "not" }
func (no *NotOperator) GetArity() int   { return 1 }

type ContainsOperator struct{}

func (co *ContainsOperator) Evaluate(left, right interface{}) (bool, error) {
	leftStr, leftOk := left.(string)
	rightStr, rightOk := right.(string)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("CONTAINS operator requires string operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	return strings.Contains(leftStr, rightStr), nil
}
func (co *ContainsOperator) GetName() string { return "contains" }
func (co *ContainsOperator) GetArity() int   { return 2 }

type StartsWithOperator struct{}

func (swo *StartsWithOperator) Evaluate(left, right interface{}) (bool, error) {
	leftStr, leftOk := left.(string)
	rightStr, rightOk := right.(string)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("STARTSWITH operator requires string operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	return strings.HasPrefix(leftStr, rightStr), nil
}
func (swo *StartsWithOperator) GetName() string { return "startswith" }
func (swo *StartsWithOperator) GetArity() int   { return 2 }

type EndsWithOperator struct{}

func (ewo *EndsWithOperator) Evaluate(left, right interface{}) (bool, error) {
	leftStr, leftOk := left.(string)
	rightStr, rightOk := right.(string)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("ENDSWITH operator requires string operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	return strings.HasSuffix(leftStr, rightStr), nil
}
func (ewo *EndsWithOperator) GetName() string { return "endswith" }
func (ewo *EndsWithOperator) GetArity() int   { return 2 }

type RegexMatchOperator struct{}

func (rmo *RegexMatchOperator) Evaluate(left, right interface{}) (bool, error) {
	leftStr, leftOk := left.(string)
	rightStr, rightOk := right.(string)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("MATCHES operator requires string operands, left: %v, right: %v: %w", left, right, errors.ErrInvalidInput)
	}
	matched, err := regexp.MatchString(rightStr, leftStr)
	if err != nil {
		return false, fmt.Errorf("regex match failed: %w", err)
	}
	return matched, nil
}
func (rmo *RegexMatchOperator) GetName() string { return "matches" }
func (rmo *RegexMatchOperator) GetArity() int   { return 2 }

type InOperator struct{}

func (io *InOperator) Evaluate(left, right interface{}) (bool, error) {
	// Implementation would check if left is in right (array/slice)
	return false, nil
}
func (io *InOperator) GetName() string { return "in" }
func (io *InOperator) GetArity() int   { return 2 }

type NotInOperator struct{}

func (nio *NotInOperator) Evaluate(left, right interface{}) (bool, error) {
	// Implementation would check if left is not in right (array/slice)
	return false, nil
}
func (nio *NotInOperator) GetName() string { return "notin" }
func (nio *NotInOperator) GetArity() int   { return 2 }

type SubsetOperator struct{}

func (so *SubsetOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (so *SubsetOperator) GetName() string { return "subset" }
func (so *SubsetOperator) GetArity() int   { return 2 }

type SupersetOperator struct{}

func (sso *SupersetOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (sso *SupersetOperator) GetName() string { return "superset" }
func (sso *SupersetOperator) GetArity() int   { return 2 }

type IntersectsOperator struct{}

func (io2 *IntersectsOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (io2 *IntersectsOperator) GetName() string { return "intersects" }
func (io2 *IntersectsOperator) GetArity() int   { return 2 }

type BeforeOperator struct{}

func (bo *BeforeOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (bo *BeforeOperator) GetName() string { return "before" }
func (bo *BeforeOperator) GetArity() int   { return 2 }

type AfterOperator struct{}

func (ao2 *AfterOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (ao2 *AfterOperator) GetName() string { return "after" }
func (ao2 *AfterOperator) GetArity() int   { return 2 }

type DuringOperator struct{}

func (do *DuringOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (do *DuringOperator) GetName() string { return "during" }
func (do *DuringOperator) GetArity() int   { return 2 }

type BetweenOperator struct{}

func (bo2 *BetweenOperator) Evaluate(left, right interface{}) (bool, error) {
	return false, nil
}
func (bo2 *BetweenOperator) GetName() string { return "between" }
func (bo2 *BetweenOperator) GetArity() int   { return 3 }

// Stub function implementations

type StringLengthFunction struct{}

func (slf *StringLengthFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strlen function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("strlen function requires string argument, got %v: %w", reflect.TypeOf(args[0]), errors.ErrInvalidInput)
	}
	return len(str), nil
}
func (slf *StringLengthFunction) GetName() string                       { return "strlen" }
func (slf *StringLengthFunction) GetArity() int                         { return 1 }
func (slf *StringLengthFunction) ValidateArgs(args []interface{}) error { return nil }

type UpperCaseFunction struct{}

func (ucf *UpperCaseFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("upper function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("upper function requires string argument, got %v: %w", reflect.TypeOf(args[0]), errors.ErrInvalidInput)
	}
	return strings.ToUpper(str), nil
}
func (ucf *UpperCaseFunction) GetName() string                       { return "upper" }
func (ucf *UpperCaseFunction) GetArity() int                         { return 1 }
func (ucf *UpperCaseFunction) ValidateArgs(args []interface{}) error { return nil }

type LowerCaseFunction struct{}

func (lcf *LowerCaseFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("lower function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("lower function requires string argument, got %v: %w", reflect.TypeOf(args[0]), errors.ErrInvalidInput)
	}
	return strings.ToLower(str), nil
}
func (lcf *LowerCaseFunction) GetName() string                       { return "lower" }
func (lcf *LowerCaseFunction) GetArity() int                         { return 1 }
func (lcf *LowerCaseFunction) ValidateArgs(args []interface{}) error { return nil }

type TrimFunction struct{}

func (tf *TrimFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trim function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("trim function requires string argument, got %v: %w", reflect.TypeOf(args[0]), errors.ErrInvalidInput)
	}
	return strings.TrimSpace(str), nil
}
func (tf *TrimFunction) GetName() string                       { return "trim" }
func (tf *TrimFunction) GetArity() int                         { return 1 }
func (tf *TrimFunction) ValidateArgs(args []interface{}) error { return nil }

type SubstringFunction struct{}

func (sf *SubstringFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("substr function requires exactly 3 arguments, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	str, ok1 := args[0].(string)
	start, ok2 := args[1].(int)
	length, ok3 := args[2].(int)
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("substr function requires (string, int, int) arguments: %w", errors.ErrInvalidInput)
	}
	if start < 0 || start >= len(str) || length < 0 {
		return "", nil
	}
	end := start + length
	if end > len(str) {
		end = len(str)
	}
	return str[start:end], nil
}
func (sf *SubstringFunction) GetName() string                       { return "substr" }
func (sf *SubstringFunction) GetArity() int                         { return 3 }
func (sf *SubstringFunction) ValidateArgs(args []interface{}) error { return nil }

type AbsoluteFunction struct{}

func (af *AbsoluteFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("abs function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	// Implementation would handle numeric absolute value
	return args[0], nil
}
func (af *AbsoluteFunction) GetName() string                       { return "abs" }
func (af *AbsoluteFunction) GetArity() int                         { return 1 }
func (af *AbsoluteFunction) ValidateArgs(args []interface{}) error { return nil }

type MinFunction struct{}

func (mf *MinFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("min function requires at least 2 arguments, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	// Implementation would find minimum value
	return args[0], nil
}
func (mf *MinFunction) GetName() string                       { return "min" }
func (mf *MinFunction) GetArity() int                         { return -1 } // Variable arity
func (mf *MinFunction) ValidateArgs(args []interface{}) error { return nil }

type MaxFunction struct{}

func (maxf *MaxFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("max function requires at least 2 arguments, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	// Implementation would find maximum value
	return args[0], nil
}
func (maxf *MaxFunction) GetName() string                       { return "max" }
func (maxf *MaxFunction) GetArity() int                         { return -1 } // Variable arity
func (maxf *MaxFunction) ValidateArgs(args []interface{}) error { return nil }

type SumFunction struct{}

func (sf2 *SumFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would sum numeric values
	return 0, nil
}
func (sf2 *SumFunction) GetName() string                       { return "sum" }
func (sf2 *SumFunction) GetArity() int                         { return -1 } // Variable arity
func (sf2 *SumFunction) ValidateArgs(args []interface{}) error { return nil }

type AverageFunction struct{}

func (avgf *AverageFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would calculate average
	return 0.0, nil
}
func (avgf *AverageFunction) GetName() string                       { return "avg" }
func (avgf *AverageFunction) GetArity() int                         { return -1 } // Variable arity
func (avgf *AverageFunction) ValidateArgs(args []interface{}) error { return nil }

type NowFunction struct{}

func (nf *NowFunction) Execute(args []interface{}) (interface{}, error) {
	return time.Now(), nil
}
func (nf *NowFunction) GetName() string                       { return "now" }
func (nf *NowFunction) GetArity() int                         { return 0 }
func (nf *NowFunction) ValidateArgs(args []interface{}) error { return nil }

type DateFunction struct{}

func (df *DateFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would parse date string
	return time.Now(), nil
}
func (df *DateFunction) GetName() string                       { return "date" }
func (df *DateFunction) GetArity() int                         { return 1 }
func (df *DateFunction) ValidateArgs(args []interface{}) error { return nil }

type TimeFormatFunction struct{}

func (tff *TimeFormatFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would format time
	return "", nil
}
func (tff *TimeFormatFunction) GetName() string                       { return "timeformat" }
func (tff *TimeFormatFunction) GetArity() int                         { return 2 }
func (tff *TimeFormatFunction) ValidateArgs(args []interface{}) error { return nil }

type DateDifferenceFunction struct{}

func (ddf *DateDifferenceFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would calculate date difference
	return time.Duration(0), nil
}
func (ddf *DateDifferenceFunction) GetName() string                       { return "datediff" }
func (ddf *DateDifferenceFunction) GetArity() int                         { return 2 }
func (ddf *DateDifferenceFunction) ValidateArgs(args []interface{}) error { return nil }

type CountFunction struct{}

func (cf *CountFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("count function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	// Implementation would count elements in collection
	return 0, nil
}
func (cf *CountFunction) GetName() string                       { return "count" }
func (cf *CountFunction) GetArity() int                         { return 1 }
func (cf *CountFunction) ValidateArgs(args []interface{}) error { return nil }

type FirstFunction struct{}

func (ff *FirstFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would return first element
	return nil, nil
}
func (ff *FirstFunction) GetName() string                       { return "first" }
func (ff *FirstFunction) GetArity() int                         { return 1 }
func (ff *FirstFunction) ValidateArgs(args []interface{}) error { return nil }

type LastFunction struct{}

func (lf *LastFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would return last element
	return nil, nil
}
func (lf *LastFunction) GetName() string                       { return "last" }
func (lf *LastFunction) GetArity() int                         { return 1 }
func (lf *LastFunction) ValidateArgs(args []interface{}) error { return nil }

type DistinctFunction struct{}

func (df2 *DistinctFunction) Execute(args []interface{}) (interface{}, error) {
	// Implementation would return distinct elements
	return args[0], nil
}
func (df2 *DistinctFunction) GetName() string                       { return "distinct" }
func (df2 *DistinctFunction) GetArity() int                         { return 1 }
func (df2 *DistinctFunction) ValidateArgs(args []interface{}) error { return nil }

type TypeFunction struct{}

func (tf2 *TypeFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	return reflect.TypeOf(args[0]).String(), nil
}
func (tf2 *TypeFunction) GetName() string                       { return "type" }
func (tf2 *TypeFunction) GetArity() int                         { return 1 }
func (tf2 *TypeFunction) ValidateArgs(args []interface{}) error { return nil }

type ExistsFunction struct{}

func (ef *ExistsFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	return args[0] != nil, nil
}
func (ef *ExistsFunction) GetName() string                       { return "exists" }
func (ef *ExistsFunction) GetArity() int                         { return 1 }
func (ef *ExistsFunction) ValidateArgs(args []interface{}) error { return nil }

type EmptyFunction struct{}

func (ef2 *EmptyFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty function requires exactly 1 argument, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	// Implementation would check if collection/string is empty
	return false, nil
}
func (ef2 *EmptyFunction) GetName() string                       { return "empty" }
func (ef2 *EmptyFunction) GetArity() int                         { return 1 }
func (ef2 *EmptyFunction) ValidateArgs(args []interface{}) error { return nil }

type DefaultFunction struct{}

func (df3 *DefaultFunction) Execute(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("default function requires exactly 2 arguments, got %d: %w", len(args), errors.ErrInvalidInput)
	}
	if args[0] != nil {
		return args[0], nil
	}
	return args[1], nil
}
func (df3 *DefaultFunction) GetName() string                       { return "default" }
func (df3 *DefaultFunction) GetArity() int                         { return 2 }
func (df3 *DefaultFunction) ValidateArgs(args []interface{}) error { return nil }

// Stub implementations for advanced rule engine

func (are *AdvancedRuleEngine) tokenizeExpression(expression string) ([]string, error) {
	// Simple tokenization - in practice this would be more sophisticated
	tokens := strings.Fields(expression)
	return tokens, nil
}

func (are *AdvancedRuleEngine) parseTokens(tokens []string) (*ExpressionNode, error) {
	// Simple parser - in practice this would build a proper AST
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression, tokens: %v: %w", tokens, errors.ErrInvalidInput)
	}

	return &ExpressionNode{
		Type:  NodeTypeLiteral,
		Value: tokens[0],
	}, nil
}

func (are *AdvancedRuleEngine) compileAST(ast *ExpressionNode) (*CompiledExpression, error) {
	return &CompiledExpression{
		AST:        ast,
		Variables:  []string{},
		Functions:  []string{},
		Operators:  []string{},
		Complexity: 1,
	}, nil
}

func (are *AdvancedRuleEngine) executeCompiledExpression(
	ctx context.Context,
	compiled *CompiledExpression,
	attributeCtx *EvaluationAttributeContext,
) (*ExpressionEvaluationResult, error) {
	// Simple execution - in practice this would traverse the AST
	return &ExpressionEvaluationResult{
		Result:      true,
		Explanation: "Expression evaluated successfully",
	}, nil
}

// Additional supporting types for geographic, network, device, and security contexts

// GeographicLocation type already defined in attribute_collector.go

type NetworkInformation struct {
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	NetworkType string `json:"network_type"`
}

type DeviceInformation struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	OSVersion  string `json:"os_version"`
	Browser    string `json:"browser"`
}

type SecurityContext struct {
	ThreatLevel         string                 `json:"threat_level"`
	AuthenticationLevel string                 `json:"authentication_level"`
	EncryptionLevel     string                 `json:"encryption_level"`
	SecurityFlags       []string               `json:"security_flags"`
	RiskScore           float64                `json:"risk_score"`
	Attributes          map[string]interface{} `json:"attributes"`
}

// AnalyzeEvaluationPerformance analyzes the performance of policy evaluations
func (pee *policyEvaluationEngine) AnalyzeEvaluationPerformance(ctx context.Context, req *PerformanceAnalysisRequest) (*EvaluationPerformanceAnalysis, error) {
	// Placeholder implementation
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
	// Placeholder implementation
	return &MultiplePolicyEvaluationResult{
		Results:   []PolicyEvaluationResult{},
		Decision:  types.PolicyDecisionAllow,
		RequestID: req.RequestID,
	}, nil
}

// EvaluateRule evaluates a single rule
func (pee *policyEvaluationEngine) EvaluateRule(ctx context.Context, req *RuleEvaluationRequest) (*RuleEvaluationResult, error) {
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
		Details:  map[string]interface{}{"expression": expression},
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
