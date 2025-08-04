package main

import (
	"context"

	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type Infrastructure struct {
	Config  *config.Config
	Tracing tracing.TracingService
	Metrics *metrics.MetricsService
}

func InitializeInfrastructure() (*Infrastructure, error) {
	// Initialize logger from environment
	if err := logger.InitializeFromEnv(); err != nil {
		return nil, err
	}

	logger.Info("Starting Awo ERP server", logger.Fields{
		"service": "awo-server",
		"version": "1.0.0",
	})

	// Load configuration
	cfg := config.Load()
	logger.Info("Configuration loaded", logger.Fields{
		"server_port": cfg.Server.Port,
		"db_host":     cfg.Database.Host,
		"redis_host":  cfg.Redis.Host,
	})

	// Initialize tracing
	tracingService, err := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    "awo-server",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		ExporterType:   tracing.StdoutExporter,
		SamplingRatio:  1.0,
	})
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	metricsService, err := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: "awo-erp",
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		return nil, err
	}

	return &Infrastructure{
		Config:  cfg,
		Tracing: tracingService,
		Metrics: metricsService,
	}, nil
}

func (i *Infrastructure) Shutdown(ctx context.Context) error {
	return i.Tracing.Shutdown(ctx)
}
