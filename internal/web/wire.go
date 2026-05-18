// Package web wires the UI pipeline: stages → registry → pipeline builder.
//
// Call NewUIPipeline at application startup to get a configured *pipeline.PipelineBuilder
// ready to serve AMIS schema requests. Pass the builder to handler.NewSchemaHandler.
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
	"awo.so/internal/pipeline"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"awo.so/internal/web/authz"
	"awo.so/internal/web/stages"
)

// NewUIPipeline builds a StageRegistry populated with all UI pipeline stages
// and returns a PipelineBuilder ready for use by SchemaHandler.
//
// Parameters:
//   - authzSvc: resolves permissions per request via Casbin (required)
//   - cacheSvc: tenant-aware Redis+memory cache (required; pass nil to disable caching)
//   - tracer:   OTel tracing service (optional; nil = no tracing)
//   - mp:       metrics provider (optional; nil = no metrics)
//   - log:      structured logger (optional; nil = no logging)
func NewUIPipeline(
	authzSvc authz.UIAuthzService,
	cacheSvc cache.Service,
	tracer tracing.Service,
	mp metrics.MetricsProvider,
	log logger.Logger,
) *pipeline.PipelineBuilder {
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

	return pipeline.NewPipelineBuilder(reg, nil) // nil txRunner: UI has no DB transactions
}
