package main

import (
	"context"

	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type Infrastructure struct {
	Config   *config.Config
	Tracing  tracing.TracingService
	Metrics  *metrics.MetricsService
	Logger   logger.Logger
	Temporal *temporal.Platform
}

func InitializeInfrastructure() (*Infrastructure, error) {
	// Initialize logger from environment
	if err := logger.InitializeFromEnv(); err != nil {
		return nil, err
	}
	log := logger.WithFields(logger.Fields{})

	// Load configuration
	cfg := config.Load()
	log.Info("Configuration loaded", logger.Fields{
		"server_port": cfg.Server.Port,
		"db_host":     cfg.Database.Host,
		"redis_host":  cfg.Redis.Host,
	})

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

	// Initialize Temporal platform
	temporalPlatform, err := temporal.NewPlatform(&cfg.Temporal, log)
	if err != nil {
		return nil, err
	}

	return &Infrastructure{
		Config:   cfg,
		Tracing:  tracingService,
		Metrics:  metricsService,
		Logger:   log,
		Temporal: temporalPlatform,
	}, nil
}

func (i *Infrastructure) Shutdown(ctx context.Context) error {
	// Shutdown Temporal platform first
	if i.Temporal != nil {
		if err := i.Temporal.Stop(ctx); err != nil {
			i.Logger.Error("Error stopping Temporal platform", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	// Shutdown tracing
	return i.Tracing.Shutdown(ctx)
}
