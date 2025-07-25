package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/niiniyare/erp/internal/core/abac/models"
)

// PermissionEvaluationWorkflow orchestrates the evaluation of user permissions using ABAC policies
func PermissionEvaluationWorkflow(ctx workflow.Context, req *models.PermissionEvaluationWorkflowRequest) (*models.PermissionEvaluationWorkflowResult, error) {
	// Set workflow timeouts and retry policies
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 2,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Initialize result
	result := &models.PermissionEvaluationWorkflowResult{
		RequestID: req.RequestID,
		Allowed:   false,
	}

	startTime := workflow.Now(ctx)

	// Step 1: Check cache first for performance
	var cacheResult *models.PolicyEvaluationCacheEntry
	err := workflow.ExecuteActivity(ctx, "GetCachedPolicyEvaluation", req.UserID, req.ResourceName, req.ActionName, req.CacheKey).Get(ctx, &cacheResult)
	if err == nil && cacheResult != nil {
		// Cache hit - return cached result
		result.Allowed = cacheResult.Decision == "ALLOW"
		result.PolicyDecisions = cacheResult.PolicyDecisions
		result.EvaluationTimeMS = time.Since(startTime).Milliseconds()
		result.CacheHit = true
		return result, nil
	}

	// Step 2: Collect user attributes (person, employee, user data)
	var userAttributes *models.AttributeCollectionResult
	err = workflow.ExecuteActivity(ctx, "CollectUserAttributes", req.UserID).Get(ctx, &userAttributes)
	if err != nil {
		result.ErrorMessage = "Failed to collect user attributes: " + err.Error()
		return result, nil
	}

	// Step 3: Collect resource attributes
	var resourceAttributes map[string]interface{}
	err = workflow.ExecuteActivity(ctx, "CollectResourceAttributes", req.ResourceName, req.EntityID).Get(ctx, &resourceAttributes)
	if err != nil {
		result.ErrorMessage = "Failed to collect resource attributes: " + err.Error()
		return result, nil
	}

	// Step 4: Enrich environment context (time, location, device, risk)
	var envContext *models.ContextEnrichmentResult
	err = workflow.ExecuteActivity(ctx, "EnrichEnvironmentContext", req.Context).Get(ctx, &envContext)
	if err != nil {
		result.ErrorMessage = "Failed to enrich environment context: " + err.Error()
		return result, nil
	}

	// Step 5: Evaluate applicable policies
	policyEvalRequest := map[string]interface{}{
		"user_id":             req.UserID,
		"resource_name":       req.ResourceName,
		"action_name":         req.ActionName,
		"entity_id":           req.EntityID,
		"user_attributes":     userAttributes,
		"resource_attributes": resourceAttributes,
		"environment_context": envContext,
	}

	var policyResults []*models.PolicyEvaluationResult
	err = workflow.ExecuteActivity(ctx, "EvaluatePolicies", policyEvalRequest).Get(ctx, &policyResults)
	if err != nil {
		result.ErrorMessage = "Failed to evaluate policies: " + err.Error()
		return result, nil
	}

	// Step 6: Process policy results and determine final decision
	finalDecision := "DENY" // Default to DENY
	var policyDecisions []string
	var effectiveRoles []string

	for _, policyResult := range policyResults {
		policyDecisions = append(policyDecisions, policyResult.Decision)
		
		// DENY always wins (highest priority)
		if policyResult.Effect == "DENY" && policyResult.RuleResult {
			finalDecision = "DENY"
			break
		}
		
		// ALLOW if rule matches
		if policyResult.Effect == "ALLOW" && policyResult.RuleResult {
			finalDecision = "ALLOW"
		}
	}

	// Step 7: Collect effective roles for audit trail
	err = workflow.ExecuteActivity(ctx, "GetUserEffectiveRoles", req.UserID, req.EntityID).Get(ctx, &effectiveRoles)
	if err != nil {
		// Don't fail the workflow for this, just log
		workflow.GetLogger(ctx).Warn("Failed to get effective roles", "error", err)
	}

	// Step 8: Cache the result if successful
	if finalDecision != "" {
		cacheEntry := &models.PolicyEvaluationCacheEntry{
			UserID:           req.UserID,
			ResourceName:     req.ResourceName,
			ActionName:       req.ActionName,
			ContextHash:      req.CacheKey,
			Decision:         finalDecision,
			PolicyDecisions:  policyDecisions,
			EvaluationTimeMS: time.Since(startTime).Milliseconds(),
			EvaluatedAt:      workflow.Now(ctx),
			ExpiresAt:        workflow.Now(ctx).Add(time.Hour), // Cache for 1 hour
		}
		
		// Cache result (don't fail workflow if caching fails)
		workflow.ExecuteActivity(ctx, "CachePolicyEvaluation", cacheEntry)
	}

	// Step 9: Audit log the permission evaluation
	auditData := map[string]interface{}{
		"user_id":          req.UserID,
		"resource_name":    req.ResourceName,
		"action_name":      req.ActionName,
		"decision":         finalDecision,
		"policy_decisions": policyDecisions,
		"evaluation_time":  time.Since(startTime).Milliseconds(),
		"context":          req.Context,
	}
	workflow.ExecuteActivity(ctx, "LogPermissionEvaluation", auditData)

	// Build final result
	result.Allowed = finalDecision == "ALLOW"
	result.PolicyDecisions = policyDecisions
	result.EffectiveRoles = effectiveRoles
	result.EvaluationTimeMS = time.Since(startTime).Milliseconds()
	result.CacheHit = false

	return result, nil
}

// BulkPermissionEvaluationWorkflow evaluates multiple permissions in parallel
func BulkPermissionEvaluationWorkflow(ctx workflow.Context, requests []*models.PermissionEvaluationWorkflowRequest) ([]*models.PermissionEvaluationWorkflowResult, error) {
	// Set up parallel execution
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 5,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	// Execute all evaluations in parallel
	var futures []workflow.ChildWorkflowFuture
	for _, req := range requests {
		future := workflow.ExecuteChildWorkflow(ctx, PermissionEvaluationWorkflow, req)
		futures = append(futures, future)
	}

	// Collect results
	var results []*models.PermissionEvaluationWorkflowResult
	for i, future := range futures {
		var result *models.PermissionEvaluationWorkflowResult
		err := future.Get(ctx, &result)
		if err != nil {
			// Create error result for failed evaluation
			result = &models.PermissionEvaluationWorkflowResult{
				RequestID:    requests[i].RequestID,
				Allowed:      false,
				ErrorMessage: err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}