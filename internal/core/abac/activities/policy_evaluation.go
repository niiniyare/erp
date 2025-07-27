package activities

import (
	"context"
	"encoding/json"
	"fmt"
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

// PolicyEvaluationActivities handles ABAC policy evaluation operations
type PolicyEvaluationActivities struct {
	policyRepo           repository.PolicyRepository
	policyEvaluationRepo repository.PolicyEvaluationRepository
	logger               logger.Logger
	metrics              metrics.MetricsProvider
	tracer               tracing.TracingService
}

// NewPolicyEvaluationActivities creates a new PolicyEvaluationActivities instance
func NewPolicyEvaluationActivities(
	policyRepo repository.PolicyRepository,
	policyEvaluationRepo repository.PolicyEvaluationRepository,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *PolicyEvaluationActivities {
	return &PolicyEvaluationActivities{
		policyRepo:           policyRepo,
		policyEvaluationRepo: policyEvaluationRepo,
		logger:               logger,
		metrics:              metrics,
		tracer:               tracer,
	}
}

// EvaluatePoliciesActivityInput represents input for policy evaluation
type EvaluatePoliciesActivityInput struct {
	UserID       uuid.UUID              `json:"user_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Action       string                 `json:"action"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Attributes   map[string]interface{} `json:"attributes"`
	RequestID    string                 `json:"request_id"`
}

// EvaluatePoliciesActivityOutput represents output of policy evaluation
type EvaluatePoliciesActivityOutput struct {
	Decision         types.PolicyDecisionType `json:"decision"`
	PolicyDecisions  []*models.PolicyDecision `json:"policy_decisions"`
	EvaluationTimeMS int64                    `json:"evaluation_time_ms"`
	CacheHit         bool                     `json:"cache_hit"`
	RequestID        string                   `json:"request_id"`
}

// EvaluatePolicies evaluates policies for a given request
func (a *PolicyEvaluationActivities) EvaluatePolicies(ctx context.Context, input *EvaluatePoliciesActivityInput) (*EvaluatePoliciesActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.EvaluatePolicies",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", input.UserID.String()),
			tracing.StringAttribute("resource_type", input.ResourceType),
			tracing.StringAttribute("action", input.Action),
			tracing.StringAttribute("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting policy evaluation",
		logger.Fields{
			"user_id":       input.UserID,
			"resource_type": input.ResourceType,
			"action":        input.Action,
			"request_id":    input.RequestID,
		})

	// Check cache first
	cacheResult, err := a.checkEvaluationCache(ctx, input)
	if err == nil && cacheResult != nil {
		a.metrics.IncrementSuccessCount("abac_evaluation_cache_hit")
		a.logger.InfoContext(ctx, "Cache hit for policy evaluation", logger.Fields{"request_id": input.RequestID})

		return &EvaluatePoliciesActivityOutput{
			Decision:         cacheResult.Decision,
			PolicyDecisions:  cacheResult.PolicyDecisions,
			EvaluationTimeMS: time.Since(startTime).Milliseconds(),
			CacheHit:         true,
			RequestID:        input.RequestID,
		}, nil
	}

	a.metrics.IncrementCounter("abac_evaluation_cache_miss")

	// Get applicable policies
	policies, err := a.getApplicablePolicies(ctx, input)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_RETRIEVAL_FAILED", "Failed to retrieve applicable policies").WithErr(err)
	}

	// Evaluate policies
	evaluation, err := a.evaluatePoliciesList(ctx, policies, input)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_FAILED", "Failed to evaluate policies").WithErr(err)
	}

	evaluationTime := time.Since(startTime)
	evaluation.EvaluationTimeMS = evaluationTime.Milliseconds()
	evaluation.RequestID = input.RequestID

	// Cache the result
	go func() {
		if err := a.cacheEvaluationResult(context.Background(), input, evaluation); err != nil {
			a.logger.WarnContext(ctx, "Failed to cache evaluation result",
				logger.Fields{"error": err.Error(), "request_id": input.RequestID})
		}
	}()

	// Record metrics
	a.recordEvaluationMetrics(ctx, input, evaluation, evaluationTime)

	// Audit log
	a.auditPolicyEvaluation(ctx, input, evaluation)

	a.logger.InfoContext(ctx, "Policy evaluation completed",
		logger.Fields{
			"decision":           evaluation.Decision,
			"policies_evaluated": len(evaluation.PolicyDecisions),
			"evaluation_time_ms": evaluation.EvaluationTimeMS,
			"request_id":         input.RequestID,
		})

	return evaluation, nil
}

// checkEvaluationCache checks if there's a cached evaluation result
func (a *PolicyEvaluationActivities) checkEvaluationCache(ctx context.Context, input *EvaluatePoliciesActivityInput) (*EvaluatePoliciesActivityOutput, error) {
	contextHash := a.generateContextHash(input.Attributes)

	cacheReq := &repository.GetCachedEvaluationResultRequest{
		UserID:       input.UserID,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Action:       input.Action,
		ContextHash:  contextHash,
	}

	result, err := a.policyEvaluationRepo.GetCachedEvaluationResult(ctx, cacheReq)
	if err != nil {
		return nil, err
	}

	return &EvaluatePoliciesActivityOutput{
		Decision:        result.Decision,
		PolicyDecisions: result.PolicyDecisions,
		CacheHit:        true,
	}, nil
}

// getApplicablePolicies retrieves policies applicable to the evaluation request
func (a *PolicyEvaluationActivities) getApplicablePolicies(ctx context.Context, input *EvaluatePoliciesActivityInput) ([]*models.Policy, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.GetApplicablePolicies")
	defer span.End()

	req := &repository.GetPoliciesForEvaluationRequest{
		ResourceType: input.ResourceType,
		Action:       input.Action,
		EntityID:     input.EntityID,
		Context:      input.Attributes,
	}

	policies, err := a.policyRepo.GetPoliciesForEvaluation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get applicable policies: %w", err)
	}

	span.SetAttributes(tracing.IntAttribute("policies_count", len(policies)))

	return policies, nil
}

// evaluatePoliciesList evaluates a list of policies against the request
func (a *PolicyEvaluationActivities) evaluatePoliciesList(ctx context.Context, policies []*models.Policy, input *EvaluatePoliciesActivityInput) (*EvaluatePoliciesActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.EvaluatePoliciesList")
	defer span.End()

	var policyDecisions []*models.PolicyDecision
	var allowCount, denyCount int

	for _, policy := range policies {
		decision, err := a.evaluateSinglePolicy(ctx, policy, input)
		if err != nil {
			a.logger.WarnContext(ctx, "Failed to evaluate policy",
				logger.Fields{
					"policy_id": policy.ID,
					"error":     err.Error(),
				})
			continue
		}

		policyDecisions = append(policyDecisions, decision)

		switch decision.Effect {
		case types.PolicyEffectAllow:
			allowCount++
		case types.PolicyEffectDeny:
			denyCount++
		}
	}

	// Apply combining algorithm (simplified - deny overrides)
	finalDecision := a.applyCombiningAlgorithm(ctx, policyDecisions)

	span.SetAttributes(
		tracing.StringAttribute("final_decision", string(finalDecision)),
		tracing.IntAttribute("allow_policies", allowCount),
		tracing.IntAttribute("deny_policies", denyCount),
	)

	return &EvaluatePoliciesActivityOutput{
		Decision:        finalDecision,
		PolicyDecisions: policyDecisions,
		CacheHit:        false,
	}, nil
}

// evaluateSinglePolicy evaluates a single policy against the request
func (a *PolicyEvaluationActivities) evaluateSinglePolicy(ctx context.Context, policy *models.Policy, input *EvaluatePoliciesActivityInput) (*models.PolicyDecision, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.EvaluateSinglePolicy",
		tracing.WithAttributes(tracing.StringAttribute("policy_id", policy.ID.String())))
	defer span.End()

	decision := &models.PolicyDecision{
		PolicyID:   policy.ID,
		PolicyName: policy.Name,
		Effect:     types.PolicyEffectNotApplicable,
	}

	// Check target matches
	if !a.evaluateTarget(ctx, policy.Target, input) {
		decision.Reason = "Target does not match"
		return decision, nil
	}

	// Evaluate rule
	ruleResult, err := a.evaluateRule(ctx, policy.Rule, input)
	if err != nil {
		decision.Effect = types.PolicyEffectIndeterminate
		decision.Reason = fmt.Sprintf("Rule evaluation error: %v", err)
		return decision, nil
	}

	if ruleResult {
		decision.Effect = policy.Effect
		decision.Reason = "Policy rule evaluated to true"
	} else {
		decision.Reason = "Policy rule evaluated to false"
	}

	return decision, nil
}

// evaluateTarget checks if the policy target matches the request
func (a *PolicyEvaluationActivities) evaluateTarget(ctx context.Context, target map[string]interface{}, input *EvaluatePoliciesActivityInput) bool {
	// Check resource type
	if resourceType, ok := target["resource_type"].(string); ok {
		if resourceType != input.ResourceType && resourceType != "*" {
			return false
		}
	}

	// Check action
	if action, ok := target["action"].(string); ok {
		if action != input.Action && action != "*" {
			return false
		}
	}

	// Check entity (if specified)
	if entityID, ok := target["entity_id"].(string); ok && input.EntityID != nil {
		if entityID != input.EntityID.String() && entityID != "*" {
			return false
		}
	}

	return true
}

// evaluateRule evaluates the policy rule against the request attributes
func (a *PolicyEvaluationActivities) evaluateRule(ctx context.Context, rule map[string]interface{}, input *EvaluatePoliciesActivityInput) (bool, error) {
	// Handle different rule types
	if conditions, ok := rule["conditions"]; ok {
		return a.evaluateConditions(ctx, conditions, input.Attributes)
	}

	// Simple attribute matching
	for key, expectedValue := range rule {
		actualValue, exists := input.Attributes[key]
		if !exists {
			return false, nil
		}

		if !a.compareValues(expectedValue, actualValue) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateConditions evaluates complex conditions (AND, OR, NOT)
func (a *PolicyEvaluationActivities) evaluateConditions(ctx context.Context, conditions interface{}, attributes map[string]interface{}) (bool, error) {
	switch cond := conditions.(type) {
	case map[string]interface{}:
		// Handle AND, OR, NOT operators
		if andConds, ok := cond["AND"]; ok {
			return a.evaluateAndConditions(ctx, andConds, attributes)
		}
		if orConds, ok := cond["OR"]; ok {
			return a.evaluateOrConditions(ctx, orConds, attributes)
		}
		if notCond, ok := cond["NOT"]; ok {
			result, err := a.evaluateConditions(ctx, notCond, attributes)
			return !result, err
		}

		// Simple attribute comparison
		return a.evaluateAttributeCondition(ctx, cond, attributes)

	case []interface{}:
		// Array of conditions (implicit AND)
		for _, condition := range cond {
			result, err := a.evaluateConditions(ctx, condition, attributes)
			if err != nil || !result {
				return false, err
			}
		}
		return true, nil
	}

	return false, fmt.Errorf("unsupported condition type: %T", conditions)
}

// evaluateAndConditions evaluates AND conditions
func (a *PolicyEvaluationActivities) evaluateAndConditions(ctx context.Context, conditions interface{}, attributes map[string]interface{}) (bool, error) {
	condList, ok := conditions.([]interface{})
	if !ok {
		return false, fmt.Errorf("AND conditions must be an array")
	}

	for _, condition := range condList {
		result, err := a.evaluateConditions(ctx, condition, attributes)
		if err != nil || !result {
			return false, err
		}
	}
	return true, nil
}

// evaluateOrConditions evaluates OR conditions
func (a *PolicyEvaluationActivities) evaluateOrConditions(ctx context.Context, conditions interface{}, attributes map[string]interface{}) (bool, error) {
	condList, ok := conditions.([]interface{})
	if !ok {
		return false, fmt.Errorf("OR conditions must be an array")
	}

	for _, condition := range condList {
		result, err := a.evaluateConditions(ctx, condition, attributes)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

// evaluateAttributeCondition evaluates a simple attribute condition
func (a *PolicyEvaluationActivities) evaluateAttributeCondition(ctx context.Context, condition map[string]interface{}, attributes map[string]interface{}) (bool, error) {
	attrName, ok := condition["attribute"].(string)
	if !ok {
		return false, fmt.Errorf("condition must specify attribute name")
	}

	operator, ok := condition["operator"].(string)
	if !ok {
		operator = "eq" // default to equals
	}

	expectedValue := condition["value"]
	actualValue, exists := attributes[attrName]
	if !exists {
		return false, nil
	}

	return a.compareWithOperator(operator, actualValue, expectedValue)
}

// compareWithOperator compares values using the specified operator
func (a *PolicyEvaluationActivities) compareWithOperator(operator string, actual, expected interface{}) (bool, error) {
	switch operator {
	case "eq", "equals":
		return a.compareValues(actual, expected), nil
	case "ne", "not_equals":
		return !a.compareValues(actual, expected), nil
	case "gt", "greater_than":
		return a.compareNumbers(actual, expected, func(a, b float64) bool { return a > b })
	case "gte", "greater_than_equals":
		return a.compareNumbers(actual, expected, func(a, b float64) bool { return a >= b })
	case "lt", "less_than":
		return a.compareNumbers(actual, expected, func(a, b float64) bool { return a < b })
	case "lte", "less_than_equals":
		return a.compareNumbers(actual, expected, func(a, b float64) bool { return a <= b })
	case "in":
		return a.valueInList(actual, expected), nil
	case "not_in":
		return !a.valueInList(actual, expected), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// compareValues compares two values for equality
func (a *PolicyEvaluationActivities) compareValues(actual, expected interface{}) bool {
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// compareNumbers compares numeric values
func (a *PolicyEvaluationActivities) compareNumbers(actual, expected interface{}, compareFn func(float64, float64) bool) (bool, error) {
	actualNum, ok1 := actual.(float64)
	expectedNum, ok2 := expected.(float64)

	if !ok1 || !ok2 {
		// Try to convert
		actualNum, ok1 = a.toFloat64(actual)
		expectedNum, ok2 = a.toFloat64(expected)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("cannot compare non-numeric values")
		}
	}

	return compareFn(actualNum, expectedNum), nil
}

// toFloat64 converts various numeric types to float64
func (a *PolicyEvaluationActivities) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// valueInList checks if a value is in a list
func (a *PolicyEvaluationActivities) valueInList(value interface{}, list interface{}) bool {
	listSlice, ok := list.([]interface{})
	if !ok {
		return false
	}

	for _, item := range listSlice {
		if a.compareValues(value, item) {
			return true
		}
	}
	return false
}

// applyCombiningAlgorithm applies the policy combining algorithm
func (a *PolicyEvaluationActivities) applyCombiningAlgorithm(ctx context.Context, decisions []*models.PolicyDecision) types.PolicyDecisionType {
	// Simplified deny-overrides algorithm
	for _, decision := range decisions {
		if decision.Effect == types.PolicyEffectDeny {
			return types.PolicyDecisionDeny
		}
	}

	for _, decision := range decisions {
		if decision.Effect == types.PolicyEffectAllow {
			return types.PolicyDecisionAllow
		}
	}

	return types.PolicyDecisionDeny // Default deny
}

// generateContextHash generates a hash for the evaluation context
func (a *PolicyEvaluationActivities) generateContextHash(attributes map[string]interface{}) string {
	// Simple hash generation - in production, use proper cryptographic hash
	data, _ := json.Marshal(attributes)
	return fmt.Sprintf("%x", len(data))
}

// cacheEvaluationResult caches the evaluation result
func (a *PolicyEvaluationActivities) cacheEvaluationResult(ctx context.Context, input *EvaluatePoliciesActivityInput, output *EvaluatePoliciesActivityOutput) error {
	contextHash := a.generateContextHash(input.Attributes)

	// Extract policy IDs
	var policyIDs []uuid.UUID
	for _, decision := range output.PolicyDecisions {
		policyIDs = append(policyIDs, decision.PolicyID)
	}

	cacheReq := &repository.CacheEvaluationResultRequest{
		UserID:             input.UserID,
		ResourceType:       input.ResourceType,
		ResourceID:         input.ResourceID,
		Action:             input.Action,
		ContextHash:        contextHash,
		Decision:           output.Decision,
		ApplicablePolicies: policyIDs,
		EvaluationTimeMS:   output.EvaluationTimeMS,
		ExpiresAt:          time.Now().Add(15 * time.Minute), // 15 minute cache
		Result: &models.PolicyEvaluationResult{
			Decision:        output.Decision,
			PolicyDecisions: output.PolicyDecisions,
		},
	}

	return a.policyEvaluationRepo.CacheEvaluationResult(ctx, cacheReq)
}

// recordEvaluationMetrics records metrics for the evaluation
func (a *PolicyEvaluationActivities) recordEvaluationMetrics(ctx context.Context, input *EvaluatePoliciesActivityInput, output *EvaluatePoliciesActivityOutput, duration time.Duration) {
	labels := metrics.Fields{
		"decision":      string(output.Decision),
		"resource_type": input.ResourceType,
		"action":        input.Action,
		"cache_hit":     fmt.Sprintf("%t", output.CacheHit),
	}

	a.metrics.IncrementCounter("abac_policy_evaluations_total", labels)
	a.metrics.ObserveHistogram("abac_evaluation_duration_seconds", duration.Seconds(), labels)
	a.metrics.RecordGauge("abac_policies_evaluated", float64(len(output.PolicyDecisions)), labels)
}

// auditPolicyEvaluation creates an audit log entry for the evaluation
func (a *PolicyEvaluationActivities) auditPolicyEvaluation(ctx context.Context, input *EvaluatePoliciesActivityInput, output *EvaluatePoliciesActivityOutput) {
	a.logger.InfoContext(ctx, "ABAC policy evaluation audit",
		logger.Fields{
			"event_type":         "policy_evaluation",
			"user_id":            input.UserID,
			"resource_type":      input.ResourceType,
			"resource_id":        input.ResourceID,
			"action":             input.Action,
			"decision":           output.Decision,
			"policies_evaluated": len(output.PolicyDecisions),
			"evaluation_time_ms": output.EvaluationTimeMS,
			"cache_hit":          output.CacheHit,
			"request_id":         input.RequestID,
		})
}
