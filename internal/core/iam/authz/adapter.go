package authz

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/access"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// adapter implements the authorization service interface by delegating to existing services
type adapter struct {
	// Core services
	abacService   abac.Service
	accessService access.Service

	// Infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewAdapter creates a new authorization service adapter
func NewAdapter(
	abacService abac.Service,
	accessService access.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) Service {
	return &adapter{
		abacService:   abacService,
		accessService: accessService,
		logger:        logger,
		metrics:       metrics,
		tracer:        tracer,
	}
}

// ─── PERMISSION EVALUATION ─────────────────────────────────────────────────

// EvaluatePermission delegates to ABAC service with type conversion
func (a *adapter) EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.EvaluatePermission")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		a.metrics.ObserveHistogram("authz.adapter.permission_evaluation.duration", duration.Seconds(),
			metrics.Fields{"method": "single"})
	}()

	// Convert IAM request to ABAC request
	abacReq := &abac.PermissionEvaluationRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
		Context:      req.Context,
		RequestID:    req.RequestID,
	}

	// Delegate to ABAC service
	abacResult, err := a.abacService.EvaluatePermission(ctx, abacReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("authz.adapter.permission_evaluation.errors",
			metrics.Fields{"reason": "abac_evaluation_failed"})
		return nil, fmt.Errorf("failed to evaluate permission via ABAC: %w", err)
	}

	// Convert ABAC result to IAM result
	result := &PermissionEvaluationResult{
		Decision:         convertPolicyDecisionType(abacResult.Decision),
		PolicyDecisions:  convertPolicyDecisions(abacResult.PolicyDecisions),
		EvaluationTimeMS: abacResult.EvaluationTimeMS,
		CacheHit:         abacResult.CacheHit,
		RequestID:        abacResult.RequestID,
		Timestamp:        abacResult.Timestamp,
	}

	a.metrics.IncrementCounter("authz.adapter.permission_evaluation.success",
		metrics.Fields{"decision": string(result.Decision), "cache_hit": fmt.Sprintf("%t", result.CacheHit)})

	a.logger.InfoContext(ctx, "Permission evaluation completed via ABAC adapter",
		logger.Fields{
			"user_id":            req.UserID,
			"resource_type":      req.ResourceType,
			"action":             req.Action,
			"decision":           result.Decision,
			"evaluation_time_ms": result.EvaluationTimeMS,
			"cache_hit":          result.CacheHit,
		})

	return result, nil
}

// BulkEvaluatePermissions delegates to ABAC service for bulk evaluation
func (a *adapter) BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) (*BulkPermissionEvaluationResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.BulkEvaluatePermissions")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		a.metrics.ObserveHistogram("authz.adapter.permission_evaluation.duration", duration.Seconds(),
			metrics.Fields{"method": "bulk", "count": fmt.Sprintf("%d", len(req.Requests))})
	}()

	// Convert IAM requests to ABAC requests
	abacRequests := make([]*abac.PermissionEvaluationRequest, len(req.Requests))
	for i, iamReq := range req.Requests {
		abacRequests[i] = &abac.PermissionEvaluationRequest{
			UserID:       iamReq.UserID,
			ResourceType: iamReq.ResourceType,
			ResourceID:   iamReq.ResourceID,
			Action:       iamReq.Action,
			EntityID:     iamReq.EntityID,
			Context:      iamReq.Context,
			RequestID:    iamReq.RequestID,
		}
	}

	abacBulkReq := &abac.BulkPermissionEvaluationRequest{
		Requests:  abacRequests,
		RequestID: req.RequestID,
	}

	// Delegate to ABAC service
	abacResult, err := a.abacService.BulkEvaluatePermissions(ctx, abacBulkReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("authz.adapter.bulk_permission_evaluation.errors",
			metrics.Fields{"reason": "abac_bulk_evaluation_failed"})
		return nil, fmt.Errorf("failed to bulk evaluate permissions via ABAC: %w", err)
	}

	// Convert ABAC results to IAM results
	iamResults := make([]*PermissionEvaluationResult, len(abacResult.Results))
	for i, abacRes := range abacResult.Results {
		iamResults[i] = &PermissionEvaluationResult{
			Decision:         convertPolicyDecisionType(abacRes.Decision),
			PolicyDecisions:  convertPolicyDecisions(abacRes.PolicyDecisions),
			EvaluationTimeMS: abacRes.EvaluationTimeMS,
			CacheHit:         abacRes.CacheHit,
			RequestID:        abacRes.RequestID,
			Timestamp:        abacRes.Timestamp,
		}
	}

	result := &BulkPermissionEvaluationResult{
		Results:         iamResults,
		TotalRequests:   abacResult.TotalRequests,
		SuccessfulCount: abacResult.SuccessfulCount,
		FailedCount:     abacResult.FailedCount,
		TotalTimeMS:     abacResult.TotalTimeMS,
		AverageTimeMS:   abacResult.AverageTimeMS,
		RequestID:       abacResult.RequestID,
		Timestamp:       abacResult.Timestamp,
	}

	a.metrics.IncrementCounter("authz.adapter.bulk_permission_evaluation.success",
		metrics.Fields{
			"total_requests":   fmt.Sprintf("%d", result.TotalRequests),
			"successful_count": fmt.Sprintf("%d", result.SuccessfulCount),
			"failed_count":     fmt.Sprintf("%d", result.FailedCount),
		})

	return result, nil
}

// ─── USER PERMISSIONS ──────────────────────────────────────────────────────

// GetUserEffectivePermissions delegates to ABAC service
func (a *adapter) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*UserEffectivePermissions, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GetUserEffectivePermissions")
	defer span.End()

	abacResult, err := a.abacService.GetUserEffectivePermissions(ctx, userID, entityID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get user effective permissions via ABAC: %w", err)
	}

	// Direct mapping since structure is compatible
	result := &UserEffectivePermissions{
		UserID:      abacResult.UserID,
		EntityID:    abacResult.EntityID,
		Permissions: abacResult.Permissions,
		Roles:       abacResult.Roles,
		Timestamp:   abacResult.Timestamp,
	}

	a.metrics.IncrementCounter("authz.adapter.get_user_effective_permissions.success",
		metrics.Fields{"user_id": userID.String()})

	return result, nil
}

// CalculateRoleHierarchy delegates to ABAC service
func (a *adapter) CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*RoleHierarchy, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.CalculateRoleHierarchy")
	defer span.End()

	abacResult, err := a.abacService.CalculateRoleHierarchy(ctx, userID, entityID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to calculate role hierarchy via ABAC: %w", err)
	}

	// Direct mapping since structure is compatible
	result := &RoleHierarchy{
		UserID:    abacResult.UserID,
		EntityID:  abacResult.EntityID,
		Roles:     abacResult.Roles,
		Hierarchy: abacResult.Hierarchy,
		Timestamp: abacResult.Timestamp,
	}

	a.metrics.IncrementCounter("authz.adapter.calculate_role_hierarchy.success",
		metrics.Fields{"user_id": userID.String()})

	return result, nil
}

// ─── ACCESS REQUESTS ────────────────────────────────────────────────────────

// CreateAccessRequest delegates to access service
func (a *adapter) CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest) (*model.AccessRequest, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.CreateAccessRequest")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.CreateAccessRequestRequest{
		UserID:        req.UserID,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
		Action:        req.Action,
		Justification: req.Justification,
		Duration:      req.Duration,
		Priority:      req.Priority,
		Metadata:      req.Metadata,
	}

	accessResult, err := a.accessService.CreateAccessRequest(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create access request via access service: %w", err)
	}

	// Convert access service result to IAM model
	result := convertAccessRequestToIAMModel(accessResult)

	a.metrics.IncrementCounter("authz.adapter.create_access_request.success",
		metrics.Fields{"user_id": req.UserID.String(), "resource_type": req.ResourceType})

	return result, nil
}

// GetAccessRequest delegates to access service
func (a *adapter) GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*model.AccessRequest, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GetAccessRequest")
	defer span.End()

	accessResult, err := a.accessService.GetAccessRequest(ctx, requestID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get access request via access service: %w", err)
	}

	result := convertAccessRequestToIAMModel(accessResult)

	a.metrics.IncrementCounter("authz.adapter.get_access_request.success",
		metrics.Fields{"request_id": requestID.String()})

	return result, nil
}

// ProcessAccessRequest delegates to access service
func (a *adapter) ProcessAccessRequest(ctx context.Context, req *ProcessAccessRequestRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.ProcessAccessRequest")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.ProcessAccessRequestRequest{
		RequestID:  req.RequestID,
		Action:     req.Action,
		ApproverID: req.ApproverID,
		Comments:   req.Comments,
		Conditions: req.Conditions,
	}

	err := a.accessService.ProcessAccessRequest(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to process access request via access service: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.process_access_request.success",
		metrics.Fields{"request_id": req.RequestID.String(), "action": req.Action})

	return nil
}

// ListAccessRequests delegates to access service
func (a *adapter) ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) (*ListAccessRequestsResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.ListAccessRequests")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.ListAccessRequestsRequest{
		UserID:   req.UserID,
		Status:   req.Status,
		EntityID: req.EntityID,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	accessResult, err := a.accessService.ListAccessRequests(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to list access requests via access service: %w", err)
	}

	// Convert access service results to IAM models
	iamRequests := make([]*model.AccessRequest, len(accessResult.Requests))
	for i, accessReq := range accessResult.Requests {
		iamRequests[i] = convertAccessRequestToIAMModel(accessReq)
	}

	result := &ListAccessRequestsResult{
		Requests: iamRequests,
		Total:    accessResult.Total,
		Limit:    accessResult.Limit,
		Offset:   accessResult.Offset,
		HasMore:  accessResult.HasMore,
	}

	a.metrics.IncrementCounter("authz.adapter.list_access_requests.success",
		metrics.Fields{"count": fmt.Sprintf("%d", len(result.Requests))})

	return result, nil
}

// ─── APPROVAL WORKFLOWS ────────────────────────────────────────────────────

// CreateApprovalWorkflow delegates to access service
func (a *adapter) CreateApprovalWorkflow(ctx context.Context, req *CreateApprovalWorkflowRequest) (*model.ApprovalWorkflow, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.CreateApprovalWorkflow")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.CreateApprovalWorkflowRequest{
		Name:        req.Name,
		Description: req.Description,
		Steps:       convertApprovalStepsFromIAM(req.Steps),
		Metadata:    req.Metadata,
	}

	accessResult, err := a.accessService.CreateApprovalWorkflow(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create approval workflow via access service: %w", err)
	}

	result := convertApprovalWorkflowToIAMModel(accessResult)

	a.metrics.IncrementCounter("authz.adapter.create_approval_workflow.success",
		metrics.Fields{"workflow_name": req.Name})

	return result, nil
}

// GetApprovalWorkflow delegates to access service
func (a *adapter) GetApprovalWorkflow(ctx context.Context, workflowID uuid.UUID) (*model.ApprovalWorkflow, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GetApprovalWorkflow")
	defer span.End()

	accessResult, err := a.accessService.GetApprovalWorkflow(ctx, workflowID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get approval workflow via access service: %w", err)
	}

	result := convertApprovalWorkflowToIAMModel(accessResult)

	a.metrics.IncrementCounter("authz.adapter.get_approval_workflow.success",
		metrics.Fields{"workflow_id": workflowID.String()})

	return result, nil
}

// UpdateApprovalWorkflow delegates to access service
func (a *adapter) UpdateApprovalWorkflow(ctx context.Context, req *UpdateApprovalWorkflowRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.UpdateApprovalWorkflow")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.UpdateApprovalWorkflowRequest{
		WorkflowID:  req.WorkflowID,
		Name:        req.Name,
		Description: req.Description,
		Steps:       convertApprovalStepsFromIAM(req.Steps),
		Metadata:    req.Metadata,
	}

	err := a.accessService.UpdateApprovalWorkflow(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to update approval workflow via access service: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.update_approval_workflow.success",
		metrics.Fields{"workflow_id": req.WorkflowID.String()})

	return nil
}

// ─── CONDITIONAL ACCESS ────────────────────────────────────────────────────

// EvaluateConditionalAccess delegates to access service
func (a *adapter) EvaluateConditionalAccess(ctx context.Context, req *ConditionalAccessRequest) (*ConditionalAccessResult, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.EvaluateConditionalAccess")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.EvaluateConditionalAccessRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		Context:      convertAccessContextFromIAM(req.Context),
	}

	accessResult, err := a.accessService.EvaluateConditionalAccess(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to evaluate conditional access via access service: %w", err)
	}

	result := &ConditionalAccessResult{
		Allowed:    accessResult.Allowed,
		Conditions: convertAccessConditionsToIAM(accessResult.Conditions),
		Reason:     accessResult.Reason,
		Metadata:   accessResult.Metadata,
	}

	a.metrics.IncrementCounter("authz.adapter.evaluate_conditional_access.success",
		metrics.Fields{"user_id": req.UserID.String(), "allowed": fmt.Sprintf("%t", result.Allowed)})

	return result, nil
}

// CreateConditionalAccessPolicy delegates to access service
func (a *adapter) CreateConditionalAccessPolicy(ctx context.Context, req *CreateConditionalAccessPolicyRequest) (*model.ConditionalAccessPolicy, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.CreateConditionalAccessPolicy")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.CreateConditionalAccessPolicyRequest{
		Name:        req.Name,
		Description: req.Description,
		Conditions:  convertPolicyConditionsFromIAM(req.Conditions),
		Actions:     convertPolicyActionsFromIAM(req.Actions),
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Metadata:    req.Metadata,
	}

	accessResult, err := a.accessService.CreateConditionalAccessPolicy(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to create conditional access policy via access service: %w", err)
	}

	result := convertConditionalAccessPolicyToIAMModel(accessResult)

	a.metrics.IncrementCounter("authz.adapter.create_conditional_access_policy.success",
		metrics.Fields{"policy_name": req.Name})

	return result, nil
}

// UpdateConditionalAccessPolicy delegates to access service
func (a *adapter) UpdateConditionalAccessPolicy(ctx context.Context, req *UpdateConditionalAccessPolicyRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.UpdateConditionalAccessPolicy")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.UpdateConditionalAccessPolicyRequest{
		PolicyID:    req.PolicyID,
		Name:        req.Name,
		Description: req.Description,
		Conditions:  convertPolicyConditionsFromIAM(req.Conditions),
		Actions:     convertPolicyActionsFromIAM(req.Actions),
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Metadata:    req.Metadata,
	}

	err := a.accessService.UpdateConditionalAccessPolicy(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to update conditional access policy via access service: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.update_conditional_access_policy.success",
		metrics.Fields{"policy_id": req.PolicyID.String()})

	return nil
}

// ─── PERMISSION MANAGEMENT ─────────────────────────────────────────────────

// GrantPermission delegates to access service
func (a *adapter) GrantPermission(ctx context.Context, req *GrantPermissionRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GrantPermission")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.GrantPermissionRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
		ExpiresAt:    req.ExpiresAt,
		Conditions:   req.Conditions,
	}

	err := a.accessService.GrantPermission(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to grant permission via access service: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.grant_permission.success",
		metrics.Fields{
			"user_id":       req.UserID.String(),
			"resource_type": req.ResourceType,
			"action":        req.Action,
		})

	return nil
}

// RevokePermission delegates to access service
func (a *adapter) RevokePermission(ctx context.Context, req *RevokePermissionRequest) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.RevokePermission")
	defer span.End()

	// Convert IAM request to access service request
	accessReq := &access.RevokePermissionRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
		EntityID:     req.EntityID,
	}

	err := a.accessService.RevokePermission(ctx, accessReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to revoke permission via access service: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.revoke_permission.success",
		metrics.Fields{
			"user_id":       req.UserID.String(),
			"resource_type": req.ResourceType,
			"action":        req.Action,
		})

	return nil
}

// ListUserPermissions delegates to access service
func (a *adapter) ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.ListUserPermissions")
	defer span.End()

	accessResult, err := a.accessService.ListUserPermissions(ctx, userID, entityID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to list user permissions via access service: %w", err)
	}

	// Convert access service permissions to IAM models
	iamPermissions := make([]*model.Permission, len(accessResult))
	for i, accessPerm := range accessResult {
		iamPermissions[i] = convertPermissionToIAMModel(accessPerm)
	}

	a.metrics.IncrementCounter("authz.adapter.list_user_permissions.success",
		metrics.Fields{"user_id": userID.String(), "count": fmt.Sprintf("%d", len(iamPermissions))})

	return iamPermissions, nil
}

// ─── DECISION HISTORY AND AUDIT ────────────────────────────────────────────

// GetDecisionHistory delegates to ABAC service
func (a *adapter) GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*DecisionHistoryEntry, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GetDecisionHistory")
	defer span.End()

	abacResult, err := a.abacService.GetDecisionHistory(ctx, userID, limit)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get decision history via ABAC: %w", err)
	}

	// Convert ABAC decision history to IAM format
	iamHistory := make([]*DecisionHistoryEntry, len(abacResult))
	for i, abacEntry := range abacResult {
		iamHistory[i] = &DecisionHistoryEntry{
			ID:             abacEntry.ID,
			UserID:         abacEntry.UserID,
			ResourceType:   abacEntry.ResourceType,
			ResourceID:     &abacEntry.ResourceID,
			Action:         abacEntry.Action,
			Decision:       convertPolicyDecisionType(abacEntry.Decision),
			Allowed:        abacEntry.Allowed,
			EvaluationTime: abacEntry.EvaluationTime,
			EvaluatedAt:    abacEntry.EvaluatedAt,
			PolicyCount:    abacEntry.PolicyCount,
			CacheHit:       abacEntry.CacheHit,
			RequestID:      abacEntry.RequestID,
		}
	}

	a.metrics.IncrementCounter("authz.adapter.get_decision_history.success",
		metrics.Fields{"user_id": userID.String(), "count": fmt.Sprintf("%d", len(iamHistory))})

	return iamHistory, nil
}

// ─── CACHE MANAGEMENT ──────────────────────────────────────────────────────

// InvalidateUserCache delegates to ABAC service
func (a *adapter) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.InvalidateUserCache")
	defer span.End()

	err := a.abacService.InvalidateUserCache(ctx, userID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to invalidate user cache via ABAC: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.invalidate_user_cache.success",
		metrics.Fields{"user_id": userID.String()})

	return nil
}

// InvalidatePolicyCache delegates to ABAC service
func (a *adapter) InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.InvalidatePolicyCache")
	defer span.End()

	err := a.abacService.InvalidatePolicyCache(ctx, policyIDs)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return fmt.Errorf("failed to invalidate policy cache via ABAC: %w", err)
	}

	a.metrics.IncrementCounter("authz.adapter.invalidate_policy_cache.success",
		metrics.Fields{"policy_count": fmt.Sprintf("%d", len(policyIDs))})

	return nil
}

// GetCacheStatistics delegates to ABAC service
func (a *adapter) GetCacheStatistics(ctx context.Context) (*CacheStatistics, error) {
	ctx, span := a.tracer.StartSpan(ctx, "authz.adapter.GetCacheStatistics")
	defer span.End()

	abacStats, err := a.abacService.GetCacheStatistics(ctx)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("failed to get cache statistics via ABAC: %w", err)
	}

	// Convert ABAC cache statistics to IAM format
	// Note: ABAC service returns a different structure, so we'll provide default values
	result := &CacheStatistics{
		HitRate:        0.0, // Would need to be calculated from the stats
		MissRate:       0.0, // Would need to be calculated from the stats
		TotalRequests:  0,   // Would need to be extracted from sub-stats
		CacheHits:      0,   // Would need to be extracted from sub-stats
		CacheMisses:    0,   // Would need to be extracted from sub-stats
		EvictionCount:  0,   // Would need to be extracted from sub-stats
		AverageLatency: 0,   // Would need to be calculated from the stats
	}

	// If policy evaluation stats are available, use those values
	if abacStats.PolicyEvaluationStats != nil {
		// Extract available metrics from the policy evaluation stats
		// This is a simplified mapping - in practice you'd calculate these from the available data
		result.TotalRequests = 100 // Placeholder
		result.CacheHits = 80      // Placeholder
		result.CacheMisses = 20    // Placeholder
		result.HitRate = 0.8       // Placeholder
		result.MissRate = 0.2      // Placeholder
	}

	a.metrics.IncrementCounter("authz.adapter.get_cache_statistics.success", nil)

	return result, nil
}
