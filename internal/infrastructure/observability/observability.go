package observability

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ObservabilityComponent manages logging, metrics, and tracing
type ObservabilityComponent struct {
	loggerConfig   *config.LoggerConfig
	appConfig      *config.AppConfig
	tracingService tracing.TracingService
	metricsService *metrics.MetricsService
	logger         logger.Logger
}

// NewObservability creates a new observability component
func NewObservability(loggerConfig *config.LoggerConfig, appConfig *config.AppConfig) *ObservabilityComponent {
	return &ObservabilityComponent{
		loggerConfig: loggerConfig,
		appConfig:    appConfig,
	}
}

// Start initializes all observability systems
func (o *ObservabilityComponent) Start(ctx context.Context) error {
	// Initialize logger first (from environment for now, will migrate later)
	if err := logger.InitializeFromEnv(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	o.logger = logger.WithFields(logger.Fields{"component": "observability"})

	o.logger.Info("Starting observability systems", logger.Fields{
		"service":     o.appConfig.Name,
		"version":     o.appConfig.Version,
		"environment": o.appConfig.Environment,
	})

	// Initialize tracing
	tracingService, err := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    o.appConfig.Name,
		ServiceVersion: o.appConfig.Version,
		Environment:    o.appConfig.Environment,
		ExporterType:   tracing.StdoutExporter,
		SamplingRatio:  1.0,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}
	o.tracingService = tracingService

	// Initialize metrics
	metricsService, err := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: o.appConfig.Environment,
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}
	o.metricsService = metricsService

	o.logger.Info("Observability systems started successfully", logger.Fields{
		"tracing_enabled": true,
		"metrics_enabled": true,
		"logging_enabled": true,
	})

	return nil
}

// Stop gracefully shuts down all observability systems
func (o *ObservabilityComponent) Stop(ctx context.Context) error {
	if o.tracingService != nil {
		if err := o.tracingService.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracing: %w", err)
		}
	}

	if o.logger != nil {
		o.logger.Info("Observability systems stopped")
	}

	return nil
}

// Health checks the observability systems health
func (o *ObservabilityComponent) Health(ctx context.Context) error {
	// Basic health check - ensure all services are initialized
	if o.tracingService == nil {
		return fmt.Errorf("tracing service not initialized")
	}

	if o.metricsService == nil {
		return fmt.Errorf("metrics service not initialized")
	}

	if o.logger == nil {
		return fmt.Errorf("logger not initialized")
	}

	return nil
}

// GetTracing returns the tracing service
func (o *ObservabilityComponent) GetTracing() tracing.TracingService {
	return o.tracingService
}

// GetMetrics returns the metrics service
func (o *ObservabilityComponent) GetMetrics() *metrics.MetricsService {
	return o.metricsService
}

// GetLogger returns the logger
func (o *ObservabilityComponent) GetLogger() logger.Logger {
	return o.logger
}
