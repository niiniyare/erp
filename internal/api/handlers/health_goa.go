package handlers

import (
	"context"

	"github.com/niiniyare/erp/internal/api/gen/health"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// HealthGoaHandler implements the GOA health service
type HealthGoaHandler struct {
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewHealthGoaHandler creates a new GOA health handler
func NewHealthGoaHandler(tracing tracing.TracingService, metrics metrics.MetricsProvider) health.Service {
	return &HealthGoaHandler{
		tracing: tracing,
		metrics: metrics,
	}
}

// Health implements the health check endpoint
func (h *HealthGoaHandler) Health(ctx context.Context) (*health.HealthStatus, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "health.check",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("service.name", "awo"),
			attribute.String("service.version", "1.0.0"),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("health_check_duration", metrics.Fields{
		"endpoint": "health",
	})
	defer timer.Stop()

	// Log health check request
	logger.DebugContext(ctx, "Processing health check request")

	// Create health status response
	healthStatus := &health.HealthStatus{
		Status:  "ok",
		Service: "awo",
		Version: "1.0.0",
	}

	// Success metrics
	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "health",
		"status":   "success",
	})

	// Log success
	logger.DebugContext(ctx, "Health check completed successfully",
		logger.Fields{
			"status":  healthStatus.Status,
			"service": healthStatus.Service,
			"version": healthStatus.Version,
		})

	return healthStatus, nil
}

// Ready implements the readiness check endpoint
func (h *HealthGoaHandler) Ready(ctx context.Context) (*health.ReadinessStatus, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "health.ready",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("service.name", "awo"),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("readiness_check_duration", metrics.Fields{
		"endpoint": "ready",
	})
	defer timer.Stop()

	// Log readiness check request
	logger.DebugContext(ctx, "Processing readiness check request")

	// TODO: Add actual readiness checks (database, cache, etc.)
	// For now, we'll return a simple "ready" status matching the Gin handler
	checks := &health.HealthChecks{
		Database: "ok",
		Cache:    "ok",
	}

	readinessStatus := &health.ReadinessStatus{
		Status: "ready",
		Checks: checks,
	}

	// Success metrics
	h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
		"endpoint": "ready",
		"status":   "success",
	})

	// Log success
	logger.DebugContext(ctx, "Readiness check completed successfully",
		logger.Fields{
			"status":           readinessStatus.Status,
			"database_status":  checks.Database,
			"cache_status":     checks.Cache,
		})

	return readinessStatus, nil
}