package client

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/niiniyare/erp/internal/core/abac/activities"
	"github.com/niiniyare/erp/internal/core/abac/workflows"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TemporalConfig holds configuration for Temporal client
type TemporalConfig struct {
	HostPort  string `json:"host_port" default:"localhost:7233"`
	Namespace string `json:"namespace" default:"default"`
	TaskQueue string `json:"task_queue" default:"abac-task-queue"`
}

// ABACTemporalClient wraps Temporal client for ABAC operations
type ABACTemporalClient struct {
	client    client.Client
	config    *TemporalConfig
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracing   tracing.TracingService
	worker    worker.Worker
}

// NewABACTemporalClient creates a new ABAC Temporal client
func NewABACTemporalClient(
	config *TemporalConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracing tracing.TracingService,
) (*ABACTemporalClient, error) {
	// Create Temporal client options
	clientOptions := client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
		Logger:    logger,
		MetricsHandler: metrics.GetTemporalMetricsHandler(), // TODO: Implement this
	}

	// Create Temporal client
	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	abacClient := &ABACTemporalClient{
		client:  temporalClient,
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}

	// Initialize worker
	abacClient.initializeWorker()

	return abacClient, nil
}

// initializeWorker creates and configures the Temporal worker
func (c *ABACTemporalClient) initializeWorker() {
	// Create worker options
	workerOptions := worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 10,
		MaxConcurrentActivityExecutionSize:     20,
	}

	// Create worker
	c.worker = worker.New(c.client, c.config.TaskQueue, workerOptions)

	// Register workflows
	c.worker.RegisterWorkflow(workflows.PermissionEvaluationWorkflow)
	c.worker.RegisterWorkflow(workflows.BulkPermissionEvaluationWorkflow)
	// TODO: Register other workflows as they are implemented

	// Create activity instances
	attributeActivities := activities.NewAttributeCollectionActivities()
	policyActivities := activities.NewPolicyEvaluationActivities(c.logger, c.metrics, c.tracing)
	cacheActivities := activities.NewCacheActivities(c.logger, c.metrics, c.tracing)

	// Register activities
	c.worker.RegisterActivity(attributeActivities.CollectUserAttributes)
	c.worker.RegisterActivity(attributeActivities.CollectResourceAttributes)
	c.worker.RegisterActivity(attributeActivities.EnrichEnvironmentContext)
	
	c.worker.RegisterActivity(policyActivities.EvaluatePolicies)
	
	c.worker.RegisterActivity(cacheActivities.GetCachedPolicyEvaluation)
	c.worker.RegisterActivity(cacheActivities.CachePolicyEvaluation)
	c.worker.RegisterActivity(cacheActivities.InvalidatePolicyCache)
	c.worker.RegisterActivity(cacheActivities.GetUserEffectiveRoles)
	c.worker.RegisterActivity(cacheActivities.LogPermissionEvaluation)
	c.worker.RegisterActivity(cacheActivities.GenerateContextHash)
	c.worker.RegisterActivity(cacheActivities.CleanupExpiredCache)
	c.worker.RegisterActivity(cacheActivities.WarmupCache)

	c.logger.Info("Temporal worker initialized", 
		"task_queue", c.config.TaskQueue,
		"workflows_registered", 2,
		"activities_registered", 11)
}

// StartWorker starts the Temporal worker
func (c *ABACTemporalClient) StartWorker(ctx context.Context) error {
	c.logger.Info("Starting ABAC Temporal worker", "task_queue", c.config.TaskQueue)
	
	// Start worker in a goroutine
	go func() {
		err := c.worker.Run(worker.InterruptCh())
		if err != nil {
			c.logger.Error("Temporal worker stopped with error", "error", err)
		} else {
			c.logger.Info("Temporal worker stopped gracefully")
		}
	}()

	return nil
}

// StopWorker stops the Temporal worker
func (c *ABACTemporalClient) StopWorker() {
	c.logger.Info("Stopping ABAC Temporal worker")
	c.worker.Stop()
}

// ExecutePermissionEvaluationWorkflow executes a permission evaluation workflow
func (c *ABACTemporalClient) ExecutePermissionEvaluationWorkflow(
	ctx context.Context,
	workflowID string,
	request interface{},
) (client.WorkflowRun, error) {
	// Set workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: c.config.TaskQueue,
		WorkflowExecutionTimeout: time.Minute * 5,
		WorkflowTaskTimeout:      time.Minute * 1,
	}

	// Execute workflow
	workflowRun, err := c.client.ExecuteWorkflow(
		ctx,
		workflowOptions,
		workflows.PermissionEvaluationWorkflow,
		request,
	)
	
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.permission_evaluation.start_error", 1)
		return nil, fmt.Errorf("failed to start permission evaluation workflow: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.permission_evaluation.started", 1)
	c.logger.Debug("Started permission evaluation workflow", 
		"workflow_id", workflowID,
		"run_id", workflowRun.GetRunID())

	return workflowRun, nil
}

// ExecuteBulkPermissionEvaluationWorkflow executes a bulk permission evaluation workflow
func (c *ABACTemporalClient) ExecuteBulkPermissionEvaluationWorkflow(
	ctx context.Context,
	workflowID string,
	requests interface{},
) (client.WorkflowRun, error) {
	// Set workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: c.config.TaskQueue,
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 2,
	}

	// Execute workflow
	workflowRun, err := c.client.ExecuteWorkflow(
		ctx,
		workflowOptions,
		workflows.BulkPermissionEvaluationWorkflow,
		requests,
	)
	
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.bulk_permission_evaluation.start_error", 1)
		return nil, fmt.Errorf("failed to start bulk permission evaluation workflow: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.bulk_permission_evaluation.started", 1)
	c.logger.Debug("Started bulk permission evaluation workflow", 
		"workflow_id", workflowID,
		"run_id", workflowRun.GetRunID())

	return workflowRun, nil
}

// GetWorkflowResult gets the result of a completed workflow
func (c *ABACTemporalClient) GetWorkflowResult(
	ctx context.Context,
	workflowID string,
	runID string,
	result interface{},
) error {
	// Get workflow run
	workflowRun := c.client.GetWorkflow(ctx, workflowID, runID)
	
	// Get result
	err := workflowRun.Get(ctx, result)
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.get_result.error", 1)
		return fmt.Errorf("failed to get workflow result: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.get_result.success", 1)
	return nil
}

// CancelWorkflow cancels a running workflow
func (c *ABACTemporalClient) CancelWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
) error {
	err := c.client.CancelWorkflow(ctx, workflowID, runID)
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.cancel.error", 1)
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.cancel.success", 1)
	c.logger.Info("Workflow cancelled", 
		"workflow_id", workflowID,
		"run_id", runID)

	return nil
}

// SignalWorkflow sends a signal to a running workflow
func (c *ABACTemporalClient) SignalWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	signalName string,
	arg interface{},
) error {
	err := c.client.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.signal.error", 1)
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.signal.success", 1)
	c.logger.Debug("Workflow signaled", 
		"workflow_id", workflowID,
		"run_id", runID,
		"signal_name", signalName)

	return nil
}

// QueryWorkflow queries a running workflow
func (c *ABACTemporalClient) QueryWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	queryType string,
	result interface{},
) error {
	// Get workflow run
	workflowRun := c.client.GetWorkflow(ctx, workflowID, runID)
	
	// Query workflow
	value, err := workflowRun.Query(ctx, queryType)
	if err != nil {
		c.metrics.RecordCounter("abac.workflow.query.error", 1)
		return fmt.Errorf("failed to query workflow: %w", err)
	}

	// Decode result
	err = value.Get(result)
	if err != nil {
		return fmt.Errorf("failed to decode query result: %w", err)
	}

	c.metrics.RecordCounter("abac.workflow.query.success", 1)
	return nil
}

// Close closes the Temporal client
func (c *ABACTemporalClient) Close() {
	c.logger.Info("Closing ABAC Temporal client")
	
	// Stop worker first
	c.StopWorker()
	
	// Close client
	c.client.Close()
	
	c.logger.Info("ABAC Temporal client closed")
}

// Health check for the Temporal client
func (c *ABACTemporalClient) HealthCheck(ctx context.Context) error {
	// Check if we can connect to Temporal
	_, err := c.client.CheckHealth(ctx, nil)
	if err != nil {
		return fmt.Errorf("Temporal health check failed: %w", err)
	}
	
	return nil
}