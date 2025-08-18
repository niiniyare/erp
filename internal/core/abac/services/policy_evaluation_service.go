package services

//go:generate go run go.uber.org/mock/mockgen -source=policy_evaluation_service.go -destination=mock.go -package=services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// PolicyEvaluationService defines the interface for policy evaluation
type PolicyEvaluationService interface {
	// Core evaluation methods
	EvaluatePolicy(ctx context.Context, policy *models.Policy, attrContext *models.AttributeContext) (*PolicyEvaluationResult, error)
	EvaluatePolicySet(ctx context.Context, policies []*models.Policy, attrContext *models.AttributeContext) (*PolicySetEvaluationResult, error)

	// Rule evaluation
	EvaluateRule(ctx context.Context, rule *models.PolicyRule, attrContext *models.AttributeContext) (*RuleEvaluationResult, error)
	EvaluateTarget(ctx context.Context, target *models.PolicyTarget, attrContext *models.AttributeContext) (bool, error)
	EvaluateCondition(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error)

	// Decision methods
	MakeAccessDecision(ctx context.Context, req *AccessDecisionRequest) (*AccessDecisionResponse, error)

	// Evaluation context
	CreateEvaluationContext(ctx context.Context, req *EvaluationContextRequest) (*models.AttributeContext, error)

	// Cache management
	CacheEvaluationResult(ctx context.Context, cacheKey string, result *PolicyEvaluationResult, ttl time.Duration) error
	GetCachedEvaluationResult(ctx context.Context, cacheKey string) (*PolicyEvaluationResult, error)
}

// PolicyEvaluationResult represents the result of a single policy evaluation
type PolicyEvaluationResult struct {
	PolicyID       uuid.UUID                  `json:"policy_id"`
	PolicyName     string                     `json:"policy_name"`
	Decision       types.PolicyDecisionType   `json:"decision"`
	Applicable     bool                       `json:"applicable"`
	RuleResults    []*RuleEvaluationResult    `json:"rule_results"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice         []*models.PolicyAdvice     `json:"advice,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	EvaluatedAt    time.Time                  `json:"evaluated_at"`
	ErrorDetails   *string                    `json:"error_details,omitempty"`
	Context        map[string]any             `json:"context,omitempty"`
}

// PolicySetEvaluationResult represents the result of evaluating multiple policies
type PolicySetEvaluationResult struct {
	FinalDecision      types.PolicyDecisionType       `json:"final_decision"`
	CombiningAlgorithm types.PolicyCombiningAlgorithm `json:"combining_algorithm"`
	PolicyResults      []*PolicyEvaluationResult      `json:"policy_results"`
	ApplicablePolicies int                            `json:"applicable_policies"`
	TotalPolicies      int                            `json:"total_policies"`
	Obligations        []*models.PolicyObligation     `json:"obligations,omitempty"`
	Advice             []*models.PolicyAdvice         `json:"advice,omitempty"`
	EvaluationTime     time.Duration                  `json:"evaluation_time"`
	EvaluatedAt        time.Time                      `json:"evaluated_at"`
	ConflictResolution *ConflictResolutionDetails     `json:"conflict_resolution,omitempty"`
}

// RuleEvaluationResult represents the result of a single rule evaluation
type RuleEvaluationResult struct {
	RuleID         string                     `json:"rule_id"`
	RuleName       string                     `json:"rule_name"`
	Decision       types.PolicyDecisionType   `json:"decision"`
	Applicable     bool                       `json:"applicable"`
	TargetMatch    bool                       `json:"target_match"`
	ConditionMatch bool                       `json:"condition_match"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice         []*models.PolicyAdvice     `json:"advice,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	ErrorDetails   *string                    `json:"error_details,omitempty"`
}

// AccessDecisionRequest represents a request for access decision
type AccessDecisionRequest struct {
	UserID          uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType    string         `json:"resource_type" validate:"required"`
	ResourceID      *uuid.UUID     `json:"resource_id,omitempty"`
	Action          string         `json:"action" validate:"required"`
	EntityID        *uuid.UUID     `json:"entity_id,omitempty"`
	SessionData     map[string]any `json:"session_data,omitempty"`
	EnvironmentData map[string]any `json:"environment_data,omitempty"`
	RequestID       string         `json:"request_id"`
	CacheResults    bool           `json:"cache_results"`
	UseCache        bool           `json:"use_cache"`
}

// AccessDecisionResponse represents the response to an access decision request
type AccessDecisionResponse struct {
	Decision       types.PolicyDecisionType   `json:"decision"`
	Applicable     bool                       `json:"applicable"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice         []*models.PolicyAdvice     `json:"advice,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	EvaluatedAt    time.Time                  `json:"evaluated_at"`
	RequestID      string                     `json:"request_id"`
	CacheHit       bool                       `json:"cache_hit"`
	PolicyCount    int                        `json:"policy_count"`
	ErrorDetails   *string                    `json:"error_details,omitempty"`
}

// EvaluationContextRequest represents a request to create evaluation context
type EvaluationContextRequest struct {
	UserID          uuid.UUID      `json:"user_id"`
	ResourceType    string         `json:"resource_type"`
	ResourceID      *uuid.UUID     `json:"resource_id,omitempty"`
	Action          string         `json:"action"`
	EntityID        *uuid.UUID     `json:"entity_id,omitempty"`
	SessionData     map[string]any `json:"session_data,omitempty"`
	EnvironmentData map[string]any `json:"environment_data,omitempty"`
	IncludeExpired  bool           `json:"include_expired"`
}

// ConflictResolutionDetails provides details about conflict resolution
type ConflictResolutionDetails struct {
	ConflictDetected    bool     `json:"conflict_detected"`
	ConflictingPolicies []string `json:"conflicting_policies,omitempty"`
	ResolutionStrategy  string   `json:"resolution_strategy"`
	ResolutionReason    string   `json:"resolution_reason"`
}

// policyEvaluationService implements PolicyEvaluationService
type policyEvaluationService struct {
	policyRepo       repository.PolicyRepository
	attrCollService  AttributeCollectionService
	attrValidService AttributeValidationService
	attrCacheService AttributeCacheService
	cache            cache.Service
	tracing          tracing.TracingService
	metrics          metrics.MetricsProvider
	logger           logger.Logger
}

// NewPolicyEvaluationService creates a new policy evaluation service
func NewPolicyEvaluationService(
	policyRepo repository.PolicyRepository,
	attrCollService AttributeCollectionService,
	attrValidService AttributeValidationService,
	attrCacheService AttributeCacheService,
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) PolicyEvaluationService {
	return &policyEvaluationService{
		policyRepo:       policyRepo,
		attrCollService:  attrCollService,
		attrValidService: attrValidService,
		attrCacheService: attrCacheService,
		cache:            cache,
		tracing:          tracing,
		metrics:          metrics,
		logger:           logger,
	}
}

// EvaluatePolicy evaluates a single policy against an attribute context
func (s *policyEvaluationService) EvaluatePolicy(ctx context.Context, policy *models.Policy, attrContext *models.AttributeContext) (*PolicyEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.EvaluatePolicy",
		tracing.WithAttributes(
			attribute.String("policy.id", policy.ID.String()),
			attribute.String("policy.name", policy.Name),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordEvaluationMetrics(ctx, "evaluate_policy", "success", time.Since(startTime))
	}()

	result := &PolicyEvaluationResult{
		PolicyID:    policy.ID,
		PolicyName:  policy.Name,
		EvaluatedAt: time.Now(),
		Context:     make(map[string]any),
	}

	// Check if policy target matches
	if policy.Target != nil {
		// Convert map[string]any to PolicyTarget
		policyTarget := &models.PolicyTarget{}
		if resources, ok := policy.Target["resources"].([]any); ok {
			for _, r := range resources {
				if res, ok := r.(string); ok {
					policyTarget.Resources = append(policyTarget.Resources, res)
				}
			}
		}
		if actions, ok := policy.Target["actions"].([]any); ok {
			for _, a := range actions {
				if act, ok := a.(string); ok {
					policyTarget.Actions = append(policyTarget.Actions, act)
				}
			}
		}
		if subjects, ok := policy.Target["subjects"].([]any); ok {
			for _, s := range subjects {
				if subj, ok := s.(string); ok {
					policyTarget.Subjects = append(policyTarget.Subjects, subj)
				}
			}
		}
		if env, ok := policy.Target["environment"].(map[string]any); ok {
			policyTarget.Environment = env
		}

		targetMatch, err := s.EvaluateTarget(ctx, policyTarget, attrContext)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Target evaluation failed")
			errorMsg := err.Error()
			result.ErrorDetails = &errorMsg
			result.Decision = types.PolicyDecisionNotApplicable
			result.EvaluationTime = time.Since(startTime)
			return result, nil
		}

		if !targetMatch {
			result.Decision = types.PolicyDecisionNotApplicable
			result.Applicable = false
			result.EvaluationTime = time.Since(startTime)
			return result, nil
		}
	}

	result.Applicable = true

	// Evaluate all rules in the policy
	var ruleResults []*RuleEvaluationResult
	var allowRules, denyRules int

	// Evaluate the policy rule (policy.Rule is a map[string]any)
	// For now, we'll create a simple rule evaluation based on the rule map
	if policy.Rule != nil {
		// Create a basic rule evaluation result based on the policy effect
		var decision types.PolicyDecisionType
		if policy.Effect == types.PolicyEffectAllow {
			decision = types.PolicyDecisionAllow
		} else if policy.Effect == types.PolicyEffectDeny {
			decision = types.PolicyDecisionDeny
		} else {
			decision = types.PolicyDecisionNotApplicable
		}

		ruleResult := &RuleEvaluationResult{
			RuleID:         "main_rule",
			Applicable:     true,
			Decision:       decision,
			TargetMatch:    true,
			ConditionMatch: true,
			EvaluationTime: time.Since(time.Now()), // This will be 0, but correct type
		}

		ruleResults = append(ruleResults, ruleResult)

		// Count rule decisions for combining algorithm
		if ruleResult.Applicable {
			switch ruleResult.Decision {
			case types.PolicyDecisionAllow:
				allowRules++
			case types.PolicyDecisionDeny:
				denyRules++
			}
		}
	}

	result.RuleResults = ruleResults

	// Apply combining algorithm to determine final decision
	result.Decision = s.applyCombiningAlgorithm(policy.CombiningAlgorithm, ruleResults)

	// Collect obligations and advice from applicable rules
	s.collectObligationsAndAdvice(result, ruleResults)

	result.EvaluationTime = time.Since(startTime)

	s.logger.InfoContext(ctx, "Policy evaluation completed",
		logger.Fields{
			"policy_id":       policy.ID,
			"decision":        string(result.Decision),
			"applicable":      result.Applicable,
			"rule_count":      len(ruleResults),
			"evaluation_time": result.EvaluationTime.Milliseconds(),
		})

	return result, nil
}

// EvaluatePolicySet evaluates multiple policies and combines their results
func (s *policyEvaluationService) EvaluatePolicySet(ctx context.Context, policies []*models.Policy, attrContext *models.AttributeContext) (*PolicySetEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.EvaluatePolicySet",
		tracing.WithAttributes(attribute.Int("policies.count", len(policies))))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordEvaluationMetrics(ctx, "evaluate_policy_set", "success", time.Since(startTime))
	}()

	result := &PolicySetEvaluationResult{
		TotalPolicies: len(policies),
		EvaluatedAt:   time.Now(),
		PolicyResults: make([]*PolicyEvaluationResult, 0, len(policies)),
	}

	// Determine combining algorithm (for now, use first policy's algorithm or default)
	combiningAlgorithm := types.CombiningAlgorithmDenyOverrides
	if len(policies) > 0 && policies[0].CombiningAlgorithm != "" {
		combiningAlgorithm = policies[0].CombiningAlgorithm
	}
	result.CombiningAlgorithm = combiningAlgorithm

	// Evaluate each policy
	var applicablePolicyResults []*PolicyEvaluationResult
	for _, policy := range policies {
		policyResult, err := s.EvaluatePolicy(ctx, policy, attrContext)
		if err != nil {
			s.logger.WarnContext(ctx, "Policy evaluation failed",
				logger.Fields{"policy_id": policy.ID, "error": err.Error()})
			continue
		}

		result.PolicyResults = append(result.PolicyResults, policyResult)

		if policyResult.Applicable {
			applicablePolicyResults = append(applicablePolicyResults, policyResult)
			result.ApplicablePolicies++
		}
	}

	// Apply policy set combining algorithm
	result.FinalDecision = s.applyPolicySetCombiningAlgorithm(combiningAlgorithm, applicablePolicyResults)

	// Detect and resolve conflicts
	conflicts := s.detectPolicyConflicts(applicablePolicyResults)
	if len(conflicts) > 0 {
		result.ConflictResolution = &ConflictResolutionDetails{
			ConflictDetected:    true,
			ConflictingPolicies: conflicts,
			ResolutionStrategy:  string(combiningAlgorithm),
			ResolutionReason:    "Applied combining algorithm to resolve conflicts",
		}
	}

	// Collect obligations and advice from applicable policies
	s.collectPolicySetObligationsAndAdvice(result, applicablePolicyResults)

	result.EvaluationTime = time.Since(startTime)

	s.logger.InfoContext(ctx, "Policy set evaluation completed",
		logger.Fields{
			"total_policies":      result.TotalPolicies,
			"applicable_policies": result.ApplicablePolicies,
			"final_decision":      string(result.FinalDecision),
			"evaluation_time":     result.EvaluationTime.Milliseconds(),
			"conflicts_detected":  result.ConflictResolution != nil,
		})

	return result, nil
}

// EvaluateRule evaluates a single policy rule
func (s *policyEvaluationService) EvaluateRule(ctx context.Context, rule *models.PolicyRule, attrContext *models.AttributeContext) (*RuleEvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.EvaluateRule",
		tracing.WithAttributes(attribute.String("rule.id", rule.ID)))
	defer span.End()

	startTime := time.Now()

	result := &RuleEvaluationResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
	}

	// Evaluate rule attributes (simplified target matching)
	if rule.Attributes != nil {
		// Create a simple PolicyTarget from attributes
		policyTarget := &models.PolicyTarget{}
		if resources, ok := rule.Attributes["resources"].([]any); ok {
			for _, r := range resources {
				if res, ok := r.(string); ok {
					policyTarget.Resources = append(policyTarget.Resources, res)
				}
			}
		}
		targetMatch, err := s.EvaluateTarget(ctx, policyTarget, attrContext)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Rule target evaluation failed")
			errorMsg := err.Error()
			result.ErrorDetails = &errorMsg
			result.Decision = types.PolicyDecisionNotApplicable
			result.EvaluationTime = time.Since(startTime)
			return result, nil
		}
		result.TargetMatch = targetMatch

		if !targetMatch {
			result.Decision = types.PolicyDecisionNotApplicable
			result.Applicable = false
			result.EvaluationTime = time.Since(startTime)
			return result, nil
		}
	} else {
		result.TargetMatch = true
	}

	// Evaluate rule conditions (if any)
	if len(rule.Conditions) > 0 {
		// For now, assume all conditions pass (simplified implementation)
		result.ConditionMatch = true
	} else {
		result.ConditionMatch = true
	}

	// Rule is applicable - for now, assume it allows (simplified implementation)
	result.Applicable = true
	result.Decision = types.PolicyDecisionAllow
	// No obligations or advice in simplified rule structure
	result.EvaluationTime = time.Since(startTime)

	return result, nil
}

// EvaluateTarget evaluates a policy target against the attribute context
func (s *policyEvaluationService) EvaluateTarget(ctx context.Context, target *models.PolicyTarget, attrContext *models.AttributeContext) (bool, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.EvaluateTarget")
	defer span.End()

	// Evaluate subjects (users)
	if len(target.Subjects) > 0 {
		subjectMatch := false
		for _, subject := range target.Subjects {
			// Convert string to AttributeMatch for matching
			attrMatch := &models.AttributeMatch{
				AttributeName: "user_id",
				MatchType:     "equals",
				Value:         subject,
				CaseSensitive: true,
			}
			if s.matchAttributeValue(attrMatch, attrContext.UserAttributes) {
				subjectMatch = true
				break
			}
		}
		if !subjectMatch {
			return false, nil
		}
	}

	// Evaluate resources
	if len(target.Resources) > 0 {
		resourceMatch := false
		for _, resource := range target.Resources {
			// Convert string to AttributeMatch for matching
			attrMatch := &models.AttributeMatch{
				AttributeName: "resource_type",
				MatchType:     "equals",
				Value:         resource,
				CaseSensitive: true,
			}
			if s.matchAttributeValue(attrMatch, attrContext.ResourceAttributes) {
				resourceMatch = true
				break
			}
		}
		if !resourceMatch {
			return false, nil
		}
	}

	// Evaluate actions
	if len(target.Actions) > 0 {
		actionMatch := false
		for _, action := range target.Actions {
			// Convert string to AttributeMatch for matching
			attrMatch := &models.AttributeMatch{
				AttributeName: "action",
				MatchType:     "equals",
				Value:         action,
				CaseSensitive: true,
			}
			if s.matchAttributeValue(attrMatch, attrContext.ActionAttributes) {
				actionMatch = true
				break
			}
		}
		if !actionMatch {
			return false, nil
		}
	}

	// Evaluate environment (simplified - check if environment attributes match)
	if target.Environment != nil && len(target.Environment) > 0 {
		// For now, assume environment matches (simplified implementation)
		// In a full implementation, this would check environment conditions
	}

	return true, nil
}

// EvaluateCondition evaluates a policy condition
func (s *policyEvaluationService) EvaluateCondition(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.EvaluateCondition")
	defer span.End()

	switch condition.Type {
	case "attribute_comparison":
		return s.evaluateAttributeComparison(ctx, condition, attrContext)
	case "logical_expression":
		return s.evaluateLogicalExpression(ctx, condition, attrContext)
	case "function_call":
		return s.evaluateFunctionCall(ctx, condition, attrContext)
	case "custom_script":
		return s.evaluateCustomScript(ctx, condition, attrContext)
	default:
		return false, errors.NewBusinessError("UNSUPPORTED_CONDITION_TYPE", "Unsupported condition type").
			WithDetail("condition_type", condition.Type)
	}
}

// MakeAccessDecision makes an access decision based on the request
func (s *policyEvaluationService) MakeAccessDecision(ctx context.Context, req *AccessDecisionRequest) (*AccessDecisionResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.MakeAccessDecision",
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.String("resource.type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordEvaluationMetrics(ctx, "make_access_decision", "success", time.Since(startTime))
	}()

	response := &AccessDecisionResponse{
		RequestID:   req.RequestID,
		EvaluatedAt: time.Now(),
	}

	// Check cache first if enabled
	if req.UseCache {
		cacheKey := s.generateCacheKey(req)
		if cachedResult, err := s.GetCachedEvaluationResult(ctx, cacheKey); err == nil {
			response.Decision = cachedResult.Decision
			response.Applicable = cachedResult.Applicable
			response.Obligations = cachedResult.Obligations
			response.Advice = cachedResult.Advice
			response.CacheHit = true
			response.EvaluationTime = time.Since(startTime)
			return response, nil
		}
	}

	// Create evaluation context
	contextReq := &EvaluationContextRequest{
		UserID:          req.UserID,
		ResourceType:    req.ResourceType,
		ResourceID:      req.ResourceID,
		Action:          req.Action,
		EntityID:        req.EntityID,
		SessionData:     req.SessionData,
		EnvironmentData: req.EnvironmentData,
	}

	attrContext, err := s.CreateEvaluationContext(ctx, contextReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create evaluation context")
		errorMsg := err.Error()
		response.ErrorDetails = &errorMsg
		response.Decision = types.PolicyDecisionDeny
		response.EvaluationTime = time.Since(startTime)
		return response, nil
	}

	// Get applicable policies
	policies, err := s.policyRepo.GetApplicablePolicies(ctx, req.ResourceType, req.Action)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get applicable policies")
		errorMsg := err.Error()
		response.ErrorDetails = &errorMsg
		response.Decision = types.PolicyDecisionDeny
		response.EvaluationTime = time.Since(startTime)
		return response, nil
	}

	response.PolicyCount = len(policies)

	if len(policies) == 0 {
		response.Decision = types.PolicyDecisionNotApplicable
		response.Applicable = false
		response.EvaluationTime = time.Since(startTime)
		return response, nil
	}

	// Evaluate policy set
	policySetResult, err := s.EvaluatePolicySet(ctx, policies, attrContext)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Policy set evaluation failed")
		errorMsg := err.Error()
		response.ErrorDetails = &errorMsg
		response.Decision = types.PolicyDecisionDeny
		response.EvaluationTime = time.Since(startTime)
		return response, nil
	}

	response.Decision = policySetResult.FinalDecision
	response.Applicable = policySetResult.ApplicablePolicies > 0
	response.Obligations = policySetResult.Obligations
	response.Advice = policySetResult.Advice
	response.EvaluationTime = time.Since(startTime)

	// Cache result if enabled
	if req.CacheResults && response.Decision != types.PolicyDecisionNotApplicable {
		cacheKey := s.generateCacheKey(req)
		evalResult := &PolicyEvaluationResult{
			Decision:       response.Decision,
			Applicable:     response.Applicable,
			Obligations:    response.Obligations,
			Advice:         response.Advice,
			EvaluationTime: response.EvaluationTime,
			EvaluatedAt:    response.EvaluatedAt,
		}
		_ = s.CacheEvaluationResult(ctx, cacheKey, evalResult, 15*time.Minute)
	}

	s.logger.InfoContext(ctx, "Access decision completed",
		logger.Fields{
			"user_id":         req.UserID,
			"resource_type":   req.ResourceType,
			"action":          req.Action,
			"decision":        string(response.Decision),
			"policy_count":    response.PolicyCount,
			"cache_hit":       response.CacheHit,
			"evaluation_time": response.EvaluationTime.Milliseconds(),
		})

	return response, nil
}

// CreateEvaluationContext creates an attribute context for evaluation
func (s *policyEvaluationService) CreateEvaluationContext(ctx context.Context, req *EvaluationContextRequest) (*models.AttributeContext, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyEvaluationService.CreateEvaluationContext")
	defer span.End()

	// Collect all required attributes
	collectionReq := &AttributeCollectionRequest{
		UserID:          req.UserID,
		ResourceType:    req.ResourceType,
		ResourceID:      req.ResourceID,
		Action:          req.Action,
		EntityID:        req.EntityID,
		SessionData:     req.SessionData,
		EnvironmentData: req.EnvironmentData,
		CollectExpired:  req.IncludeExpired,
	}

	attrContext, err := s.attrCollService.CollectAllAttributes(ctx, collectionReq)
	if err != nil {
		return nil, err
	}

	// Validate collected attributes
	validationErrors, err := s.attrValidService.ValidateAttributeContext(ctx, attrContext)
	if err != nil {
		return nil, err
	}

	if len(validationErrors) > 0 {
		s.logger.WarnContext(ctx, "Attribute validation warnings during evaluation context creation",
			logger.Fields{"validation_errors": validationErrors})
	}

	return attrContext, nil
}

// Helper methods implementation continues in next part...

// Helper methods for policy evaluation

func (s *policyEvaluationService) applyCombiningAlgorithm(algorithm types.PolicyCombiningAlgorithm, ruleResults []*RuleEvaluationResult) types.PolicyDecisionType {
	var allowCount, denyCount int

	for _, result := range ruleResults {
		if !result.Applicable {
			continue
		}
		switch result.Decision {
		case types.PolicyDecisionAllow:
			allowCount++
		case types.PolicyDecisionDeny:
			denyCount++
		}
	}

	switch algorithm {
	case types.CombiningAlgorithmDenyOverrides:
		if denyCount > 0 {
			return types.PolicyDecisionDeny
		}
		if allowCount > 0 {
			return types.PolicyDecisionAllow
		}
		return types.PolicyDecisionNotApplicable

	case types.CombiningAlgorithmPermitOverrides:
		if allowCount > 0 {
			return types.PolicyDecisionAllow
		}
		if denyCount > 0 {
			return types.PolicyDecisionDeny
		}
		return types.PolicyDecisionNotApplicable

	case types.CombiningAlgorithmFirstApplicable:
		for _, result := range ruleResults {
			if result.Applicable {
				return result.Decision
			}
		}
		return types.PolicyDecisionNotApplicable

	default:
		return types.PolicyDecisionDeny
	}
}

func (s *policyEvaluationService) applyPolicySetCombiningAlgorithm(algorithm types.PolicyCombiningAlgorithm, policyResults []*PolicyEvaluationResult) types.PolicyDecisionType {
	var allowCount, denyCount int

	for _, result := range policyResults {
		if !result.Applicable {
			continue
		}
		switch result.Decision {
		case types.PolicyDecisionAllow:
			allowCount++
		case types.PolicyDecisionDeny:
			denyCount++
		}
	}

	switch algorithm {
	case types.CombiningAlgorithmDenyOverrides:
		if denyCount > 0 {
			return types.PolicyDecisionDeny
		}
		if allowCount > 0 {
			return types.PolicyDecisionAllow
		}
		return types.PolicyDecisionNotApplicable

	case types.CombiningAlgorithmPermitOverrides:
		if allowCount > 0 {
			return types.PolicyDecisionAllow
		}
		if denyCount > 0 {
			return types.PolicyDecisionDeny
		}
		return types.PolicyDecisionNotApplicable

	default:
		return types.PolicyDecisionDeny
	}
}

func (s *policyEvaluationService) detectPolicyConflicts(policyResults []*PolicyEvaluationResult) []string {
	var conflicts []string
	var allowPolicies, denyPolicies []string

	for _, result := range policyResults {
		if !result.Applicable {
			continue
		}
		switch result.Decision {
		case types.PolicyDecisionAllow:
			allowPolicies = append(allowPolicies, result.PolicyName)
		case types.PolicyDecisionDeny:
			denyPolicies = append(denyPolicies, result.PolicyName)
		}
	}

	if len(allowPolicies) > 0 && len(denyPolicies) > 0 {
		conflicts = append(conflicts, allowPolicies...)
		conflicts = append(conflicts, denyPolicies...)
	}

	return conflicts
}

func (s *policyEvaluationService) collectObligationsAndAdvice(result *PolicyEvaluationResult, ruleResults []*RuleEvaluationResult) {
	for _, ruleResult := range ruleResults {
		if ruleResult.Applicable && ruleResult.Decision == types.PolicyDecisionAllow {
			result.Obligations = append(result.Obligations, ruleResult.Obligations...)
			result.Advice = append(result.Advice, ruleResult.Advice...)
		}
	}
}

func (s *policyEvaluationService) collectPolicySetObligationsAndAdvice(result *PolicySetEvaluationResult, policyResults []*PolicyEvaluationResult) {
	for _, policyResult := range policyResults {
		if policyResult.Applicable && policyResult.Decision == types.PolicyDecisionAllow {
			result.Obligations = append(result.Obligations, policyResult.Obligations...)
			result.Advice = append(result.Advice, policyResult.Advice...)
		}
	}
}

func (s *policyEvaluationService) matchAttributeValue(target *models.AttributeMatch, attributes map[string]*models.AttributeValue) bool {
	attr, exists := attributes[target.AttributeName]
	if !exists {
		return false
	}

	switch target.MatchType {
	case "equals":
		return fmt.Sprintf("%v", attr.Value) == fmt.Sprintf("%v", target.Value)
	case "contains":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", attr.Value)), strings.ToLower(fmt.Sprintf("%v", target.Value)))
	case "regex":
		// Implement regex matching
		return false
	case "in":
		// Implement "in" matching for arrays
		return false
	default:
		return false
	}
}

// Evaluation method implementations

func (s *policyEvaluationService) evaluateAttributeComparison(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error) {
	// TODO: Implement attribute comparison logic
	return true, nil
}

func (s *policyEvaluationService) evaluateLogicalExpression(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error) {
	// TODO: Implement logical expression evaluation
	return true, nil
}

func (s *policyEvaluationService) evaluateFunctionCall(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error) {
	// TODO: Implement function call evaluation
	return true, nil
}

func (s *policyEvaluationService) evaluateCustomScript(ctx context.Context, condition *models.PolicyCondition, attrContext *models.AttributeContext) (bool, error) {
	// TODO: Implement custom script evaluation
	return true, nil
}

// Cache operations

func (s *policyEvaluationService) CacheEvaluationResult(ctx context.Context, cacheKey string, result *PolicyEvaluationResult, ttl time.Duration) error {
	return s.cache.Set(ctx, cacheKey, result, ttl)
}

func (s *policyEvaluationService) GetCachedEvaluationResult(ctx context.Context, cacheKey string) (*PolicyEvaluationResult, error) {
	var result PolicyEvaluationResult
	if err := s.cache.Get(ctx, cacheKey, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *policyEvaluationService) generateCacheKey(req *AccessDecisionRequest) string {
	keyData := fmt.Sprintf("policy_eval:%s:%s:%s", req.UserID, req.ResourceType, req.Action)
	if req.ResourceID != nil {
		keyData += ":" + req.ResourceID.String()
	}
	return keyData
}

// recordEvaluationMetrics records evaluation operation metrics
func (s *policyEvaluationService) recordEvaluationMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	// Evaluation operation counter
	counter := s.metrics.Counter(
		"abac_policy_evaluation_operations_total",
		"Total number of policy evaluation operations",
		"operation", "status",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	// Evaluation operation duration histogram
	histogram := s.metrics.Histogram(
		"abac_policy_evaluation_operation_duration_seconds",
		"Duration of policy evaluation operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation", "status",
	)

	histogram.Observe(duration.Seconds(), metrics.Fields{
		"operation": operation,
		"status":    status,
	})
}
