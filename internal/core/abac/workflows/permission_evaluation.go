package workflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"

	"erp/internal/core/abac"
	"erp/internal/shared/errors"
)

// PermissionEvaluationWorkflow is a Temporal workflow that orchestrates the ABAC permission evaluation process.
func PermissionEvaluationWorkflow(ctx workflow.Context, request abac.PermissionEvaluationWorkflowRequest) (*abac.PermissionEvaluationWorkflowResponse, error) {
	// Set up workflow options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				string(errors.CategoryValidation),
				string(errors.CategorySecurity),
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Initialize response
	response := &abac.PermissionEvaluationWorkflowResponse{
		Allowed: false,
		PolicyDecisions: []abac.PolicyDecision{},
		EffectiveRoles:  []string{},
	}

	startTime := workflow.Now(ctx)

	// 1. Collect User Attributes Activity
	var userAttrsOutput abac.CollectUserAttributesActivityOutput
	userAttrsInput := abac.CollectUserAttributesActivityInput{
		UserID:   request.UserID,
		TenantID: request.TenantID,
	}
	if err := workflow.ExecuteActivity(ctx, "CollectUserAttributesActivity", userAttrsInput).Get(ctx, &userAttrsOutput); err != nil {
		response.Error = fmt.Sprintf("Failed to collect user attributes: %v", err)
		return response, err
	}

	// 2. Collect Resource Attributes Activity
	var resourceAttrsOutput abac.CollectResourceAttributesActivityOutput
	resourceAttrsInput := abac.CollectResourceAttributesActivityInput{
		ResourceID: request.ResourceID,
		TenantID:   request.TenantID,
	}
	if err := workflow.ExecuteActivity(ctx, "CollectResourceAttributesActivity", resourceAttrsInput).Get(ctx, &resourceAttrsOutput); err != nil {
		response.Error = fmt.Sprintf("Failed to collect resource attributes: %v", err)
		return response, err
	}

	// 3. Build Environment Context Activity
	var envContextOutput abac.BuildEnvironmentContextActivityOutput
	envContextInput := abac.BuildEnvironmentContextActivityInput{
		// Populate with data from request or other sources if available
		Timestamp: workflow.Now(ctx),
	}
	if err := workflow.ExecuteActivity(ctx, "BuildEnvironmentContextActivity", envContextInput).Get(ctx, &envContextOutput); err != nil {
		response.Error = fmt.Sprintf("Failed to build environment context: %v", err)
		return response, err
	}

	// Combine all attributes
	allAttributes := make(map[string]interface{})
	for k, v := range userAttrsOutput.Attributes {
		allAttributes["user." + k] = v
	}
	for k, v := range resourceAttrsOutput.Attributes {
		allAttributes["resource." + k] = v
	}
	for k, v := range envContextOutput.Context {
		allAttributes["env." + k] = v
	}
	// Add request context attributes directly
	for k, v := range request.Context {
		allAttributes[k] = v
	}

	// 4. Evaluate Policies Activity
	var policyEvalOutput abac.EvaluatePoliciesActivityOutput
	policyEvalInput := abac.EvaluatePoliciesActivityInput{
		TenantID:   request.TenantID,
		ResourceID: request.ResourceID,
		ActionID:   request.ActionID,
		Attributes: allAttributes,
	}
	if err := workflow.ExecuteActivity(ctx, "EvaluatePoliciesActivity", policyEvalInput).Get(ctx, &policyEvalOutput); err != nil {
		response.Error = fmt.Sprintf("Failed to evaluate policies: %v", err)
		return response, err
	}

	response.Allowed = policyEvalOutput.Allowed
	response.PolicyDecisions = policyEvalOutput.PolicyDecisions

	// 5. Cache Policy Result Activity (fire and forget, or handle errors if critical)
	cacheInput := abac.CachePolicyResultActivityInput{
		TenantID:         request.TenantID,
		UserID:           request.UserID,
		ResourceID:       request.ResourceID,
		ActionID:         request.ActionID,
		ContextHash:      generateContextHash(allAttributes), // Implement this helper
		Decision:         fmt.Sprintf("%v", policyEvalOutput.Allowed), // Convert bool to string
		ApplicablePolicies: extractPolicyIDs(policyEvalOutput.PolicyDecisions),
		EvaluationTimeMS: int(workflow.Since(ctx, startTime).Milliseconds()),
		ExpiresAt:        workflow.Now(ctx).Add(1 * time.Hour), // Cache for 1 hour
	}
	_ = workflow.ExecuteActivity(ctx, "CachePolicyResultActivity", cacheInput).Get(ctx, nil) // Ignore error for caching

	response.EvaluationTimeMS = int(workflow.Since(ctx, startTime).Milliseconds())

	return response, nil
}

// Helper to generate a hash from context attributes (simplified for example)
func generateContextHash(attrs map[string]interface{}) string {
	// In a real scenario, use a cryptographic hash (e.g., SHA256) of a canonical JSON representation
	// For simplicity, this is a placeholder.
	return fmt.Sprintf("hash_%v", len(attrs))
}

// Helper to extract policy IDs from policy decisions
func extractPolicyIDs(decisions []abac.PolicyDecision) []uuid.UUID {
	ids := make([]uuid.UUID, len(decisions))
	for i, d := range decisions {
		ids[i] = d.PolicyID
	}
	return ids
}
