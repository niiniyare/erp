package temporal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/niiniyare/erp/internal/platform/config"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkerManager manages Temporal workers for all modules
type WorkerManager struct {
	client       client.Client
	config       *config.TemporalConfig
	logger       loggerPkg.Logger
	workers      map[string]worker.Worker
	activityReg  *ActivityRegistry
	workflowReg  *WorkflowRegistry
	mu           sync.RWMutex
	stopChannels map[string]chan struct{}
}

// ActivityRegistry manages activity registrations across modules
type ActivityRegistry struct {
	activities map[string]any
	mu         sync.RWMutex
}

// WorkflowRegistry manages workflow registrations across modules
type WorkflowRegistry struct {
	workflows map[string]any
	mu        sync.RWMutex
}

// NewActivityRegistry creates a new activity registry
func NewActivityRegistry() *ActivityRegistry {
	return &ActivityRegistry{
		activities: make(map[string]any),
	}
}

// NewWorkflowRegistry creates a new workflow registry
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]any),
	}
}

// RegisterActivity registers an activity with the registry
func (ar *ActivityRegistry) RegisterActivity(name string, activity any) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.activities[name] = activity
}

// RegisterWorkflow registers a workflow with the registry
func (wr *WorkflowRegistry) RegisterWorkflow(name string, workflow any) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.workflows[name] = workflow
}

// GetActivities returns all registered activities
func (ar *ActivityRegistry) GetActivities() map[string]any {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	activities := make(map[string]any)
	for k, v := range ar.activities {
		activities[k] = v
	}
	return activities
}

// GetWorkflows returns all registered workflows
func (wr *WorkflowRegistry) GetWorkflows() map[string]any {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	workflows := make(map[string]any)
	for k, v := range wr.workflows {
		workflows[k] = v
	}
	return workflows
}

// NewWorkerManager creates a new Temporal worker manager
func NewWorkerManager(client client.Client, cfg *config.TemporalConfig, logger loggerPkg.Logger) *WorkerManager {
	return &WorkerManager{
		client:       client,
		config:       cfg,
		logger:       logger,
		workers:      make(map[string]worker.Worker),
		activityReg:  NewActivityRegistry(),
		workflowReg:  NewWorkflowRegistry(),
		stopChannels: make(map[string]chan struct{}),
	}
}

// RegisterModuleActivities registers activities for a specific module
func (wm *WorkerManager) RegisterModuleActivities(moduleName string, activities map[string]any) {
	for name, activity := range activities {
		activityName := fmt.Sprintf("%s.%s", moduleName, name)
		wm.activityReg.RegisterActivity(activityName, activity)
	}
	
	wm.logger.Info("Registered module activities", loggerPkg.Fields{
		"module":           moduleName,
		"activity_count":   len(activities),
		"registered_names": getActivityNames(activities),
	})
}

// RegisterModuleWorkflows registers workflows for a specific module
func (wm *WorkerManager) RegisterModuleWorkflows(moduleName string, workflows map[string]any) {
	for name, workflowFunc := range workflows {
		workflowName := fmt.Sprintf("%s.%s", moduleName, name)
		wm.workflowReg.RegisterWorkflow(workflowName, workflowFunc)
	}
	
	wm.logger.Info("Registered module workflows", loggerPkg.Fields{
		"module":          moduleName,
		"workflow_count":  len(workflows),
		"registered_names": getWorkflowNames(workflows),
	})
}

// StartWorkers starts all enabled workers based on configuration
func (wm *WorkerManager) StartWorkers(ctx context.Context) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Start system workers
	if err := wm.startSystemWorkers(ctx); err != nil {
		return fmt.Errorf("failed to start system workers: %w", err)
	}

	// Start module workers for enabled modules
	enabledModules := wm.config.GetEnabledModules()
	for _, moduleName := range enabledModules {
		if err := wm.startModuleWorkers(ctx, moduleName); err != nil {
			wm.logger.Error("Failed to start module workers", loggerPkg.Fields{
				"module": moduleName,
				"error":  err.Error(),
			})
			continue
		}
	}

	wm.logger.Info("✅ All Temporal Workers Started", loggerPkg.Fields{
		"total_workers":     len(wm.workers),
		"enabled_modules":   enabledModules,
		"system_workers":    []string{"system", "notifications", "analytics"},
		"status":           "running",
	})

	return nil
}

// startSystemWorkers starts system-wide workers
func (wm *WorkerManager) startSystemWorkers(ctx context.Context) error {
	systemWorkers := []struct {
		name   string
		config config.TemporalTaskQueueConfig
	}{
		{"system", wm.config.Workers.System},
		{"notifications", wm.config.Workers.Notifications},
		{"analytics", wm.config.Workers.Analytics},
	}

	for _, sw := range systemWorkers {
		if !sw.config.Enabled {
			wm.logger.Debug("System worker disabled", loggerPkg.Fields{
				"worker": sw.name,
			})
			continue
		}

		if err := wm.startWorker(ctx, sw.name, sw.config, true); err != nil {
			return fmt.Errorf("failed to start system worker %s: %w", sw.name, err)
		}
	}

	return nil
}

// startModuleWorkers starts workers for a specific module
func (wm *WorkerManager) startModuleWorkers(ctx context.Context, moduleName string) error {
	moduleConfig := wm.config.GetModuleConfig(moduleName)
	if moduleConfig == nil || !moduleConfig.Enabled {
		return nil
	}

	for queueName, queueConfig := range moduleConfig.TaskQueues {
		if !queueConfig.Enabled {
			continue
		}

		workerName := fmt.Sprintf("%s.%s", moduleName, queueName)
		if err := wm.startWorker(ctx, workerName, queueConfig, false); err != nil {
			return fmt.Errorf("failed to start worker %s: %w", workerName, err)
		}
	}

	return nil
}

// startWorker starts a single worker with the given configuration
func (wm *WorkerManager) startWorker(ctx context.Context, workerName string, config config.TemporalTaskQueueConfig, isSystemWorker bool) error {
	// Create worker options with proper configuration
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:      config.MaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize:  config.MaxConcurrentWorkflows,
		MaxConcurrentLocalActivityExecutionSize: config.MaxConcurrentLocalActivities,
		EnableLoggingInReplay:                   wm.config.Workers.EnableLoggingInReplay,
		StickyScheduleToStartTimeout:            wm.config.Workers.StickyScheduleToStartTimeout,
		WorkerStopTimeout:                       wm.config.Workers.WorkerStopTimeout,
	}

	// Create task queue name
	taskQueue := workerName
	if isSystemWorker {
		taskQueue = fmt.Sprintf("system.%s", workerName)
	}

	// Create worker
	w := worker.New(wm.client, taskQueue, workerOptions)

	// Register activities and workflows
	wm.registerWorkerActivities(w, workerName, isSystemWorker)
	wm.registerWorkerWorkflows(w, workerName, isSystemWorker)

	// Start worker in background
	stopChannel := make(chan struct{})
	wm.workers[workerName] = w
	wm.stopChannels[workerName] = stopChannel

	go func() {
		wm.logger.Info("Starting Temporal worker", loggerPkg.Fields{
			"worker":     workerName,
			"task_queue": taskQueue,
			"type":       getWorkerType(isSystemWorker),
		})

		if err := w.Run(worker.InterruptCh()); err != nil {
			wm.logger.Error("Temporal worker stopped with error", loggerPkg.Fields{
				"worker": workerName,
				"error":  err.Error(),
			})
		}

		wm.logger.Info("Temporal worker stopped", loggerPkg.Fields{
			"worker": workerName,
		})
	}()

	return nil
}

// registerWorkerActivities registers activities for a worker
func (wm *WorkerManager) registerWorkerActivities(w worker.Worker, workerName string, isSystemWorker bool) {
	activities := wm.activityReg.GetActivities()
	
	if isSystemWorker {
		// System workers get all activities
		for _, activity := range activities {
			w.RegisterActivity(activity)
		}
	} else {
		// Module workers get module-specific activities
		moduleName := getModuleFromWorkerName(workerName)
		for name, activity := range activities {
			if isActivityForModule(name, moduleName) {
				w.RegisterActivity(activity)
			}
		}
	}
}

// registerWorkerWorkflows registers workflows for a worker
func (wm *WorkerManager) registerWorkerWorkflows(w worker.Worker, workerName string, isSystemWorker bool) {
	workflows := wm.workflowReg.GetWorkflows()
	
	if isSystemWorker {
		// System workers get all workflows
		for _, workflowFunc := range workflows {
			w.RegisterWorkflow(workflowFunc)
		}
	} else {
		// Module workers get module-specific workflows
		moduleName := getModuleFromWorkerName(workerName)
		for name, workflowFunc := range workflows {
			if isWorkflowForModule(name, moduleName) {
				w.RegisterWorkflow(workflowFunc)
			}
		}
	}
}

// StopWorkers gracefully stops all workers
func (wm *WorkerManager) StopWorkers(ctx context.Context) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if len(wm.workers) == 0 {
		return nil
	}

	wm.logger.Info("Stopping Temporal workers", loggerPkg.Fields{
		"worker_count": len(wm.workers),
	})

	// Create timeout context
	stopCtx, cancel := context.WithTimeout(ctx, wm.config.Workers.WorkerStopTimeout)
	defer cancel()

	// Stop all workers concurrently
	var wg sync.WaitGroup
	for name, stopChannel := range wm.stopChannels {
		wg.Add(1)
		go func(workerName string, ch chan struct{}) {
			defer wg.Done()
			
			select {
			case ch <- struct{}{}:
				wm.logger.Debug("Sent stop signal to worker", loggerPkg.Fields{
					"worker": workerName,
				})
			case <-stopCtx.Done():
				wm.logger.Warn("Timeout sending stop signal to worker", loggerPkg.Fields{
					"worker": workerName,
				})
			}
		}(name, stopChannel)
	}

	// Wait for all stop signals or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wm.logger.Info("All Temporal workers stopped gracefully")
	case <-stopCtx.Done():
		wm.logger.Warn("Timeout waiting for workers to stop")
	}

	// Clear workers
	wm.workers = make(map[string]worker.Worker)
	wm.stopChannels = make(map[string]chan struct{})

	return nil
}

// GetWorkerStatus returns status of all workers
func (wm *WorkerManager) GetWorkerStatus() map[string]any {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	status := map[string]any{
		"total_workers":      len(wm.workers),
		"enabled_modules":    wm.config.GetEnabledModules(),
		"registered_activities": len(wm.activityReg.GetActivities()),
		"registered_workflows":  len(wm.workflowReg.GetWorkflows()),
		"workers": make(map[string]string),
	}

	for name := range wm.workers {
		status["workers"].(map[string]string)[name] = "running"
	}

	return status
}

// Helper functions

func getActivityNames(activities map[string]any) []string {
	names := make([]string, 0, len(activities))
	for name := range activities {
		names = append(names, name)
	}
	return names
}

func getWorkflowNames(workflows map[string]any) []string {
	names := make([]string, 0, len(workflows))
	for name := range workflows {
		names = append(names, name)
	}
	return names
}

func getWorkerType(isSystemWorker bool) string {
	if isSystemWorker {
		return "system"
	}
	return "module"
}

func getModuleFromWorkerName(workerName string) string {
	// Extract module name from worker name (e.g., "finance.standard" -> "finance")
	if idx := len(workerName); idx > 0 {
		for i, c := range workerName {
			if c == '.' {
				return workerName[:i]
			}
		}
	}
	return workerName
}

func isActivityForModule(activityName, moduleName string) bool {
	return len(activityName) > len(moduleName) && activityName[:len(moduleName)] == moduleName
}

func isWorkflowForModule(workflowName, moduleName string) bool {
	return len(workflowName) > len(moduleName) && workflowName[:len(moduleName)] == moduleName
}

// WorkflowStarter provides utilities for starting workflows
type WorkflowStarter struct {
	client client.Client
	logger loggerPkg.Logger
}

// NewWorkflowStarter creates a new workflow starter
func NewWorkflowStarter(client client.Client, logger loggerPkg.Logger) *WorkflowStarter {
	return &WorkflowStarter{
		client: client,
		logger: logger,
	}
}

// StartWorkflow starts a workflow with proper logging and error handling
func (ws *WorkflowStarter) StartWorkflow(
	ctx context.Context,
	workflowType string,
	taskQueue string,
	input any,
	options ...client.StartWorkflowOptions,
) (client.WorkflowRun, error) {
	// Set default options
	defaultOptions := client.StartWorkflowOptions{
		TaskQueue: taskQueue,
		ID:        fmt.Sprintf("%s-%d", workflowType, time.Now().UnixNano()),
	}

	// Merge with provided options
	if len(options) > 0 {
		opts := options[0]
		if opts.TaskQueue != "" {
			defaultOptions.TaskQueue = opts.TaskQueue
		}
		if opts.ID != "" {
			defaultOptions.ID = opts.ID
		}
		if opts.WorkflowExecutionTimeout != 0 {
			defaultOptions.WorkflowExecutionTimeout = opts.WorkflowExecutionTimeout
		}
		if opts.WorkflowRunTimeout != 0 {
			defaultOptions.WorkflowRunTimeout = opts.WorkflowRunTimeout
		}
		if opts.WorkflowTaskTimeout != 0 {
			defaultOptions.WorkflowTaskTimeout = opts.WorkflowTaskTimeout
		}
	}

	// Start workflow
	workflowRun, err := ws.client.ExecuteWorkflow(ctx, defaultOptions, workflowType, input)
	if err != nil {
		ws.logger.Error("Failed to start workflow", loggerPkg.Fields{
			"workflow_type": workflowType,
			"workflow_id":   defaultOptions.ID,
			"task_queue":    defaultOptions.TaskQueue,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to start workflow %s: %w", workflowType, err)
	}

	ws.logger.Info("Started workflow", loggerPkg.Fields{
		"workflow_type": workflowType,
		"workflow_id":   workflowRun.GetID(),
		"run_id":        workflowRun.GetRunID(),
		"task_queue":    defaultOptions.TaskQueue,
	})

	return workflowRun, nil
}