package handlers

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/api/gen/health"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// HealthGoaHandler implements the GOA health service
// NOTE: Enhanced with dependency health checking
// TODO: Add external service health checks (notification services, etc.)
type HealthGoaHandler struct {
	healthChecker HealthChecker
	tracing       tracing.TracingService
	metrics       metrics.MetricsProvider
}

// NewHealthGoaHandler creates a new GOA health handler with health checking
func NewHealthGoaHandler(
	healthChecker HealthChecker,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
) health.Service {
	return &HealthGoaHandler{
		healthChecker: healthChecker,
		tracing:       tracing,
		metrics:       metrics,
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

// Ready implements the readiness check endpoint with dependency validation
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

	// Perform dependency health checks
	healthResults := h.healthChecker.CheckDependencies(ctx)

	// Convert health check results to API response format
	checks := &health.HealthChecks{
		Database: convertHealthStatus(healthResults["database"]),
		Cache:    convertHealthStatus(healthResults["cache"]),
	}

	// Determine overall readiness status
	overallStatus := "ready"
	hasWarnings := false
	hasCritical := false

	for component, result := range healthResults {
		if result.Status == "critical" {
			hasCritical = true
			// Add component failure details to span
			span.SetAttributes(
				attribute.String(fmt.Sprintf("health.%s.status", component), "critical"),
				attribute.String(fmt.Sprintf("health.%s.message", component), result.Message),
			)
		} else if result.Status == "warning" {
			hasWarnings = true
			span.SetAttributes(
				attribute.String(fmt.Sprintf("health.%s.status", component), "warning"),
				attribute.String(fmt.Sprintf("health.%s.message", component), result.Message),
			)
		}
	}

	// Set overall status based on critical failures
	if hasCritical {
		overallStatus = "not_ready"
		h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
			"endpoint": "ready",
			"status":   "critical",
		})
	} else if hasWarnings {
		overallStatus = "ready_with_warnings"
		h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
			"endpoint": "ready",
			"status":   "warning",
		})
	} else {
		h.metrics.IncrementCounter("health_checks_total", metrics.Fields{
			"endpoint": "ready",
			"status":   "success",
		})
	}

	readinessStatus := &health.ReadinessStatus{
		Status: overallStatus,
		Checks: checks,
	}

	// Log readiness check completion
	logger.InfoContext(ctx, "Readiness check completed",
		logger.Fields{
			"status":          readinessStatus.Status,
			"database_status": checks.Database,
			"cache_status":    checks.Cache,
			"has_warnings":    hasWarnings,
			"has_critical":    hasCritical,
		})

	// Return appropriate HTTP status based on health
	if hasCritical {
		// NOTE: In production, you might want to return an error here to signal unhealthy state
		// For now, we return success but with "not_ready" status for monitoring
		logger.WarnContext(ctx, "Service has critical health issues but returning success for monitoring")
	}

	return readinessStatus, nil
}

// convertHealthStatus converts internal health check status to API response status
func convertHealthStatus(result HealthCheckResult) string {
	switch result.Status {
	case "ok":
		return "ok"
	case "warning":
		return "degraded"
	case "critical":
		return "failed"
	default:
		return "unknown"
	}
}
