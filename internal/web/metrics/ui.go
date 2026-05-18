// Package metrics defines the canonical UI pipeline metric names and registers
// histograms with appropriate buckets via the shared MetricsProvider.
//
// Call RegisterUIMetrics(mp) once at startup (in NewUIPipeline) so histograms
// have precise bucket boundaries. Counters auto-register on first increment.
package metrics

import (
	"awo.so/internal/shared/metrics"
)

// ─── METRIC NAMES ────────────────────────────────────────────────────────────
//
// All UI pipeline metrics share the "ui_" prefix to avoid collisions with
// business-domain metrics. Use these constants everywhere — never inline strings.

const (
	// MetricCompileDurationMs measures end-to-end schema compilation latency
	// (registry lookup → compile → normalize → validate) in milliseconds.
	// Labels: route, tenant_id, cache_hit, ast_compiled.
	MetricCompileDurationMs = "ui_compile_duration_ms"

	// MetricStageDurationMs measures per-stage execution latency in milliseconds.
	// Labels: stage, route, tenant_id, status ("completed"/"skipped"/"failed").
	MetricStageDurationMs = "ui_stage_execution_duration_ms"

	// MetricSchemaValidationFailures counts schema validation rejections from
	// ValidateStage. Each increment represents a programmer error in a PageFn.
	// Labels: route, tenant_id, rule (VALIDATE_* code).
	MetricSchemaValidationFailures = "ui_schema_validation_failures_total"

	// MetricCacheGenerationMismatch counts cache misses caused by a schema or
	// policy generation increment — i.e. forced soft-invalidations.
	// Labels: route, tenant_id, reason ("policy_gen"/"schema_gen").
	MetricCacheGenerationMismatch = "ui_cache_generation_mismatch_total"

	// MetricInvalidationEvents counts cache invalidation calls dispatched via
	// InvalidateSchemaCache, broken down by scope.
	// Labels: tenant_id, scope ("tenant"/"module"/"policy"/"hard").
	MetricInvalidationEvents = "ui_invalidation_events_total"

	// MetricRegistryResolutionFailures counts route-not-found responses from
	// RegistryStage. Increments indicate a missing page registration or a bad
	// client-side route reference.
	// Labels: route, tenant_id.
	MetricRegistryResolutionFailures = "ui_registry_resolution_failures_total"
)

// ─── HISTOGRAM BUCKETS ───────────────────────────────────────────────────────

// compileDurationBuckets are appropriate for end-to-end schema compilation.
// Most cache hits complete in < 5 ms; full compiles typically take 5–50 ms.
var compileDurationBuckets = []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}

// stageDurationBuckets are appropriate for individual stage execution.
// Most stages complete in < 2 ms; the Authz (Casbin) stage may take 5–20 ms.
var stageDurationBuckets = []float64{0.5, 1, 2, 5, 10, 20, 50, 100, 200}

// ─── REGISTRATION ─────────────────────────────────────────────────────────────

// RegisterUIMetrics pre-registers UI pipeline histograms with bucket boundaries.
// Call once at startup before any stage executes to ensure Prometheus exports
// all buckets even before traffic arrives.
//
// Safe to call with a nil MetricsProvider — silently returns without panicking.
func RegisterUIMetrics(mp metrics.MetricsProvider) {
	if mp == nil {
		return
	}

	mp.Histogram(
		MetricCompileDurationMs,
		"End-to-end UI schema compilation latency in milliseconds",
		compileDurationBuckets,
		"route", "tenant_id", "cache_hit", "ast_compiled",
	)

	mp.Histogram(
		MetricStageDurationMs,
		"Per-stage execution latency in milliseconds",
		stageDurationBuckets,
		"stage", "route", "tenant_id", "status",
	)
}
