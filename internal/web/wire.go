// Package web wires the UI pipeline: stages → registry → pipeline builder.
//
// Call NewUIPipeline at application startup to get a configured *UIPipeline
// ready to serve AMIS schema requests. Pass it to handler.NewSchemaHandler.
//
// Dependency graph:
//
//	cache.Service  ──────────────────────────────────────────────────────┐
//	authz.UIAuthzService  ──────────────────────────────────────────────┐ │
//	                                                                     ↓ ↓
//	SessionStage(10) → AuthzStage(20) → CacheLookup(30) → Registry(40)
//	  → Compile(50) → Normalize(60) → Validate(70) → CacheStore(80) → Response(90)
//
// Observability:
//   - Pass non-nil tracer/metrics/log to wrap every stage with InstrumentedStage.
//   - Pass nil to skip (e.g. in tests).
package web

import (
	"fmt"

	"awo.so/internal/pipeline"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"awo.so/internal/web/authz"
	uicache "awo.so/internal/web/cache"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/stages"
	"awo.so/internal/web/ui"
)

// NewUIPipeline builds a StageRegistry populated with all UI pipeline stages
// and returns a PipelineBuilder ready for use by SchemaHandler.
//
// Parameters:
//   - authzSvc: resolves permissions per request via Casbin (required)
//   - cacheSvc: tenant-aware Redis+memory cache (required; pass nil to disable caching)
//   - versions: generation-aware cache key components (pass uicache.DefaultVersions()
//               in tests/dev; production must inject real values from config/env)
//   - tracer:   OTel tracing service (optional; nil = no tracing)
//   - mp:       metrics provider (optional; nil = no metrics)
//   - log:      structured logger (optional; nil = no logging)
func NewUIPipeline(
	authzSvc authz.UIAuthzService,
	cacheSvc cache.Service,
	versions uicache.CacheVersions,
	tracer tracing.Service,
	mp metrics.MetricsProvider,
	log logger.Logger,
) *UIPipeline {
	reg := pipeline.NewStageRegistry()

	instrument := func(s pipeline.Stage) pipeline.Stage {
		if tracer == nil && mp == nil && log == nil {
			return s
		}
		return stages.Instrument(s, tracer, mp, log)
	}

	reg.Register(
		instrument(stages.NewSessionStage()),
		instrument(stages.NewAuthzStage(authzSvc)),
		instrument(stages.NewRegistryStage()),
		instrument(stages.NewCompileStage()),
		instrument(stages.NewNormalizeStage()),
		instrument(stages.NewValidateStage()),
		instrument(stages.NewResponseStage()),
	)

	// Cache stages are optional — omit if cacheSvc is nil.
	if cacheSvc != nil {
		reg.Register(
			instrument(stages.NewCacheLookupStage(cacheSvc)),
			instrument(stages.NewCacheStoreStage(cacheSvc)),
		)
	}

	// Validate the page registry — all RegisterPage() calls must have complete metadata.
	// Panics if any registration is missing Module, Title, or both Fn and ASTFn.
	if err := registry.ValidateRegistry(); err != nil {
		panic(fmt.Sprintf("UI page registry invalid: %v", err))
	}

	// Validate the stage dependency graph at startup.
	// Panics if any stage declares a dependency on an unregistered stage or
	// if a cycle exists — an invalid DAG must never reach production.
	if err := reg.ValidateDAG(ui.OperationKey); err != nil {
		panic(fmt.Sprintf("UI pipeline DAG invalid: %v", err))
	}

	pb := pipeline.NewPipelineBuilder(reg, nil) // nil txRunner: UI has no DB transactions
	return &UIPipeline{builder: pb, versions: versions}
}

// UIPipeline wraps PipelineBuilder to automatically inject CacheVersions into
// every OperationContext before Run() is called. This keeps the versions
// concern out of SchemaHandler — the handler just calls Run as before.
type UIPipeline struct {
	builder  *pipeline.PipelineBuilder
	versions uicache.CacheVersions
}

// Run pre-populates DataKeyCacheVersions then delegates to the underlying builder.
func (p *UIPipeline) Run(opCtx *pipeline.OperationContext) error {
	opCtx.Data[ui.DataKeyCacheVersions] = p.versions
	return p.builder.Run(opCtx)
}

// Builder exposes the underlying PipelineBuilder for callers that need it
// (e.g. RunFrom for suspended pipeline resumption).
func (p *UIPipeline) Builder() *pipeline.PipelineBuilder {
	return p.builder
}
