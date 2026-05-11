//go:build ignore

package featureflag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/core/access/conditional"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// AdvancedEvaluationEngine provides sophisticated feature flag evaluation
// integrated with conditional access system
type AdvancedEvaluationEngine interface {
	// Core evaluation with conditional access integration
	EvaluateWithConditionalAccess(ctx context.Context, request *AdvancedEvaluationRequest) (*AdvancedEvaluationResult, error)

	// Complex rule evaluation
	EvaluateComplexRules(ctx context.Context, flagID uuid.UUID, rules *ComplexEvaluationRules, context *AdvancedEvaluationContext) (*AdvancedEvaluationResult, error)

	// Context-aware evaluation
	EvaluateWithContext(ctx context.Context, flagID uuid.UUID, userID uuid.UUID, contextData map[string]any) (*ContextualEvaluationResult, error)

	// Bulk evaluation for performance
	BulkEvaluate(ctx context.Context, requests []*AdvancedEvaluationRequest) ([]*AdvancedEvaluationResult, error)

	// A/B testing integration
	EvaluateVariant(ctx context.Context, request *VariantEvaluationRequest) (*VariantEvaluationResult, error)
}

// AdvancedEvaluationRequest represents a request for advanced feature flag evaluation
type AdvancedEvaluationRequest struct {
	FlagID      uuid.UUID      `json:"flag_id"`
	UserID      uuid.UUID      `json:"user_id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	SessionID   *uuid.UUID     `json:"session_id,omitempty"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `json:"user_agent"`
	Context     map[string]any `json:"context,omitempty"`
	RequestTime time.Time      `json:"request_time"`

	// Conditional access context
	AccessContext *conditional.AccessContext `json:"access_context,omitempty"`
}

// AdvancedEvaluationResult represents the result of advanced feature flag evaluation
type AdvancedEvaluationResult struct {
	FlagID  uuid.UUID `json:"flag_id"`
	UserID  uuid.UUID `json:"user_id"`
	Enabled bool      `json:"enabled"`
	Value   any       `json:"value,omitempty"`
	Variant *string   `json:"variant,omitempty"`

	// Evaluation details
	Reason            string                   `json:"reason"`
	EvaluationPath    []string                 `json:"evaluation_path"`
	MatchedRules      []string                 `json:"matched_rules"`
	ConditionalAccess *ConditionalAccessResult `json:"conditional_access,omitempty"`

	// Performance metrics
	EvaluationTimeMS int  `json:"evaluation_time_ms"`
	CacheHit         bool `json:"cache_hit"`

	// Context information
	EvaluatedAt time.Time      `json:"evaluated_at"`
	Context     map[string]any `json:"context,omitempty"`

	// A/B testing information
	ExperimentID      *uuid.UUID         `json:"experiment_id,omitempty"`
	VariantAssignment *VariantAssignment `json:"variant_assignment,omitempty"`
}

// ConditionalAccessResult represents conditional access evaluation result
type ConditionalAccessResult struct {
	Decision        string                                `json:"decision"`
	MatchedRules    []string                              `json:"matched_rules"`
	RequiredActions []conditional.ConditionalAccessAction `json:"required_actions"`
	RiskScore       int                                   `json:"risk_score"`
	RiskLevel       string                                `json:"risk_level"`
	AccessGranted   bool                                  `json:"access_granted"`
}

// ComplexEvaluationRules represents complex boolean logic rules for feature flag evaluation
type ComplexEvaluationRules struct {
	RootRule *EvaluationRule `json:"root_rule"`
}

// EvaluationRule represents a single evaluation rule with boolean logic
type EvaluationRule struct {
	ID          uuid.UUID         `json:"id"`
	Type        RuleType          `json:"type"`
	Operator    LogicalOperator   `json:"operator,omitempty"`
	Condition   *RuleCondition    `json:"condition,omitempty"`
	Children    []*EvaluationRule `json:"children,omitempty"`
	Weight      *float64          `json:"weight,omitempty"`
	Description string            `json:"description"`
}

// RuleType represents the type of evaluation rule
type RuleType string

const (
	RuleTypeLogical    RuleType = "LOGICAL"    // AND, OR, NOT operations
	RuleTypeCondition  RuleType = "CONDITION"  // Individual condition
	RuleTypePercentage RuleType = "PERCENTAGE" // Percentage rollout
	RuleTypeTime       RuleType = "TIME"       // Time-based rules
	RuleTypeLocation   RuleType = "LOCATION"   // Location-based rules
	RuleTypeDevice     RuleType = "DEVICE"     // Device-based rules
	RuleTypeNetwork    RuleType = "NETWORK"    // Network-based rules
	RuleTypeRisk       RuleType = "RISK"       // Risk-based rules
	RuleTypeCustom     RuleType = "CUSTOM"     // Custom evaluation logic
)

// LogicalOperator represents logical operators for combining rules
type LogicalOperator string

const (
	OperatorAND LogicalOperator = "AND"
	OperatorOR  LogicalOperator = "OR"
	OperatorNOT LogicalOperator = "NOT"
)

// RuleCondition represents a condition within an evaluation rule
type RuleCondition struct {
	Field         string            `json:"field"`            // user.role, context.department, etc.
	Operator      ConditionOperator `json:"operator"`         // EQUALS, CONTAINS, GREATER_THAN, etc.
	Value         any               `json:"value"`            // Expected value
	Values        []any             `json:"values,omitempty"` // Multiple values for IN operator
	CaseSensitive bool              `json:"case_sensitive"`
}

// ConditionOperator represents operators for individual conditions
type ConditionOperator string

const (
	OpEquals       ConditionOperator = "EQUALS"
	OpNotEquals    ConditionOperator = "NOT_EQUALS"
	OpContains     ConditionOperator = "CONTAINS"
	OpNotContains  ConditionOperator = "NOT_CONTAINS"
	OpStartsWith   ConditionOperator = "STARTS_WITH"
	OpEndsWith     ConditionOperator = "ENDS_WITH"
	OpIn           ConditionOperator = "IN"
	OpNotIn        ConditionOperator = "NOT_IN"
	OpGreaterThan  ConditionOperator = "GREATER_THAN"
	OpLessThan     ConditionOperator = "LESS_THAN"
	OpGreaterEqual ConditionOperator = "GREATER_EQUAL"
	OpLessEqual    ConditionOperator = "LESS_EQUAL"
	OpRegexMatch   ConditionOperator = "REGEX_MATCH"
	OpExists       ConditionOperator = "EXISTS"
	OpNotExists    ConditionOperator = "NOT_EXISTS"
)

// AdvancedEvaluationContext represents the context for advanced feature flag evaluation
type AdvancedEvaluationContext struct {
	UserID       uuid.UUID                 `json:"user_id"`
	UserData     map[string]any            `json:"user_data,omitempty"`
	SessionData  map[string]any            `json:"session_data,omitempty"`
	RequestData  map[string]any            `json:"request_data,omitempty"`
	DeviceInfo   *conditional.DeviceInfo   `json:"device_info,omitempty"`
	LocationInfo *conditional.LocationInfo `json:"location_info,omitempty"`
	NetworkInfo  *conditional.NetworkInfo  `json:"network_info,omitempty"`
	TimeContext  *conditional.TimeContext  `json:"time_context,omitempty"`
	RiskContext  *conditional.RiskContext  `json:"risk_context,omitempty"`
	CustomData   map[string]any            `json:"custom_data,omitempty"`
}

// ContextualEvaluationResult represents the result of contextual evaluation
type ContextualEvaluationResult struct {
	*AdvancedEvaluationResult
	ContextScore      int      `json:"context_score"`
	ContextFactors    []string `json:"context_factors"`
	RecommendedAction string   `json:"recommended_action"`
}

// VariantEvaluationRequest represents a request for A/B testing variant evaluation
type VariantEvaluationRequest struct {
	ExperimentID uuid.UUID      `json:"experiment_id"`
	UserID       uuid.UUID      `json:"user_id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	Context      map[string]any `json:"context,omitempty"`
	ForceVariant *string        `json:"force_variant,omitempty"`
}

// VariantEvaluationResult represents the result of variant evaluation
type VariantEvaluationResult struct {
	ExperimentID      uuid.UUID          `json:"experiment_id"`
	UserID            uuid.UUID          `json:"user_id"`
	AssignedVariant   string             `json:"assigned_variant"`
	VariantConfig     map[string]any     `json:"variant_config"`
	Assignment        *VariantAssignment `json:"assignment"`
	IsControl         bool               `json:"is_control"`
	TrafficAllocation float64            `json:"traffic_allocation"`
}

// VariantAssignment represents a user's variant assignment
type VariantAssignment struct {
	ID            uuid.UUID      `json:"id"`
	ExperimentID  uuid.UUID      `json:"experiment_id"`
	UserID        uuid.UUID      `json:"user_id"`
	VariantName   string         `json:"variant_name"`
	AssignedAt    time.Time      `json:"assigned_at"`
	Sticky        bool           `json:"sticky"`
	TrafficBucket int            `json:"traffic_bucket"`
	Properties    map[string]any `json:"properties,omitempty"`
}

// advancedEvaluationEngine implements AdvancedEvaluationEngine
type advancedEvaluationEngine struct {
	conditionalAccessService conditional.ConditionalAccessService
	simpleService            Service
	logger                   logger.Logger
	metrics                  metrics.MetricsProvider
	tracing                  tracing.Service
}

// NewAdvancedEvaluationEngine creates a new advanced evaluation engine
func NewAdvancedEvaluationEngine(
	conditionalAccessService conditional.ConditionalAccessService,
	simpleService Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracing tracing.Service,
) AdvancedEvaluationEngine {
	return &advancedEvaluationEngine{
		conditionalAccessService: conditionalAccessService,
		simpleService:            simpleService,
		logger:                   logger,
		metrics:                  metrics,
		tracing:                  tracing,
	}
}

// EvaluateWithConditionalAccess evaluates a feature flag with conditional access integration
func (e *advancedEvaluationEngine) EvaluateWithConditionalAccess(ctx context.Context, request *AdvancedEvaluationRequest) (*AdvancedEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "advancedEvaluationEngine.EvaluateWithConditionalAccess")
	defer span.End()

	startTime := time.Now()

	// Create or enrich access context
	accessContext := request.AccessContext
	if accessContext == nil {
		accessContext = &conditional.AccessContext{
			UserID:            request.UserID,
			IPAddress:         request.IPAddress,
			UserAgent:         request.UserAgent,
			RequestedResource: fmt.Sprintf("feature_flag:%s", request.FlagID.String()),
			RequestedAction:   "evaluate",
			Metadata:          request.Context,
		}
		if request.SessionID != nil {
			accessContext.SessionID = *request.SessionID
		}
	}

	// Evaluate conditional access first
	conditionalResult, err := e.conditionalAccessService.EvaluateAccess(ctx, accessContext)
	if err != nil {
		e.logger.Error("Conditional access evaluation failed", logger.Fields{
			"error":   err.Error(),
			"user_id": request.UserID,
			"flag_id": request.FlagID,
		})
		// Continue with basic evaluation if conditional access fails
	}

	// Check if conditional access blocks the evaluation
	accessGranted := true
	var conditionalAccessResult *ConditionalAccessResult

	if conditionalResult != nil {
		accessGranted = conditionalResult.Decision == conditional.EffectAllow ||
			conditionalResult.Decision == conditional.EffectAuditOnly

		conditionalAccessResult = &ConditionalAccessResult{
			Decision:        string(conditionalResult.Decision),
			MatchedRules:    e.extractRuleNames(conditionalResult.MatchedRules),
			RequiredActions: conditionalResult.RequiredActions,
			RiskScore:       conditionalResult.RiskScore,
			AccessGranted:   accessGranted,
		}

		// Determine risk level from risk score
		switch {
		case conditionalResult.RiskScore >= 70:
			conditionalAccessResult.RiskLevel = "CRITICAL"
		case conditionalResult.RiskScore >= 50:
			conditionalAccessResult.RiskLevel = "HIGH"
		case conditionalResult.RiskScore >= 25:
			conditionalAccessResult.RiskLevel = "MEDIUM"
		default:
			conditionalAccessResult.RiskLevel = "LOW"
		}
	}

	// Perform basic feature flag evaluation
	basicResult, err := e.evaluateBasicFlag(ctx, request.FlagID, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("basic feature flag evaluation failed: %w", err)
	}

	// Apply conditional access decision
	finalEnabled := basicResult.Enabled && accessGranted
	reason := basicResult.Reason

	if !accessGranted {
		reason = fmt.Sprintf("Access denied by conditional access: %s", conditionalResult.Reason)
	}

	result := &AdvancedEvaluationResult{
		FlagID:            request.FlagID,
		UserID:            request.UserID,
		Enabled:           finalEnabled,
		Value:             basicResult.Value,
		Reason:            reason,
		EvaluationPath:    []string{"conditional_access", "basic_evaluation"},
		MatchedRules:      []string{}, // TODO: Extract from basic result
		ConditionalAccess: conditionalAccessResult,
		EvaluationTimeMS:  int(time.Since(startTime).Milliseconds()),
		CacheHit:          false, // Advanced evaluation is not cached
		EvaluatedAt:       time.Now(),
		Context:           request.Context,
	}

	// Record metrics
	e.recordEvaluationMetrics(ctx, request, result, conditionalResult)

	// Log evaluation
	e.logger.Info("Advanced feature flag evaluation completed", logger.Fields{
		"flag_id":            request.FlagID,
		"user_id":            request.UserID,
		"enabled":            finalEnabled,
		"conditional_access": accessGranted,
		"risk_score":         conditionalResult.RiskScore,
		"evaluation_time_ms": result.EvaluationTimeMS,
	})

	return result, nil
}

// evaluateBasicFlag performs basic feature flag evaluation
func (e *advancedEvaluationEngine) evaluateBasicFlag(ctx context.Context, flagID, userID uuid.UUID) (*EvaluationResult, error) {
	// Get the feature flag
	flag, err := e.simpleService.GetFeatureFlagByID(ctx, flagID)
	if err != nil {
		return nil, err
	}

	// Create evaluation context for basic evaluation
	evalCtx := &EvaluationContext{
		TenantID:   flag.TenantID,
		UserID:     &userID,
		Attributes: map[string]string{},
	}

	// Use the simple service evaluation
	result, err := e.simpleService.EvaluateFlag(ctx, flag.Name, evalCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// EvaluateComplexRules evaluates feature flags using complex boolean logic rules
func (e *advancedEvaluationEngine) EvaluateComplexRules(ctx context.Context, flagID uuid.UUID, rules *ComplexEvaluationRules, context *AdvancedEvaluationContext) (*AdvancedEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "advancedEvaluationEngine.EvaluateComplexRules")
	defer span.End()

	if rules == nil || rules.RootRule == nil {
		return &AdvancedEvaluationResult{
			FlagID:      flagID,
			UserID:      context.UserID,
			Enabled:     false,
			Reason:      "No evaluation rules provided",
			CacheHit:    false,
			EvaluatedAt: time.Now(),
		}, nil
	}

	// Evaluate the root rule recursively
	result, evaluationPath := e.evaluateRule(ctx, rules.RootRule, context)

	return &AdvancedEvaluationResult{
		FlagID:         flagID,
		UserID:         context.UserID,
		Enabled:        result,
		Reason:         strings.Join(evaluationPath, " -> "),
		EvaluationPath: evaluationPath,
		CacheHit:       false, // Complex rules are not cached
		EvaluatedAt:    time.Now(),
	}, nil
}

// EvaluateWithContext evaluates a feature flag with rich contextual information
func (e *advancedEvaluationEngine) EvaluateWithContext(ctx context.Context, flagID uuid.UUID, userID uuid.UUID, contextData map[string]any) (*ContextualEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "advancedEvaluationEngine.EvaluateWithContext")
	defer span.End()

	// Create advanced evaluation request
	request := &AdvancedEvaluationRequest{
		FlagID:      flagID,
		UserID:      userID,
		Context:     contextData,
		RequestTime: time.Now(),
	}

	// Extract network information from context
	if ipAddress, ok := contextData["ip_address"].(string); ok {
		request.IPAddress = ipAddress
	}
	if userAgent, ok := contextData["user_agent"].(string); ok {
		request.UserAgent = userAgent
	}

	// Perform advanced evaluation
	result, err := e.EvaluateWithConditionalAccess(ctx, request)
	if err != nil {
		return nil, err
	}

	// Calculate context score based on available context information
	contextScore := e.calculateContextScore(contextData)
	contextFactors := e.extractContextFactors(contextData)
	recommendedAction := e.determineRecommendedAction(result, contextScore)

	return &ContextualEvaluationResult{
		AdvancedEvaluationResult: result,
		ContextScore:             contextScore,
		ContextFactors:           contextFactors,
		RecommendedAction:        recommendedAction,
	}, nil
}

// BulkEvaluate performs bulk evaluation for multiple feature flags
func (e *advancedEvaluationEngine) BulkEvaluate(ctx context.Context, requests []*AdvancedEvaluationRequest) ([]*AdvancedEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "advancedEvaluationEngine.BulkEvaluate")
	defer span.End()

	results := make([]*AdvancedEvaluationResult, len(requests))

	// Evaluate each request in parallel
	// TODO: Implement concurrency with semaphore limiting
	for i, request := range requests {
		result, err := e.EvaluateWithConditionalAccess(ctx, request)
		if err != nil {
			e.logger.Error("Bulk evaluation failed for request", logger.Fields{
				"error":   err.Error(),
				"index":   i,
				"flag_id": request.FlagID,
				"user_id": request.UserID,
			})
			// Continue with other requests
			results[i] = &AdvancedEvaluationResult{
				FlagID:      request.FlagID,
				UserID:      request.UserID,
				Enabled:     false,
				Reason:      fmt.Sprintf("Evaluation failed: %v", err),
				EvaluatedAt: time.Now(),
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// EvaluateVariant evaluates A/B testing variants (placeholder implementation)
func (e *advancedEvaluationEngine) EvaluateVariant(ctx context.Context, request *VariantEvaluationRequest) (*VariantEvaluationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "advancedEvaluationEngine.EvaluateVariant")
	defer span.End()

	// TODO: Implement A/B testing variant evaluation
	// This is a placeholder implementation

	return &VariantEvaluationResult{
		ExperimentID:      request.ExperimentID,
		UserID:            request.UserID,
		AssignedVariant:   "control",
		VariantConfig:     map[string]any{"enabled": false},
		IsControl:         true,
		TrafficAllocation: 50.0,
		Assignment: &VariantAssignment{
			ID:            uuid.New(),
			ExperimentID:  request.ExperimentID,
			UserID:        request.UserID,
			VariantName:   "control",
			AssignedAt:    time.Now(),
			Sticky:        true,
			TrafficBucket: 0,
		},
	}, nil
}

// Helper methods

// evaluateRule recursively evaluates a rule
func (e *advancedEvaluationEngine) evaluateRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	switch rule.Type {
	case RuleTypeLogical:
		return e.evaluateLogicalRule(ctx, rule, context)
	case RuleTypeCondition:
		return e.evaluateConditionRule(ctx, rule, context)
	case RuleTypePercentage:
		return e.evaluatePercentageRule(ctx, rule, context)
	case RuleTypeTime:
		return e.evaluateTimeRule(ctx, rule, context)
	case RuleTypeLocation:
		return e.evaluateLocationRule(ctx, rule, context)
	case RuleTypeDevice:
		return e.evaluateDeviceRule(ctx, rule, context)
	case RuleTypeNetwork:
		return e.evaluateNetworkRule(ctx, rule, context)
	case RuleTypeRisk:
		return e.evaluateRiskRule(ctx, rule, context)
	case RuleTypeCustom:
		return e.evaluateCustomRule(ctx, rule, context)
	default:
		return false, []string{fmt.Sprintf("unknown_rule_type:%s", rule.Type)}
	}
}

// evaluateLogicalRule evaluates logical operators (AND, OR, NOT)
func (e *advancedEvaluationEngine) evaluateLogicalRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	var paths []string

	switch rule.Operator {
	case OperatorAND:
		for _, child := range rule.Children {
			result, childPaths := e.evaluateRule(ctx, child, context)
			paths = append(paths, childPaths...)
			if !result {
				return false, append(paths, "AND:false")
			}
		}
		return true, append(paths, "AND:true")

	case OperatorOR:
		allFalse := true
		for _, child := range rule.Children {
			result, childPaths := e.evaluateRule(ctx, child, context)
			paths = append(paths, childPaths...)
			if result {
				allFalse = false
				return true, append(paths, "OR:true")
			}
		}
		if allFalse {
			return false, append(paths, "OR:false")
		}

	case OperatorNOT:
		if len(rule.Children) > 0 {
			result, childPaths := e.evaluateRule(ctx, rule.Children[0], context)
			paths = append(paths, childPaths...)
			return !result, append(paths, fmt.Sprintf("NOT:%t", !result))
		}
	}

	return false, append(paths, "logical_error")
}

// evaluateConditionRule evaluates individual conditions
func (e *advancedEvaluationEngine) evaluateConditionRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if rule.Condition == nil {
		return false, []string{"condition_missing"}
	}

	// Extract field value from context
	fieldValue := e.extractFieldValue(rule.Condition.Field, context)
	expectedValue := rule.Condition.Value

	// Evaluate condition based on operator
	result := e.evaluateConditionOperator(rule.Condition.Operator, fieldValue, expectedValue, rule.Condition.Values, rule.Condition.CaseSensitive)

	path := fmt.Sprintf("condition:%s:%s:%v=%t", rule.Condition.Field, rule.Condition.Operator, expectedValue, result)
	return result, []string{path}
}

// evaluatePercentageRule evaluates percentage-based rollout
func (e *advancedEvaluationEngine) evaluatePercentageRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if rule.Weight == nil {
		return false, []string{"percentage_weight_missing"}
	}

	// Use user ID for consistent percentage calculation
	userHash := e.hashUserID(context.UserID)
	percentage := float64(userHash%100) / 100.0

	result := percentage < *rule.Weight
	path := fmt.Sprintf("percentage:%.2f<%.2f=%t", percentage, *rule.Weight, result)

	return result, []string{path}
}

// evaluateTimeRule evaluates time-based rules using conditional access time context
func (e *advancedEvaluationEngine) evaluateTimeRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if context.TimeContext == nil {
		return false, []string{"time_context_missing"}
	}

	// Example time-based evaluation (can be extended)
	isWorkingHours := context.TimeContext.IsWorkingHours
	isWeekend := context.TimeContext.IsWeekend

	// Default: allow during working hours, deny on weekends
	result := isWorkingHours && !isWeekend
	path := fmt.Sprintf("time:working_hours=%t,weekend=%t,result=%t", isWorkingHours, isWeekend, result)

	return result, []string{path}
}

// evaluateLocationRule evaluates location-based rules
func (e *advancedEvaluationEngine) evaluateLocationRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if context.LocationInfo == nil {
		return false, []string{"location_context_missing"}
	}

	// Example location-based evaluation
	isTrusted := context.LocationInfo.IsTrusted
	path := fmt.Sprintf("location:trusted=%t", isTrusted)

	return isTrusted, []string{path}
}

// evaluateDeviceRule evaluates device-based rules
func (e *advancedEvaluationEngine) evaluateDeviceRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if context.DeviceInfo == nil {
		return false, []string{"device_context_missing"}
	}

	// Example device-based evaluation
	isCompliant := context.DeviceInfo.IsCompliant
	isManaged := context.DeviceInfo.IsManaged

	result := isCompliant && isManaged
	path := fmt.Sprintf("device:compliant=%t,managed=%t,result=%t", isCompliant, isManaged, result)

	return result, []string{path}
}

// evaluateNetworkRule evaluates network-based rules
func (e *advancedEvaluationEngine) evaluateNetworkRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if context.NetworkInfo == nil {
		return false, []string{"network_context_missing"}
	}

	// Example network-based evaluation
	isCorporate := context.NetworkInfo.IsCorporate
	isSecure := context.NetworkInfo.IsSecure

	result := isCorporate && isSecure
	path := fmt.Sprintf("network:corporate=%t,secure=%t,result=%t", isCorporate, isSecure, result)

	return result, []string{path}
}

// evaluateRiskRule evaluates risk-based rules
func (e *advancedEvaluationEngine) evaluateRiskRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	if context.RiskContext == nil {
		return false, []string{"risk_context_missing"}
	}

	// Example risk-based evaluation
	riskScore := context.RiskContext.RiskScore
	isLowRisk := riskScore < 25 // Risk score threshold

	path := fmt.Sprintf("risk:score=%d,low_risk=%t", riskScore, isLowRisk)

	return isLowRisk, []string{path}
}

// evaluateCustomRule evaluates custom rules (placeholder)
func (e *advancedEvaluationEngine) evaluateCustomRule(ctx context.Context, rule *EvaluationRule, context *AdvancedEvaluationContext) (bool, []string) {
	// TODO: Implement custom rule evaluation logic
	// This could involve calling external services, complex business logic, etc.

	return false, []string{"custom_rule_not_implemented"}
}

// extractFieldValue extracts a field value from the evaluation context
func (e *advancedEvaluationEngine) extractFieldValue(field string, context *AdvancedEvaluationContext) any {
	// Parse dot notation field paths
	parts := strings.Split(field, ".")

	switch parts[0] {
	case "user":
		if len(parts) > 1 && context.UserData != nil {
			return context.UserData[parts[1]]
		}
		return nil
	case "session":
		if len(parts) > 1 && context.SessionData != nil {
			return context.SessionData[parts[1]]
		}
		return nil
	case "request":
		if len(parts) > 1 && context.RequestData != nil {
			return context.RequestData[parts[1]]
		}
		return nil
	case "device":
		if len(parts) > 1 && context.DeviceInfo != nil {
			switch parts[1] {
			case "type":
				return context.DeviceInfo.DeviceType
			case "os":
				return context.DeviceInfo.OperatingSystem
			case "browser":
				return context.DeviceInfo.Browser
			case "is_managed":
				return context.DeviceInfo.IsManaged
			case "is_compliant":
				return context.DeviceInfo.IsCompliant
			}
		}
		return nil
	case "location":
		if len(parts) > 1 && context.LocationInfo != nil {
			switch parts[1] {
			case "country":
				return context.LocationInfo.Country
			case "region":
				return context.LocationInfo.Region
			case "city":
				return context.LocationInfo.City
			case "is_trusted":
				return context.LocationInfo.IsTrusted
			}
		}
		return nil
	case "custom":
		if len(parts) > 1 && context.CustomData != nil {
			return context.CustomData[parts[1]]
		}
		return nil
	}

	return nil
}

// evaluateConditionOperator evaluates a condition based on the operator
func (e *advancedEvaluationEngine) evaluateConditionOperator(operator ConditionOperator, actual, expected any, values []any, caseSensitive bool) bool {
	// Convert to strings for comparison if needed
	actualStr := e.toString(actual, caseSensitive)
	expectedStr := e.toString(expected, caseSensitive)

	switch operator {
	case OpEquals:
		return actualStr == expectedStr
	case OpNotEquals:
		return actualStr != expectedStr
	case OpContains:
		return strings.Contains(actualStr, expectedStr)
	case OpNotContains:
		return !strings.Contains(actualStr, expectedStr)
	case OpStartsWith:
		return strings.HasPrefix(actualStr, expectedStr)
	case OpEndsWith:
		return strings.HasSuffix(actualStr, expectedStr)
	case OpIn:
		for _, value := range values {
			if actualStr == e.toString(value, caseSensitive) {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, value := range values {
			if actualStr == e.toString(value, caseSensitive) {
				return false
			}
		}
		return true
	case OpExists:
		return actual != nil
	case OpNotExists:
		return actual == nil
	// TODO: Implement numeric and regex operators
	default:
		return false
	}
}

// Helper utility methods

func (e *advancedEvaluationEngine) toString(value any, caseSensitive bool) string {
	if value == nil {
		return ""
	}

	str := fmt.Sprintf("%v", value)
	if !caseSensitive {
		str = strings.ToLower(str)
	}

	return str
}

func (e *advancedEvaluationEngine) hashUserID(userID uuid.UUID) uint32 {
	// Simple hash function for consistent percentage calculation
	bytes := []byte(userID.String())
	var hash uint32
	for _, b := range bytes {
		hash = hash*31 + uint32(b)
	}
	return hash
}

func (e *advancedEvaluationEngine) extractRuleNames(rules []conditional.ConditionalAccessRule) []string {
	names := make([]string, len(rules))
	for i, rule := range rules {
		names[i] = rule.Name
	}
	return names
}

func (e *advancedEvaluationEngine) calculateContextScore(contextData map[string]any) int {
	// Calculate a score based on available context information
	score := 0

	if _, ok := contextData["ip_address"]; ok {
		score += 10
	}
	if _, ok := contextData["user_agent"]; ok {
		score += 10
	}
	if _, ok := contextData["location"]; ok {
		score += 15
	}
	if _, ok := contextData["device_info"]; ok {
		score += 15
	}
	if _, ok := contextData["session_data"]; ok {
		score += 10
	}
	if _, ok := contextData["user_attributes"]; ok {
		score += 20
	}

	// Additional context factors
	score += len(contextData) * 2 // 2 points per additional context field

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

func (e *advancedEvaluationEngine) extractContextFactors(contextData map[string]any) []string {
	factors := make([]string, 0, len(contextData))
	for key := range contextData {
		factors = append(factors, key)
	}
	return factors
}

func (e *advancedEvaluationEngine) determineRecommendedAction(result *AdvancedEvaluationResult, contextScore int) string {
	if !result.Enabled {
		return "DENIED"
	}

	if result.ConditionalAccess != nil && result.ConditionalAccess.RiskScore > 50 {
		return "MONITOR"
	}

	if contextScore < 30 {
		return "COLLECT_MORE_CONTEXT"
	}

	return "ALLOW"
}

func (e *advancedEvaluationEngine) recordEvaluationMetrics(ctx context.Context, request *AdvancedEvaluationRequest, result *AdvancedEvaluationResult, conditionalResult *conditional.ConditionalAccessEvaluationResult) {
	labels := map[string]any{
		"flag_id":   request.FlagID.String(),
		"enabled":   result.Enabled,
		"cache_hit": result.CacheHit,
	}

	if conditionalResult != nil {
		labels["conditional_access"] = string(conditionalResult.Decision)
		labels["risk_score"] = conditionalResult.RiskScore
	}

	e.metrics.IncrementCounter("advanced_feature_flag_evaluations_total", labels)
	e.metrics.ObserveHistogram("advanced_feature_flag_evaluation_duration_ms", float64(result.EvaluationTimeMS), labels)

	if conditionalResult != nil {
		e.metrics.ObserveHistogram("conditional_access_risk_score", float64(conditionalResult.RiskScore), labels)
	}
}
