package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// HealthChecker provides comprehensive health check functionality for dependencies
// NOTE: This service validates critical system dependencies including database and cache
// TODO: Add external service health checks (auth providers, notification services)
type HealthChecker interface {
	CheckDatabase(ctx context.Context) HealthCheckResult
	CheckCache(ctx context.Context) HealthCheckResult
	CheckDependencies(ctx context.Context) map[string]HealthCheckResult
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    string                 `json:"status"` // "ok", "warning", "critical"
	Message   string                 `json:"message,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// healthChecker implements HealthChecker with comprehensive dependency validation
type healthChecker struct {
	store   db.Store
	redis   redis.Cmdable
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// Health check configuration
	dbTimeout    time.Duration
	cacheTimeout time.Duration
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(
	store db.Store,
	redis redis.Cmdable,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) HealthChecker {
	return &healthChecker{
		store:        store,
		redis:        redis,
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
		dbTimeout:    5 * time.Second,
		cacheTimeout: 3 * time.Second,
	}
}

// CheckDatabase performs comprehensive database health checks
func (h *healthChecker) CheckDatabase(ctx context.Context) HealthCheckResult {
	ctx, span := h.tracer.StartSpan(ctx, "health.check_database")
	defer span.End()

	start := time.Now()
	result := HealthCheckResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Create timeout context for database operations
	dbCtx, cancel := context.WithTimeout(ctx, h.dbTimeout)
	defer cancel()

	// Test database connection with ping
	if err := h.store.GetPool().Ping(dbCtx); err != nil {
		duration := time.Since(start)
		result.Status = "critical"
		result.Message = fmt.Sprintf("Database ping failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "failed"

		h.metrics.IncrementCounter("health_check_failures",
			metrics.Fields{"component": "database", "type": "connection"})
		h.logger.ErrorContext(ctx, "Database health check failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	// Test database query execution with a simple query
	// Use a simple SELECT 1 query to test database connectivity
	_, err := h.store.GetPool().Query(dbCtx, "SELECT 1")
	if err != nil {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = fmt.Sprintf("Database query test failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "ok"
		result.Details["query_execution"] = "failed"

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "database", "type": "query"})
		h.logger.WarnContext(ctx, "Database query test failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	// Test tenant context functionality
	testTenantID := uuid.New()
	err = h.store.WithTenant(dbCtx, testTenantID, func(ctx context.Context, store db.Store) error {
		// Simple test within tenant context
		return nil
	})
	if err != nil {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = fmt.Sprintf("Database tenant context test failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "ok"
		result.Details["query_execution"] = "ok"
		result.Details["tenant_context"] = "failed"

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "database", "type": "tenant_context"})
		h.logger.WarnContext(ctx, "Database tenant context test failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	// All database checks passed
	duration := time.Since(start)
	result.Status = "ok"
	result.Message = "Database is healthy"
	result.Duration = duration
	result.Details["connection"] = "ok"
	result.Details["query_execution"] = "ok"
	result.Details["tenant_context"] = "ok"

	h.metrics.IncrementCounter("health_check_success",
		metrics.Fields{"component": "database"})
	h.metrics.ObserveHistogram("health_check_duration_seconds", duration.Seconds(),
		metrics.Fields{"component": "database"})

	h.logger.DebugContext(ctx, "Database health check successful",
		logger.Fields{"duration": duration})

	return result
}

// CheckCache performs comprehensive Redis cache health checks
func (h *healthChecker) CheckCache(ctx context.Context) HealthCheckResult {
	ctx, span := h.tracer.StartSpan(ctx, "health.check_cache")
	defer span.End()

	start := time.Now()
	result := HealthCheckResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	if h.redis == nil {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = "Redis cache not configured"
		result.Duration = duration
		result.Details["configured"] = false

		h.logger.WarnContext(ctx, "Redis cache not configured for health check")
		return result
	}

	// Create timeout context for cache operations
	cacheCtx, cancel := context.WithTimeout(ctx, h.cacheTimeout)
	defer cancel()

	// Test cache connection with ping
	pong, err := h.redis.Ping(cacheCtx).Result()
	if err != nil {
		duration := time.Since(start)
		result.Status = "critical"
		result.Message = fmt.Sprintf("Redis ping failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "failed"
		result.Details["configured"] = true

		h.metrics.IncrementCounter("health_check_failures",
			metrics.Fields{"component": "cache", "type": "connection"})
		h.logger.ErrorContext(ctx, "Redis health check failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	if pong != "PONG" {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = fmt.Sprintf("Redis ping returned unexpected response: %s", pong)
		result.Duration = duration
		result.Details["ping_response"] = pong
		result.Details["connection"] = "ok"
		result.Details["configured"] = true

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "cache", "type": "ping_response"})
		h.logger.WarnContext(ctx, "Redis ping returned unexpected response",
			logger.Fields{"response": pong, "duration": duration})

		return result
	}

	// Test cache operations with a test key
	testKey := fmt.Sprintf("health_check:%s", uuid.New().String())
	testValue := "health_check_value"

	// Test SET operation
	err = h.redis.Set(cacheCtx, testKey, testValue, 10*time.Second).Err()
	if err != nil {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = fmt.Sprintf("Redis SET operation failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "ok"
		result.Details["set_operation"] = "failed"
		result.Details["configured"] = true

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "cache", "type": "set_operation"})
		h.logger.WarnContext(ctx, "Redis SET operation failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	// Test GET operation
	retrievedValue, err := h.redis.Get(cacheCtx, testKey).Result()
	if err != nil {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = fmt.Sprintf("Redis GET operation failed: %v", err)
		result.Duration = duration
		result.Details["error"] = err.Error()
		result.Details["connection"] = "ok"
		result.Details["set_operation"] = "ok"
		result.Details["get_operation"] = "failed"
		result.Details["configured"] = true

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "cache", "type": "get_operation"})
		h.logger.WarnContext(ctx, "Redis GET operation failed",
			logger.Fields{"error": err.Error(), "duration": duration})

		return result
	}

	if retrievedValue != testValue {
		duration := time.Since(start)
		result.Status = "warning"
		result.Message = "Redis GET operation returned incorrect value"
		result.Duration = duration
		result.Details["expected_value"] = testValue
		result.Details["retrieved_value"] = retrievedValue
		result.Details["connection"] = "ok"
		result.Details["set_operation"] = "ok"
		result.Details["get_operation"] = "inconsistent"
		result.Details["configured"] = true

		h.metrics.IncrementCounter("health_check_warnings",
			metrics.Fields{"component": "cache", "type": "data_consistency"})
		h.logger.WarnContext(ctx, "Redis data consistency check failed",
			logger.Fields{"expected": testValue, "retrieved": retrievedValue, "duration": duration})

		return result
	}

	// Clean up test key
	h.redis.Del(cacheCtx, testKey)

	// All cache checks passed
	duration := time.Since(start)
	result.Status = "ok"
	result.Message = "Redis cache is healthy"
	result.Duration = duration
	result.Details["connection"] = "ok"
	result.Details["set_operation"] = "ok"
	result.Details["get_operation"] = "ok"
	result.Details["data_consistency"] = "ok"
	result.Details["configured"] = true

	h.metrics.IncrementCounter("health_check_success",
		metrics.Fields{"component": "cache"})
	h.metrics.ObserveHistogram("health_check_duration_seconds", duration.Seconds(),
		metrics.Fields{"component": "cache"})

	h.logger.DebugContext(ctx, "Redis cache health check successful",
		logger.Fields{"duration": duration})

	return result
}

// CheckDependencies performs health checks on all system dependencies
func (h *healthChecker) CheckDependencies(ctx context.Context) map[string]HealthCheckResult {
	ctx, span := h.tracer.StartSpan(ctx, "health.check_all_dependencies")
	defer span.End()

	results := make(map[string]HealthCheckResult)

	// Run database and cache checks concurrently
	dbResult := make(chan HealthCheckResult, 1)
	cacheResult := make(chan HealthCheckResult, 1)

	// Database check goroutine
	go func() {
		result := h.CheckDatabase(ctx)
		dbResult <- result
	}()

	// Cache check goroutine
	go func() {
		result := h.CheckCache(ctx)
		cacheResult <- result
	}()

	// Collect results
	results["database"] = <-dbResult
	results["cache"] = <-cacheResult

	// Log overall health status
	overallHealthy := true
	for component, result := range results {
		if result.Status == "critical" {
			overallHealthy = false
			h.logger.ErrorContext(ctx, "Critical health check failure",
				logger.Fields{"component": component, "message": result.Message})
		} else if result.Status == "warning" {
			h.logger.WarnContext(ctx, "Health check warning",
				logger.Fields{"component": component, "message": result.Message})
		}
	}

	if overallHealthy {
		h.logger.InfoContext(ctx, "All health checks passed successfully")
		h.metrics.IncrementCounter("health_check_success",
			metrics.Fields{"component": "overall"})
	} else {
		h.logger.ErrorContext(ctx, "One or more health checks failed")
		h.metrics.IncrementCounter("health_check_failures",
			metrics.Fields{"component": "overall"})
	}

	return results
}
