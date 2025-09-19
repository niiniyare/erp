package messaging

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/config"
	platformConfig "github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/api/workflowservice/v1"
)

// TemporalComponent manages Temporal workflow engine connections and lifecycle
type TemporalComponent struct {
	platform *temporal.Platform
	config   *config.TemporalSettings
	logger   logger.Logger
}

// NewTemporal creates a new Temporal component
func NewTemporal(cfg *config.TemporalSettings, logger logger.Logger) *TemporalComponent {
	return &TemporalComponent{
		config: cfg,
		logger: logger,
	}
}

// Start initializes the Temporal platform
func (t *TemporalComponent) Start(ctx context.Context) error {
	t.logger.Info("Starting Temporal platform", logger.Fields{
		"host_port": t.config.HostPort,
		"namespace": t.config.Namespace,
	})

	// Convert new config format to platform temporal config format
	platformTemporalConfig := &platformConfig.TemporalConfig{
		HostPort:  t.config.HostPort,
		Namespace: t.config.Namespace,
		Workers: platformConfig.TemporalWorkersConfig{
			MaxConcurrentActivities: t.config.Workers.MaxConcurrentActivities,
			MaxConcurrentWorkflows:  t.config.Workers.MaxConcurrentWorkflows,
			WorkerStopTimeout:       t.config.Workers.WorkerStopTimeout,
		},
		Client: platformConfig.TemporalClientConfig{
			Identity:          t.config.Client.Identity,
			ConnectionTimeout: t.config.Client.ConnectionTimeout,
		},
	}

	// Create Temporal platform
	platform, err := temporal.NewPlatform(platformTemporalConfig, t.logger)
	if err != nil {
		t.logger.Error("Failed to create Temporal platform", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create Temporal platform: %w", err)
	}

	t.platform = platform

	t.logger.Info("Temporal platform initialized successfully", logger.Fields{
		"host_port": t.config.HostPort,
		"namespace": t.config.Namespace,
	})

	return nil
}

// Stop gracefully shuts down the Temporal platform
func (t *TemporalComponent) Stop(ctx context.Context) error {
	if t.platform != nil {
		if err := t.platform.Stop(ctx); err != nil {
			t.logger.Error("Failed to stop Temporal platform", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
		t.logger.Info("Temporal platform stopped")
	}
	return nil
}

// Health checks the Temporal platform health
func (t *TemporalComponent) Health(ctx context.Context) error {
	if t.platform == nil {
		return fmt.Errorf("Temporal platform not initialized")
	}

	// Use the platform's built-in health check through client manager
	client := t.platform.GetClient()
	if client == nil {
		return fmt.Errorf("Temporal client not available")
	}

	// Use WorkflowService to check health (similar to platform's CheckConnection)
	_, err := client.WorkflowService().GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		return fmt.Errorf("Temporal health check failed: %w", err)
	}

	return nil
}

// GetPlatform returns the Temporal platform
func (t *TemporalComponent) GetPlatform() *temporal.Platform {
	return t.platform
}
