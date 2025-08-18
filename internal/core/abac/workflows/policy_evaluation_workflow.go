package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/types"
)

// PolicyEvaluationWorkflowName is the name of the policy evaluation workflow
const PolicyEvaluationWorkflowName = "PolicyEvaluationWorkflow"

// PolicyEvaluationWorkflowInput represents input for the policy evaluation workflow
type PolicyEvaluationWorkflowInput struct {
	EvaluationRequest *activities.EvaluatePoliciesActivityInput `json:"evaluation_request"`
	CacheTTLMinutes   int32                                     `json:"cache_ttl_minutes"`
	SkipCache         bool                                      `json:"skip_cache"`
}

// PolicyEvaluationWorkflowOutput represents output of the policy evaluation workflow
type PolicyEvaluationWorkflowOutput struct {
	Decision         types.PolicyDecisionType           `json:"decision"`
	PolicyDecisions  []*activities.PolicyDecisionResult `json:"policy_decisions"`
	EvaluationTimeMS int64                              `json:"evaluation_time_ms"`
	CacheHit         bool                               `json:"cache_hit"`
	RequestID        string                             `json:"request_id"`
}

// PolicyEvaluationWorkflow orchestrates ABAC policy evaluation with caching
func PolicyEvaluationWorkflow(ctx workflow.Context, input PolicyEvaluationWorkflowInput) (*PolicyEvaluationWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Starting ABAC policy evaluation workflow",
		"user_id", input.EvaluationRequest.UserID,
		"resource_type", input.EvaluationRequest.ResourceType,
		"action", input.EvaluationRequest.Action,
		"request_id", input.EvaluationRequest.RequestID)

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Check cache first (unless skipped)
	var cacheResult *activities.GetCachedPolicyEvaluationOutput
	if !input.SkipCache {
		contextHash := generateContextHash(input.EvaluationRequest.Attributes)

		getCacheInput := &activities.GetCachedPolicyEvaluationInput{
			UserID:       input.EvaluationRequest.UserID,
			ResourceType: input.EvaluationRequest.ResourceType,
			ResourceID:   input.EvaluationRequest.ResourceID,
			Action:       input.EvaluationRequest.Action,
			ContextHash:  contextHash,
			RequestID:    input.EvaluationRequest.RequestID,
		}

		err := workflow.ExecuteActivity(ctx, "CacheActivities.GetCachedPolicyEvaluation", getCacheInput).Get(ctx, &cacheResult)
		if err == nil && cacheResult.Found {
			logger.Info("Policy evaluation cache hit", "request_id", input.EvaluationRequest.RequestID)

			return &PolicyEvaluationWorkflowOutput{
				Decision:         cacheResult.Decision,
				PolicyDecisions:  convertToWorkflowPolicyDecisions(cacheResult.PolicyDecisions),
				EvaluationTimeMS: 0, // Cache hit - minimal time
				CacheHit:         true,
				RequestID:        input.EvaluationRequest.RequestID,
			}, nil
		}
	}

	logger.Info("Policy evaluation cache miss - proceeding with full evaluation")

	// Step 2: Collect user attributes
	var userAttrs *activities.CollectUserAttributesOutput
	userAttrInput := &activities.CollectUserAttributesInput{
		UserID:    input.EvaluationRequest.UserID,
		EntityID:  input.EvaluationRequest.EntityID,
		RequestID: input.EvaluationRequest.RequestID,
	}

	err := workflow.ExecuteActivity(ctx, "AttributeCollectionActivities.CollectUserAttributes", userAttrInput).Get(ctx, &userAttrs)
	if err != nil {
		logger.Error("Failed to collect user attributes", "error", err, "request_id", input.EvaluationRequest.RequestID)
		return nil, err
	}

	// Step 3: Collect resource attributes
	var resourceAttrs *activities.CollectResourceAttributesOutput
	resourceAttrInput := &activities.CollectResourceAttributesInput{
		ResourceType: input.EvaluationRequest.ResourceType,
		ResourceID:   input.EvaluationRequest.ResourceID,
		EntityID:     input.EvaluationRequest.EntityID,
		RequestID:    input.EvaluationRequest.RequestID,
	}

	err = workflow.ExecuteActivity(ctx, "AttributeCollectionActivities.CollectResourceAttributes", resourceAttrInput).Get(ctx, &resourceAttrs)
	if err != nil {
		logger.Error("Failed to collect resource attributes", "error", err, "request_id", input.EvaluationRequest.RequestID)
		return nil, err
	}

	// Step 4: Build environment context
	var envContext *activities.BuildEnvironmentContextOutput
	envContextInput := &activities.BuildEnvironmentContextInput{
		RequestTime: workflow.Now(ctx),
		RequestID:   input.EvaluationRequest.RequestID,
	}

	err = workflow.ExecuteActivity(ctx, "AttributeCollectionActivities.BuildEnvironmentContext", envContextInput).Get(ctx, &envContext)
	if err != nil {
		logger.Error("Failed to build environment context", "error", err, "request_id", input.EvaluationRequest.RequestID)
		return nil, err
	}

	// Step 5: Combine all attributes
	allAttributes := make(map[string]any)

	// Add user attributes
	for k, v := range userAttrs.Attributes {
		allAttributes[k] = v
	}

	// Add resource attributes
	for k, v := range resourceAttrs.Attributes {
		allAttributes[k] = v
	}

	// Add environment context
	for k, v := range envContext.Context {
		allAttributes[k] = v
	}

	// Add original request attributes
	for k, v := range input.EvaluationRequest.Attributes {
		allAttributes[k] = v
	}

	// Update evaluation request with combined attributes
	input.EvaluationRequest.Attributes = allAttributes

	// Step 6: Evaluate policies
	var policyResult *activities.EvaluatePoliciesActivityOutput
	err = workflow.ExecuteActivity(ctx, "PolicyEvaluationActivities.EvaluatePolicies", input.EvaluationRequest).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to evaluate policies", "error", err, "request_id", input.EvaluationRequest.RequestID)
		return nil, err
	}

	// Step 7: Cache the result (fire and forget)
	if !input.SkipCache && input.CacheTTLMinutes > 0 {
		contextHash := generateContextHash(allAttributes)

		cacheInput := &activities.CachePolicyEvaluationInput{
			UserID:           input.EvaluationRequest.UserID,
			ResourceType:     input.EvaluationRequest.ResourceType,
			ResourceID:       input.EvaluationRequest.ResourceID,
			Action:           input.EvaluationRequest.Action,
			ContextHash:      contextHash,
			Decision:         policyResult.Decision,
			PolicyDecisions:  policyResult.PolicyDecisions,
			EvaluationTimeMS: policyResult.EvaluationTimeMS,
			TTLMinutes:       input.CacheTTLMinutes,
			RequestID:        input.EvaluationRequest.RequestID,
		}

		// Execute cache activity asynchronously (fire and forget)
		workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
			RetryPolicy:         nil, // Don't retry cache operations
		}), "CacheActivities.CachePolicyEvaluation", cacheInput)
	}

	logger.Info("Policy evaluation workflow completed",
		"decision", policyResult.Decision,
		"policies_evaluated", len(policyResult.PolicyDecisions),
		"evaluation_time_ms", policyResult.EvaluationTimeMS,
		"request_id", input.EvaluationRequest.RequestID)

	return &PolicyEvaluationWorkflowOutput{
		Decision:         policyResult.Decision,
		PolicyDecisions:  convertToWorkflowPolicyDecisions(policyResult.PolicyDecisions),
		EvaluationTimeMS: policyResult.EvaluationTimeMS,
		CacheHit:         false,
		RequestID:        input.EvaluationRequest.RequestID,
	}, nil
}

// BulkPolicyEvaluationWorkflowName is the name of the bulk policy evaluation workflow
const BulkPolicyEvaluationWorkflowName = "BulkPolicyEvaluationWorkflow"

// BulkPolicyEvaluationWorkflowInput represents input for bulk policy evaluation
type BulkPolicyEvaluationWorkflowInput struct {
	EvaluationRequests []*activities.EvaluatePoliciesActivityInput `json:"evaluation_requests"`
	CacheTTLMinutes    int32                                       `json:"cache_ttl_minutes"`
	MaxConcurrency     int                                         `json:"max_concurrency"`
	SkipCache          bool                                        `json:"skip_cache"`
}

// BulkPolicyEvaluationWorkflowOutput represents output of bulk policy evaluation
type BulkPolicyEvaluationWorkflowOutput struct {
	Results         []*PolicyEvaluationWorkflowOutput `json:"results"`
	TotalRequests   int                               `json:"total_requests"`
	SuccessfulCount int                               `json:"successful_count"`
	FailedCount     int                               `json:"failed_count"`
	TotalTimeMS     int64                             `json:"total_time_ms"`
}

// BulkPolicyEvaluationWorkflow processes multiple policy evaluations concurrently
func BulkPolicyEvaluationWorkflow(ctx workflow.Context, input BulkPolicyEvaluationWorkflowInput) (*BulkPolicyEvaluationWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)

	logger.Info("Starting bulk policy evaluation workflow",
		"request_count", len(input.EvaluationRequests),
		"max_concurrency", input.MaxConcurrency)

	if input.MaxConcurrency <= 0 {
		input.MaxConcurrency = 10 // Default concurrency
	}

	results := make([]*PolicyEvaluationWorkflowOutput, len(input.EvaluationRequests))
	successCount := 0
	failedCount := 0

	// Create a semaphore to limit concurrency
	sem := workflow.NewBufferedChannel(ctx, input.MaxConcurrency)
	futures := make([]workflow.Future, len(input.EvaluationRequests))

	// Start concurrent evaluations
	for i, evalRequest := range input.EvaluationRequests {
		// Acquire semaphore
		workflow.Go(ctx, func(ctx workflow.Context) {
			sem.Send(ctx, true)
			defer func() { sem.Receive(ctx, nil) }()

			// Create child workflow input
			childInput := PolicyEvaluationWorkflowInput{
				EvaluationRequest: evalRequest,
				CacheTTLMinutes:   input.CacheTTLMinutes,
				SkipCache:         input.SkipCache,
			}

			// Execute child workflow
			childOptions := workflow.ChildWorkflowOptions{
				WorkflowID: "policy-eval-" + evalRequest.RequestID,
			}
			childCtx := workflow.WithChildOptions(ctx, childOptions)

			future := workflow.ExecuteChildWorkflow(childCtx, PolicyEvaluationWorkflowName, childInput)
			futures[i] = future
		})
	}

	// Wait for all evaluations to complete
	for i, future := range futures {
		var result *PolicyEvaluationWorkflowOutput
		err := future.Get(ctx, &result)
		if err != nil {
			logger.Error("Bulk evaluation failed for request",
				"index", i,
				"error", err)
			failedCount++

			// Create error result
			results[i] = &PolicyEvaluationWorkflowOutput{
				Decision:         types.PolicyDecisionDeny,
				PolicyDecisions:  []*activities.PolicyDecisionResult{},
				EvaluationTimeMS: 0,
				CacheHit:         false,
				RequestID:        input.EvaluationRequests[i].RequestID,
			}
		} else {
			results[i] = result
			successCount++
		}
	}

	totalTime := workflow.Now(ctx).Sub(startTime)

	logger.Info("Bulk policy evaluation workflow completed",
		"total_requests", len(input.EvaluationRequests),
		"successful_count", successCount,
		"failed_count", failedCount,
		"total_time_ms", totalTime.Milliseconds())

	return &BulkPolicyEvaluationWorkflowOutput{
		Results:         results,
		TotalRequests:   len(input.EvaluationRequests),
		SuccessfulCount: successCount,
		FailedCount:     failedCount,
		TotalTimeMS:     totalTime.Milliseconds(),
	}, nil
}

// Helper functions

// generateContextHash creates a simple hash of the attributes for caching
func generateContextHash(attributes map[string]any) string {
	// Simple hash generation - in production, use proper cryptographic hash
	hash := ""
	for k, v := range attributes {
		hash += k + ":" + fmt.Sprintf("%v", v)
	}
	return hash
}

// convertToWorkflowPolicyDecisions converts repository policy decisions to workflow format
func convertToWorkflowPolicyDecisions(decisions []*models.PolicyDecision) []*activities.PolicyDecisionResult {
	results := make([]*activities.PolicyDecisionResult, len(decisions))
	for i, decision := range decisions {
		results[i] = &activities.PolicyDecisionResult{
			PolicyID:      decision.PolicyID,
			Decision:      decision.Decision,
			Reason:        decision.Reason,
			MatchedRule:   decision.MatchedRule,
			EvaluationMS:  decision.EvaluationMS,
			TargetMatched: decision.TargetMatched,
		}
	}
	return results
}
