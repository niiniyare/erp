package handlers

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	// Generated Goa types
	goaABAC "github.com/niiniyare/erp/internal/api/gen/abac"

	// Domain services and models
	abacModels "github.com/niiniyare/erp/internal/core/abac/models"
	abacServices "github.com/niiniyare/erp/internal/core/abac/services"
	"github.com/niiniyare/erp/internal/shared/errors"
)

// Conversion methods for ABACGoaHandler

// convertToDecisionRequest converts Goa policy evaluation request to domain decision request
func (h *ABACGoaHandler) convertToDecisionRequest(p *goaABAC.PolicyEvaluationRequest) (*abacServices.DecisionRequest, error) {
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

// convertToBulkDecisionRequest converts Goa bulk request to domain bulk request
func (h *ABACGoaHandler) convertToBulkDecisionRequest(p *goaABAC.BulkPolicyEvaluationRequest) (*abacServices.BulkDecisionRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	domainReq := &abacServices.BulkDecisionRequest{
		UserID:       userID,
		RequestID:    *p.RequestID,
		FailFast:     *p.FailFast,
		UseCache:     *p.UseCache,
		CacheResults: *p.CacheResults,
	}

	// Convert individual requests
	domainReq.Decisions = make([]*abacServices.DecisionRequest, len(p.Requests))
	for i, req := range p.Requests {
		decisionReq, err := h.convertToDecisionRequest(req)
		if err != nil {
			return nil, fmt.Errorf("invalid request at index %d: %w", i, err)
		}
		domainReq.Decisions[i] = decisionReq
	}

	return domainReq, nil
}

// convertToPermissionCheckRequest converts Goa authorization request to domain permission check request
func (h *ABACGoaHandler) convertToPermissionCheckRequest(p *goaABAC.AuthorizationRequest) (*abacServices.PermissionCheckRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	permReq := &abacServices.PermissionCheckRequest{
		UserID:       userID,
		ResourceType: p.ResourceType,
		Action:       p.Action,
		Context:      p.Context,
	}

	if p.ResourceID != nil {
		resourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid resource ID: %w", err)
		}
		permReq.ResourceID = &resourceID
	}

	return permReq, nil
}

// convertToDecisionExplanationRequest converts Goa explanation request to domain explanation request
func (h *ABACGoaHandler) convertToDecisionExplanationRequest(p *goaABAC.PolicyExplanationRequest) (*abacServices.DecisionExplanationRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	explanationReq := &abacServices.DecisionExplanationRequest{
		UserID:       userID,
		ResourceType: p.ResourceType,
		Action:       p.Action,
		Context:      p.Context,
		DetailLevel:  *p.DetailLevel,
	}

	if p.ResourceID != nil {
		resourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid resource ID: %w", err)
		}
		explanationReq.ResourceID = &resourceID
	}

	return explanationReq, nil
}

// convertToPolicyDiscoveryRequest converts Goa discovery request to domain discovery request
func (h *ABACGoaHandler) convertToPolicyDiscoveryRequest(p *goaABAC.PolicyDiscoveryRequest) (*abacServices.PolicyDiscoveryRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return &abacServices.PolicyDiscoveryRequest{
		UserID:       userID,
		ResourceType: p.ResourceType,
		Action:       p.Action,
		Context:      p.Context,
	}, nil
}

// convertToAttributeCollectionRequest converts Goa attribute collection request to domain request
func (h *ABACGoaHandler) convertToAttributeCollectionRequest(p *goaABAC.AttributeCollectionRequest) (*abacServices.AttributeCollectionRequest, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	domainReq := &abacServices.AttributeCollectionRequest{
		UserID:         userID,
		ResourceType:   p.ResourceType,
		Action:         p.Action,
		CollectExpired: *p.IncludeExpired,
	}

	if p.ResourceID != nil {
		resourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid resource ID: %w", err)
		}
		domainReq.ResourceID = &resourceID
	}

	if p.EntityID != nil {
		entityID, err := uuid.Parse(*p.EntityID)
		if err != nil {
			return nil, fmt.Errorf("invalid entity ID: %w", err)
		}
		domainReq.EntityID = &entityID
	}

	return domainReq, nil
}

// convertToAuditDecisionsRequest converts Goa audit request to domain parameters
func (h *ABACGoaHandler) convertToAuditDecisionsRequest(p *goaABAC.AuditDecisionsPayload) (uuid.UUID, int, error) {
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	limit := int(*p.Limit)
	return userID, limit, nil
}

// Response conversion methods

// convertToPolicyEvaluationResponse converts domain decision response to Goa response
func (h *ABACGoaHandler) convertToPolicyEvaluationResponse(resp *abacServices.DecisionResponse) *goaABAC.PolicyEvaluationResponse {
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
			goaResp.Obligations[i] = h.convertPolicyObligation(obligation)
		}
	}

	// Convert advice
	if len(resp.Advice) > 0 {
		goaResp.Advice = make([]*goaABAC.PolicyAdvice, len(resp.Advice))
		for i, advice := range resp.Advice {
			goaResp.Advice[i] = h.convertPolicyAdvice(advice)
		}
	}

	// Convert explanation if present
	if resp.Explanation != nil {
		goaResp.Explanation = h.convertPolicyExplanation(resp.Explanation)
	}

	// Convert audit trail if present
	if resp.AuditTrail != nil {
		goaResp.AuditTrail = h.convertDecisionAuditTrail(resp.AuditTrail)
	}

	return goaResp
}

// convertToBulkPolicyEvaluationResponse converts domain bulk response to Goa response
func (h *ABACGoaHandler) convertToBulkPolicyEvaluationResponse(resp *abacServices.BulkDecisionResponse) *goaABAC.BulkPolicyEvaluationResponse {
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
		goaResp.Responses[i] = h.convertToPolicyEvaluationResponse(response)
	}

	return goaResp
}

// convertToAuthorizationResponse converts domain permission response to Goa response
func (h *ABACGoaHandler) convertToAuthorizationResponse(permResp *abacServices.PermissionCheckResponse) *goaABAC.AuthorizationResponse {
	return &goaABAC.AuthorizationResponse{
		Allowed:          permResp.Allowed,
		Decision:         string(permResp.Decision),
		EvaluationTimeMs: uint(permResp.EvaluationTime.Milliseconds()),
		RequestID:        &permResp.RequestID,
	}
}

// convertToPolicyExplanationResponse converts domain explanation to Goa response
func (h *ABACGoaHandler) convertToPolicyExplanationResponse(explanation *abacServices.DecisionExplanation) *goaABAC.PolicyExplanationResponse {
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
			goaResp.PolicyEvaluations[i] = h.convertPolicyEvaluationSummary(eval)
		}
	}

	// Convert conflict resolution
	if explanation.ConflictResolution != nil {
		goaResp.ConflictResolution = h.convertConflictResolutionSummary(explanation.ConflictResolution)
	}

	return goaResp
}

// convertToPolicyDiscoveryResponse converts domain policies to Goa response
func (h *ABACGoaHandler) convertToPolicyDiscoveryResponse(policies []*abacServices.PolicySummary) *goaABAC.PolicyDiscoveryResponse {
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

	return &goaABAC.PolicyDiscoveryResponse{
		Policies:           goaPolicies,
		TotalPolicies:      uint(len(policies)), // TODO: Get actual total from service
		ApplicablePolicies: uint(len(policies)),
	}
}

// convertToAttributeCollectionResponse converts domain attribute context to Goa response
func (h *ABACGoaHandler) convertToAttributeCollectionResponse(attrContext *abacModels.AttributeContext) *goaABAC.AttributeCollectionResponse {
	goaResp := &goaABAC.AttributeCollectionResponse{
		CollectionTimeMs: 0, // TODO: Calculate actual collection time
	}

	// Count total attributes
	totalAttrs := 0

	// Convert user attributes
	if attrContext.UserAttributes != nil {
		goaResp.UserAttributes = h.convertAttributeMap(attrContext.UserAttributes)
		totalAttrs += len(attrContext.UserAttributes)
	}

	// Convert resource attributes
	if attrContext.ResourceAttributes != nil {
		goaResp.ResourceAttributes = h.convertAttributeMap(attrContext.ResourceAttributes)
		totalAttrs += len(attrContext.ResourceAttributes)
	}

	// Convert environment attributes
	if attrContext.EnvironmentAttributes != nil {
		goaResp.EnvironmentAttributes = h.convertAttributeMap(attrContext.EnvironmentAttributes)
		totalAttrs += len(attrContext.EnvironmentAttributes)
	}

	// Convert action attributes
	if attrContext.ActionAttributes != nil {
		goaResp.ActionAttributes = h.convertAttributeMap(attrContext.ActionAttributes)
		totalAttrs += len(attrContext.ActionAttributes)
	}

	// Convert entity attributes
	if attrContext.EntityAttributes != nil {
		goaResp.EntityAttributes = h.convertAttributeMap(attrContext.EntityAttributes)
		totalAttrs += len(attrContext.EntityAttributes)
	}

	// Convert session attributes
	if attrContext.SessionAttributes != nil {
		goaResp.SessionAttributes = h.convertAttributeMap(attrContext.SessionAttributes)
		totalAttrs += len(attrContext.SessionAttributes)
	}

	goaResp.TotalAttributes = uint(totalAttrs)

	return goaResp
}

// convertToDecisionAuditResponse converts domain audit logs to Goa response
func (h *ABACGoaHandler) convertToDecisionAuditResponse(decisions []*abacServices.DecisionAuditLog) *goaABAC.DecisionAuditResponse {
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

	return &goaABAC.DecisionAuditResponse{
		Decisions:  goaEntries,
		TotalCount: uint(len(decisions)), // TODO: Get actual total count
	}
}

// Helper conversion methods

func (h *ABACGoaHandler) convertPolicyObligation(obligation *abacModels.PolicyObligation) *goaABAC.PolicyObligation {
	return &goaABAC.PolicyObligation{
		ID:          obligation.ID,
		Type:        obligation.Type,
		Description: obligation.Description,
		Parameters:  obligation.Parameters,
	}
}

func (h *ABACGoaHandler) convertPolicyAdvice(advice *abacModels.PolicyAdvice) *goaABAC.PolicyAdvice {
	return &goaABAC.PolicyAdvice{
		ID:          advice.ID,
		Type:        advice.Type,
		Description: advice.Description,
		Severity:    advice.Severity,
		Parameters:  advice.Parameters,
	}
}

func (h *ABACGoaHandler) convertPolicyExplanation(explanation *abacServices.DecisionExplanation) *goaABAC.PolicyExplanation {
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
			goaExplanation.PolicyEvaluations[i] = h.convertPolicyEvaluationSummary(eval)
		}
	}

	// Convert conflict resolution
	if explanation.ConflictResolution != nil {
		goaExplanation.ConflictResolution = h.convertConflictResolutionSummary(explanation.ConflictResolution)
	}

	return goaExplanation
}

func (h *ABACGoaHandler) convertPolicyEvaluationSummary(eval *abacServices.PolicyEvaluationSummary) *goaABAC.PolicyEvaluationSummary {
	return &goaABAC.PolicyEvaluationSummary{
		PolicyName:   eval.PolicyName,
		Decision:     string(eval.Decision),
		Applicable:   eval.Applicable,
		MatchedRules: eval.MatchedRules,
		FailedRules:  eval.FailedRules,
		Reason:       eval.Reason,
	}
}

func (h *ABACGoaHandler) convertConflictResolutionSummary(resolution *abacServices.ConflictResolutionSummary) *goaABAC.ConflictResolutionSummary {
	return &goaABAC.ConflictResolutionSummary{
		ConflictDetected:    resolution.ConflictDetected,
		ConflictingPolicies: resolution.ConflictingPolicies,
		ResolutionMethod:    resolution.ResolutionMethod,
		WinningPolicy:       &resolution.WinningPolicy,
		Explanation:         resolution.Explanation,
	}
}

func (h *ABACGoaHandler) convertDecisionAuditTrail(trail *abacServices.DecisionAuditTrail) *goaABAC.DecisionAuditTrail {
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

func (h *ABACGoaHandler) convertAttributeMap(attributes map[string]*abacModels.AttributeValue) map[string]*goaABAC.AttributeValue {
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

// convertDomainErrorToGoa converts domain errors to Goa errors
func (h *ABACGoaHandler) convertDomainErrorToGoa(err error) error {
	switch {
	case errors.Is(err, errors.ErrNotFound):
		return goaABAC.MakeNotFound(err)
	case errors.Is(err, errors.ErrUnauthorized):
		return goaABAC.MakeUnauthorized(err)
	case errors.Is(err, errors.ErrForbidden):
		return goaABAC.MakeForbidden(err)
	case errors.Is(err, errors.ErrValidation):
		return goaABAC.MakeBadRequest(err)
	default:
		return goaABAC.MakeInternalError(err)
	}
}
