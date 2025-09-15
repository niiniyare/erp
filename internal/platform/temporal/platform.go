package temporal

import (
	"context"
	"fmt"
	"sync"

	"github.com/niiniyare/erp/internal/platform/config"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/sdk/client"
)

// Platform represents the complete Temporal platform for the ERP system
type Platform struct {
	clientManager   *ClientManager
	workerManager   *WorkerManager
	workflowStarter *WorkflowStarter
	config          *config.TemporalConfig
	logger          loggerPkg.Logger
	started         bool
	mu              sync.RWMutex
}

// NewPlatform creates a new Temporal platform instance
func NewPlatform(cfg *config.TemporalConfig, logger loggerPkg.Logger) (*Platform, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid temporal configuration: %w", err)
	}

	// Create client manager
	clientManager, err := NewClientManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create client manager: %w", err)
	}

	// Create worker manager
	workerManager := NewWorkerManager(clientManager.GetClient(), cfg, logger)

	// Create workflow starter
	workflowStarter := NewWorkflowStarter(clientManager.GetClient(), logger)

	platform := &Platform{
		clientManager:   clientManager,
		workerManager:   workerManager,
		workflowStarter: workflowStarter,
		config:          cfg,
		logger:          logger,
	}

	logger.Info("✅ Temporal Platform Initialized", loggerPkg.Fields{
		"host_port":       cfg.HostPort,
		"namespace":       cfg.Namespace,
		"enabled_modules": cfg.GetEnabledModules(),
		"tls_enabled":     cfg.TLS.Enabled,
		"metrics_enabled": cfg.Metrics.Enabled,
		"status":          "initialized",
	})

	return platform, nil
}

// Start starts the Temporal platform (client connection + workers)
func (p *Platform) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("temporal platform already started")
	}

	// Test client connection
	if err := p.clientManager.CheckConnection(ctx); err != nil {
		return fmt.Errorf("failed to verify temporal connection: %w", err)
	}

	// Start all workers
	if err := p.workerManager.StartWorkers(ctx); err != nil {
		return fmt.Errorf("failed to start temporal workers: %w", err)
	}

	p.started = true

	p.logger.Info("🚀 Temporal Platform Started", loggerPkg.Fields{
		"namespace":    p.config.Namespace,
		"host_port":    p.config.HostPort,
		"worker_count": len(p.workerManager.workers),
		"status":       "running",
	})

	return nil
}

// Stop gracefully stops the Temporal platform
func (p *Platform) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return nil
	}

	p.logger.Info("Stopping Temporal Platform...")

	// Stop workers first
	if err := p.workerManager.StopWorkers(ctx); err != nil {
		p.logger.Error("Error stopping temporal workers", loggerPkg.Fields{
			"error": err.Error(),
		})
	}

	// Close client connection
	p.clientManager.Close()

	p.started = false

	p.logger.Info("✅ Temporal Platform Stopped")
	return nil
}

// RegisterModuleActivities registers activities for a specific module
func (p *Platform) RegisterModuleActivities(moduleName string, activities map[string]any) {
	if !p.config.IsModuleEnabled(moduleName) {
		p.logger.Debug("Module not enabled, skipping activity registration", loggerPkg.Fields{
			"module": moduleName,
		})
		return
	}

	p.workerManager.RegisterModuleActivities(moduleName, activities)
}

// RegisterModuleWorkflows registers workflows for a specific module
func (p *Platform) RegisterModuleWorkflows(moduleName string, workflows map[string]any) {
	if !p.config.IsModuleEnabled(moduleName) {
		p.logger.Debug("Module not enabled, skipping workflow registration", loggerPkg.Fields{
			"module": moduleName,
		})
		return
	}

	p.workerManager.RegisterModuleWorkflows(moduleName, workflows)
}

// GetClient returns the Temporal client for workflow operations
func (p *Platform) GetClient() client.Client {
	return p.clientManager.GetClient()
}

// GetWorkflowStarter returns the workflow starter for launching workflows
func (p *Platform) GetWorkflowStarter() *WorkflowStarter {
	return p.workflowStarter
}

// GetStatus returns detailed status information about the platform
func (p *Platform) GetStatus() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := map[string]any{
		"platform_started": p.started,
		"namespace":        p.config.Namespace,
		"host_port":        p.config.HostPort,
		"enabled_modules":  p.config.GetEnabledModules(),
		"tls_enabled":      p.config.TLS.Enabled,
		"metrics_enabled":  p.config.Metrics.Enabled,
	}

	// Add worker status if available
	if p.started {
		status["workers"] = p.workerManager.GetWorkerStatus()
	}

	return status
}

// IsStarted returns whether the platform is currently running
func (p *Platform) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}

// HealthCheck performs a health check of the platform
func (p *Platform) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("temporal platform not started")
	}

	// Check client connection
	if err := p.clientManager.CheckConnection(ctx); err != nil {
		return fmt.Errorf("temporal client health check failed: %w", err)
	}

	// Check worker status (basic check - could be enhanced)
	workerStatus := p.workerManager.GetWorkerStatus()
	if workerCount, ok := workerStatus["total_workers"].(int); ok && workerCount == 0 {
		return fmt.Errorf("no temporal workers running")
	}

	p.logger.Debug("Temporal platform health check passed", loggerPkg.Fields{
		"worker_count": workerStatus["total_workers"],
	})

	return nil
}

// GetConfig returns the current Temporal configuration
func (p *Platform) GetConfig() *config.TemporalConfig {
	return p.config
}

// ModuleRegistrar provides a convenient interface for modules to register their components
type ModuleRegistrar struct {
	platform   *Platform
	moduleName string
}

// NewModuleRegistrar creates a module registrar for a specific module
func (p *Platform) NewModuleRegistrar(moduleName string) *ModuleRegistrar {
	return &ModuleRegistrar{
		platform:   p,
		moduleName: moduleName,
	}
}

// RegisterActivities registers activities for the module
func (mr *ModuleRegistrar) RegisterActivities(activities map[string]any) {
	mr.platform.RegisterModuleActivities(mr.moduleName, activities)
}

// RegisterWorkflows registers workflows for the module
func (mr *ModuleRegistrar) RegisterWorkflows(workflows map[string]any) {
	mr.platform.RegisterModuleWorkflows(mr.moduleName, workflows)
}

// IsEnabled returns whether the module is enabled in configuration
func (mr *ModuleRegistrar) IsEnabled() bool {
	return mr.platform.config.IsModuleEnabled(mr.moduleName)
}

// GetModuleName returns the module name
func (mr *ModuleRegistrar) GetModuleName() string {
	return mr.moduleName
}
