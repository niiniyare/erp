package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// PolicyDecisionService defines the interface for making policy decisions
type PolicyDecisionService interface {
	// Core decision methods
	MakeDecision(ctx context.Context, req *DecisionRequest) (*DecisionResponse, error)
	MakeBulkDecision(ctx context.Context, req *BulkDecisionRequest) (*BulkDecisionResponse, error)

	// Authorization checks
	IsAuthorized(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, action string, context map[string]interface{}) (bool, error)
	CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error)

	// Batch operations
	CheckMultiplePermissions(ctx context.Context, requests []*PermissionCheckRequest) ([]*PermissionCheckResponse, error)

	// Decision audit and logging
	LogDecision(ctx context.Context, decision *DecisionAuditLog) error
	GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*DecisionAuditLog, error)

	// Policy discovery
	GetApplicablePolicies(ctx context.Context, req *PolicyDiscoveryRequest) ([]*PolicySummary, error)
	ExplainDecision(ctx context.Context, req *DecisionExplanationRequest) (*DecisionExplanation, error)
}

// DecisionRequest represents a policy decision request
type DecisionRequest struct {
	UserID          uuid.UUID              `json:"user_id" validate:"required"`
	ResourceType    string                 `json:"resource_type" validate:"required"`
	ResourceID      *uuid.UUID             `json:"resource_id,omitempty"`
	Action          string                 `json:"action" validate:"required"`
	EntityID        *uuid.UUID             `json:"entity_id,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
	RequestID       string                 `json:"request_id"`
	UseCache        bool                   `json:"use_cache"`
	CacheResults    bool                   `json:"cache_results"`
	IncludeDetails  bool                   `json:"include_details"`
	IncludeAdvice   bool                   `json:"include_advice"`
	ExplainDecision bool                   `json:"explain_decision"`
}

// DecisionResponse represents a policy decision response
type DecisionResponse struct {
	Decision       types.PolicyDecisionType   `json:"decision"`
	Allowed        bool                       `json:"allowed"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice         []*models.PolicyAdvice     `json:"advice,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	EvaluatedAt    time.Time                  `json:"evaluated_at"`
	RequestID      string                     `json:"request_id"`
	CacheHit       bool                       `json:"cache_hit"`
	PolicyCount    int                        `json:"policy_count"`
	Explanation    *DecisionExplanation       `json:"explanation,omitempty"`
	AuditTrail     *DecisionAuditTrail        `json:"audit_trail,omitempty"`
	ErrorDetails   *string                    `json:"error_details,omitempty"`
}

// BulkDecisionRequest represents a request for multiple decisions
type BulkDecisionRequest struct {
	UserID       uuid.UUID          `json:"user_id" validate:"required"`
	Decisions    []*DecisionRequest `json:"decisions" validate:"required,min=1"`
	RequestID    string             `json:"request_id"`
	FailFast     bool               `json:"fail_fast"`
	UseCache     bool               `json:"use_cache"`
	CacheResults bool               `json:"cache_results"`
}

// BulkDecisionResponse represents a response for multiple decisions
type BulkDecisionResponse struct {
	Responses      []*DecisionResponse `json:"responses"`
	SuccessCount   int                 `json:"success_count"`
	ErrorCount     int                 `json:"error_count"`
	TotalRequests  int                 `json:"total_requests"`
	EvaluationTime time.Duration       `json:"evaluation_time"`
	RequestID      string              `json:"request_id"`
	PartialFailure bool                `json:"partial_failure"`
}

// PermissionCheckRequest represents a permission check request
type PermissionCheckRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Action       string                 `json:"action" validate:"required"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	RequestID    string                 `json:"request_id"`
}

// PermissionCheckResponse represents a permission check response
type PermissionCheckResponse struct {
	Allowed        bool                       `json:"allowed"`
	Decision       types.PolicyDecisionType   `json:"decision"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	Advice         []*models.PolicyAdvice     `json:"advice,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	RequestID      string                     `json:"request_id"`
	ErrorDetails   *string                    `json:"error_details,omitempty"`
}

// DecisionAuditLog represents an audit log entry for a decision
type DecisionAuditLog struct {
	ID             uuid.UUID                  `json:"id"`
	UserID         uuid.UUID                  `json:"user_id"`
	ResourceType   string                     `json:"resource_type"`
	ResourceID     *uuid.UUID                 `json:"resource_id,omitempty"`
	Action         string                     `json:"action"`
	Decision       types.PolicyDecisionType   `json:"decision"`
	Allowed        bool                       `json:"allowed"`
	Obligations    []*models.PolicyObligation `json:"obligations,omitempty"`
	EvaluationTime time.Duration              `json:"evaluation_time"`
	EvaluatedAt    time.Time                  `json:"evaluated_at"`
	PolicyCount    int                        `json:"policy_count"`
	CacheHit       bool                       `json:"cache_hit"`
	RequestID      string                     `json:"request_id"`
	TenantID       uuid.UUID                  `json:"tenant_id"`
	Context        map[string]interface{}     `json:"context,omitempty"`
}

// DecisionAuditTrail provides detailed audit information
type DecisionAuditTrail struct {
	EvaluationSteps []*EvaluationStep `json:"evaluation_steps"`
	PoliciesApplied []string          `json:"policies_applied"`
	AttributesSeen  map[string]string `json:"attributes_seen"`
	CacheEvents     []*CacheEvent     `json:"cache_events,omitempty"`
	Timing          *EvaluationTiming `json:"timing"`
}

// EvaluationStep represents a step in the evaluation process
type EvaluationStep struct {
	StepType    string                 `json:"step_type"`
	Description string                 `json:"description"`
	Result      interface{}            `json:"result"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CacheEvent represents a cache-related event during evaluation
type CacheEvent struct {
	EventType string    `json:"event_type"`
	CacheKey  string    `json:"cache_key"`
	Hit       bool      `json:"hit"`
	Timestamp time.Time `json:"timestamp"`
}

// EvaluationTiming provides detailed timing information
type EvaluationTiming struct {
	AttributeCollectionTime time.Duration `json:"attribute_collection_time"`
	AttributeValidationTime time.Duration `json:"attribute_validation_time"`
	PolicyRetrievalTime     time.Duration `json:"policy_retrieval_time"`
	PolicyEvaluationTime    time.Duration `json:"policy_evaluation_time"`
	CacheOperationTime      time.Duration `json:"cache_operation_time"`
	TotalTime               time.Duration `json:"total_time"`
}

// PolicyDiscoveryRequest represents a request to discover applicable policies
type PolicyDiscoveryRequest struct {
	UserID       uuid.UUID              `json:"user_id"`
	ResourceType string                 `json:"resource_type"`
	Action       string                 `json:"action"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// PolicySummary provides a summary of a policy
type PolicySummary struct {
	ID          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Effect      types.PolicyDecisionType `json:"effect"`
	Priority    int                      `json:"priority"`
	Applicable  bool                     `json:"applicable"`
}

// DecisionExplanationRequest represents a request for decision explanation
type DecisionExplanationRequest struct {
	UserID       uuid.UUID              `json:"user_id" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Action       string                 `json:"action" validate:"required"`
	Context      map[string]interface{} `json:"context,omitempty"`
	DetailLevel  string                 `json:"detail_level"` // "basic", "detailed", "verbose"
}

// DecisionExplanation provides an explanation of how a decision was made
type DecisionExplanation struct {
	FinalDecision      types.PolicyDecisionType   `json:"final_decision"`
	ReasoningSummary   string                     `json:"reasoning_summary"`
	PolicyEvaluations  []*PolicyEvaluationSummary `json:"policy_evaluations"`
	AttributesUsed     map[string]interface{}     `json:"attributes_used"`
	ConflictResolution *ConflictResolutionSummary `json:"conflict_resolution,omitempty"`
	CombiningAlgorithm string                     `json:"combining_algorithm"`
	Recommendations    []string                   `json:"recommendations,omitempty"`
}

// PolicyEvaluationSummary provides a summary of a policy evaluation
type PolicyEvaluationSummary struct {
	PolicyName   string                   `json:"policy_name"`
	Decision     types.PolicyDecisionType `json:"decision"`
	Applicable   bool                     `json:"applicable"`
	MatchedRules []string                 `json:"matched_rules"`
	FailedRules  []string                 `json:"failed_rules"`
	Reason       string                   `json:"reason"`
}

// ConflictResolutionSummary provides details about how conflicts were resolved
type ConflictResolutionSummary struct {
	ConflictDetected    bool     `json:"conflict_detected"`
	ConflictingPolicies []string `json:"conflicting_policies"`
	ResolutionMethod    string   `json:"resolution_method"`
	WinningPolicy       string   `json:"winning_policy"`
	Explanation         string   `json:"explanation"`
}

// policyDecisionService implements PolicyDecisionService
type policyDecisionService struct {
	policyEvalService PolicyEvaluationService
	auditRepo         repository.PolicyEvaluationRepository
	cache             cache.Service
	tracing           tracing.TracingService
	metrics           metrics.MetricsProvider
	logger            logger.Logger
}

// NewPolicyDecisionService creates a new policy decision service
func NewPolicyDecisionService(
	policyEvalService PolicyEvaluationService,
	auditRepo repository.PolicyEvaluationRepository,
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) PolicyDecisionService {
	return &policyDecisionService{
		policyEvalService: policyEvalService,
		auditRepo:         auditRepo,
		cache:             cache,
		tracing:           tracing,
		metrics:           metrics,
		logger:            logger,
	}
}

// MakeDecision makes a comprehensive policy decision
func (s *policyDecisionService) MakeDecision(ctx context.Context, req *DecisionRequest) (*DecisionResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.MakeDecision",
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.String("resource.type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordDecisionMetrics(ctx, "make_decision", "success", time.Since(startTime))
	}()

	response := &DecisionResponse{
		RequestID:   req.RequestID,
		EvaluatedAt: time.Now(),
	}

	// Create audit trail if detailed logging is requested
	var auditTrail *DecisionAuditTrail
	if req.IncludeDetails {
		auditTrail = &DecisionAuditTrail{
			EvaluationSteps: make([]*EvaluationStep, 0),
			AttributesSeen:  make(map[string]string),
			CacheEvents:     make([]*CacheEvent, 0),
			Timing:          &EvaluationTiming{},
		}
		response.AuditTrail = auditTrail
	}

	// Convert to access decision request
	accessReq := &AccessDecisionRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
		SessionData:  req.Context,
		RequestID:    req.RequestID,
		UseCache:     req.UseCache,
		CacheResults: req.CacheResults,
	}

	// Make access decision
	accessResp, err := s.policyEvalService.MakeAccessDecision(ctx, accessReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to make access decision")
		errorMsg := err.Error()
		response.ErrorDetails = &errorMsg
		response.Decision = types.PolicyDecisionDeny
		response.Allowed = false
		response.EvaluationTime = time.Since(startTime)
		return response, nil
	}

	// Map access response to decision response
	response.Decision = accessResp.Decision
	response.Allowed = accessResp.Decision == types.PolicyDecisionAllow
	response.Obligations = accessResp.Obligations
	response.CacheHit = accessResp.CacheHit
	response.PolicyCount = accessResp.PolicyCount
	response.EvaluationTime = time.Since(startTime)

	// Include advice if requested
	if req.IncludeAdvice {
		response.Advice = accessResp.Advice
	}

	// Generate explanation if requested
	if req.ExplainDecision {
		explanationReq := &DecisionExplanationRequest{
			UserID:       req.UserID,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
			Action:       req.Action,
			Context:      req.Context,
			DetailLevel:  "detailed",
		}
		explanation, err := s.ExplainDecision(ctx, explanationReq)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to generate decision explanation",
				logger.Fields{"error": err.Error()})
		} else {
			response.Explanation = explanation
		}
	}

	// Log decision for audit
	auditLog := &DecisionAuditLog{
		ID:             uuid.New(),
		UserID:         req.UserID,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		Action:         req.Action,
		Decision:       response.Decision,
		Allowed:        response.Allowed,
		Obligations:    response.Obligations,
		EvaluationTime: response.EvaluationTime,
		EvaluatedAt:    response.EvaluatedAt,
		PolicyCount:    response.PolicyCount,
		CacheHit:       response.CacheHit,
		RequestID:      req.RequestID,
		Context:        req.Context,
	}

	if err := s.LogDecision(ctx, auditLog); err != nil {
		s.logger.WarnContext(ctx, "Failed to log decision audit",
			logger.Fields{"error": err.Error()})
	}

	s.logger.InfoContext(ctx, "Policy decision completed",
		logger.Fields{
			"user_id":         req.UserID,
			"resource_type":   req.ResourceType,
			"action":          req.Action,
			"decision":        string(response.Decision),
			"allowed":         response.Allowed,
			"policy_count":    response.PolicyCount,
			"cache_hit":       response.CacheHit,
			"evaluation_time": response.EvaluationTime.Milliseconds(),
		})

	return response, nil
}

// MakeBulkDecision makes multiple policy decisions efficiently
func (s *policyDecisionService) MakeBulkDecision(ctx context.Context, req *BulkDecisionRequest) (*BulkDecisionResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.MakeBulkDecision",
		tracing.WithAttributes(
			attribute.String("user.id", req.UserID.String()),
			attribute.Int("request.count", len(req.Decisions)),
		))
	defer span.End()

	startTime := time.Now()
	defer func() {
		s.recordDecisionMetrics(ctx, "make_bulk_decision", "success", time.Since(startTime))
	}()

	response := &BulkDecisionResponse{
		Responses:     make([]*DecisionResponse, 0, len(req.Decisions)),
		TotalRequests: len(req.Decisions),
		RequestID:     req.RequestID,
	}

	// Process each decision request
	for _, decisionReq := range req.Decisions {
		// Ensure user ID matches the bulk request
		decisionReq.UserID = req.UserID
		decisionReq.UseCache = req.UseCache
		decisionReq.CacheResults = req.CacheResults

		decisionResp, err := s.MakeDecision(ctx, decisionReq)
		if err != nil {
			response.ErrorCount++

			if req.FailFast {
				span.RecordError(err)
				span.SetStatus(codes.Error, "Bulk decision failed fast")
				return response, err
			}

			// Create error response
			errorMsg := err.Error()
			decisionResp = &DecisionResponse{
				Decision:     types.PolicyDecisionDeny,
				Allowed:      false,
				RequestID:    decisionReq.RequestID,
				ErrorDetails: &errorMsg,
				EvaluatedAt:  time.Now(),
			}
		} else {
			response.SuccessCount++
		}

		response.Responses = append(response.Responses, decisionResp)
	}

	response.EvaluationTime = time.Since(startTime)
	response.PartialFailure = response.ErrorCount > 0 && response.SuccessCount > 0

	s.logger.InfoContext(ctx, "Bulk policy decision completed",
		logger.Fields{
			"user_id":         req.UserID,
			"total_requests":  response.TotalRequests,
			"success_count":   response.SuccessCount,
			"error_count":     response.ErrorCount,
			"partial_failure": response.PartialFailure,
			"evaluation_time": response.EvaluationTime.Milliseconds(),
		})

	return response, nil
}

// IsAuthorized checks if a user is authorized for an action
func (s *policyDecisionService) IsAuthorized(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, action string, context map[string]interface{}) (bool, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.IsAuthorized")
	defer span.End()

	req := &DecisionRequest{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Context:      context,
		UseCache:     true,
		CacheResults: true,
	}

	response, err := s.MakeDecision(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// CheckPermission checks a specific permission
func (s *policyDecisionService) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.CheckPermission")
	defer span.End()

	decisionReq := &DecisionRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
		Context:      req.Context,
		RequestID:    req.RequestID,
		UseCache:     true,
		CacheResults: true,
	}

	decisionResp, err := s.MakeDecision(ctx, decisionReq)
	if err != nil {
		errorMsg := err.Error()
		return &PermissionCheckResponse{
			Allowed:      false,
			Decision:     types.PolicyDecisionDeny,
			RequestID:    req.RequestID,
			ErrorDetails: &errorMsg,
		}, nil
	}

	return &PermissionCheckResponse{
		Allowed:        decisionResp.Allowed,
		Decision:       decisionResp.Decision,
		Obligations:    decisionResp.Obligations,
		Advice:         decisionResp.Advice,
		EvaluationTime: decisionResp.EvaluationTime,
		RequestID:      req.RequestID,
	}, nil
}

// CheckMultiplePermissions checks multiple permissions efficiently
func (s *policyDecisionService) CheckMultiplePermissions(ctx context.Context, requests []*PermissionCheckRequest) ([]*PermissionCheckResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.CheckMultiplePermissions",
		tracing.WithAttributes(attribute.Int("request.count", len(requests))))
	defer span.End()

	responses := make([]*PermissionCheckResponse, 0, len(requests))

	for _, req := range requests {
		resp, err := s.CheckPermission(ctx, req)
		if err != nil {
			// Continue processing other requests
			errorMsg := err.Error()
			resp = &PermissionCheckResponse{
				Allowed:      false,
				Decision:     types.PolicyDecisionDeny,
				RequestID:    req.RequestID,
				ErrorDetails: &errorMsg,
			}
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// LogDecision logs a decision for audit purposes
func (s *policyDecisionService) LogDecision(ctx context.Context, decision *DecisionAuditLog) error {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.LogDecision")
	defer span.End()

	// Store in audit repository
	// Note: Audit logging would need to be implemented via a separate audit repository
	// For now, we'll log the audit entry instead of persisting it
	s.logger.InfoContext(ctx, "Policy decision audit entry created",
		logger.Fields{
			"decision_id":   decision.ID.String(),
			"user_id":       decision.UserID.String(),
			"resource_type": decision.ResourceType,
			"action":        decision.Action,
			"decision":      string(decision.Decision),
		})

	// Record audit metrics
	s.recordAuditMetrics(ctx, decision)

	return nil
}

// GetDecisionHistory retrieves decision history for a user
func (s *policyDecisionService) GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*DecisionAuditLog, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.GetDecisionHistory",
		tracing.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

	// Use GetUserEvaluationHistory with proper request structure
	req := &repository.GetUserEvaluationHistoryRequest{
		UserID: userID,
		Limit:  limit,
	}
	evaluations, err := s.auditRepo.GetUserEvaluationHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	auditLogs := make([]*DecisionAuditLog, 0, len(evaluations))
	for _, eval := range evaluations {
		auditLog := &DecisionAuditLog{
			ID:           eval.ID,
			UserID:       eval.UserID,
			ResourceType: eval.ResourceType,
			ResourceID:   eval.ResourceID,
			Action:       eval.Action,
			Decision:     eval.Decision,
			Allowed:      eval.Decision == types.PolicyDecisionAllow,
			EvaluatedAt:  eval.EvaluatedAt,
			Context:      make(map[string]interface{}), // No context field in PolicyEvaluation
			CacheHit:     false,                        // No cache hit field in PolicyEvaluation
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, nil
}

// GetApplicablePolicies discovers applicable policies for a request
func (s *policyDecisionService) GetApplicablePolicies(ctx context.Context, req *PolicyDiscoveryRequest) ([]*PolicySummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.GetApplicablePolicies")
	defer span.End()

	// TODO: Implement policy discovery logic
	// This would involve querying policies that match the request criteria

	summaries := []*PolicySummary{
		{
			ID:          uuid.New(),
			Name:        "Sample Policy",
			Description: "A sample policy for demonstration",
			Effect:      types.PolicyDecisionAllow,
			Priority:    1,
			Applicable:  true,
		},
	}

	return summaries, nil
}

// ExplainDecision provides an explanation of how a decision was made
func (s *policyDecisionService) ExplainDecision(ctx context.Context, req *DecisionExplanationRequest) (*DecisionExplanation, error) {
	ctx, span := s.tracing.StartSpan(ctx, "policyDecisionService.ExplainDecision")
	defer span.End()

	// TODO: Implement decision explanation logic
	// This would involve re-evaluating the decision with detailed tracking

	explanation := &DecisionExplanation{
		FinalDecision:      types.PolicyDecisionAllow,
		ReasoningSummary:   "User has the required permissions for this action",
		CombiningAlgorithm: "deny-overrides",
		PolicyEvaluations: []*PolicyEvaluationSummary{
			{
				PolicyName:   "User Access Policy",
				Decision:     types.PolicyDecisionAllow,
				Applicable:   true,
				MatchedRules: []string{"rule-1", "rule-2"},
				Reason:       "User matches the required attributes",
			},
		},
		AttributesUsed: map[string]interface{}{
			"user.role":     "manager",
			"resource.type": req.ResourceType,
			"action.name":   req.Action,
		},
		Recommendations: []string{
			"Consider adding additional constraints for sensitive operations",
		},
	}

	return explanation, nil
}

// Helper methods

func (s *policyDecisionService) recordDecisionMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	// Decision operation counter
	counter := s.metrics.Counter(
		"abac_policy_decision_operations_total",
		"Total number of policy decision operations",
		"operation", "status",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	// Decision operation duration histogram
	histogram := s.metrics.Histogram(
		"abac_policy_decision_operation_duration_seconds",
		"Duration of policy decision operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation", "status",
	)

	histogram.Observe(duration.Seconds(), metrics.Fields{
		"operation": operation,
		"status":    status,
	})
}

func (s *policyDecisionService) recordAuditMetrics(ctx context.Context, decision *DecisionAuditLog) {
	// Decision counter by result
	counter := s.metrics.Counter(
		"abac_policy_decisions_total",
		"Total number of policy decisions by result",
		"decision", "resource_type", "action",
	)

	counter.Inc(metrics.Fields{
		"decision":      string(decision.Decision),
		"resource_type": decision.ResourceType,
		"action":        decision.Action,
	})

	// Cache hit ratio
	cacheCounter := s.metrics.Counter(
		"abac_policy_decision_cache_total",
		"Total number of policy decision cache operations",
		"cache_result",
	)

	cacheResult := "miss"
	if decision.CacheHit {
		cacheResult = "hit"
	}

	cacheCounter.Inc(metrics.Fields{
		"cache_result": cacheResult,
	})
}
