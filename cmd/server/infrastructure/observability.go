package infrastructure

import (
	"context"

	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type Observability struct {
	Tracing tracing.TracingService
	Metrics *metrics.MetricsService
	Logger  logger.Logger
}

func InitializeObservability(cfg *config.Config) (*Observability, error) {
	// Initialize logger from environment
	if err := logger.InitializeFromEnv(); err != nil {
		return nil, err
	}
	log := logger.WithFields(logger.Fields{})

	log.Info("Starting Awo ERP server", logger.Fields{
		"service": cfg.App.Name,
		"version": cfg.App.Version,
	})

	// Initialize tracing
	tracingService, err := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    cfg.App.Name,
		ServiceVersion: cfg.App.Version,
		Environment:    cfg.App.Environment,
		ExporterType:   tracing.StdoutExporter,
		SamplingRatio:  1.0,
	})
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	metricsService, err := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: cfg.App.Environment,
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		return nil, err
	}

	return &Observability{
		Tracing: tracingService,
		Metrics: metricsService,
		Logger:  log,
	}, nil
}

func (o *Observability) Shutdown(ctx context.Context) error {
	return o.Tracing.Shutdown(ctx)
}

// GetLogger returns the logger
func (o *Observability) GetLogger() logger.Logger {
	return o.Logger
}

// GetMetrics returns the metrics service
func (o *Observability) GetMetrics() *metrics.MetricsService {
	return o.Metrics
}

// GetTracing returns the tracing service
func (o *Observability) GetTracing() tracing.TracingService {
	return o.Tracing
}
