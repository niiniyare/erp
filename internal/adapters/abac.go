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
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	req := &abacServices.DecisionRequest{
		UserID:          userID,
		ResourceType:    *p.ResourceType,
		Action:          *p.Action,
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

	// Note: EntityID field doesn't exist in PolicyEvaluationRequest
	// Remove this section if not needed or add the field to the Goa design

	if p.RequestID != nil {
		req.RequestID = *p.RequestID
	}

	return req, nil
}

// ConvertToPolicyEvaluationResponse converts domain decision response to Goa response
func (a *ABACAdapter) ConvertToPolicyEvaluationResponse(resp *abacServices.DecisionResponse) *goaABAC.PolicyEvaluationResponse {
	decision := string(resp.Decision)
	allowed := resp.Allowed
	evaluationTimeMs := uint(resp.EvaluationTime.Milliseconds())
	evaluatedAt := resp.EvaluatedAt.Format(time.RFC3339)
	cacheHit := resp.CacheHit
	policyCount := uint(resp.PolicyCount)

	goaResp := &goaABAC.PolicyEvaluationResponse{
		Decision:         &decision,
		Allowed:          &allowed,
		EvaluationTimeMs: &evaluationTimeMs,
		EvaluatedAt:      &evaluatedAt,
		RequestID:        &resp.RequestID,
		CacheHit:         &cacheHit,
		PolicyCount:      &policyCount,
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
	successCount := uint(resp.SuccessCount)
	errorCount := uint(resp.ErrorCount)
	totalRequests := uint(resp.TotalRequests)
	evaluationTimeMs := uint(resp.EvaluationTime.Milliseconds())
	partialFailure := resp.PartialFailure

	goaResp := &goaABAC.BulkPolicyEvaluationResponse{
		SuccessCount:     &successCount,
		ErrorCount:       &errorCount,
		TotalRequests:    &totalRequests,
		EvaluationTimeMs: &evaluationTimeMs,
		RequestID:        &resp.RequestID,
		PartialFailure:   &partialFailure,
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
	finalDecision := string(explanation.FinalDecision)
	reasoningSummary := explanation.ReasoningSummary
	combiningAlgorithm := explanation.CombiningAlgorithm

	goaResp := &goaABAC.PolicyExplanationResponse{
		FinalDecision:      &finalDecision,
		ReasoningSummary:   &reasoningSummary,
		CombiningAlgorithm: &combiningAlgorithm,
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
		id := policy.ID.String()
		name := policy.Name
		effect := string(policy.Effect)
		priority := uint(policy.Priority)
		applicable := policy.Applicable

		goaPolicies[i] = &goaABAC.PolicySummary{
			ID:          &id,
			Name:        &name,
			Description: &policy.Description,
			Effect:      &effect,
			Priority:    &priority,
			Applicable:  &applicable,
		}
	}
	return goaPolicies
}

// ConvertToAttributeCollectionResponse converts domain attribute context to Goa response
func (a *ABACAdapter) ConvertToAttributeCollectionResponse(attrContext *abacModels.AttributeContext, collectionTime time.Duration) *goaABAC.AttributeCollectionResponse {
	collectionTimeMs := uint(collectionTime.Milliseconds())
	goaResp := &goaABAC.AttributeCollectionResponse{
		CollectionTimeMs: &collectionTimeMs,
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

	totalAttributes := uint(totalAttrs)
	goaResp.TotalAttributes = &totalAttributes

	return goaResp
}

// ConvertToDecisionAuditEntries converts domain audit logs to Goa entries
func (a *ABACAdapter) ConvertToDecisionAuditEntries(decisions []*abacServices.DecisionAuditLog) []*goaABAC.DecisionAuditEntry {
	goaEntries := make([]*goaABAC.DecisionAuditEntry, len(decisions))
	for i, decision := range decisions {
		id := decision.ID.String()
		userID := decision.UserID.String()
		resourceType := decision.ResourceType
		action := decision.Action
		dec := string(decision.Decision)
		allowed := decision.Allowed
		evaluationTimeMs := uint(decision.EvaluationTime.Milliseconds())
		evaluatedAt := decision.EvaluatedAt.Format(time.RFC3339)
		policyCount := uint(decision.PolicyCount)
		cacheHit := decision.CacheHit
		createdAt := decision.EvaluatedAt.Format(time.RFC3339)
		updatedAt := decision.EvaluatedAt.Format(time.RFC3339)

		goaEntries[i] = &goaABAC.DecisionAuditEntry{
			ID:               &id,
			UserID:           &userID,
			ResourceType:     &resourceType,
			Action:           &action,
			Decision:         &dec,
			Allowed:          &allowed,
			EvaluationTimeMs: &evaluationTimeMs,
			EvaluatedAt:      &evaluatedAt,
			PolicyCount:      &policyCount,
			CacheHit:         &cacheHit,
			RequestID:        &decision.RequestID,
			CreatedAt:        &createdAt,
			UpdatedAt:        &updatedAt,
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
	id := obligation.ID
	type_ := obligation.Type
	description := obligation.Description

	return &goaABAC.PolicyObligation{
		ID:          &id,
		Type:        &type_,
		Description: &description,
		Parameters:  obligation.Parameters,
	}
}

func (a *ABACAdapter) convertPolicyAdvice(advice *abacModels.PolicyAdvice) *goaABAC.PolicyAdvice {
	id := advice.ID
	type_ := advice.Type
	description := advice.Description
	// Note: Severity field doesn't exist in models.PolicyAdvice, using Type as severity
	severity := advice.Type

	return &goaABAC.PolicyAdvice{
		ID:          &id,
		Type:        &type_,
		Description: &description,
		Severity:    &severity,
		Parameters:  advice.Parameters,
	}
}

func (a *ABACAdapter) convertPolicyExplanation(explanation *abacServices.DecisionExplanation) *goaABAC.PolicyExplanation {
	finalDecision := string(explanation.FinalDecision)
	reasoningSummary := explanation.ReasoningSummary
	combiningAlgorithm := explanation.CombiningAlgorithm

	goaExplanation := &goaABAC.PolicyExplanation{
		FinalDecision:      &finalDecision,
		ReasoningSummary:   &reasoningSummary,
		CombiningAlgorithm: &combiningAlgorithm,
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
	policyName := eval.PolicyName
	decision := string(eval.Decision)
	applicable := eval.Applicable
	reason := eval.Reason

	return &goaABAC.PolicyEvaluationSummary{
		PolicyName:   &policyName,
		Decision:     &decision,
		Applicable:   &applicable,
		MatchedRules: eval.MatchedRules,
		FailedRules:  eval.FailedRules,
		Reason:       &reason,
	}
}

func (a *ABACAdapter) convertConflictResolutionSummary(resolution *abacServices.ConflictResolutionSummary) *goaABAC.ConflictResolutionSummary {
	conflictDetected := resolution.ConflictDetected
	resolutionMethod := resolution.ResolutionMethod
	explanation := resolution.Explanation

	return &goaABAC.ConflictResolutionSummary{
		ConflictDetected:    &conflictDetected,
		ConflictingPolicies: resolution.ConflictingPolicies,
		ResolutionMethod:    &resolutionMethod,
		WinningPolicy:       &resolution.WinningPolicy,
		Explanation:         &explanation,
	}
}

func (a *ABACAdapter) convertDecisionAuditTrail(trail *abacServices.DecisionAuditTrail) *goaABAC.DecisionAuditTrail {
	// Convert AttributesSeen from map[string]string to map[string]any
	attributesSeen := make(map[string]any)
	for k, v := range trail.AttributesSeen {
		attributesSeen[k] = v
	}

	goaTrail := &goaABAC.DecisionAuditTrail{
		PoliciesApplied: trail.PoliciesApplied,
		AttributesSeen:  attributesSeen,
	}

	// Convert evaluation steps
	if len(trail.EvaluationSteps) > 0 {
		goaTrail.EvaluationSteps = make([]*goaABAC.EvaluationStep, len(trail.EvaluationSteps))
		for i, step := range trail.EvaluationSteps {
			stepType := step.StepType
			description := step.Description
			durationMs := uint(step.Duration.Milliseconds())

			// Convert result to string pointer if it's not nil
			var resultPtr *string
			if step.Result != nil {
				resultStr := fmt.Sprintf("%v", step.Result)
				resultPtr = &resultStr
			}

			goaTrail.EvaluationSteps[i] = &goaABAC.EvaluationStep{
				StepType:    &stepType,
				Description: &description,
				Result:      resultPtr,
				DurationMs:  &durationMs,
				Metadata:    step.Metadata,
			}
		}
	}

	// Convert cache events
	if len(trail.CacheEvents) > 0 {
		goaTrail.CacheEvents = make([]*goaABAC.CacheEvent, len(trail.CacheEvents))
		for i, event := range trail.CacheEvents {
			eventType := event.EventType
			hit := event.Hit
			timestamp := event.Timestamp.Format(time.RFC3339)

			goaTrail.CacheEvents[i] = &goaABAC.CacheEvent{
				EventType: &eventType,
				CacheKey:  &event.CacheKey,
				Hit:       &hit,
				Timestamp: &timestamp,
			}
		}
	}

	// Convert timing
	if trail.Timing != nil {
		attributeCollectionMs := uint(trail.Timing.AttributeCollectionTime.Milliseconds())
		attributeValidationMs := uint(trail.Timing.AttributeValidationTime.Milliseconds())
		policyRetrievalMs := uint(trail.Timing.PolicyRetrievalTime.Milliseconds())
		policyEvaluationMs := uint(trail.Timing.PolicyEvaluationTime.Milliseconds())
		cacheOperationMs := uint(trail.Timing.CacheOperationTime.Milliseconds())
		totalMs := uint(trail.Timing.TotalTime.Milliseconds())

		goaTrail.Timing = &goaABAC.EvaluationTiming{
			AttributeCollectionMs: &attributeCollectionMs,
			AttributeValidationMs: &attributeValidationMs,
			PolicyRetrievalMs:     &policyRetrievalMs,
			PolicyEvaluationMs:    &policyEvaluationMs,
			CacheOperationMs:      &cacheOperationMs,
			TotalMs:               &totalMs,
		}
	}

	return goaTrail
}

func (a *ABACAdapter) convertAttributeMap(attributes map[string]*abacModels.AttributeValue) map[string]*goaABAC.AttributeValue {
	goaAttrs := make(map[string]*goaABAC.AttributeValue)
	for attrName, attr := range attributes {
		name := attr.Name
		dataType := string(attr.DataType)
		category := string(attr.Category)
		source := string(attr.Source)
		collectedAt := attr.Timestamp.Format(time.RFC3339)

		goaAttrs[attrName] = &goaABAC.AttributeValue{
			Name:        &name,
			Value:       attr.Value,
			DataType:    &dataType,
			Category:    &category,
			Source:      &source,
			CollectedAt: &collectedAt,
		}

		if attr.ExpiresAt != nil {
			expiresAt := attr.ExpiresAt.Format(time.RFC3339)
			goaAttrs[attrName].ExpiresAt = &expiresAt
		}
	}
	return goaAttrs
}
