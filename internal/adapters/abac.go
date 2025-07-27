package adapters

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	// Generated Goa types
	goaABAC "github.com/niiniyare/erp/internal/api/gen/abac"

	// Domain services and models
	abacModels "github.com/niiniyare/erp/internal/core/abac/models"
	abacServices "github.com/niiniyare/erp/internal/core/abac/services"
)

// ABACAdapter handles conversions between Goa types and domain types
type ABACAdapter struct{}

// NewABACAdapter creates a new ABAC adapter
func NewABACAdapter() *ABACAdapter {
	return &ABACAdapter{}
}

// ConvertToDecisionRequest converts Goa policy evaluation request to domain decision request
func (a *ABACAdapter) ConvertToDecisionRequest(p *goaABAC.PolicyEvaluationRequest) (*abacServices.DecisionRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	req := &abacServices.DecisionRequest{
		UserID:          userID,
		ResourceType:    p.ResourceType,
		Action:          p.Action,
		Context:         p.Context,
		UseCache:        *p.UseCache,
		CacheResults:    *p.CacheResults,
		IncludeAdvice:   *p.IncludeAdvice,
		ExplainDecision: *p.ExplainDecision,
	}

	if p.ResourceID != nil {
		resourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid resource ID: %w", err)
		}
		req.ResourceID = &resourceID
	}

	if p.EntityID != nil {
		entityID, err := uuid.Parse(*p.EntityID)
		if err != nil {
			return nil, fmt.Errorf("invalid entity ID: %w", err)
		}
		req.EntityID = &entityID
	}

	if p.RequestID != nil {
		req.RequestID = *p.RequestID
	}

	return req, nil
}

// ConvertToPolicyEvaluationResponse converts domain decision response to Goa response
func (a *ABACAdapter) ConvertToPolicyEvaluationResponse(resp *abacServices.DecisionResponse) *goaABAC.PolicyEvaluationResponse {
	goaResp := &goaABAC.PolicyEvaluationResponse{
		Decision:         string(resp.Decision),
		Allowed:          resp.Allowed,
		EvaluationTimeMs: uint(resp.EvaluationTime.Milliseconds()),
		EvaluatedAt:      resp.EvaluatedAt.Format(time.RFC3339),
		RequestID:        &resp.RequestID,
		CacheHit:         resp.CacheHit,
		PolicyCount:      uint(resp.PolicyCount),
	}

	// Convert obligations
	if len(resp.Obligations) > 0 {
		goaResp.Obligations = make([]*goaABAC.PolicyObligation, len(resp.Obligations))
		for i, obligation := range resp.Obligations {
			goaResp.Obligations[i] = a.convertPolicyObligation(obligation)
		}
	}

	// Convert advice
	if len(resp.Advice) > 0 {
		goaResp.Advice = make([]*goaABAC.PolicyAdvice, len(resp.Advice))
		for i, advice := range resp.Advice {
			goaResp.Advice[i] = a.convertPolicyAdvice(advice)
		}
	}

	// Convert explanation if present
	if resp.Explanation != nil {
		goaResp.Explanation = a.convertPolicyExplanation(resp.Explanation)
	}

	// Convert audit trail if present
	if resp.AuditTrail != nil {
		goaResp.AuditTrail = a.convertDecisionAuditTrail(resp.AuditTrail)
	}

	return goaResp
}

// ConvertToBulkPolicyEvaluationResponse converts domain bulk response to Goa response
func (a *ABACAdapter) ConvertToBulkPolicyEvaluationResponse(resp *abacServices.BulkDecisionResponse) *goaABAC.BulkPolicyEvaluationResponse {
	goaResp := &goaABAC.BulkPolicyEvaluationResponse{
		SuccessCount:     uint(resp.SuccessCount),
		ErrorCount:       uint(resp.ErrorCount),
		TotalRequests:    uint(resp.TotalRequests),
		EvaluationTimeMs: uint(resp.EvaluationTime.Milliseconds()),
		RequestID:        &resp.RequestID,
		PartialFailure:   resp.PartialFailure,
	}

	// Convert individual responses
	goaResp.Responses = make([]*goaABAC.PolicyEvaluationResponse, len(resp.Responses))
	for i, response := range resp.Responses {
		goaResp.Responses[i] = a.ConvertToPolicyEvaluationResponse(response)
	}

	return goaResp
}

// ConvertToPolicyExplanationResponse converts domain explanation to Goa response
func (a *ABACAdapter) ConvertToPolicyExplanationResponse(explanation *abacServices.DecisionExplanation) *goaABAC.PolicyExplanationResponse {
	goaResp := &goaABAC.PolicyExplanationResponse{
		FinalDecision:      string(explanation.FinalDecision),
		ReasoningSummary:   explanation.ReasoningSummary,
		CombiningAlgorithm: explanation.CombiningAlgorithm,
		AttributesUsed:     explanation.AttributesUsed,
		Recommendations:    explanation.Recommendations,
	}

	// Convert policy evaluations
	if len(explanation.PolicyEvaluations) > 0 {
		goaResp.PolicyEvaluations = make([]*goaABAC.PolicyEvaluationSummary, len(explanation.PolicyEvaluations))
		for i, eval := range explanation.PolicyEvaluations {
			goaResp.PolicyEvaluations[i] = a.convertPolicyEvaluationSummary(eval)
		}
	}

	// Convert conflict resolution
	if explanation.ConflictResolution != nil {
		goaResp.ConflictResolution = a.convertConflictResolutionSummary(explanation.ConflictResolution)
	}

	return goaResp
}

// ConvertToPolicySummaries converts domain policy summaries to Goa types
func (a *ABACAdapter) ConvertToPolicySummaries(policies []*abacServices.PolicySummary) []*goaABAC.PolicySummary {
	goaPolicies := make([]*goaABAC.PolicySummary, len(policies))
	for i, policy := range policies {
		goaPolicies[i] = &goaABAC.PolicySummary{
			ID:          policy.ID.String(),
			Name:        policy.Name,
			Description: &policy.Description,
			Effect:      string(policy.Effect),
			Priority:    uint(policy.Priority),
			Applicable:  policy.Applicable,
		}
	}
	return goaPolicies
}

// ConvertToAttributeCollectionResponse converts domain attribute context to Goa response
func (a *ABACAdapter) ConvertToAttributeCollectionResponse(attrContext *abacModels.AttributeContext, collectionTime time.Duration) *goaABAC.AttributeCollectionResponse {
	goaResp := &goaABAC.AttributeCollectionResponse{
		CollectionTimeMs: uint(collectionTime.Milliseconds()),
	}

	// Count total attributes
	totalAttrs := 0

	// Convert user attributes
	if attrContext.UserAttributes != nil {
		goaResp.UserAttributes = a.convertAttributeMap(attrContext.UserAttributes)
		totalAttrs += len(attrContext.UserAttributes)
	}

	// Convert resource attributes
	if attrContext.ResourceAttributes != nil {
		goaResp.ResourceAttributes = a.convertAttributeMap(attrContext.ResourceAttributes)
		totalAttrs += len(attrContext.ResourceAttributes)
	}

	// Convert environment attributes
	if attrContext.EnvironmentAttributes != nil {
		goaResp.EnvironmentAttributes = a.convertAttributeMap(attrContext.EnvironmentAttributes)
		totalAttrs += len(attrContext.EnvironmentAttributes)
	}

	// Convert action attributes
	if attrContext.ActionAttributes != nil {
		goaResp.ActionAttributes = a.convertAttributeMap(attrContext.ActionAttributes)
		totalAttrs += len(attrContext.ActionAttributes)
	}

	// Convert entity attributes
	if attrContext.EntityAttributes != nil {
		goaResp.EntityAttributes = a.convertAttributeMap(attrContext.EntityAttributes)
		totalAttrs += len(attrContext.EntityAttributes)
	}

	// Convert session attributes
	if attrContext.SessionAttributes != nil {
		goaResp.SessionAttributes = a.convertAttributeMap(attrContext.SessionAttributes)
		totalAttrs += len(attrContext.SessionAttributes)
	}

	goaResp.TotalAttributes = uint(totalAttrs)

	return goaResp
}

// ConvertToDecisionAuditEntries converts domain audit logs to Goa entries
func (a *ABACAdapter) ConvertToDecisionAuditEntries(decisions []*abacServices.DecisionAuditLog) []*goaABAC.DecisionAuditEntry {
	goaEntries := make([]*goaABAC.DecisionAuditEntry, len(decisions))
	for i, decision := range decisions {
		goaEntries[i] = &goaABAC.DecisionAuditEntry{
			ID:               decision.ID.String(),
			UserID:           decision.UserID.String(),
			ResourceType:     decision.ResourceType,
			Action:           decision.Action,
			Decision:         string(decision.Decision),
			Allowed:          decision.Allowed,
			EvaluationTimeMs: uint(decision.EvaluationTime.Milliseconds()),
			EvaluatedAt:      decision.EvaluatedAt.Format(time.RFC3339),
			PolicyCount:      uint(decision.PolicyCount),
			CacheHit:         decision.CacheHit,
			RequestID:        &decision.RequestID,
			CreatedAt:        decision.EvaluatedAt.Format(time.RFC3339),
			UpdatedAt:        decision.EvaluatedAt.Format(time.RFC3339),
		}

		if decision.ResourceID != nil {
			resourceID := decision.ResourceID.String()
			goaEntries[i].ResourceID = &resourceID
		}

		if decision.TenantID != uuid.Nil {
			createdBy := decision.TenantID.String()
			goaEntries[i].CreatedBy = &createdBy
			goaEntries[i].UpdatedBy = &createdBy
		}
	}
	return goaEntries
}

// Helper conversion methods

func (a *ABACAdapter) convertPolicyObligation(obligation *abacModels.PolicyObligation) *goaABAC.PolicyObligation {
	return &goaABAC.PolicyObligation{
		ID:          obligation.ID,
		Type:        obligation.Type,
		Description: obligation.Description,
		Parameters:  obligation.Parameters,
	}
}

func (a *ABACAdapter) convertPolicyAdvice(advice *abacModels.PolicyAdvice) *goaABAC.PolicyAdvice {
	return &goaABAC.PolicyAdvice{
		ID:          advice.ID,
		Type:        advice.Type,
		Description: advice.Description,
		Severity:    advice.Severity,
		Parameters:  advice.Parameters,
	}
}

func (a *ABACAdapter) convertPolicyExplanation(explanation *abacServices.DecisionExplanation) *goaABAC.PolicyExplanation {
	goaExplanation := &goaABAC.PolicyExplanation{
		FinalDecision:      string(explanation.FinalDecision),
		ReasoningSummary:   explanation.ReasoningSummary,
		CombiningAlgorithm: explanation.CombiningAlgorithm,
		AttributesUsed:     explanation.AttributesUsed,
		Recommendations:    explanation.Recommendations,
	}

	// Convert policy evaluations
	if len(explanation.PolicyEvaluations) > 0 {
		goaExplanation.PolicyEvaluations = make([]*goaABAC.PolicyEvaluationSummary, len(explanation.PolicyEvaluations))
		for i, eval := range explanation.PolicyEvaluations {
			goaExplanation.PolicyEvaluations[i] = a.convertPolicyEvaluationSummary(eval)
		}
	}

	// Convert conflict resolution
	if explanation.ConflictResolution != nil {
		goaExplanation.ConflictResolution = a.convertConflictResolutionSummary(explanation.ConflictResolution)
	}

	return goaExplanation
}

func (a *ABACAdapter) convertPolicyEvaluationSummary(eval *abacServices.PolicyEvaluationSummary) *goaABAC.PolicyEvaluationSummary {
	return &goaABAC.PolicyEvaluationSummary{
		PolicyName:   eval.PolicyName,
		Decision:     string(eval.Decision),
		Applicable:   eval.Applicable,
		MatchedRules: eval.MatchedRules,
		FailedRules:  eval.FailedRules,
		Reason:       eval.Reason,
	}
}

func (a *ABACAdapter) convertConflictResolutionSummary(resolution *abacServices.ConflictResolutionSummary) *goaABAC.ConflictResolutionSummary {
	return &goaABAC.ConflictResolutionSummary{
		ConflictDetected:    resolution.ConflictDetected,
		ConflictingPolicies: resolution.ConflictingPolicies,
		ResolutionMethod:    resolution.ResolutionMethod,
		WinningPolicy:       &resolution.WinningPolicy,
		Explanation:         resolution.Explanation,
	}
}

func (a *ABACAdapter) convertDecisionAuditTrail(trail *abacServices.DecisionAuditTrail) *goaABAC.DecisionAuditTrail {
	goaTrail := &goaABAC.DecisionAuditTrail{
		PoliciesApplied: trail.PoliciesApplied,
		AttributesSeen:  trail.AttributesSeen,
	}

	// Convert evaluation steps
	if len(trail.EvaluationSteps) > 0 {
		goaTrail.EvaluationSteps = make([]*goaABAC.EvaluationStep, len(trail.EvaluationSteps))
		for i, step := range trail.EvaluationSteps {
			goaTrail.EvaluationSteps[i] = &goaABAC.EvaluationStep{
				StepType:    step.StepType,
				Description: step.Description,
				Result:      step.Result,
				DurationMs:  uint(step.Duration.Milliseconds()),
				Metadata:    step.Metadata,
			}
		}
	}

	// Convert cache events
	if len(trail.CacheEvents) > 0 {
		goaTrail.CacheEvents = make([]*goaABAC.CacheEvent, len(trail.CacheEvents))
		for i, event := range trail.CacheEvents {
			goaTrail.CacheEvents[i] = &goaABAC.CacheEvent{
				EventType: event.EventType,
				CacheKey:  &event.CacheKey,
				Hit:       event.Hit,
				Timestamp: event.Timestamp.Format(time.RFC3339),
			}
		}
	}

	// Convert timing
	if trail.Timing != nil {
		goaTrail.Timing = &goaABAC.EvaluationTiming{
			AttributeCollectionMs: uint(trail.Timing.AttributeCollectionTime.Milliseconds()),
			AttributeValidationMs: uint(trail.Timing.AttributeValidationTime.Milliseconds()),
			PolicyRetrievalMs:     uint(trail.Timing.PolicyRetrievalTime.Milliseconds()),
			PolicyEvaluationMs:    uint(trail.Timing.PolicyEvaluationTime.Milliseconds()),
			CacheOperationMs:      uint(trail.Timing.CacheOperationTime.Milliseconds()),
			TotalMs:               uint(trail.Timing.TotalTime.Milliseconds()),
		}
	}

	return goaTrail
}

func (a *ABACAdapter) convertAttributeMap(attributes map[string]*abacModels.AttributeValue) map[string]*goaABAC.AttributeValue {
	goaAttrs := make(map[string]*goaABAC.AttributeValue)
	for name, attr := range attributes {
		goaAttrs[name] = &goaABAC.AttributeValue{
			Name:        attr.Name,
			Value:       attr.Value,
			DataType:    string(attr.DataType),
			Category:    string(attr.Category),
			Source:      &attr.Source,
			CollectedAt: attr.CollectedAt.Format(time.RFC3339),
		}

		if attr.ExpiresAt != nil {
			expiresAt := attr.ExpiresAt.Format(time.RFC3339)
			goaAttrs[name].ExpiresAt = &expiresAt
		}
	}
	return goaAttrs
}
