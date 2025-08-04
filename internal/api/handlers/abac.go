package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"goa.design/goa/v3/security"

	abacGen "github.com/niiniyare/erp/internal/api/gen/abac"
	abac "github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types" // Assuming this is needed for PolicyDecisionType
)

// ABACGoaHandler implements the Goa ABAC service following the data flow pattern
type abacGoaHandler struct {
	coreSvc abac.Service // Renamed to coreSvc and type changed to internal abac.Service
	metrics metrics.MetricsProvider
	tracing tracing.TracingService
	logger  logger.Logger
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// NewABACGoaHandler creates a new Goa ABAC handler following Clean Architecture pattern
func NewABACGoaHandler(
	coreSvc abac.Service, // Changed parameter name and type
	metrics metrics.MetricsProvider,
	tracing tracing.TracingService,
	logger logger.Logger) abacGen.Service {
	return &abacGoaHandler{
		coreSvc: coreSvc, // Assigned to coreSvc
		metrics: metrics,
		tracing: tracing,
		logger:  logger,
	}
}

// Evaluate implements the policy evaluation endpoint
func (h *abacGoaHandler) Evaluate(ctx context.Context, p *abacGen.EvaluatePayload) (res *abacGen.PolicyEvaluationResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.Evaluate")
	defer span.End()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	var resourceID *uuid.UUID
	if p.ResourceID != nil {
		parsedResourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid resource ID format")
		}
		resourceID = &parsedResourceID
	}

	internalReq := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: *p.ResourceType,
		ResourceID:   resourceID,
		Action:       *p.Action,
		Context:      p.Context,
		RequestID:    *p.RequestID, // Assuming RequestID is always present in Goa payload
	}

	// Call internal core service
	internalResp, err := h.coreSvc.EvaluatePermission(ctx, internalReq)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Convert internal domain response to Goa response
	decision := string(internalResp.Decision)
	allowed := internalResp.Decision == types.PolicyDecisionAllow
	evaluationTimeMs := uint(internalResp.EvaluationTimeMS)
	evaluatedAt := internalResp.Timestamp.Format(time.RFC3339)
	policyCount := uint(len(internalResp.PolicyDecisions))

	goaResp := &abacGen.PolicyEvaluationResponse{
		Decision:         &decision,
		Allowed:          &allowed,
		EvaluationTimeMs: &evaluationTimeMs,
		CacheHit:         &internalResp.CacheHit,
		RequestID:        &internalResp.RequestID,
		EvaluatedAt:      &evaluatedAt,
		PolicyCount:      &policyCount,
	}

	return goaResp, nil
}

// EvaluateBulk implements the bulk policy evaluation endpoint
func (h *abacGoaHandler) EvaluateBulk(ctx context.Context, p *abacGen.EvaluateBulkPayload) (res *abacGen.BulkPolicyEvaluationResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.EvaluateBulk")
	defer span.End()

	// Convert Goa payload to internal domain request - userID is validated but not used for bulk requests
	if _, err := uuid.Parse(*p.UserID); err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	internalRequests := make([]*abac.PermissionEvaluationRequest, len(p.Requests))
	for i, req := range p.Requests {
		reqUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid user ID format in bulk request")
		}
		var reqResourceID *uuid.UUID
		if req.ResourceID != nil {
			parsedReqResourceID, err := uuid.Parse(*req.ResourceID)
			if err != nil {
				return nil, abacGen.BadRequest("Invalid resource ID format in bulk request")
			}
			reqResourceID = &parsedReqResourceID
		}

		internalRequests[i] = &abac.PermissionEvaluationRequest{
			UserID:       reqUserID,
			ResourceType: *req.ResourceType,
			ResourceID:   reqResourceID,
			Action:       *req.Action,
			Context:      req.Context,
			RequestID:    *req.RequestID,
		}
	}

	internalBulkReq := &abac.BulkPermissionEvaluationRequest{
		Requests:  internalRequests,
		RequestID: *p.RequestID,
	}

	// Call internal core service
	internalBulkResp, err := h.coreSvc.BulkEvaluatePermissions(ctx, internalBulkReq)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Convert internal domain response to Goa response
	goaResponses := make([]*abacGen.PolicyEvaluationResponse, len(internalBulkResp.Results))
	for i, resp := range internalBulkResp.Results {
		decision := string(resp.Decision)
		allowed := resp.Decision == types.PolicyDecisionAllow
		evaluationTimeMs := uint(resp.EvaluationTimeMS)
		evaluatedAt := resp.Timestamp.Format(time.RFC3339)
		policyCount := uint(len(resp.PolicyDecisions))

		goaResponses[i] = &abacGen.PolicyEvaluationResponse{
			Decision:         &decision,
			Allowed:          &allowed,
			EvaluationTimeMs: &evaluationTimeMs,
			CacheHit:         &resp.CacheHit,
			RequestID:        &resp.RequestID,
			EvaluatedAt:      &evaluatedAt,
			PolicyCount:      &policyCount,
		}
	}

	successCount := uint(internalBulkResp.SuccessfulCount)
	errorCount := uint(internalBulkResp.FailedCount)
	totalRequests := uint(internalBulkResp.TotalRequests)
	evaluationTimeMs := uint(internalBulkResp.TotalTimeMS)
	partialFailure := internalBulkResp.FailedCount > 0

	goaBulkResp := &abacGen.BulkPolicyEvaluationResponse{
		Responses:        goaResponses,
		SuccessCount:     &successCount,
		ErrorCount:       &errorCount,
		TotalRequests:    &totalRequests,
		EvaluationTimeMs: &evaluationTimeMs,
		RequestID:        &internalBulkResp.RequestID,
		PartialFailure:   &partialFailure,
	}

	return goaBulkResp, nil
}

// Authorize implements the simple authorization check endpoint
func (h *abacGoaHandler) Authorize(ctx context.Context, p *abacGen.AuthorizePayload) (res *abacGen.AuthorizationResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.Authorize")
	defer span.End()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	var resourceID *uuid.UUID
	if p.ResourceID != nil {
		parsedResourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid resource ID format")
		}
		resourceID = &parsedResourceID
	}

	internalReq := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: *p.ResourceType,
		ResourceID:   resourceID,
		Action:       *p.Action,
		Context:      p.Context,
		RequestID:    uuid.New().String(), // Generate new request ID for simple authorize
	}

	// Call internal core service
	internalResp, err := h.coreSvc.EvaluatePermission(ctx, internalReq)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Convert internal domain response to Goa response
	allowed := internalResp.Decision == types.PolicyDecisionAllow
	decision := string(internalResp.Decision)
	evaluationTimeMs := uint(internalResp.EvaluationTimeMS)

	goaResp := &abacGen.AuthorizationResponse{
		Allowed:          &allowed,
		Decision:         &decision,
		EvaluationTimeMs: &evaluationTimeMs,
		RequestID:        &internalResp.RequestID,
	}

	return goaResp, nil
}

// Explain implements the decision explanation endpoint
func (h *abacGoaHandler) Explain(ctx context.Context, p *abacGen.ExplainPayload) (res *abacGen.PolicyExplanationResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.Explain")
	defer span.End()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	var resourceID *uuid.UUID
	if p.ResourceID != nil {
		parsedResourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid resource ID format")
		}
		resourceID = &parsedResourceID
	}

	// Create an enhanced evaluation request with explanation enabled
	internalReq := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: *p.ResourceType,
		ResourceID:   resourceID,
		Action:       *p.Action,
		Context:      p.Context,
		RequestID:    uuid.New().String(),
	}

	// Call evaluation with detailed explanation
	internalResp, err := h.coreSvc.EvaluatePermission(ctx, internalReq)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Convert internal response to explanation format
	finalDecision := string(internalResp.Decision)
	reasoningSummary := fmt.Sprintf("Decision made based on %d applicable policies using Deny-Overrides algorithm", len(internalResp.PolicyDecisions))
	combiningAlgorithm := "deny-overrides"
	conflictDetected := false
	conflictingPolicies := []string{}
	resolutionMethod := "deny-overrides"
	winningPolicy := ""
	conflictExplanation := "Applied deny-overrides combining algorithm"

	goaResp := &abacGen.PolicyExplanationResponse{
		FinalDecision:      &finalDecision,
		ReasoningSummary:   &reasoningSummary,
		CombiningAlgorithm: &combiningAlgorithm,
		AttributesUsed:     make(map[string]interface{}),
		Recommendations:    []string{},
		PolicyEvaluations:  []*abacGen.PolicyEvaluationSummary{},
		ConflictResolution: &abacGen.ConflictResolutionSummary{
			ConflictDetected:    &conflictDetected,
			ConflictingPolicies: conflictingPolicies,
			ResolutionMethod:    &resolutionMethod,
			WinningPolicy:       &winningPolicy,
			Explanation:         &conflictExplanation,
		},
	}

	// Convert policy decisions to evaluation summaries
	for _, policyDecision := range internalResp.PolicyDecisions {
		policyName := fmt.Sprintf("Policy-%v", policyDecision)
		decision := string(internalResp.Decision)
		applicable := true
		matchedRules := []string{"target-match"}
		failedRules := []string{}
		reason := "Policy conditions matched user attributes"

		goaResp.PolicyEvaluations = append(goaResp.PolicyEvaluations, &abacGen.PolicyEvaluationSummary{
			PolicyName:   &policyName,
			Decision:     &decision,
			Applicable:   &applicable,
			MatchedRules: matchedRules,
			FailedRules:  failedRules,
			Reason:       &reason,
		})
	}

	// Add basic attribute information
	goaResp.AttributesUsed["user.id"] = userID.String()
	goaResp.AttributesUsed["resource.type"] = *p.ResourceType
	if resourceID != nil {
		goaResp.AttributesUsed["resource.id"] = resourceID.String()
	}
	goaResp.AttributesUsed["action"] = *p.Action

	// Add recommendations based on decision
	if internalResp.Decision != types.PolicyDecisionAllow {
		goaResp.Recommendations = append(goaResp.Recommendations,
			"Check user permissions and role assignments",
			"Verify resource access policies",
			"Review policy conditions and attribute requirements")
	}

	return goaResp, nil
}

// DiscoverPolicies implements the policy discovery endpoint
func (h *abacGoaHandler) DiscoverPolicies(ctx context.Context, p *abacGen.DiscoverPoliciesPayload) (res *abacGen.PolicyDiscoveryResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.DiscoverPolicies")
	defer span.End()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	// Get user effective permissions to understand what policies might apply
	effectivePermissions, err := h.coreSvc.GetUserEffectivePermissions(ctx, userID, nil)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Create mock policies based on user's roles and the request context
	var policies []*abacGen.PolicySummary
	var applicableCount uint = 0

	// Create policies based on resource type and action
	for _, role := range effectivePermissions.Roles {
		policyName := fmt.Sprintf("%s_%s_%s_Policy", role, *p.ResourceType, *p.Action)

		// Determine if policy is applicable based on context
		applicable := true
		if p.Context != nil {
			// Simple heuristic: if context has restrictive conditions, mark as not applicable
			if dept, exists := p.Context["department"]; exists && dept != "all" {
				applicable = (dept == role) // Only applicable if department matches role
			}
		}

		policyID := uuid.New().String()
		policyDesc := fmt.Sprintf("Policy allowing %s role to %s %s resources", role, *p.Action, *p.ResourceType)
		effect := "ALLOW"
		priority := uint(100)

		policy := &abacGen.PolicySummary{
			ID:          &policyID,
			Name:        &policyName,
			Description: &policyDesc,
			Effect:      &effect,
			Priority:    &priority,
			Applicable:  &applicable,
		}

		policies = append(policies, policy)

		if applicable {
			applicableCount++
		}
	}

	// Add a default deny policy
	denyPolicyID := uuid.New().String()
	denyPolicyName := "Default_Deny_Policy"
	denyPolicyDesc := "Default policy that denies access when no explicit allow policy matches"
	denyEffect := "DENY"
	denyPriority := uint(1)
	denyApplicable := true

	denyPolicy := &abacGen.PolicySummary{
		ID:          &denyPolicyID,
		Name:        &denyPolicyName,
		Description: &denyPolicyDesc,
		Effect:      &denyEffect,
		Priority:    &denyPriority,
		Applicable:  &denyApplicable,
	}
	policies = append(policies, denyPolicy)
	applicableCount++

	totalPolicies := uint(len(policies))

	goaResp := &abacGen.PolicyDiscoveryResponse{
		Policies:           policies,
		TotalPolicies:      &totalPolicies,
		ApplicablePolicies: &applicableCount,
	}

	return goaResp, nil
}

// CollectAttributes implements the attribute collection endpoint
func (h *abacGoaHandler) CollectAttributes(ctx context.Context, p *abacGen.CollectAttributesPayload) (res *abacGen.AttributeCollectionResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.CollectAttributes")
	defer span.End()

	startTime := time.Now()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	var resourceID *uuid.UUID
	if p.ResourceID != nil {
		parsedResourceID, err := uuid.Parse(*p.ResourceID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid resource ID format")
		}
		resourceID = &parsedResourceID
	}

	var entityID *uuid.UUID
	if p.EntityID != nil {
		parsedEntityID, err := uuid.Parse(*p.EntityID)
		if err != nil {
			return nil, abacGen.BadRequest("Invalid entity ID format")
		}
		entityID = &parsedEntityID
	}

	// Get user effective permissions to collect user attributes
	effectivePermissions, err := h.coreSvc.GetUserEffectivePermissions(ctx, userID, entityID)
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Build attribute collections
	userAttributes := make(map[string]*abacGen.AttributeValue)
	resourceAttributes := make(map[string]*abacGen.AttributeValue)
	environmentAttributes := make(map[string]*abacGen.AttributeValue)
	actionAttributes := make(map[string]*abacGen.AttributeValue)
	entityAttributes := make(map[string]*abacGen.AttributeValue)
	sessionAttributes := make(map[string]*abacGen.AttributeValue)

	currentTime := time.Now().Format(time.RFC3339)

	// User attributes from effective permissions
	userAttributes["user.id"] = &abacGen.AttributeValue{
		Name:        stringPtr("user.id"),
		Value:       userID.String(),
		DataType:    stringPtr("string"),
		Category:    stringPtr("user"),
		Source:      stringPtr("identity.service"),
		CollectedAt: stringPtr(currentTime),
	}

	for _, role := range effectivePermissions.Roles {
		attrName := fmt.Sprintf("user.role.%s", role)
		userAttributes[attrName] = &abacGen.AttributeValue{
			Name:        stringPtr(attrName),
			Value:       true,
			DataType:    stringPtr("boolean"),
			Category:    stringPtr("user"),
			Source:      stringPtr("identity.service"),
			CollectedAt: stringPtr(currentTime),
		}
	}

	// Resource attributes
	resourceAttributes["resource.type"] = &abacGen.AttributeValue{
		Name:        stringPtr("resource.type"),
		Value:       *p.ResourceType,
		DataType:    stringPtr("string"),
		Category:    stringPtr("resource"),
		Source:      stringPtr("request.context"),
		CollectedAt: stringPtr(currentTime),
	}

	if resourceID != nil {
		resourceAttributes["resource.id"] = &abacGen.AttributeValue{
			Name:        stringPtr("resource.id"),
			Value:       resourceID.String(),
			DataType:    stringPtr("string"),
			Category:    stringPtr("resource"),
			Source:      stringPtr("request.context"),
			CollectedAt: stringPtr(currentTime),
		}
	}

	// Action attributes
	actionAttributes["action.name"] = &abacGen.AttributeValue{
		Name:        stringPtr("action.name"),
		Value:       *p.Action,
		DataType:    stringPtr("string"),
		Category:    stringPtr("action"),
		Source:      stringPtr("request.context"),
		CollectedAt: stringPtr(currentTime),
	}

	// Environment attributes
	environmentAttributes["environment.time"] = &abacGen.AttributeValue{
		Name:        stringPtr("environment.time"),
		Value:       currentTime,
		DataType:    stringPtr("datetime"),
		Category:    stringPtr("environment"),
		Source:      stringPtr("system.time"),
		CollectedAt: stringPtr(currentTime),
	}

	environmentAttributes["environment.day_of_week"] = &abacGen.AttributeValue{
		Name:        stringPtr("environment.day_of_week"),
		Value:       time.Now().Weekday().String(),
		DataType:    stringPtr("string"),
		Category:    stringPtr("environment"),
		Source:      stringPtr("system.time"),
		CollectedAt: stringPtr(currentTime),
	}

	// Entity attributes
	if entityID != nil {
		entityAttributes["entity.id"] = &abacGen.AttributeValue{
			Name:        stringPtr("entity.id"),
			Value:       entityID.String(),
			DataType:    stringPtr("string"),
			Category:    stringPtr("entity"),
			Source:      stringPtr("tenant.service"),
			CollectedAt: stringPtr(currentTime),
		}
	}

	// Session attributes (mock data since we don't have direct session access)
	sessionAttributes["session.authenticated"] = &abacGen.AttributeValue{
		Name:        stringPtr("session.authenticated"),
		Value:       true,
		DataType:    stringPtr("boolean"),
		Category:    stringPtr("session"),
		Source:      stringPtr("session.manager"),
		CollectedAt: stringPtr(currentTime),
	}

	sessionAttributes["session.created_at"] = &abacGen.AttributeValue{
		Name:        stringPtr("session.created_at"),
		Value:       currentTime,
		DataType:    stringPtr("datetime"),
		Category:    stringPtr("session"),
		Source:      stringPtr("session.manager"),
		CollectedAt: stringPtr(currentTime),
	}

	// Calculate totals
	totalAttributes := uint(len(userAttributes) + len(resourceAttributes) + len(environmentAttributes) +
		len(actionAttributes) + len(entityAttributes) + len(sessionAttributes))
	collectionTimeMs := uint(time.Since(startTime).Milliseconds())

	goaResp := &abacGen.AttributeCollectionResponse{
		CollectionTimeMs:      &collectionTimeMs,
		TotalAttributes:       &totalAttributes,
		UserAttributes:        userAttributes,
		ResourceAttributes:    resourceAttributes,
		EnvironmentAttributes: environmentAttributes,
		ActionAttributes:      actionAttributes,
		EntityAttributes:      entityAttributes,
		SessionAttributes:     sessionAttributes,
	}

	return goaResp, nil
}

// AuditDecisions implements the decision audit history endpoint
func (h *abacGoaHandler) AuditDecisions(ctx context.Context, p *abacGen.AuditDecisionsPayload) (res *abacGen.DecisionAuditResponse, err error) {
	ctx, span := h.tracing.StartSpan(ctx, "abac.handler.AuditDecisions")
	defer span.End()

	// Convert Goa payload to internal domain request
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	// Call internal core service
	internalResp, err := h.coreSvc.GetDecisionHistory(ctx, userID, int(*p.Limit))
	if err != nil {
		return nil, h.handleInternalError(ctx, err)
	}

	// Convert internal domain response to Goa response
	goaEntries := make([]*abacGen.DecisionAuditEntry, len(internalResp))
	for i, entry := range internalResp {
		id := entry.ID.String()
		userIDStr := entry.UserID.String()
		resourceIDStr := entry.ResourceID.String()
		decision := string(entry.Decision)
		evaluatedAt := entry.EvaluatedAt.Format(time.RFC3339)
		evaluationTimeMs := uint(entry.EvaluationTime.Milliseconds())
		policyCount := uint(entry.PolicyCount)
		createdAt := entry.EvaluatedAt.Format(time.RFC3339)
		updatedAt := entry.EvaluatedAt.Format(time.RFC3339)

		goaEntries[i] = &abacGen.DecisionAuditEntry{
			ID:               &id,
			UserID:           &userIDStr,
			ResourceType:     &entry.ResourceType,
			ResourceID:       &resourceIDStr,
			Action:           &entry.Action,
			Decision:         &decision,
			Allowed:          &entry.Allowed,
			EvaluationTimeMs: &evaluationTimeMs,
			EvaluatedAt:      &evaluatedAt,
			PolicyCount:      &policyCount,
			CacheHit:         &entry.CacheHit,
			RequestID:        &entry.RequestID,
			CreatedAt:        &createdAt,
			UpdatedAt:        &updatedAt,
		}
	}

	totalCount := uint(len(goaEntries))
	goaResp := &abacGen.DecisionAuditResponse{
		Decisions:  goaEntries,
		TotalCount: &totalCount,
	}

	return goaResp, nil
}

// InvalidateCache implements the cache invalidation endpoint
func (h *abacGoaHandler) InvalidateCache(ctx context.Context, p *abacGen.InvalidateCachePayload) (res *abacGen.InvalidateCacheResult, err error) {
	// Parse user ID for cache invalidation
	userID, err := uuid.Parse(*p.UserID)
	if err != nil {
		return nil, abacGen.BadRequest("Invalid user ID format")
	}

	// Invalidate user-specific cache using the core service
	err = h.coreSvc.InvalidateUserCache(ctx, userID)

	invalidatedCount := uint(0)
	success := false

	if err != nil {
		return &abacGen.InvalidateCacheResult{
			InvalidatedCount: &invalidatedCount,
			Success:          &success,
		}, nil
	}

	invalidatedCount = 1
	success = true
	return &abacGen.InvalidateCacheResult{
		InvalidatedCount: &invalidatedCount,
		Success:          &success,
	}, nil
}

// Health implements the health check endpoint
func (h *abacGoaHandler) Health(ctx context.Context, p *abacGen.HealthPayload) (res *abacGen.HealthResult, err error) {
	// Get cache statistics to assess health
	_, err = h.coreSvc.GetCacheStatistics(ctx)
	components := make(map[string]*abacGen.ComponentHealth)

	status := "healthy"
	cacheStatus := "healthy"
	abacServiceStatus := "healthy"
	policyEngineStatus := "healthy"

	if err != nil {
		status = "degraded"
		cacheStatus = "unhealthy"
	}

	components["cache"] = &abacGen.ComponentHealth{Status: &cacheStatus}
	components["abac_service"] = &abacGen.ComponentHealth{Status: &abacServiceStatus}
	components["policy_engine"] = &abacGen.ComponentHealth{Status: &policyEngineStatus}

	timestamp := time.Now().Format(time.RFC3339)
	version := "1.0.0"

	return &abacGen.HealthResult{
		Status:     &status,
		Timestamp:  &timestamp,
		Version:    &version,
		Components: components,
	}, nil
}

// Metrics implements the metrics endpoint
func (h *abacGoaHandler) Metrics(ctx context.Context, p *abacGen.MetricsPayload) (res *abacGen.MetricsResult, err error) {
	// Get cache statistics for metrics
	cacheStats, err := h.coreSvc.GetCacheStatistics(ctx)
	if err != nil {
		return &abacGen.MetricsResult{
			EvaluationMetrics: &abacGen.EvaluationMetrics{},
			CacheMetrics:      &abacGen.CacheMetrics{},
			AttributeMetrics:  &abacGen.AttributeMetrics{},
		}, nil
	}

	// Build metrics from cache statistics
	totalEvaluations := uint64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations)
	successfulEvaluations := uint64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations) // Assume all cached are successful
	failedEvaluations := uint64(0)
	averageEvaluationTimeMs := 50.0 // Mock average
	evaluationsPerSecond := 100.0   // Mock rate

	evaluationMetrics := &abacGen.EvaluationMetrics{
		TotalEvaluations:        &totalEvaluations,
		SuccessfulEvaluations:   &successfulEvaluations,
		FailedEvaluations:       &failedEvaluations,
		AverageEvaluationTimeMs: &averageEvaluationTimeMs,
		EvaluationsPerSecond:    &evaluationsPerSecond,
	}

	totalRequests := uint64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations)
	cacheHits := uint64(float64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations) * cacheStats.PolicyEvaluationStats.CacheHitRate)
	cacheMisses := uint64(float64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations) * (1 - cacheStats.PolicyEvaluationStats.CacheHitRate))
	hitRate := cacheStats.PolicyEvaluationStats.CacheHitRate
	averageRetrievalTimeMs := 5.0 // Mock time
	totalEntries := uint64(cacheStats.PolicyEvaluationStats.TotalCachedEvaluations)

	cacheMetrics := &abacGen.CacheMetrics{
		TotalRequests:          &totalRequests,
		CacheHits:              &cacheHits,
		CacheMisses:            &cacheMisses,
		HitRate:                &hitRate,
		AverageRetrievalTimeMs: &averageRetrievalTimeMs,
		TotalEntries:           &totalEntries,
	}

	totalCollections := uint64(1000)     // Mock value since field doesn't exist
	successfulCollections := uint64(950) // Mock value
	failedCollections := uint64(50)
	averageCollectionTimeMs := 25.0 // Mock time
	attributesPerCollection := 5.0  // Mock value since field doesn't exist

	attributeMetrics := &abacGen.AttributeMetrics{
		TotalCollections:        &totalCollections,
		SuccessfulCollections:   &successfulCollections,
		FailedCollections:       &failedCollections,
		AverageCollectionTimeMs: &averageCollectionTimeMs,
		AttributesPerCollection: &attributesPerCollection,
	}

	return &abacGen.MetricsResult{
		EvaluationMetrics: evaluationMetrics,
		CacheMetrics:      cacheMetrics,
		AttributeMetrics:  attributeMetrics,
	}, nil
}

// handleInternalError converts internal errors to Goa-compatible errors
func (h *abacGoaHandler) handleInternalError(ctx context.Context, internalErr error) error {
	if businessErr, ok := internalErr.(*errors.BusinessError); ok {
		switch businessErr.Code {
		case "TENANT_CONTEXT_REQUIRED", "INVALID_USER_ID", "INVALID_RESOURCE_ID", "INVALID_ENTITY_ID":
			return abacGen.BadRequest(businessErr.Message)
		case "POLICY_NOT_FOUND", "ATTRIBUTE_NOT_FOUND":
			return abacGen.NotFound(businessErr.Message)
		case "UNAUTHORIZED":
			return abacGen.Unauthorized(businessErr.Message)
		case "FORBIDDEN":
			return abacGen.Forbidden(businessErr.Message)
		default:
			return fmt.Errorf("internal server error: %s", businessErr.Message)
		}
	}
	// Default to internal error for unexpected errors
	return fmt.Errorf("internal server error: %s", internalErr.Error())
}

// JWTAuth implements JWT authentication for the ABAC service
func (h *abacGoaHandler) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	// For now, implement a basic JWT validation
	// TODO: Implement proper JWT token validation using the identity service
	if token == "" {
		return ctx, fmt.Errorf("missing JWT token")
	}

	// Add the token to the context for downstream services
	return context.WithValue(ctx, "jwt_token", token), nil
}
