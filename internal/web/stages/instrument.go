package stages

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"awo.so/internal/pipeline"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	uimetrics "awo.so/internal/web/metrics"
	"awo.so/internal/web/ui"
)

// ─── InstrumentedStage ────────────────────────────────────────────────────────
//
// InstrumentedStage wraps any pipeline.Stage and adds:
//   - An OTel trace span per Execute call (name: "ui.stage.<stage_name>")
//   - Stage duration recorded via ui_stage_execution_duration_ms histogram
//   - Failure counters for specific stages (validate, registry)
//   - Structured log on stage failure (error level) and slow execution (warn level)
//
// Nil tracer, metrics, or log are handled gracefully — the observation simply
// does not occur for that dimension.
//
// InstrumentedStage is transparent: all Stage interface methods delegate to
// the inner stage unchanged. Only Execute is intercepted.

// UIStageAttributes is the canonical set of OTel span attributes emitted per stage.
type UIStageAttributes struct {
	StageName    string
	Route        string
	TenantID     string
	OperationKey string
	CacheHit     bool
	ASTCompiled  bool
}

func (a UIStageAttributes) toKeyValues() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("ui.stage", a.StageName),
		attribute.String("ui.route", a.Route),
		attribute.String("ui.tenant_id", a.TenantID),
		attribute.String("ui.operation_key", a.OperationKey),
		attribute.Bool("ui.cache_hit", a.CacheHit),
		attribute.Bool("ui.ast_compiled", a.ASTCompiled),
	}
}

// slowStageThreshold is the duration above which a stage execution is logged
// at WARN level. Tuned for the expected P99 of the Casbin (authz) stage.
const slowStageThreshold = 50 * time.Millisecond

// InstrumentedStage wraps a Stage with tracing, metrics, and logging.
type InstrumentedStage struct {
	inner   pipeline.Stage
	tracer  tracing.Service
	metrics metrics.MetricsProvider
	log     logger.Logger
}

// Instrument wraps s with observability instrumentation.
// If all three of tracer, mp, log are nil the original stage is returned unwrapped.
func Instrument(s pipeline.Stage, tracer tracing.Service, mp metrics.MetricsProvider, log logger.Logger) pipeline.Stage {
	if tracer == nil && mp == nil && log == nil {
		return s
	}
	return &InstrumentedStage{
		inner:   s,
		tracer:  tracer,
		metrics: mp,
		log:     log,
	}
}

// ── Stage interface delegation ────────────────────────────────────────────────

func (w *InstrumentedStage) Name() string         { return w.inner.Name() }
func (w *InstrumentedStage) Operations() []string { return w.inner.Operations() }
func (w *InstrumentedStage) FeatureFlag() string  { return w.inner.FeatureFlag() }
func (w *InstrumentedStage) Priority() int        { return w.inner.Priority() }
func (w *InstrumentedStage) Required() bool       { return w.inner.Required() }
func (w *InstrumentedStage) RunCondition() string { return w.inner.RunCondition() }
func (w *InstrumentedStage) DependsOn() []string  { return w.inner.DependsOn() }

// Execute intercepts the inner stage's Execute to add tracing, metrics, and logging.
func (w *InstrumentedStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	start := time.Now()

	// ── Collect context for attributes ───────────────────────────────────
	route := ""
	if input, ok := opCtx.Input.(ui.UISchemaInput); ok {
		route = input.Route
	}
	tenantID := opCtx.TenantID.String()
	cacheHit, _ := opCtx.Data[ui.DataKeyCacheHit].(bool)
	astCompiled, _ := opCtx.Data[ui.DataKeyASTCompiled].(bool)

	attrs := UIStageAttributes{
		StageName:    w.inner.Name(),
		Route:        route,
		TenantID:     tenantID,
		OperationKey: opCtx.OperationKey,
		CacheHit:     cacheHit,
		ASTCompiled:  astCompiled,
	}

	// ── Start trace span ─────────────────────────────────────────────────
	ctx := opCtx.Ctx
	var span tracing.Span
	if w.tracer != nil {
		ctx, span = w.tracer.StartSpan(ctx, "ui.stage."+w.inner.Name(),
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(attrs.toKeyValues()...),
		)
		opCtx.Ctx = ctx
		defer span.End()
	}

	// ── Execute inner stage ───────────────────────────────────────────────
	result, err := w.inner.Execute(opCtx)
	dur := time.Since(start)
	durMs := float64(dur.Milliseconds())

	status := result.Status
	if err != nil {
		status = "failed"
	}

	// ── Record stage duration histogram ──────────────────────────────────
	if w.metrics != nil {
		w.metrics.ObserveHistogram(uimetrics.MetricStageDurationMs, durMs, metrics.Fields{
			"stage":     w.inner.Name(),
			"route":     route,
			"tenant_id": tenantID,
			"status":    status,
		})
	}

	// ── Error path ───────────────────────────────────────────────────────
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		w.emitFailureCounter(route, tenantID, err)

		if w.log != nil {
			w.log.ErrorContext(ctx, "ui stage failed: "+w.inner.Name(), logger.Fields{
				"stage":     w.inner.Name(),
				"route":     route,
				"tenant_id": tenantID,
				"error":     err.Error(),
				"duration":  dur.String(),
			})
		}
		return result, err
	}

	// ── Success path ─────────────────────────────────────────────────────
	if span != nil {
		span.SetStatus(codes.Ok, "")
		if result.Status == "skipped" {
			span.SetAttributes(attribute.String("ui.stage.skip_reason", result.Message))
		}
	}

	// Warn on unexpectedly slow stages.
	if dur > slowStageThreshold && w.log != nil {
		w.log.WarnContext(ctx, "slow ui stage: "+w.inner.Name(), logger.Fields{
			"stage":     w.inner.Name(),
			"route":     route,
			"tenant_id": tenantID,
			"duration":  dur.String(),
			"threshold": slowStageThreshold.String(),
		})
	}

	return result, nil
}

// emitFailureCounter increments the stage-specific failure counter.
func (w *InstrumentedStage) emitFailureCounter(route, tenantID string, err error) {
	if w.metrics == nil {
		return
	}
	switch w.inner.Name() {
	case "ui.validate":
		rule := "unknown"
		if be, ok := err.(*sharedErrors.BusinessError); ok {
			rule = be.Code
		}
		w.metrics.IncrementCounter(uimetrics.MetricSchemaValidationFailures, metrics.Fields{
			"route":     route,
			"tenant_id": tenantID,
			"rule":      rule,
		})
	case "ui.registry":
		w.metrics.IncrementCounter(uimetrics.MetricRegistryResolutionFailures, metrics.Fields{
			"route":     route,
			"tenant_id": tenantID,
		})
	}
}

var _ pipeline.Stage = (*InstrumentedStage)(nil)
