package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	// "github.com/niiniyare/erp/internal/core/abac" // TODO: Fix import path
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TemporalClient defines the interface for interacting with ABAC Temporal workflows.
type TemporalClient interface {
	StartPermissionEvaluationWorkflow(ctx context.Context, tenantID uuid.UUID, workflowID string, request abac.PermissionEvaluationWorkflowRequest) (client.WorkflowRun, error)
	// Add other ABAC workflow starters here (e.g., StartAccessRequestWorkflow)
}

// temporalClient implements the TemporalClient interface.
type temporalClient struct {
	temporalClient client.Client
	tenantService  tenant.Service
	logger         logger.Logger
	metrics        metrics.MetricsProvider
	tracer         tracing.TracingService
}

// NewTemporalClient creates a new TemporalClient instance.
func NewTemporalClient(
	temporalClient client.Client,
	tenantService tenant.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) TemporalClient {
	return &temporalClient{
		temporalClient: temporalClient,
		tenantService:  tenantService,
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
	}
}

// StartPermissionEvaluationWorkflow starts a permission evaluation workflow.
func (c *temporalClient) StartPermissionEvaluationWorkflow(
	ctx context.Context, tenantID uuid.UUID, workflowID string, request abac.PermissionEvaluationWorkflowRequest,
) (client.WorkflowRun, error) {
	ctx, span := c.tracer.StartSpan(ctx, "abac.client.StartPermissionEvaluationWorkflow", tracing.WithSpanKind(tracing.SpanKindClient))
	defer span.End()

	// Set tenant context in the database session for the duration of this operation.
	// This ensures RLS is applied correctly for any database interactions within the workflow/activities.
	if err := c.tenantService.SetTenant(ctx, tenantID); err != nil {
		c.logger.ErrorContext(ctx, "Failed to set tenant context for workflow", logger.Fields{"tenant_id": tenantID, "error": err})
		c.metrics.IncrementErrorCount("abac_workflow_start", "set_tenant_context_failed")
		c.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewRepositoryErrorWithContext(ctx, "TENANT_CONTEXT_ERROR", "Failed to set tenant context", err)
	}
	// Ensure tenant context is reset after the operation, regardless of success or failure.
	defer func() {
		if err := c.tenantService.ResetTenant(ctx); err != nil {
			c.logger.ErrorContext(ctx, "Failed to reset tenant context after workflow start", logger.Fields{"tenant_id": tenantID, "error": err})
			c.metrics.IncrementErrorCount("abac_workflow_start", "reset_tenant_context_failed")
			c.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		}
	}()

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: abac.PermissionEvaluationTaskQueue,
	}

	run, err := c.temporalClient.ExecuteWorkflow(ctx, options, abac.PermissionEvaluationWorkflow, request)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to start permission evaluation workflow", logger.Fields{"workflow_id": workflowID, "error": err})
		c.metrics.IncrementErrorCount("abac_workflow_start", "execute_workflow_failed")
		c.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "TEMPORAL_WORKFLOW_START_FAILED", "Failed to start permission evaluation workflow").
			WithDetail("workflow_id", workflowID).WithDetail("task_queue", abac.PermissionEvaluationTaskQueue).WithErr(err)
	}

	c.logger.InfoContext(ctx, "Permission evaluation workflow started", logger.Fields{"workflow_id": run.GetID(), "run_id": run.GetRunID()})
	c.metrics.IncrementSuccessCount("abac_workflow_start")

	return run, nil
}
