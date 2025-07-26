package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/niiniyare/erp/db/sqlc"
	// "github.com/niiniyare/erp/internal/core/abac" // TODO: Fix import path
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// PolicyEvaluationActivities implements the Temporal activities for evaluating policies and caching results.
type PolicyEvaluationActivities struct {
	policyRepo     abac.PolicyRepository
	policyEvalRepo abac.PolicyEvaluationRepository
	logger         logger.Logger
	ruleEngine     PolicyRuleEngine
}

// NewPolicyEvaluationActivities creates a new instance of PolicyEvaluationActivities.
func NewPolicyEvaluationActivities(
	policyRepo abac.PolicyRepository,
	policyEvalRepo abac.PolicyEvaluationRepository,
	logger logger.Logger,
) *PolicyEvaluationActivities {
	return &PolicyEvaluationActivities{
		policyRepo:     policyRepo,
		policyEvalRepo: policyEvalRepo,
		logger:         logger,
		ruleEngine:     newSimpleRuleEngine(logger),
	}
}

// EvaluatePoliciesActivity evaluates all applicable policies for a given request.
func (a *PolicyEvaluationActivities) EvaluatePoliciesActivity(ctx context.Context, input abac.EvaluatePoliciesActivityInput) (*abac.EvaluatePoliciesActivityOutput, error) {
	a.logger.InfoContext(ctx, "Starting EvaluatePoliciesActivity", logger.Fields{"resource_id": input.ResourceID, "action_id": input.ActionID})

	// 1. Load applicable policies
	policies, err := a.policyRepo.GetApplicablePolicies(ctx, input.ResourceID.String(), input.ActionID.String())
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to get applicable policies", logger.Fields{"error": err})
		return nil, errors.WrapSimpleError(err, "POLICY_FETCH_FAILED", 500)
	}

	if len(policies) == 0 {
		a.logger.InfoContext(ctx, "No applicable policies found, defaulting to DENY")
		return &abac.EvaluatePoliciesActivityOutput{Allowed: false}, nil
	}

	// Sort policies by priority (higher priority first)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	var decisions []abac.PolicyDecision
	finalDecision := false // Default to deny

	// 2. Evaluate policies
	for _, policy := range policies {
		decision, err := a.ruleEngine.Evaluate(ctx, policy, input.Attributes)
		if err != nil {
			a.logger.WarnContext(ctx, "Failed to evaluate policy rule", logger.Fields{"policy_id": policy.ID, "error": err})
			continue // Skip policies with evaluation errors
		}
		decisions = append(decisions, *decision)

		// 3. Handle priorities and conflicts: Deny overrides Allow
		if decision.Effect == "DENY" {
			finalDecision = false
			break // A DENY decision is final
		}
		if decision.Effect == "ALLOW" {
			finalDecision = true
		}
	}

	output := &abac.EvaluatePoliciesActivityOutput{
		Allowed:         finalDecision,
		PolicyDecisions: decisions,
	}

	a.logger.InfoContext(ctx, "Finished EvaluatePoliciesActivity", logger.Fields{"decision": finalDecision, "policy_count": len(policies)})
	return output, nil
}

// CachePolicyResultActivity stores the result of a policy evaluation.
func (a *PolicyEvaluationActivities) CachePolicyResultActivity(ctx context.Context, input abac.CachePolicyResultActivityInput) (*abac.CachePolicyResultActivityOutput, error) {
	a.logger.InfoContext(ctx, "Starting CachePolicyResultActivity", logger.Fields{"user_id": input.UserID, "resource_id": input.ResourceID})

	params := sqlc.CreatePolicyEvaluationParams{
		UserID:             input.UserID,
		ResourceID:         input.ResourceID,
		ActionID:           input.ActionID,
		ContextHash:        input.ContextHash,
		Decision:           input.Decision,
		ApplicablePolicies: input.ApplicablePolicies,
		EvaluationTimeMs:   int32(input.EvaluationTimeMS),
		ExpiresAt:          input.ExpiresAt,
	}

	_, err := a.policyEvalRepo.CreatePolicyEvaluation(ctx, params)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to cache policy evaluation", logger.Fields{"error": err})
		// Non-critical error, so we don't return it to the workflow
	}

	return &abac.CachePolicyResultActivityOutput{Success: err == nil}, nil
}

// PolicyRuleEngine defines the interface for a policy rule evaluation engine.
type PolicyRuleEngine interface {
	Evaluate(ctx context.Context, policy sqlc.Policy, attributes map[string]interface{}) (*abac.PolicyDecision, error)
}

// simpleRuleEngine is a basic implementation of a policy rule engine.
type simpleRuleEngine struct {
	logger logger.Logger
}

func newSimpleRuleEngine(logger logger.Logger) *simpleRuleEngine {
	return &simpleRuleEngine{logger: logger}
}

// Evaluate evaluates a policy rule.
func (e *simpleRuleEngine) Evaluate(ctx context.Context, policy sqlc.Policy, attributes map[string]interface{}) (*abac.PolicyDecision, error) {
	var rule map[string]interface{}
	if err := json.Unmarshal(policy.Rule, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy rule: %w", err)
	}

	// Basic evaluation: check for a simple "allow" or "deny" rule.
	// This is a placeholder for a real rule engine.
	if allow, ok := rule["allow"].(bool); ok && allow {
		return &abac.PolicyDecision{PolicyID: policy.ID, PolicyName: policy.Name, Effect: "ALLOW"}, nil
	}
	if deny, ok := rule["deny"].(bool); ok && deny {
		return &abac.PolicyDecision{PolicyID: policy.ID, PolicyName: policy.Name, Effect: "DENY"}, nil
	}

	// Default to not applicable if no simple rule matches
	return &abac.PolicyDecision{PolicyID: policy.ID, PolicyName: policy.Name, Effect: "NOT_APPLICABLE"}, nil
}
