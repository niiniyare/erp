package handlers

import (
	"context"
	"fmt"

	// Generated Goa interfaces
	goaABAC "github.com/niiniyare/erp/internal/api/gen/abac"

	// Core ABAC services
	abacServices "github.com/niiniyare/erp/internal/core/abac/services"

	// Shared infrastructure
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// ABACGoaHandler implements the Goa ABAC service following the data flow pattern
type ABACGoaHandler struct {
	policyDecisionService abacServices.PolicyDecisionService
	policyEvalService     abacServices.PolicyEvaluationService
	attrCollectionService abacServices.AttributeCollectionService
	tracing               tracing.TracingService
	metrics               *metrics.MetricsService
}

// NewABACGoaHandler creates a new Goa ABAC handler following Clean Architecture pattern
func NewABACGoaHandler(
	policyDecisionService abacServices.PolicyDecisionService,
	policyEvalService abacServices.PolicyEvaluationService,
	attrCollectionService abacServices.AttributeCollectionService,
	tracing tracing.TracingService,
	metrics *metrics.MetricsService,
) goaABAC.Service {
	return &ABACGoaHandler{
		policyDecisionService: policyDecisionService,
		policyEvalService:     policyEvalService,
		attrCollectionService: attrCollectionService,
		tracing:               tracing,
		metrics:               metrics,
	}
}

// Evaluate implements the policy evaluation endpoint
func (h *ABACGoaHandler) Evaluate(ctx context.Context, p *goaABAC.PolicyEvaluationRequest) (*goaABAC.PolicyEvaluationResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.evaluate",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.String("resource_type", p.ResourceType),
			attribute.String("action", p.Action),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_evaluate_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "ABAC policy evaluation request", logger.Fields{
		"user_id":       p.UserID,
		"resource_type": p.ResourceType,
		"action":        p.Action,
		"request_id":    p.RequestID,
	})

	// Convert Goa request to domain request
	domainReq, err := h.convertToDecisionRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid request: %w", err))
	}

	// Make access decision using domain service
	domainResp, err := h.policyDecisionService.MakeDecision(ctx, domainReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert domain response to Goa response
	goaResp := h.convertToPolicyEvaluationResponse(domainResp)

	// Record metrics
	h.metrics.IncrementCounter("abac_evaluate_total", metrics.Fields{
		"decision": string(domainResp.Decision),
		"allowed":  fmt.Sprintf("%v", domainResp.Allowed),
	})

	span.SetAttributes(
		attribute.String("result.decision", string(domainResp.Decision)),
		attribute.Bool("result.allowed", domainResp.Allowed),
		attribute.Int("result.policy_count", domainResp.PolicyCount),
	)

	logger.InfoContext(ctx, "ABAC policy evaluation completed", logger.Fields{
		"decision":        string(domainResp.Decision),
		"allowed":         domainResp.Allowed,
		"policy_count":    domainResp.PolicyCount,
		"cache_hit":       domainResp.CacheHit,
		"evaluation_time": domainResp.EvaluationTime.Milliseconds(),
	})

	return goaResp, nil
}

// EvaluateBulk implements the bulk policy evaluation endpoint
func (h *ABACGoaHandler) EvaluateBulk(ctx context.Context, p *goaABAC.BulkPolicyEvaluationRequest) (*goaABAC.BulkPolicyEvaluationResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.evaluate_bulk",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.Int("request_count", len(p.Requests)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_evaluate_bulk_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "ABAC bulk policy evaluation request", logger.Fields{
		"user_id":       p.UserID,
		"request_count": len(p.Requests),
		"request_id":    p.RequestID,
	})

	// Convert to domain bulk request
	domainReq, err := h.convertToBulkDecisionRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid bulk request: %w", err))
	}

	// Execute bulk decision
	domainResp, err := h.policyDecisionService.MakeBulkDecision(ctx, domainReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert response
	goaResp := h.convertToBulkPolicyEvaluationResponse(domainResp)

	// Record metrics
	h.metrics.IncrementCounter("abac_evaluate_bulk_total", metrics.Fields{
		"total_requests": fmt.Sprintf("%d", domainResp.TotalRequests),
		"success_count":  fmt.Sprintf("%d", domainResp.SuccessCount),
		"error_count":    fmt.Sprintf("%d", domainResp.ErrorCount),
	})

	span.SetAttributes(
		attribute.Int("result.total_requests", domainResp.TotalRequests),
		attribute.Int("result.success_count", domainResp.SuccessCount),
		attribute.Int("result.error_count", domainResp.ErrorCount),
	)

	logger.InfoContext(ctx, "ABAC bulk policy evaluation completed", logger.Fields{
		"total_requests":  domainResp.TotalRequests,
		"success_count":   domainResp.SuccessCount,
		"error_count":     domainResp.ErrorCount,
		"evaluation_time": domainResp.EvaluationTime.Milliseconds(),
	})

	return goaResp, nil
}

// Authorize implements the simple authorization check endpoint
func (h *ABACGoaHandler) Authorize(ctx context.Context, p *goaABAC.AuthorizationRequest) (*goaABAC.AuthorizationResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.authorize",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.String("resource_type", p.ResourceType),
			attribute.String("action", p.Action),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_authorize_duration", metrics.Fields{})
	defer timer.Stop()

	// Convert to permission check request
	permReq, err := h.convertToPermissionCheckRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid authorization request: %w", err))
	}

	// Check permission
	permResp, err := h.policyDecisionService.CheckPermission(ctx, permReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert response
	goaResp := h.convertToAuthorizationResponse(permResp)

	// Record metrics
	h.metrics.IncrementCounter("abac_authorize_total", metrics.Fields{
		"decision": string(permResp.Decision),
		"allowed":  fmt.Sprintf("%v", permResp.Allowed),
	})

	span.SetAttributes(
		attribute.String("result.decision", string(permResp.Decision)),
		attribute.Bool("result.allowed", permResp.Allowed),
	)

	return goaResp, nil
}

// Explain implements the decision explanation endpoint
func (h *ABACGoaHandler) Explain(ctx context.Context, p *goaABAC.PolicyExplanationRequest) (*goaABAC.PolicyExplanationResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.explain",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.String("resource_type", p.ResourceType),
			attribute.String("action", p.Action),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_explain_duration", metrics.Fields{})
	defer timer.Stop()

	// Convert to explanation request
	explanationReq, err := h.convertToDecisionExplanationRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid explanation request: %w", err))
	}

	// Get explanation
	explanation, err := h.policyDecisionService.ExplainDecision(ctx, explanationReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert to Goa response
	goaResp := h.convertToPolicyExplanationResponse(explanation)

	// Record metrics
	h.metrics.IncrementCounter("abac_explain_total", metrics.Fields{})

	return goaResp, nil
}

// DiscoverPolicies implements the policy discovery endpoint
func (h *ABACGoaHandler) DiscoverPolicies(ctx context.Context, p *goaABAC.PolicyDiscoveryRequest) (*goaABAC.PolicyDiscoveryResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.discover_policies",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.String("resource_type", p.ResourceType),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_discover_policies_duration", metrics.Fields{})
	defer timer.Stop()

	// Convert to domain request
	domainReq, err := h.convertToPolicyDiscoveryRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid discovery request: %w", err))
	}

	// Discover policies
	policies, err := h.policyDecisionService.GetApplicablePolicies(ctx, domainReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert response
	goaResp := h.convertToPolicyDiscoveryResponse(policies)

	// Record metrics
	h.metrics.IncrementCounter("abac_discover_policies_total", metrics.Fields{
		"policy_count": fmt.Sprintf("%d", len(policies)),
	})

	span.SetAttributes(attribute.Int("result.policy_count", len(policies)))

	return goaResp, nil
}

// CollectAttributes implements the attribute collection endpoint
func (h *ABACGoaHandler) CollectAttributes(ctx context.Context, p *goaABAC.AttributeCollectionRequest) (*goaABAC.AttributeCollectionResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.collect_attributes",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
			attribute.String("resource_type", p.ResourceType),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_collect_attributes_duration", metrics.Fields{})
	defer timer.Stop()

	// Convert to domain request
	domainReq, err := h.convertToAttributeCollectionRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid attribute collection request: %w", err))
	}

	// Collect attributes
	attrContext, err := h.attrCollectionService.CollectAllAttributes(ctx, domainReq)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert response
	goaResp := h.convertToAttributeCollectionResponse(attrContext)

	// Record metrics
	h.metrics.IncrementCounter("abac_collect_attributes_total", metrics.Fields{})

	return goaResp, nil
}

// AuditDecisions implements the decision audit history endpoint
func (h *ABACGoaHandler) AuditDecisions(ctx context.Context, p *goaABAC.AuditDecisionsPayload) (*goaABAC.DecisionAuditResponse, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.audit_decisions",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("user_id", p.UserID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_audit_decisions_duration", metrics.Fields{})
	defer timer.Stop()

	// Convert to domain request
	userID, limit, err := h.convertToAuditDecisionsRequest(p)
	if err != nil {
		span.RecordError(err)
		return nil, goaABAC.MakeBadRequest(fmt.Errorf("invalid audit request: %w", err))
	}

	// Get decision history
	decisions, err := h.policyDecisionService.GetDecisionHistory(ctx, userID, limit)
	if err != nil {
		span.RecordError(err)
		return nil, h.convertDomainErrorToGoa(err)
	}

	// Convert response
	goaResp := h.convertToDecisionAuditResponse(decisions)

	// Record metrics
	h.metrics.IncrementCounter("abac_audit_decisions_total", metrics.Fields{})

	span.SetAttributes(attribute.Int("result.decision_count", len(decisions)))

	return goaResp, nil
}

// InvalidateCache implements the cache invalidation endpoint
func (h *ABACGoaHandler) InvalidateCache(ctx context.Context, p *goaABAC.InvalidateCachePayload) (*goaABAC.InvalidateCacheResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.invalidate_cache",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_invalidate_cache_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "ABAC cache invalidation requested", logger.Fields{
		"user_id":       p.UserID,
		"resource_type": p.ResourceType,
		"pattern":       p.Pattern,
	})

	// TODO: Implement cache invalidation logic
	// This would typically call a cache management service

	// For now, return a mock response
	goaResp := &goaABAC.InvalidateCacheResult{
		InvalidatedCount: 0, // TODO: Return actual count
		Success:          true,
	}

	// Record metrics
	h.metrics.IncrementCounter("abac_invalidate_cache_total", metrics.Fields{})

	return goaResp, nil
}

// Health implements the health check endpoint
func (h *ABACGoaHandler) Health(ctx context.Context) (*goaABAC.HealthResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.health",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_health_duration", metrics.Fields{})
	defer timer.Stop()

	// TODO: Implement actual health checks for dependencies
	goaResp := &goaABAC.HealthResult{
		Status:    "healthy",
		Timestamp: "2024-01-01T00:00:00Z", // TODO: Use actual timestamp
		Version:   "1.0.0",
		Components: map[string]*goaABAC.ComponentHealth{
			"policy_engine": {
				Status: "healthy",
			},
			"attribute_service": {
				Status: "healthy",
			},
			"cache": {
				Status: "healthy",
			},
		},
	}

	// Record metrics
	h.metrics.IncrementCounter("abac_health_total", metrics.Fields{})

	return goaResp, nil
}

// Metrics implements the metrics endpoint
func (h *ABACGoaHandler) Metrics(ctx context.Context) (*goaABAC.MetricsResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "abac.metrics",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("abac_metrics_duration", metrics.Fields{})
	defer timer.Stop()

	// TODO: Implement actual metrics collection
	goaResp := &goaABAC.MetricsResult{
		EvaluationMetrics: &goaABAC.EvaluationMetrics{
			TotalEvaluations:        0,
			SuccessfulEvaluations:   0,
			FailedEvaluations:       0,
			AverageEvaluationTimeMs: 0,
			EvaluationsPerSecond:    0,
		},
		CacheMetrics: &goaABAC.CacheMetrics{
			TotalRequests:          0,
			CacheHits:              0,
			CacheMisses:            0,
			HitRate:                0.0,
			AverageRetrievalTimeMs: 0,
			TotalEntries:           0,
		},
		AttributeMetrics: &goaABAC.AttributeMetrics{
			TotalCollections:        0,
			SuccessfulCollections:   0,
			FailedCollections:       0,
			AverageCollectionTimeMs: 0,
			AttributesPerCollection: 0.0,
		},
	}

	// Record metrics
	h.metrics.IncrementCounter("abac_metrics_total", metrics.Fields{})

	return goaResp, nil
}
