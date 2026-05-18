package stages

// ─── TASK 8b — OBSERVABILITY WRAPPER ─────────────────────────────────────────
//
// InstrumentedStage wraps any pipeline.Stage and emits:
//   - OpenTelemetry trace spans (one per stage)
//   - Prometheus/OTel metrics (duration histogram, failure counter)
//   - Structured log entries via logger.Logger
//
// This satisfies PHASE 8 — Observability without modifying individual stage code.
// Stages focus on logic; InstrumentedStage adds the observability envelope.
//
// DESCRIPTION:
// Wraps stage.Execute() with span start/end, attribute tagging, and metric
// recording. On error, marks span as failed and increments error counter.
// On success, records duration in histogram.
//
// WHY:
// Per-stage observability enables:
//   - Identifying which specific stage is slow (AuthzStage vs CompileStage)
//   - Cache hit ratio tracking
//   - IAM resolution latency separate from schema compile latency
//   - Distributed trace correlation across service boundaries
//
// USAGE:
// Wrap stages at wire time:
//   registry.Register(stages.Instrument(stages.NewAuthzStage(svc), tracer, metrics, log))

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"awo.so/internal/pipeline"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// InstrumentedStage wraps a Stage with tracing, metrics, and logging.
type InstrumentedStage struct {
	inner   pipeline.Stage
	tracer  tracing.Service
	metrics metrics.MetricsProvider
	log     logger.Logger
}

// Instrument wraps a stage with observability. All three providers are optional;
// nil providers are skipped gracefully.
func Instrument(
	stage pipeline.Stage,
	tracer tracing.Service,
	mp metrics.MetricsProvider,
	log logger.Logger,
) pipeline.Stage {
	return &InstrumentedStage{
		inner:   stage,
		tracer:  tracer,
		metrics: mp,
		log:     log,
	}
}

// ── Stage interface delegation ────────────────────────────────────────────────

func (s *InstrumentedStage) Name() string         { return s.inner.Name() }
func (s *InstrumentedStage) Operations() []string { return s.inner.Operations() }
func (s *InstrumentedStage) FeatureFlag() string  { return s.inner.FeatureFlag() }
func (s *InstrumentedStage) Priority() int        { return s.inner.Priority() }
func (s *InstrumentedStage) Required() bool       { return s.inner.Required() }
func (s *InstrumentedStage) RunCondition() string { return s.inner.RunCondition() }
func (s *InstrumentedStage) DependsOn() []string  { return s.inner.DependsOn() }

// Execute wraps the inner Execute with span, metrics, and log.
func (s *InstrumentedStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	stageName := s.inner.Name()
	started := time.Now()

	// ── Tracing ──────────────────────────────────────────────────────────────
	var spanCtx = opCtx.Ctx
	var span tracing.Span
	if s.tracer != nil {
		spanCtx, span = s.tracer.StartSpan(opCtx.Ctx, "ui.pipeline.stage."+stageName)
		defer span.End()
		span.SetAttributes(
			attribute.String("ui.stage", stageName),
			attribute.String("ui.tenant_id", opCtx.TenantID.String()),
			attribute.String("ui.operation", opCtx.OperationKey),
		)
		opCtx.Ctx = spanCtx
	}

	// ── Execute inner stage ───────────────────────────────────────────────────
	result, err := s.inner.Execute(opCtx)

	duration := time.Since(started)

	// ── Metrics ───────────────────────────────────────────────────────────────
	if s.metrics != nil {
		labels := metrics.Fields{
			"stage":     stageName,
			"tenant_id": opCtx.TenantID.String(),
			"status":    statusLabel(result, err),
		}
		s.metrics.ObserveHistogram("ui_stage_duration_ms", float64(duration.Milliseconds()), labels)

		if err != nil {
			s.metrics.IncrementCounter("ui_stage_failure_total", labels)
		}

		// Cache-specific metrics
		if stageName == "ui.cache_lookup" {
			if hit, _ := opCtx.Data["ui.cache.hit"].(bool); hit {
				s.metrics.IncrementCounter("ui_cache_hit_total", metrics.Fields{
					"tenant_id": opCtx.TenantID.String(),
				})
			} else {
				s.metrics.IncrementCounter("ui_cache_miss_total", metrics.Fields{
					"tenant_id": opCtx.TenantID.String(),
				})
			}
		}
	}

	// ── Tracing error recording ───────────────────────────────────────────────
	if s.tracer != nil && span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.SetAttributes(
			attribute.Int64("ui.stage.duration_ms", duration.Milliseconds()),
			attribute.Bool("ui.cache_hit", boolData(opCtx, "ui.cache.hit")),
		)
	}

	// ── Logging ───────────────────────────────────────────────────────────────
	if s.log != nil {
		fields := logger.Fields{
			"stage":       stageName,
			"duration_ms": duration.Milliseconds(),
			"status":      statusLabel(result, err),
			"tenant_id":   opCtx.TenantID.String(),
			"operation":   opCtx.OperationKey,
		}
		if err != nil {
			fields["error"] = err.Error()
			s.log.ErrorContext(opCtx.Ctx, fmt.Sprintf("ui stage failed: %s", stageName), fields)
		} else {
			s.log.InfoContext(opCtx.Ctx, fmt.Sprintf("ui stage completed: %s", stageName), fields)
		}
	}

	return result, err
}

func statusLabel(result pipeline.StageResult, err error) string {
	if err != nil {
		return "failed"
	}
	if result.Status != "" {
		return result.Status
	}
	return "completed"
}

func boolData(opCtx *pipeline.OperationContext, key string) bool {
	v, _ := opCtx.Data[key].(bool)
	return v
}
