package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/abac/workflows"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ABACWorkerConfig represents configuration for the ABAC worker
type ABACWorkerConfig struct {
	TaskQueue          string `json:"task_queue"`
	MaxConcurrentTasks int    `json:"max_concurrent_tasks"`
	MaxPollers         int    `json:"max_pollers"`
}

// ABACWorker represents an ABAC Temporal worker
type ABACWorker struct {
	client client.Client
	worker worker.Worker
	config ABACWorkerConfig
	logger loggerPkg.Logger
}

// NewABACWorker creates a new ABAC Temporal worker
func NewABACWorker(
	temporalClient client.Client,
	config ABACWorkerConfig,
	policyRepo repository.PolicyRepository,
	attributeRepo repository.AttributeRepository,
	policyEvaluationRepo repository.PolicyEvaluationRepository,
	identityService identity.Service,
	tenantService tenant.Service,
	logger loggerPkg.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *ABACWorker {
	// Set default task queue if not provided
	if config.TaskQueue == "" {
		config.TaskQueue = "abac-task-queue"
	}

	// Set default concurrency limits
	if config.MaxConcurrentTasks == 0 {
		config.MaxConcurrentTasks = 100
	}
	if config.MaxPollers == 0 {
		config.MaxPollers = 10
	}

	// Create worker options
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:     config.MaxConcurrentTasks,
		MaxConcurrentWorkflowTaskExecutionSize: config.MaxPollers,
	}

	// Create the worker
	w := worker.New(temporalClient, config.TaskQueue, workerOptions)

	// Register workflows
	w.RegisterWorkflow(workflows.PolicyEvaluationWorkflow)
	w.RegisterWorkflow(workflows.BulkPolicyEvaluationWorkflow)
	w.RegisterWorkflow(workflows.CacheCleanupWorkflow)
	w.RegisterWorkflow(workflows.CacheWarmupWorkflow)
	w.RegisterWorkflow(workflows.CacheInvalidationWorkflow)

	// Create activity instances
	policyEvaluationActivities := activities.NewPolicyEvaluationActivities(
		policyRepo, policyEvaluationRepo, logger, metrics, tracer)

	attributeCollectionActivities := activities.NewAttributeCollectionActivities(
		attributeRepo, identityService, tenantService, logger, metrics, tracer)

	cacheActivities := activities.NewCacheActivities(
		policyEvaluationRepo, attributeRepo, logger, metrics, tracer)

	// Register activities
	w.RegisterActivity(policyEvaluationActivities)
	w.RegisterActivity(attributeCollectionActivities)
	w.RegisterActivity(cacheActivities)

	logger.Info("ABAC Temporal worker created",
		loggerPkg.Fields{
			"task_queue":           config.TaskQueue,
			"max_concurrent_tasks": config.MaxConcurrentTasks,
			"max_pollers":          config.MaxPollers,
		})

	return &ABACWorker{
		client: temporalClient,
		worker: w,
		config: config,
		logger: logger,
	}
}

// Start starts the ABAC worker
func (w *ABACWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting ABAC Temporal worker",
		loggerPkg.Fields{
			"task_queue": w.config.TaskQueue,
		})

	// Start the worker
	err := w.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start ABAC worker: %w", err)
	}

	w.logger.Info("ABAC Temporal worker started successfully")
	return nil
}

// Stop stops the ABAC worker gracefully
func (w *ABACWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping ABAC Temporal worker")

	// Stop the worker
	w.worker.Stop()

	w.logger.Info("ABAC Temporal worker stopped")
	return nil
}

// GetTaskQueue returns the task queue name
func (w *ABACWorker) GetTaskQueue() string {
	return w.config.TaskQueue
}

// WorkflowClient provides access to the Temporal client for workflow operations
type WorkflowClient struct {
	Client    client.Client
	taskQueue string
	logger    loggerPkg.Logger
}

// NewWorkflowClient creates a new workflow client for ABAC operations
func NewWorkflowClient(
	temporalClient client.Client,
	taskQueue string,
	logger loggerPkg.Logger,
) *WorkflowClient {
	if taskQueue == "" {
		taskQueue = "abac-task-queue"
	}

	return &WorkflowClient{
		Client:    temporalClient,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

// StartPolicyEvaluationWorkflow starts a policy evaluation workflow
func (wc *WorkflowClient) StartPolicyEvaluationWorkflow(
	ctx context.Context,
	workflowID string,
	input workflows.PolicyEvaluationWorkflowInput,
) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: wc.taskQueue,
	}

	wc.logger.InfoContext(ctx, "Starting policy evaluation workflow",
		loggerPkg.Fields{
			"workflow_id": workflowID,
			"user_id":     input.EvaluationRequest.UserID,
			"resource":    input.EvaluationRequest.ResourceType,
			"action":      input.EvaluationRequest.Action,
		})

	return wc.Client.ExecuteWorkflow(ctx, options, workflows.PolicyEvaluationWorkflowName, input)
}

// StartBulkPolicyEvaluationWorkflow starts a bulk policy evaluation workflow
func (wc *WorkflowClient) StartBulkPolicyEvaluationWorkflow(
	ctx context.Context,
	workflowID string,
	input workflows.BulkPolicyEvaluationWorkflowInput,
) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: wc.taskQueue,
	}

	wc.logger.InfoContext(ctx, "Starting bulk policy evaluation workflow",
		loggerPkg.Fields{
			"workflow_id":     workflowID,
			"request_count":   len(input.EvaluationRequests),
			"max_concurrency": input.MaxConcurrency,
		})

	return wc.Client.ExecuteWorkflow(ctx, options, workflows.BulkPolicyEvaluationWorkflowName, input)
}

// StartCacheCleanupWorkflow starts a cache cleanup workflow
func (wc *WorkflowClient) StartCacheCleanupWorkflow(
	ctx context.Context,
	workflowID string,
	input workflows.CacheCleanupWorkflowInput,
) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: wc.taskQueue,
		// Cache cleanup workflows can be long-running
		WorkflowExecutionTimeout: 0, // No timeout for cleanup workflows
	}

	wc.logger.InfoContext(ctx, "Starting cache cleanup workflow",
		loggerPkg.Fields{
			"workflow_id":      workflowID,
			"cleanup_interval": input.CleanupIntervalHours,
			"max_age":          input.MaxAge,
		})

	return wc.Client.ExecuteWorkflow(ctx, options, workflows.CacheCleanupWorkflowName, input)
}

// StartCacheWarmupWorkflow starts a cache warmup workflow
func (wc *WorkflowClient) StartCacheWarmupWorkflow(
	ctx context.Context,
	workflowID string,
	input workflows.CacheWarmupWorkflowInput,
) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: wc.taskQueue,
	}

	wc.logger.InfoContext(ctx, "Starting cache warmup workflow",
		loggerPkg.Fields{
			"workflow_id":  workflowID,
			"config_count": len(input.WarmupConfigs),
		})

	return wc.Client.ExecuteWorkflow(ctx, options, workflows.CacheWarmupWorkflowName, input)
}

// StartCacheInvalidationWorkflow starts a cache invalidation workflow
func (wc *WorkflowClient) StartCacheInvalidationWorkflow(
	ctx context.Context,
	workflowID string,
	input workflows.CacheInvalidationWorkflowInput,
) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: wc.taskQueue,
	}

	wc.logger.InfoContext(ctx, "Starting cache invalidation workflow",
		loggerPkg.Fields{
			"workflow_id":   workflowID,
			"request_count": len(input.InvalidationRequests),
		})

	return wc.Client.ExecuteWorkflow(ctx, options, workflows.CacheInvalidationWorkflowName, input)
}

// GetWorkflow gets a workflow execution
func (wc *WorkflowClient) GetWorkflow(ctx context.Context, workflowID, runID string) client.WorkflowRun {
	return wc.Client.GetWorkflow(ctx, workflowID, runID)
}
