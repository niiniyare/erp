package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// PolicyEvaluationActivities contains all activities related to evaluating ABAC policies
type PolicyEvaluationActivities struct {
	// Dependencies
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracing tracing.TracingService
	// policyRepo repository.PolicyRepository (will be added later)
	// cacheRepo  repository.CacheRepository  (will be added later)
}

// NewPolicyEvaluationActivities creates a new instance of policy evaluation activities
func NewPolicyEvaluationActivities(
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracing tracing.TracingService,
) *PolicyEvaluationActivities {
	return &PolicyEvaluationActivities{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// EvaluatePolicies evaluates all applicable ABAC policies for a permission request
func (a *PolicyEvaluationActivities) EvaluatePolicies(ctx context.Context, evalRequest map[string]interface{}) ([]*models.PolicyEvaluationResult, error) {
	ctx, span := a.tracing.StartSpan(ctx, "policy_evaluation.evaluate_policies")
	defer span.End()

	startTime := time.Now()
	defer func() {
		a.metrics.RecordHistogram("abac.policy_evaluation.duration_ms", time.Since(startTime).Seconds()*1000)
	}()

	// Extract request parameters
	userID, _ := uuid.Parse(evalRequest["user_id"].(string))
	resourceName := evalRequest["resource_name"].(string)
	actionName := evalRequest["action_name"].(string)
	entityID, _ := uuid.Parse(evalRequest["entity_id"].(string))

	userAttributes := evalRequest["user_attributes"].(*models.AttributeCollectionResult)
	resourceAttributes := evalRequest["resource_attributes"].(map[string]interface{})
	envContext := evalRequest["environment_context"].(*models.ContextEnrichmentResult)

	// TODO: Replace with actual policy repository call
	// policies, err := a.policyRepo.GetApplicablePolicies(ctx, resourceName, actionName, entityID)
	// For now, use mock policies
	policies := a.getMockPolicies(resourceName, actionName)

	var results []*models.PolicyEvaluationResult

	for _, policy := range policies {
		result, err := a.evaluatePolicy(ctx, policy, userAttributes, resourceAttributes, envContext)
		if err != nil {
			a.logger.Error("Failed to evaluate policy", 
				"policy_id", policy["id"], 
				"error", err,
				"user_id", userID,
				"resource", resourceName,
				"action", actionName)
			continue
		}
		results = append(results, result)
	}

	a.metrics.RecordCounter("abac.policies_evaluated", float64(len(results)))
	return results, nil
}

// evaluatePolicy evaluates a single ABAC policy
func (a *PolicyEvaluationActivities) evaluatePolicy(
	ctx context.Context,
	policy map[string]interface{},
	userAttrs *models.AttributeCollectionResult,
	resourceAttrs map[string]interface{},
	envContext *models.ContextEnrichmentResult,
) (*models.PolicyEvaluationResult, error) {
	startTime := time.Now()
	
	policyID, _ := uuid.Parse(policy["id"].(string))
	policyName := policy["name"].(string)
	effect := policy["effect"].(string)
	priority := int(policy["priority"].(float64))

	result := &models.PolicyEvaluationResult{
		PolicyID:   policyID,
		PolicyName: policyName,
		Effect:     effect,
		Priority:   priority,
		Decision:   "NOT_APPLICABLE",
		Details:    make(map[string]interface{}),
	}

	// Step 1: Check if policy target matches the request
	target := policy["target"].(map[string]interface{})
	targetMatches, err := a.evaluateTarget(target, userAttrs, resourceAttrs, envContext)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate policy target: %w", err)
	}

	result.TargetMatches = targetMatches
	if !targetMatches {
		result.Decision = "NOT_APPLICABLE"
		result.EvaluationTimeMS = time.Since(startTime).Milliseconds()
		return result, nil
	}

	// Step 2: Evaluate policy rule
	rule := policy["rule"].(map[string]interface{})
	ruleResult, err := a.evaluateRule(rule, userAttrs, resourceAttrs, envContext)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate policy rule: %w", err)
	}

	result.RuleResult = ruleResult
	if ruleResult {
		result.Decision = effect
	} else {
		result.Decision = "NOT_APPLICABLE"
	}

	result.EvaluationTimeMS = time.Since(startTime).Milliseconds()
	return result, nil
}

// evaluateTarget checks if the policy target matches the current request context
func (a *PolicyEvaluationActivities) evaluateTarget(
	target map[string]interface{},
	userAttrs *models.AttributeCollectionResult,
	resourceAttrs map[string]interface{},
	envContext *models.ContextEnrichmentResult,
) (bool, error) {
	// Combine all attributes for evaluation
	allAttrs := a.combineAttributes(userAttrs, resourceAttrs, envContext)

	// Evaluate each target condition
	for key, value := range target {
		match, err := a.evaluateCondition(key, value, allAttrs)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}

	return true, nil
}

// evaluateRule evaluates the policy rule logic (AND/OR conditions)
func (a *PolicyEvaluationActivities) evaluateRule(
	rule map[string]interface{},
	userAttrs *models.AttributeCollectionResult,
	resourceAttrs map[string]interface{},
	envContext *models.ContextEnrichmentResult,
) (bool, error) {
	// Combine all attributes for evaluation
	allAttrs := a.combineAttributes(userAttrs, resourceAttrs, envContext)

	// Handle AND conditions
	if andConditions, ok := rule["and"].([]interface{}); ok {
		for _, condition := range andConditions {
			condMap := condition.(map[string]interface{})
			for key, value := range condMap {
				match, err := a.evaluateCondition(key, value, allAttrs)
				if err != nil {
					return false, err
				}
				if !match {
					return false, nil // AND: all must be true
				}
			}
		}
		return true, nil
	}

	// Handle OR conditions
	if orConditions, ok := rule["or"].([]interface{}); ok {
		for _, condition := range orConditions {
			condMap := condition.(map[string]interface{})
			for key, value := range condMap {
				match, err := a.evaluateCondition(key, value, allAttrs)
				if err != nil {
					return false, err
				}
				if match {
					return true, nil // OR: any can be true
				}
			}
		}
		return false, nil
	}

	// Handle direct conditions (no AND/OR wrapper)
	for key, value := range rule {
		match, err := a.evaluateCondition(key, value, allAttrs)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}

	return true, nil
}

// evaluateCondition evaluates a single condition against the attribute context
func (a *PolicyEvaluationActivities) evaluateCondition(
	key string,
	value interface{},
	allAttrs map[string]interface{},
) (bool, error) {
	// Get the actual attribute value
	attrValue, exists := allAttrs[key]
	if !exists {
		return false, nil // Attribute doesn't exist, condition fails
	}

	// Handle different condition types
	switch v := value.(type) {
	case string:
		// Direct string comparison
		return fmt.Sprintf("%v", attrValue) == v, nil
	case float64:
		// Numeric comparison
		if attrNum, ok := attrValue.(float64); ok {
			return attrNum == v, nil
		}
		if attrInt, ok := attrValue.(int); ok {
			return float64(attrInt) == v, nil
		}
		return false, nil
	case bool:
		// Boolean comparison
		if attrBool, ok := attrValue.(bool); ok {
			return attrBool == v, nil
		}
		return false, nil
	case map[string]interface{}:
		// Complex condition with operators
		return a.evaluateComplexCondition(attrValue, v)
	case []interface{}:
		// Array contains check
		attrStr := fmt.Sprintf("%v", attrValue)
		for _, item := range v {
			if fmt.Sprintf("%v", item) == attrStr {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported condition type: %T", value)
	}
}

// evaluateComplexCondition handles complex conditions with operators
func (a *PolicyEvaluationActivities) evaluateComplexCondition(
	attrValue interface{},
	condition map[string]interface{},
) (bool, error) {
	for operator, operand := range condition {
		switch operator {
		case "eq":
			return fmt.Sprintf("%v", attrValue) == fmt.Sprintf("%v", operand), nil
		case "ne":
			return fmt.Sprintf("%v", attrValue) != fmt.Sprintf("%v", operand), nil
		case "gt":
			return a.compareNumbers(attrValue, operand, ">")
		case "gte":
			return a.compareNumbers(attrValue, operand, ">=")
		case "lt":
			return a.compareNumbers(attrValue, operand, "<")
		case "lte":
			return a.compareNumbers(attrValue, operand, "<=")
		case "in":
			if operandArray, ok := operand.([]interface{}); ok {
				attrStr := fmt.Sprintf("%v", attrValue)
				for _, item := range operandArray {
					if fmt.Sprintf("%v", item) == attrStr {
						return true, nil
					}
				}
			}
			return false, nil
		case "contains":
			attrStr := strings.ToLower(fmt.Sprintf("%v", attrValue))
			operandStr := strings.ToLower(fmt.Sprintf("%v", operand))
			return strings.Contains(attrStr, operandStr), nil
		case "starts_with":
			attrStr := strings.ToLower(fmt.Sprintf("%v", attrValue))
			operandStr := strings.ToLower(fmt.Sprintf("%v", operand))
			return strings.HasPrefix(attrStr, operandStr), nil
		case "ends_with":
			attrStr := strings.ToLower(fmt.Sprintf("%v", attrValue))
			operandStr := strings.ToLower(fmt.Sprintf("%v", operand))
			return strings.HasSuffix(attrStr, operandStr), nil
		case "between":
			if operandArray, ok := operand.([]interface{}); ok && len(operandArray) == 2 {
				min := operandArray[0]
				max := operandArray[1]
				gtResult, _ := a.compareNumbers(attrValue, min, ">=")
				ltResult, _ := a.compareNumbers(attrValue, max, "<=")
				return gtResult && ltResult, nil
			}
			return false, nil
		default:
			return false, fmt.Errorf("unsupported operator: %s", operator)
		}
	}
	return false, nil
}

// compareNumbers compares numeric values
func (a *PolicyEvaluationActivities) compareNumbers(attr, operand interface{}, operator string) (bool, error) {
	var attrNum, operandNum float64
	var ok bool

	// Convert attr to float64
	switch a := attr.(type) {
	case float64:
		attrNum = a
	case int:
		attrNum = float64(a)
	case int32:
		attrNum = float64(a)
	case int64:
		attrNum = float64(a)
	default:
		return false, fmt.Errorf("attribute is not numeric: %T", attr)
	}

	// Convert operand to float64
	switch o := operand.(type) {
	case float64:
		operandNum = o
	case int:
		operandNum = float64(o)
	case int32:
		operandNum = float64(o)
	case int64:
		operandNum = float64(o)
	default:
		return false, fmt.Errorf("operand is not numeric: %T", operand)
	}

	switch operator {
	case ">":
		return attrNum > operandNum, nil
	case ">=":
		return attrNum >= operandNum, nil
	case "<":
		return attrNum < operandNum, nil
	case "<=":
		return attrNum <= operandNum, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

// combineAttributes combines all attribute sources into a single map
func (a *PolicyEvaluationActivities) combineAttributes(
	userAttrs *models.AttributeCollectionResult,
	resourceAttrs map[string]interface{},
	envContext *models.ContextEnrichmentResult,
) map[string]interface{} {
	combined := make(map[string]interface{})

	// Add user attributes with prefix
	for k, v := range userAttrs.UserAttributes {
		combined["user."+k] = v
	}

	// Add person attributes with prefix
	for k, v := range userAttrs.PersonAttributes {
		combined["person."+k] = v
	}

	// Add employee attributes with prefix
	for k, v := range userAttrs.EmployeeAttributes {
		combined["employee."+k] = v
	}

	// Add resource attributes with prefix
	for k, v := range resourceAttrs {
		combined["resource."+k] = v
	}

	// Add environment attributes with prefix
	for k, v := range envContext.TimeContext {
		combined["time."+k] = v
	}
	for k, v := range envContext.DeviceContext {
		combined["device."+k] = v
	}
	for k, v := range envContext.LocationContext {
		combined["location."+k] = v
	}
	for k, v := range envContext.NetworkContext {
		combined["network."+k] = v
	}
	for k, v := range envContext.RiskContext {
		combined["risk."+k] = v
	}

	return combined
}

// getMockPolicies returns mock policies for testing
func (a *PolicyEvaluationActivities) getMockPolicies(resourceName, actionName string) []map[string]interface{} {
	// TODO: Replace with actual database query
	policies := []map[string]interface{}{
		{
			"id":       uuid.New().String(),
			"name":     "business_hours_access",
			"effect":   "ALLOW",
			"priority": 100.0,
			"target": map[string]interface{}{
				"resource.resource_name": resourceName,
				"resource.sensitivity_level": map[string]interface{}{
					"in": []interface{}{"LOW", "MEDIUM"},
				},
			},
			"rule": map[string]interface{}{
				"and": []interface{}{
					map[string]interface{}{
						"time.is_business_hours": true,
					},
					map[string]interface{}{
						"employee.security_level": map[string]interface{}{
							"gte": 3.0,
						},
					},
				},
			},
		},
		{
			"id":       uuid.New().String(),
			"name":     "high_security_access",
			"effect":   "ALLOW",
			"priority": 200.0,
			"target": map[string]interface{}{
				"resource.sensitivity_level": "HIGH",
			},
			"rule": map[string]interface{}{
				"and": []interface{}{
					map[string]interface{}{
						"employee.security_level": map[string]interface{}{
							"gte": 8.0,
						},
					},
					map[string]interface{}{
						"location.is_trusted": true,
					},
					map[string]interface{}{
						"device.is_trusted": true,
					},
				},
			},
		},
		{
			"id":       uuid.New().String(),
			"name":     "deny_high_risk",
			"effect":   "DENY",
			"priority": 300.0,
			"target": map[string]interface{}{},
			"rule": map[string]interface{}{
				"risk.risk_score": map[string]interface{}{
					"gt": 80.0,
				},
			},
		},
	}

	return policies
}