package abac

//go:generate go run go.uber.org/mock/mockgen -source=temporal_service.go -destination=mock.go -package=abac


import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	workflowservice "go.temporal.io/api/workflowservice/v1"

	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/worker"
	"github.com/niiniyare/erp/internal/core/abac/workflows"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TemporalService extends the main ABAC service with Temporal workflow capabilities
type TemporalService interface {
	Service

	// Workflow-based operations
	EvaluatePermissionAsync(ctx context.Context, req *PermissionEvaluationRequest) (*AsyncEvaluationResult, error)
	BulkEvaluatePermissionsAsync(ctx context.Context, req *BulkPermissionEvaluationRequest) (*AsyncBulkEvaluationResult, error)

	// Cache management workflows
	StartCacheCleanupWorkflow(ctx context.Context) (*WorkflowInfo, error)
	StartCacheWarmupWorkflow(ctx context.Context, configs []*CacheWarmupConfig) (*WorkflowInfo, error)
	InvalidateCacheAsync(ctx context.Context, req *TemporalCacheInvalidationRequest) (*WorkflowInfo, error)

	// Workflow monitoring
	GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
	CancelWorkflow(ctx context.Context, workflowID string) error
}

// temporalService implements TemporalService
type temporalService struct {
	Service        // Embed the base service
	workflowClient *worker.WorkflowClient
	logger         logger.Logger
	metrics        metrics.MetricsProvider
	tracer         tracing.TracingService
}

// NewTemporalService creates a new ABAC service with Temporal capabilities
func NewTemporalService(
	baseService Service,
	workflowClient *worker.WorkflowClient,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) TemporalService {
	return &temporalService{
		Service:        baseService,
		workflowClient: workflowClient,
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
	}
}

// AsyncEvaluationResult represents the result of an async policy evaluation
type AsyncEvaluationResult struct {
	WorkflowID string                      `json:"workflow_id"`
	RunID      string                      `json:"run_id"`
	RequestID  string                      `json:"request_id"`
	Status     string                      `json:"status"`
	Result     *PermissionEvaluationResult `json:"result,omitempty"`
}

// EvaluatePermissionAsync evaluates a permission request asynchronously using workflows
func (s *temporalService) EvaluatePermissionAsync(ctx context.Context, req *PermissionEvaluationRequest) (*AsyncEvaluationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.EvaluatePermissionAsync",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("resource_type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	workflowID := fmt.Sprintf("abac-eval-%s", requestID)

	s.logger.InfoContext(ctx, "Starting async permission evaluation",
		logger.Fields{
			"workflow_id":   workflowID,
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"request_id":    requestID,
		})

	// Convert to workflow input
	workflowInput := workflows.PolicyEvaluationWorkflowInput{
		EvaluationRequest: &activities.EvaluatePoliciesActivityInput{
			UserID:       req.UserID,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
			Action:       req.Action,
			EntityID:     req.EntityID,
			Attributes:   req.Context,
			RequestID:    requestID,
		},
		CacheTTLMinutes: 15, // Default 15 minute cache
		SkipCache:       false,
	}

	// Start the workflow
	// Start the workflow
	_, err := s.workflowClient.StartPolicyEvaluationWorkflow(ctx, workflowID, workflowInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("async_policy_evaluation_errors", metrics.Fields{"reason": "workflow_start_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "ASYNC_EVALUATION_FAILED", "Failed to start async policy evaluation")
	}

	s.metrics.IncrementCounter("async_policy_evaluation_started", nil)

	return &AsyncEvaluationResult{
		WorkflowID: s.workflowClient.GetWorkflow(ctx, workflowID, "").GetID(),
		RunID:      s.workflowClient.GetWorkflow(ctx, workflowID, "").GetRunID(),
		RequestID:  requestID,
		Status:     "running",
	}, nil
}

// AsyncBulkEvaluationResult represents the result of an async bulk policy evaluation
type AsyncBulkEvaluationResult struct {
	WorkflowID   string                          `json:"workflow_id"`
	RunID        string                          `json:"run_id"`
	RequestID    string                          `json:"request_id"`
	Status       string                          `json:"status"`
	RequestCount int                             `json:"request_count"`
	Result       *BulkPermissionEvaluationResult `json:"result,omitempty"`
}

// BulkEvaluatePermissionsAsync evaluates multiple permission requests asynchronously
func (s *temporalService) BulkEvaluatePermissionsAsync(ctx context.Context, req *BulkPermissionEvaluationRequest) (*AsyncBulkEvaluationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.BulkEvaluatePermissionsAsync",
		tracing.WithAttributes(
			attribute.Int("request_count", len(req.Requests)),
		))
	defer span.End()

	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	workflowID := fmt.Sprintf("abac-bulk-eval-%s", requestID)

	s.logger.InfoContext(ctx, "Starting async bulk permission evaluation",
		logger.Fields{
			"workflow_id":   workflowID,
			"request_count": len(req.Requests),
			"request_id":    requestID,
		})

	// Convert to workflow input
	evaluationRequests := make([]*activities.EvaluatePoliciesActivityInput, len(req.Requests))
	for i, evalReq := range req.Requests {
		evaluationRequests[i] = &activities.EvaluatePoliciesActivityInput{
			UserID:       evalReq.UserID,
			ResourceType: evalReq.ResourceType,
			ResourceID:   evalReq.ResourceID,
			Action:       evalReq.Action,
			EntityID:     evalReq.EntityID,
			Attributes:   evalReq.Context,
			RequestID:    fmt.Sprintf("%s-%d", requestID, i),
		}
	}

	workflowInput := workflows.BulkPolicyEvaluationWorkflowInput{
		EvaluationRequests: evaluationRequests,
		CacheTTLMinutes:    15, // Default 15 minute cache
		MaxConcurrency:     10, // Default concurrency
		SkipCache:          false,
	}

	// Start the workflow
	// Start the workflow
	_, err := s.workflowClient.StartBulkPolicyEvaluationWorkflow(ctx, workflowID, workflowInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("async_bulk_policy_evaluation_errors", metrics.Fields{"reason": "workflow_start_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "ASYNC_BULK_EVALUATION_FAILED", "Failed to start async bulk policy evaluation")
	}

	s.metrics.IncrementCounter("async_bulk_policy_evaluation_started", nil)

	return &AsyncBulkEvaluationResult{
		WorkflowID:   s.workflowClient.GetWorkflow(ctx, workflowID, "").GetID(),
		RunID:        s.workflowClient.GetWorkflow(ctx, workflowID, "").GetRunID(),
		RequestID:    requestID,
		Status:       "running",
		RequestCount: len(req.Requests),
	}, nil
}

// CacheWarmupConfig represents configuration for cache warmup
type CacheWarmupConfig struct {
	UserIDs       []uuid.UUID `json:"user_ids"`
	ResourceTypes []string    `json:"resource_types"`
	Actions       []string    `json:"actions"`
	Priority      string      `json:"priority"`
	BatchSize     int         `json:"batch_size"`
}

// WorkflowInfo represents information about a started workflow
type WorkflowInfo struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
	RequestID  string `json:"request_id"`
	Status     string `json:"status"`
}

// StartCacheCleanupWorkflow starts a long-running cache cleanup workflow
func (s *temporalService) StartCacheCleanupWorkflow(ctx context.Context) (*WorkflowInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.StartCacheCleanupWorkflow")
	defer span.End()

	requestID := uuid.New().String()
	workflowID := fmt.Sprintf("abac-cache-cleanup-%s", requestID)

	s.logger.InfoContext(ctx, "Starting cache cleanup workflow",
		logger.Fields{
			"workflow_id": workflowID,
			"request_id":  requestID,
		})

	workflowInput := workflows.CacheCleanupWorkflowInput{
		MaxAge:               24 * time.Hour, // Clean up entries older than 24 hours
		BatchSize:            1000,           // Process 1000 entries at a time
		CleanupIntervalHours: 4,              // Run cleanup every 4 hours
		RequestID:            requestID,
	}

	// Start the workflow
	_, err := s.workflowClient.StartCacheCleanupWorkflow(ctx, workflowID, workflowInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("cache_cleanup_workflow_errors", metrics.Fields{"reason": "start_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "CACHE_CLEANUP_WORKFLOW_FAILED", "Failed to start cache cleanup workflow")
	}

	s.metrics.IncrementCounter("cache_cleanup_workflow_started", nil)

	return &WorkflowInfo{
		WorkflowID: s.workflowClient.GetWorkflow(ctx, workflowID, "").GetID(),
		RunID:      s.workflowClient.GetWorkflow(ctx, workflowID, "").GetRunID(),
		RequestID:  requestID,
		Status:     "running",
	}, nil
}

// StartCacheWarmupWorkflow starts a cache warmup workflow
func (s *temporalService) StartCacheWarmupWorkflow(ctx context.Context, configs []*CacheWarmupConfig) (*WorkflowInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.StartCacheWarmupWorkflow",
		tracing.WithAttributes(
			attribute.Int("config_count", len(configs)),
		))
	defer span.End()

	requestID := uuid.New().String()
	workflowID := fmt.Sprintf("abac-cache-warmup-%s", requestID)

	s.logger.InfoContext(ctx, "Starting cache warmup workflow",
		logger.Fields{
			"workflow_id":  workflowID,
			"config_count": len(configs),
			"request_id":   requestID,
		})

	// Convert configs to workflow format
	warmupConfigs := make([]*workflows.CacheWarmupConfig, len(configs))
	for i, config := range configs {
		// Convert UUIDs to strings for workflow
		userIDStrings := make([]string, len(config.UserIDs))
		for j, userID := range config.UserIDs {
			userIDStrings[j] = userID.String()
		}

		warmupConfigs[i] = &workflows.CacheWarmupConfig{
			UserIDs:       userIDStrings,
			ResourceTypes: config.ResourceTypes,
			Actions:       config.Actions,
			Priority:      config.Priority,
			BatchSize:     config.BatchSize,
		}
	}

	workflowInput := workflows.CacheWarmupWorkflowInput{
		WarmupConfigs: warmupConfigs,
		RequestID:     requestID,
	}

	_, err := s.workflowClient.StartCacheWarmupWorkflow(ctx, workflowID, workflowInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("cache_warmup_workflow_errors", metrics.Fields{"reason": "start_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "CACHE_WARMUP_WORKFLOW_FAILED", "Failed to start cache warmup workflow")
	}

	s.metrics.IncrementCounter("cache_warmup_workflow_started", nil)

	return &WorkflowInfo{
		WorkflowID: s.workflowClient.GetWorkflow(ctx, workflowID, "").GetID(),
		RunID:      s.workflowClient.GetWorkflow(ctx, workflowID, "").GetRunID(),
		RequestID:  requestID,
		Status:     "running",
	}, nil
}

// CacheInvalidationRequest represents a cache invalidation request
type TemporalCacheInvalidationRequest struct {
	UserID        *uuid.UUID  `json:"user_id,omitempty"`
	ResourceType  *string     `json:"resource_type,omitempty"`
	ResourceID    *uuid.UUID  `json:"resource_id,omitempty"`
	Action        *string     `json:"action,omitempty"`
	PolicyIDs     []uuid.UUID `json:"policy_ids,omitempty"`
	InvalidateAll bool        `json:"invalidate_all"`
}

// InvalidateCacheAsync invalidates cache entries asynchronously
func (s *temporalService) InvalidateCacheAsync(ctx context.Context, req *TemporalCacheInvalidationRequest) (*WorkflowInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.InvalidateCacheAsync")
	defer span.End()

	requestID := uuid.New().String()
	workflowID := fmt.Sprintf("abac-cache-invalidation-%s", requestID)

	s.logger.InfoContext(ctx, "Starting cache invalidation workflow",
		logger.Fields{
			"workflow_id":    workflowID,
			"invalidate_all": req.InvalidateAll,
			"request_id":     requestID,
		})

	invalidationRequest := &activities.InvalidatePolicyCacheInput{
		UserID:        req.UserID,
		PolicyIDs:     req.PolicyIDs,
		InvalidateAll: req.InvalidateAll,
		RequestID:     requestID,
	}

	workflowInput := workflows.CacheInvalidationWorkflowInput{
		InvalidationRequests: []*activities.InvalidatePolicyCacheInput{invalidationRequest},
		RequestID:            requestID,
	}

	_, err := s.workflowClient.StartCacheInvalidationWorkflow(ctx, workflowID, workflowInput)
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("cache_invalidation_workflow_errors", metrics.Fields{"reason": "start_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "CACHE_INVALIDATION_WORKFLOW_FAILED", "Failed to start cache invalidation workflow")
	}

	s.metrics.IncrementCounter("cache_invalidation_workflow_started", nil)

	return &WorkflowInfo{
		WorkflowID: s.workflowClient.GetWorkflow(ctx, workflowID, "").GetID(),
		RunID:      s.workflowClient.GetWorkflow(ctx, workflowID, "").GetRunID(),
		RequestID:  requestID,
		Status:     "running",
	}, nil
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus struct {
	WorkflowID string      `json:"workflow_id"`
	RunID      string      `json:"run_id"`
	Status     string      `json:"status"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	StartTime  time.Time   `json:"start_time"`
	CloseTime  *time.Time  `json:"close_time,omitempty"`
}

// GetWorkflowStatus gets the status of a workflow
func (s *temporalService) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error) {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.GetWorkflowStatus",
		tracing.WithAttributes(
			attribute.String("workflow_id", workflowID),
		))
	defer span.End()

	// workflowRun := s.workflowClient.GetWorkflow(ctx, workflowID, "")

	// Get workflow description

	var desc *workflowservice.DescribeWorkflowExecutionResponse
	// var desc *client.DescribeWorkflowExecutionResponse
	var err error
	desc, err = s.workflowClient.Client.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "WORKFLOW_STATUS_FAILED", "Failed to get workflow status")
	}

	status := &WorkflowStatus{
		WorkflowID: desc.WorkflowExecutionInfo.Execution.GetWorkflowId(),
		RunID:      desc.WorkflowExecutionInfo.Execution.GetRunId(),
		Status:     desc.WorkflowExecutionInfo.Status.String(),
		StartTime:  desc.WorkflowExecutionInfo.GetStartTime().AsTime(),
	}

	if desc.WorkflowExecutionInfo.CloseTime != nil {
		closeTime := desc.WorkflowExecutionInfo.GetCloseTime().AsTime()
		status.CloseTime = &closeTime
	}

	// Try to get result if workflow is completed
	if desc.WorkflowExecutionInfo.Status.String() == "Completed" {
		var result interface{}
		err = s.workflowClient.GetWorkflow(ctx, workflowID, "").Get(ctx, &result)
		if err != nil {
			status.Error = err.Error()
		} else {
			status.Result = result
		}
	}

	return status, nil
}

// CancelWorkflow cancels a running workflow
func (s *temporalService) CancelWorkflow(ctx context.Context, workflowID string) error {
	ctx, span := s.tracer.StartSpan(ctx, "abac.temporal_service.CancelWorkflow",
		tracing.WithAttributes(
			attribute.String("workflow_id", workflowID),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Cancelling workflow",
		logger.Fields{
			"workflow_id": workflowID,
		})

	// workflowRun := s.workflowClient.GetWorkflow(ctx, workflowID, "")
	err := s.workflowClient.Client.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		s.metrics.IncrementCounter("workflow_status_errors", metrics.Fields{"reason": "get_status_failed"})
		return errors.NewBusinessErrorWithContext(ctx, "WORKFLOW_CANCELLATION_FAILED", "Failed to cancel workflow")
	}

	s.metrics.IncrementCounter("workflow_cancellation", nil)
	s.logger.InfoContext(ctx, "Workflow cancelled successfully",
		logger.Fields{
			"workflow_id": workflowID,
		})

	return nil
}
