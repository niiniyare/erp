package services

//go:generate go run go.uber.org/mock/mockgen -source=policy_combining_service.go -destination=mock.go -package=services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"awo/internal/core/abac/models"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// PolicyCombiningService defines the interface for policy combining algorithms
type PolicyCombiningService interface {
	// Core combining methods
	CombinePolicyDecisions(ctx context.Context, req *CombiningRequest) (*CombiningResult, error)
	CombineRuleDecisions(ctx context.Context, algorithm types.PolicyCombiningAlgorithm, ruleResults []*RuleEvaluationResult) (*RuleCombiningResult, error)

	// Algorithm implementations
	ApplyDenyOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)
	ApplyAllowOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)
	ApplyFirstApplicable(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)
	ApplyOnlyOneApplicable(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)
	ApplyOrderedDenyOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)
	ApplyOrderedAllowOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error)

	// Conflict resolution
	DetectConflicts(ctx context.Context, decisions []*DecisionInput) (*ConflictAnalysis, error)
	ResolveConflicts(ctx context.Context, conflicts *ConflictAnalysis, algorithm types.PolicyCombiningAlgorithm) (*ConflictResolution, error)

	// Utility methods
	ValidateCombiningAlgorithm(ctx context.Context, algorithm types.PolicyCombiningAlgorithm) error
	GetSupportedAlgorithms(ctx context.Context) ([]types.PolicyCombiningAlgorithm, error)
	ExplainCombiningDecision(ctx context.Context, req *CombiningExplanationRequest) (*CombiningExplanation, error)
}

// CombiningRequest represents a request to combine policy decisions
type CombiningRequest struct {
	Algorithm           types.PolicyCombiningAlgorithm `json:"algorithm" validate:"required"`
	Decisions           []*DecisionInput               `json:"decisions" validate:"required,min=1"`
	IgnoreNotApplicable bool                           `json:"ignore_not_applicable"`
	StrictEvaluation    bool                           `json:"strict_evaluation"`
	CollectObligations  bool                           `json:"collect_obligations"`
	CollectAdvice       bool                           `json:"collect_advice"`
	RequestID           string                         `json:"request_id"`
	Context             map[string]any                 `json:"context,omitempty"`
}

// DecisionInput represents an input decision for combining
type DecisionInput struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Decision    types.PolicyDecisionType   `json:"decision"`
	Applicable  bool                       `json:"applicable"`
	Priority    int                        `json:"priority"`
	Weight      float64                    `json:"weight"`
	Obligations []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice      []*models.PolicyAdvice     `json:"advice,omitempty"`
	Metadata    map[string]any             `json:"metadata,omitempty"`
	EvaluatedAt time.Time                  `json:"evaluated_at"`
}

// CombiningResult represents the result of combining decisions
type CombiningResult struct {
	FinalDecision      types.PolicyDecisionType       `json:"final_decision"`
	Algorithm          types.PolicyCombiningAlgorithm `json:"algorithm"`
	ApplicableCount    int                            `json:"applicable_count"`
	AllowCount         int                            `json:"allow_count"`
	DenyCount          int                            `json:"deny_count"`
	NotApplicableCount int                            `json:"not_applicable_count"`
	Obligations        []*models.PolicyObligation     `json:"obligations,omitempty"`
	Advice             []*models.PolicyAdvice         `json:"advice,omitempty"`
	ConflictDetected   bool                           `json:"conflict_detected"`
	ConflictResolution *ConflictResolution            `json:"conflict_resolution,omitempty"`
	CombiningTrace     []*CombiningStep               `json:"combining_trace,omitempty"`
	ProcessingTime     time.Duration                  `json:"processing_time"`
	ProcessedAt        time.Time                      `json:"processed_at"`
}

// RuleCombiningResult represents the result of combining rule decisions
type RuleCombiningResult struct {
	FinalDecision   types.PolicyDecisionType       `json:"final_decision"`
	Algorithm       types.PolicyCombiningAlgorithm `json:"algorithm"`
	RulesProcessed  int                            `json:"rules_processed"`
	RulesApplicable int                            `json:"rules_applicable"`
	WinningRule     *string                        `json:"winning_rule,omitempty"`
	ProcessingTrace []*RuleCombiningStep           `json:"processing_trace,omitempty"`
}

// ConflictAnalysis represents an analysis of conflicts between decisions
type ConflictAnalysis struct {
	ConflictDetected     bool             `json:"conflict_detected"`
	ConflictType         ConflictType     `json:"conflict_type"`
	ConflictingDecisions []*DecisionInput `json:"conflicting_decisions"`
	AllowDecisions       []*DecisionInput `json:"allow_decisions"`
	DenyDecisions        []*DecisionInput `json:"deny_decisions"`
	ConflictSeverity     ConflictSeverity `json:"conflict_severity"`
	ConflictReasons      []string         `json:"conflict_reasons"`
	AnalysisMetadata     map[string]any   `json:"analysis_metadata,omitempty"`
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	ResolutionMethod    string                   `json:"resolution_method"`
	WinningDecision     *DecisionInput           `json:"winning_decision"`
	ResolvedDecision    types.PolicyDecisionType `json:"resolved_decision"`
	ResolutionReason    string                   `json:"resolution_reason"`
	OverriddenDecisions []*DecisionInput         `json:"overridden_decisions,omitempty"`
	ResolutionMetadata  map[string]any           `json:"resolution_metadata,omitempty"`
}

// CombiningStep represents a step in the combining process
type CombiningStep struct {
	StepNumber  int                      `json:"step_number"`
	Description string                   `json:"description"`
	InputCount  int                      `json:"input_count"`
	Result      types.PolicyDecisionType `json:"result"`
	Rationale   string                   `json:"rationale"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

// RuleCombiningStep represents a step in rule combining
type RuleCombiningStep struct {
	RuleID     string                   `json:"rule_id"`
	RuleName   string                   `json:"rule_name"`
	Decision   types.PolicyDecisionType `json:"decision"`
	Applicable bool                     `json:"applicable"`
	Action     string                   `json:"action"` // "evaluated", "skipped", "overridden"
	Reason     string                   `json:"reason"`
}

// CombiningExplanationRequest represents a request for combining explanation
type CombiningExplanationRequest struct {
	Algorithm   types.PolicyCombiningAlgorithm `json:"algorithm"`
	Decisions   []*DecisionInput               `json:"decisions"`
	DetailLevel string                         `json:"detail_level"` // "basic", "detailed", "verbose"
}

// CombiningExplanation provides an explanation of how decisions were combined
type CombiningExplanation struct {
	Algorithm            types.PolicyCombiningAlgorithm `json:"algorithm"`
	AlgorithmDescription string                         `json:"algorithm_description"`
	FinalDecision        types.PolicyDecisionType       `json:"final_decision"`
	StepByStepProcess    []*ExplanationStep             `json:"step_by_step_process"`
	ConflictExplanation  *ConflictExplanation           `json:"conflict_explanation,omitempty"`
	DecisionRationale    string                         `json:"decision_rationale"`
	AlternativeOutcomes  []*AlternativeOutcome          `json:"alternative_outcomes,omitempty"`
}

// ExplanationStep represents a step in the explanation
type ExplanationStep struct {
	StepNumber          int                      `json:"step_number"`
	Description         string                   `json:"description"`
	DecisionsConsidered []*DecisionInput         `json:"decisions_considered"`
	Outcome             types.PolicyDecisionType `json:"outcome"`
	Reasoning           string                   `json:"reasoning"`
}

// ConflictExplanation explains how conflicts were detected and resolved
type ConflictExplanation struct {
	ConflictType        ConflictType `json:"conflict_type"`
	ConflictingPolicies []string     `json:"conflicting_policies"`
	ResolutionStrategy  string       `json:"resolution_strategy"`
	WhyThisResolution   string       `json:"why_this_resolution"`
}

// AlternativeOutcome shows what would happen with different algorithms
type AlternativeOutcome struct {
	Algorithm types.PolicyCombiningAlgorithm `json:"algorithm"`
	Outcome   types.PolicyDecisionType       `json:"outcome"`
	Reason    string                         `json:"reason"`
}

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeAllowDeny     ConflictType = "allow_deny"
	ConflictTypeMultipleAllow ConflictType = "multiple_allow"
	ConflictTypeMultipleDeny  ConflictType = "multiple_deny"
	ConflictTypePriority      ConflictType = "priority"
	ConflictTypeNone          ConflictType = "none"
)

// ConflictSeverity represents the severity of a conflict
type ConflictSeverity string

const (
	ConflictSeverityLow      ConflictSeverity = "low"
	ConflictSeverityMedium   ConflictSeverity = "medium"
	ConflictSeverityHigh     ConflictSeverity = "high"
	ConflictSeverityCritical ConflictSeverity = "critical"
)

// policyCombiningService implements PolicyCombiningService
type policyCombiningService struct {
	tracing tracing.Service
	metrics metrics.MetricsProvider
	logger  logger.Logger
}

// NewPolicyCombiningService creates a new policy combining service
func NewPolicyCombiningService(
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) PolicyCombiningService {
	return &policyCombiningService{
		tracing: tracing,
		metrics: metrics,
		logger:  logger,
	}
}

// CombinePolicyDecisions combines multiple policy decisions using the specified algorithm
func (s *policyCombiningService) CombinePolicyDecisions(ctx context.Context, req *CombiningRequest) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.CombinePolicyDecisions",
		tracing.WithAttributes(
			attribute.String("algorithm", string(req.Algorithm)),
			attribute.Int("decision.count", len(req.Decisions)),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordCombiningMetrics(ctx, "combine_policies", string(req.Algorithm), time.Since(startTime))
	}()

	// Validate algorithm
	if err := s.ValidateCombiningAlgorithm(ctx, req.Algorithm); err != nil {
		return nil, err
	}

	result := &CombiningResult{
		Algorithm:      req.Algorithm,
		ProcessedAt:    time.Now(),
		CombiningTrace: make([]*CombiningStep, 0),
	}

	// Filter applicable decisions if requested
	decisions := req.Decisions
	if req.IgnoreNotApplicable {
		applicableDecisions := make([]*DecisionInput, 0)
		for _, decision := range req.Decisions {
			if decision.Applicable {
				applicableDecisions = append(applicableDecisions, decision)
			}
		}
		decisions = applicableDecisions
	}

	// Count decision types
	s.countDecisions(decisions, result)

	// Detect conflicts
	conflicts, err := s.DetectConflicts(ctx, decisions)
	if err != nil {
		return nil, err
	}

	result.ConflictDetected = conflicts.ConflictDetected
	if conflicts.ConflictDetected {
		resolution, err := s.ResolveConflicts(ctx, conflicts, req.Algorithm)
		if err != nil {
			return nil, err
		}
		result.ConflictResolution = resolution
	}

	// Apply the specified combining algorithm
	switch req.Algorithm {
	case types.CombiningAlgorithmDenyOverrides:
		combiningResult, err := s.ApplyDenyOverrides(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	case types.CombiningAlgorithmPermitOverrides:
		combiningResult, err := s.ApplyAllowOverrides(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	case types.CombiningAlgorithmFirstApplicable:
		combiningResult, err := s.ApplyFirstApplicable(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	case types.CombiningAlgorithmOnlyOneApplicable:
		combiningResult, err := s.ApplyOnlyOneApplicable(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	case types.CombiningAlgorithmOrderedDenyOverrides:
		combiningResult, err := s.ApplyOrderedDenyOverrides(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	case types.CombiningAlgorithmOrderedPermitOverrides:
		combiningResult, err := s.ApplyOrderedAllowOverrides(ctx, decisions)
		if err != nil {
			return nil, err
		}
		result.FinalDecision = combiningResult.FinalDecision
		result.CombiningTrace = combiningResult.CombiningTrace

	default:
		return nil, errors.NewBusinessError("UNSUPPORTED_COMBINING_ALGORITHM", "Unsupported combining algorithm").
			WithDetail("algorithm", string(req.Algorithm))
	}

	// Collect obligations and advice if requested
	if req.CollectObligations || req.CollectAdvice {
		s.collectObligationsAndAdvice(decisions, result, req.CollectObligations, req.CollectAdvice)
	}

	result.ProcessingTime = time.Since(startTime)

	s.logger.InfoContext(ctx, "Policy decisions combined",
		logger.Fields{
			"algorithm":          string(req.Algorithm),
			"input_count":        len(req.Decisions),
			"applicable_count":   result.ApplicableCount,
			"final_decision":     string(result.FinalDecision),
			"conflict_detected":  result.ConflictDetected,
			"processing_time_ms": result.ProcessingTime.Milliseconds(),
		})

	return result, nil
}

// CombineRuleDecisions combines rule evaluation results
func (s *policyCombiningService) CombineRuleDecisions(ctx context.Context, algorithm types.PolicyCombiningAlgorithm, ruleResults []*RuleEvaluationResult) (*RuleCombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.CombineRuleDecisions",
		tracing.WithAttributes(
			attribute.String("algorithm", string(algorithm)),
			attribute.Int("rule.count", len(ruleResults)),
		))
	defer span.End()

	result := &RuleCombiningResult{
		Algorithm:       algorithm,
		RulesProcessed:  len(ruleResults),
		ProcessingTrace: make([]*RuleCombiningStep, 0),
	}

	// Count applicable rules
	applicableRules := 0
	for _, ruleResult := range ruleResults {
		if ruleResult.Applicable {
			applicableRules++
		}
	}
	result.RulesApplicable = applicableRules

	// Convert rule results to decision inputs
	decisions := make([]*DecisionInput, 0, len(ruleResults))
	for _, ruleResult := range ruleResults {
		decision := &DecisionInput{
			ID:          ruleResult.RuleID,
			Name:        ruleResult.RuleName,
			Decision:    ruleResult.Decision,
			Applicable:  ruleResult.Applicable,
			Obligations: ruleResult.Obligations,
			Advice:      ruleResult.Advice,
		}
		decisions = append(decisions, decision)
	}

	// Apply combining algorithm
	combiningReq := &CombiningRequest{
		Algorithm:           algorithm,
		Decisions:           decisions,
		IgnoreNotApplicable: true,
		CollectObligations:  true,
		CollectAdvice:       true,
	}

	combiningResult, err := s.CombinePolicyDecisions(ctx, combiningReq)
	if err != nil {
		return nil, err
	}

	result.FinalDecision = combiningResult.FinalDecision

	// Convert combining trace to rule combining trace
	for _, step := range combiningResult.CombiningTrace {
		ruleStep := &RuleCombiningStep{
			Decision: step.Result,
			Action:   "evaluated",
			Reason:   step.Rationale,
		}
		result.ProcessingTrace = append(result.ProcessingTrace, ruleStep)
	}

	return result, nil
}

// ApplyDenyOverrides implements the deny-overrides combining algorithm
func (s *policyCombiningService) ApplyDenyOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyDenyOverrides")
	defer span.End()

	result := &CombiningResult{
		Algorithm:      types.CombiningAlgorithmDenyOverrides,
		CombiningTrace: make([]*CombiningStep, 0),
	}

	step := &CombiningStep{
		StepNumber:  1,
		Description: "Evaluating deny-overrides algorithm",
		InputCount:  len(decisions),
	}

	// Check for any deny decisions
	for _, decision := range decisions {
		if decision.Applicable && decision.Decision == types.PolicyDecisionDeny {
			result.FinalDecision = types.PolicyDecisionDeny
			step.Result = types.PolicyDecisionDeny
			step.Rationale = fmt.Sprintf("Deny decision found in policy: %s", decision.Name)
			result.CombiningTrace = append(result.CombiningTrace, step)
			return result, nil
		}
	}

	// Check for any allow decisions
	for _, decision := range decisions {
		if decision.Applicable && decision.Decision == types.PolicyDecisionAllow {
			result.FinalDecision = types.PolicyDecisionAllow
			step.Result = types.PolicyDecisionAllow
			step.Rationale = "Allow decision found and no deny decisions present"
			result.CombiningTrace = append(result.CombiningTrace, step)
			return result, nil
		}
	}

	// No applicable decisions
	result.FinalDecision = types.PolicyDecisionNotApplicable
	step.Result = types.PolicyDecisionNotApplicable
	step.Rationale = "No applicable decisions found"
	result.CombiningTrace = append(result.CombiningTrace, step)

	return result, nil
}

// ApplyAllowOverrides implements the allow-overrides combining algorithm
func (s *policyCombiningService) ApplyAllowOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyAllowOverrides")
	defer span.End()

	result := &CombiningResult{
		Algorithm:      types.CombiningAlgorithmPermitOverrides,
		CombiningTrace: make([]*CombiningStep, 0),
	}

	step := &CombiningStep{
		StepNumber:  1,
		Description: "Evaluating allow-overrides algorithm",
		InputCount:  len(decisions),
	}

	// Check for any allow decisions
	for _, decision := range decisions {
		if decision.Applicable && decision.Decision == types.PolicyDecisionAllow {
			result.FinalDecision = types.PolicyDecisionAllow
			step.Result = types.PolicyDecisionAllow
			step.Rationale = fmt.Sprintf("Allow decision found in policy: %s", decision.Name)
			result.CombiningTrace = append(result.CombiningTrace, step)
			return result, nil
		}
	}

	// Check for any deny decisions
	for _, decision := range decisions {
		if decision.Applicable && decision.Decision == types.PolicyDecisionDeny {
			result.FinalDecision = types.PolicyDecisionDeny
			step.Result = types.PolicyDecisionDeny
			step.Rationale = "Deny decision found and no allow decisions present"
			result.CombiningTrace = append(result.CombiningTrace, step)
			return result, nil
		}
	}

	// No applicable decisions
	result.FinalDecision = types.PolicyDecisionNotApplicable
	step.Result = types.PolicyDecisionNotApplicable
	step.Rationale = "No applicable decisions found"
	result.CombiningTrace = append(result.CombiningTrace, step)

	return result, nil
}

// ApplyFirstApplicable implements the first-applicable combining algorithm
func (s *policyCombiningService) ApplyFirstApplicable(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyFirstApplicable")
	defer span.End()

	result := &CombiningResult{
		Algorithm:      types.CombiningAlgorithmFirstApplicable,
		CombiningTrace: make([]*CombiningStep, 0),
	}

	step := &CombiningStep{
		StepNumber:  1,
		Description: "Evaluating first-applicable algorithm",
		InputCount:  len(decisions),
	}

	// Sort decisions by priority if available
	sortedDecisions := make([]*DecisionInput, len(decisions))
	copy(sortedDecisions, decisions)
	sort.Slice(sortedDecisions, func(i, j int) bool {
		return sortedDecisions[i].Priority < sortedDecisions[j].Priority
	})

	// Return the first applicable decision
	for _, decision := range sortedDecisions {
		if decision.Applicable {
			result.FinalDecision = decision.Decision
			step.Result = decision.Decision
			step.Rationale = fmt.Sprintf("First applicable decision from policy: %s", decision.Name)
			result.CombiningTrace = append(result.CombiningTrace, step)
			return result, nil
		}
	}

	// No applicable decisions
	result.FinalDecision = types.PolicyDecisionNotApplicable
	step.Result = types.PolicyDecisionNotApplicable
	step.Rationale = "No applicable decisions found"
	result.CombiningTrace = append(result.CombiningTrace, step)

	return result, nil
}

// ApplyOnlyOneApplicable implements the only-one-applicable combining algorithm
func (s *policyCombiningService) ApplyOnlyOneApplicable(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyOnlyOneApplicable")
	defer span.End()

	result := &CombiningResult{
		Algorithm:      types.CombiningAlgorithmOnlyOneApplicable,
		CombiningTrace: make([]*CombiningStep, 0),
	}

	step := &CombiningStep{
		StepNumber:  1,
		Description: "Evaluating only-one-applicable algorithm",
		InputCount:  len(decisions),
	}

	var applicableDecision *DecisionInput
	applicableCount := 0

	// Count applicable decisions
	for _, decision := range decisions {
		if decision.Applicable {
			applicableDecision = decision
			applicableCount++
		}
	}

	if applicableCount == 0 {
		result.FinalDecision = types.PolicyDecisionNotApplicable
		step.Result = types.PolicyDecisionNotApplicable
		step.Rationale = "No applicable decisions found"
	} else if applicableCount == 1 {
		result.FinalDecision = applicableDecision.Decision
		step.Result = applicableDecision.Decision
		step.Rationale = fmt.Sprintf("Exactly one applicable decision from policy: %s", applicableDecision.Name)
	} else {
		// More than one applicable decision is an error
		result.FinalDecision = types.PolicyDecisionDeny
		step.Result = types.PolicyDecisionDeny
		step.Rationale = fmt.Sprintf("Multiple applicable decisions found (%d), which violates only-one-applicable constraint", applicableCount)
	}

	result.CombiningTrace = append(result.CombiningTrace, step)
	return result, nil
}

// ApplyOrderedDenyOverrides implements the ordered deny-overrides combining algorithm
func (s *policyCombiningService) ApplyOrderedDenyOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyOrderedDenyOverrides")
	defer span.End()

	// Sort decisions by priority
	sortedDecisions := make([]*DecisionInput, len(decisions))
	copy(sortedDecisions, decisions)
	sort.Slice(sortedDecisions, func(i, j int) bool {
		return sortedDecisions[i].Priority < sortedDecisions[j].Priority
	})

	// Apply deny-overrides to the sorted decisions
	return s.ApplyDenyOverrides(ctx, sortedDecisions)
}

// ApplyOrderedAllowOverrides implements the ordered allow-overrides combining algorithm
func (s *policyCombiningService) ApplyOrderedAllowOverrides(ctx context.Context, decisions []*DecisionInput) (*CombiningResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ApplyOrderedAllowOverrides")
	defer span.End()

	// Sort decisions by priority
	sortedDecisions := make([]*DecisionInput, len(decisions))
	copy(sortedDecisions, decisions)
	sort.Slice(sortedDecisions, func(i, j int) bool {
		return sortedDecisions[i].Priority < sortedDecisions[j].Priority
	})

	// Apply allow-overrides to the sorted decisions
	return s.ApplyAllowOverrides(ctx, sortedDecisions)
}

// DetectConflicts detects conflicts between policy decisions
func (s *policyCombiningService) DetectConflicts(ctx context.Context, decisions []*DecisionInput) (*ConflictAnalysis, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.DetectConflicts")
	defer span.End()

	analysis := &ConflictAnalysis{
		AllowDecisions:   make([]*DecisionInput, 0),
		DenyDecisions:    make([]*DecisionInput, 0),
		ConflictReasons:  make([]string, 0),
		AnalysisMetadata: make(map[string]any),
	}

	// Separate allow and deny decisions
	for _, decision := range decisions {
		if !decision.Applicable {
			continue
		}

		switch decision.Decision {
		case types.PolicyDecisionAllow:
			analysis.AllowDecisions = append(analysis.AllowDecisions, decision)
		case types.PolicyDecisionDeny:
			analysis.DenyDecisions = append(analysis.DenyDecisions, decision)
		}
	}

	// Detect conflicts
	allowCount := len(analysis.AllowDecisions)
	denyCount := len(analysis.DenyDecisions)

	if allowCount > 0 && denyCount > 0 {
		analysis.ConflictDetected = true
		analysis.ConflictType = ConflictTypeAllowDeny
		analysis.ConflictingDecisions = append(analysis.AllowDecisions, analysis.DenyDecisions...)
		analysis.ConflictReasons = append(analysis.ConflictReasons,
			fmt.Sprintf("Found %d allow and %d deny decisions", allowCount, denyCount))

		// Determine severity
		if allowCount == 1 && denyCount == 1 {
			analysis.ConflictSeverity = ConflictSeverityMedium
		} else if allowCount > 3 || denyCount > 3 {
			analysis.ConflictSeverity = ConflictSeverityHigh
		} else {
			analysis.ConflictSeverity = ConflictSeverityMedium
		}
	} else if allowCount > 1 {
		analysis.ConflictType = ConflictTypeMultipleAllow
		analysis.ConflictSeverity = ConflictSeverityLow
		analysis.ConflictReasons = append(analysis.ConflictReasons,
			fmt.Sprintf("Found %d allow decisions", allowCount))
	} else if denyCount > 1 {
		analysis.ConflictType = ConflictTypeMultipleDeny
		analysis.ConflictSeverity = ConflictSeverityLow
		analysis.ConflictReasons = append(analysis.ConflictReasons,
			fmt.Sprintf("Found %d deny decisions", denyCount))
	} else {
		analysis.ConflictType = ConflictTypeNone
		analysis.ConflictSeverity = ConflictSeverityLow
	}

	analysis.AnalysisMetadata["allow_count"] = allowCount
	analysis.AnalysisMetadata["deny_count"] = denyCount
	analysis.AnalysisMetadata["total_applicable"] = allowCount + denyCount

	return analysis, nil
}

// ResolveConflicts resolves conflicts using the specified algorithm
func (s *policyCombiningService) ResolveConflicts(ctx context.Context, conflicts *ConflictAnalysis, algorithm types.PolicyCombiningAlgorithm) (*ConflictResolution, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ResolveConflicts")
	defer span.End()

	resolution := &ConflictResolution{
		ResolutionMethod:   string(algorithm),
		ResolutionMetadata: make(map[string]any),
	}

	switch algorithm {
	case types.CombiningAlgorithmDenyOverrides, types.CombiningAlgorithmOrderedDenyOverrides:
		if len(conflicts.DenyDecisions) > 0 {
			resolution.WinningDecision = conflicts.DenyDecisions[0]
			resolution.ResolvedDecision = types.PolicyDecisionDeny
			resolution.ResolutionReason = "Deny decisions override allow decisions"
			resolution.OverriddenDecisions = conflicts.AllowDecisions
		} else if len(conflicts.AllowDecisions) > 0 {
			resolution.WinningDecision = conflicts.AllowDecisions[0]
			resolution.ResolvedDecision = types.PolicyDecisionAllow
			resolution.ResolutionReason = "No deny decisions present, using allow decision"
		}

	case types.CombiningAlgorithmPermitOverrides, types.CombiningAlgorithmOrderedPermitOverrides:
		if len(conflicts.AllowDecisions) > 0 {
			resolution.WinningDecision = conflicts.AllowDecisions[0]
			resolution.ResolvedDecision = types.PolicyDecisionAllow
			resolution.ResolutionReason = "Allow decisions override deny decisions"
			resolution.OverriddenDecisions = conflicts.DenyDecisions
		} else if len(conflicts.DenyDecisions) > 0 {
			resolution.WinningDecision = conflicts.DenyDecisions[0]
			resolution.ResolvedDecision = types.PolicyDecisionDeny
			resolution.ResolutionReason = "No allow decisions present, using deny decision"
		}

	case types.CombiningAlgorithmFirstApplicable:
		// Find the first applicable decision by priority
		allDecisions := append(conflicts.AllowDecisions, conflicts.DenyDecisions...)
		sort.Slice(allDecisions, func(i, j int) bool {
			return allDecisions[i].Priority < allDecisions[j].Priority
		})
		if len(allDecisions) > 0 {
			resolution.WinningDecision = allDecisions[0]
			resolution.ResolvedDecision = allDecisions[0].Decision
			resolution.ResolutionReason = "First applicable decision by priority"
		}

	default:
		return nil, errors.NewBusinessError("UNSUPPORTED_CONFLICT_RESOLUTION", "Unsupported conflict resolution algorithm").
			WithDetail("algorithm", string(algorithm))
	}

	return resolution, nil
}

// ValidateCombiningAlgorithm validates a combining algorithm
func (s *policyCombiningService) ValidateCombiningAlgorithm(ctx context.Context, algorithm types.PolicyCombiningAlgorithm) error {
	supportedAlgorithms := []types.PolicyCombiningAlgorithm{
		types.CombiningAlgorithmDenyOverrides,
		types.CombiningAlgorithmPermitOverrides,
		types.CombiningAlgorithmFirstApplicable,
		types.CombiningAlgorithmOnlyOneApplicable,
		types.CombiningAlgorithmOrderedDenyOverrides,
		types.CombiningAlgorithmOrderedPermitOverrides,
	}

	for _, supported := range supportedAlgorithms {
		if algorithm == supported {
			return nil
		}
	}

	return errors.NewBusinessError("UNSUPPORTED_COMBINING_ALGORITHM", "Unsupported combining algorithm").
		WithDetail("algorithm", string(algorithm)).
		WithDetail("supported_algorithms", strings.Join(s.algorithmStrings(supportedAlgorithms), ", "))
}

// GetSupportedAlgorithms returns the list of supported combining algorithms
func (s *policyCombiningService) GetSupportedAlgorithms(ctx context.Context) ([]types.PolicyCombiningAlgorithm, error) {
	return []types.PolicyCombiningAlgorithm{
		types.CombiningAlgorithmDenyOverrides,
		types.CombiningAlgorithmPermitOverrides,
		types.CombiningAlgorithmFirstApplicable,
		types.CombiningAlgorithmOnlyOneApplicable,
		types.CombiningAlgorithmOrderedDenyOverrides,
		types.CombiningAlgorithmOrderedPermitOverrides,
	}, nil
}

// ExplainCombiningDecision provides an explanation of how decisions were combined
func (s *policyCombiningService) ExplainCombiningDecision(ctx context.Context, req *CombiningExplanationRequest) (*CombiningExplanation, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyCombiningService.ExplainCombiningDecision")
	defer span.End()

	// TODO: Implement detailed explanation logic
	explanation := &CombiningExplanation{
		Algorithm:            req.Algorithm,
		AlgorithmDescription: s.getAlgorithmDescription(req.Algorithm),
		StepByStepProcess:    make([]*ExplanationStep, 0),
		DecisionRationale:    "Decision made according to combining algorithm rules",
	}

	return explanation, nil
}

// Helper methods

func (s *policyCombiningService) countDecisions(decisions []*DecisionInput, result *CombiningResult) {
	for _, decision := range decisions {
		if decision.Applicable {
			result.ApplicableCount++
			switch decision.Decision {
			case types.PolicyDecisionAllow:
				result.AllowCount++
			case types.PolicyDecisionDeny:
				result.DenyCount++
			}
		} else {
			result.NotApplicableCount++
		}
	}
}

func (s *policyCombiningService) collectObligationsAndAdvice(decisions []*DecisionInput, result *CombiningResult, collectObligations, collectAdvice bool) {
	if collectObligations {
		result.Obligations = make([]*models.PolicyObligation, 0)
		for _, decision := range decisions {
			if decision.Applicable && decision.Decision == types.PolicyDecisionAllow {
				result.Obligations = append(result.Obligations, decision.Obligations...)
			}
		}
	}

	if collectAdvice {
		result.Advice = make([]*models.PolicyAdvice, 0)
		for _, decision := range decisions {
			if decision.Applicable {
				result.Advice = append(result.Advice, decision.Advice...)
			}
		}
	}
}

func (s *policyCombiningService) algorithmStrings(algorithms []types.PolicyCombiningAlgorithm) []string {
	strings := make([]string, len(algorithms))
	for i, alg := range algorithms {
		strings[i] = string(alg)
	}
	return strings
}

func (s *policyCombiningService) getAlgorithmDescription(algorithm types.PolicyCombiningAlgorithm) string {
	descriptions := map[types.PolicyCombiningAlgorithm]string{
		types.CombiningAlgorithmDenyOverrides:          "Any deny decision overrides all allow decisions",
		types.CombiningAlgorithmPermitOverrides:        "Any allow decision overrides all deny decisions",
		types.CombiningAlgorithmFirstApplicable:        "The first applicable decision is used",
		types.CombiningAlgorithmOnlyOneApplicable:      "Exactly one decision must be applicable",
		types.CombiningAlgorithmOrderedDenyOverrides:   "Deny overrides with priority ordering",
		types.CombiningAlgorithmOrderedPermitOverrides: "Allow overrides with priority ordering",
	}

	if desc, exists := descriptions[algorithm]; exists {
		return desc
	}
	return "Unknown algorithm"
}

func (s *policyCombiningService) recordCombiningMetrics(ctx context.Context, operation, algorithm string, duration time.Duration) {
	// Combining operation counter
	counter := s.metrics.Counter(
		"abac_policy_combining_operations_total",
		"Total number of policy combining operations",
		"operation", "algorithm",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
		"algorithm": algorithm,
	})

	// Combining operation duration histogram
	histogram := s.metrics.Histogram(
		"abac_policy_combining_operation_duration_seconds",
		"Duration of policy combining operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation", "algorithm",
	)

	histogram.Observe(duration.Seconds(), metrics.Fields{
		"operation": operation,
		"algorithm": algorithm,
	})
}
