package ui

// OperationKey is the pipeline operation key for all UI schema compilation requests.
// All UI pipeline stages declare Operations() []string{OperationKey}.
const OperationKey = "ui.schema.compile"

// AppOperationKey is the pipeline operation key for the app shell (nav tree) compilation.
const AppOperationKey = "ui.app.compile"

// Priority bands for UI pipeline stages.
// These fit within the standard pipeline priority scheme but use a dedicated
// range (10–90) that does not conflict with business operation stages (100–999).
// UI compilation is a read-only, non-transactional pipeline — no TxHooks,
// no compensation, no DB writes.
const (
	// PrioritySession: extract and validate the IAM contract session context.
	PrioritySession = 10

	// PriorityAuthz: bulk-resolve permissions and feature flag fingerprinting.
	PriorityAuthz = 20

	// PriorityCache: cache lookup — must run after Authz (fingerprint must exist).
	PriorityCache = 30

	// PriorityRegistry: resolve PageFn from route. Skipped on cache hit.
	PriorityRegistry = 40

	// PriorityCompile: execute PageFn(UISessionContext). Skipped on cache hit.
	PriorityCompile = 50

	// PriorityNormalize: enforce AMIS compliance rules. Skipped on cache hit.
	PriorityNormalize = 60

	// PriorityValidate: reject forbidden expressions. Skipped on cache hit.
	PriorityValidate = 70

	// PriorityCacheStore: write compiled schema to cache. Skipped on cache hit.
	PriorityCacheStore = 80

	// PriorityResponse: assemble the final API response envelope. Always runs.
	PriorityResponse = 90
)

// Data keys used by stages to communicate via opCtx.Data.
// Convention: "ui.<stage_name>.<key>"
const (
	// DataKeyRoute is set by SchemaHandler before pipeline.Run().
	DataKeyRoute = "ui.request.route"

	// DataKeyPermissions is the resolved map[string]bool set by AuthzStage.
	DataKeyPermissions = "ui.authz.permissions"

	// DataKeyPermFingerprint is the stable permission hash set by AuthzStage.
	DataKeyPermFingerprint = "ui.authz.perm_fingerprint"

	// DataKeyFlagFingerprint is the stable feature flag hash set by AuthzStage.
	DataKeyFlagFingerprint = "ui.authz.flag_fingerprint"

	// DataKeyCacheKey is the full cache key computed by CacheStage.
	DataKeyCacheKey = "ui.cache.key"

	// DataKeyCacheHit is set true by CacheStage on a cache hit.
	DataKeyCacheHit = "ui.cache.hit"

	// DataKeyPageFn is the resolved PageFn set by RegistryStage (legacy path).
	DataKeyPageFn = "ui.registry.page_fn"

	// DataKeyASTPageFn is the resolved ASTPageFn set by RegistryStage (typed AST path).
	// CompileStage checks this key first; falls back to DataKeyPageFn when absent.
	DataKeyASTPageFn = "ui.registry.ast_page_fn"

	// DataKeyRouteParams is the map[string]string of URL params extracted by RegistryStage.
	// Example: pattern "/finance/invoices/:id" matched against "/finance/invoices/abc-123"
	// produces map{"id": "abc-123"}. Empty map when route has no param segments.
	DataKeyRouteParams = "ui.registry.route_params"

	// DataKeyASTCompiled is set true by CompileStage when the schema was produced
	// via ASTPageFn + CompileTree. NormalizeStage skips structural rules that are
	// already guaranteed by the typed AST (syncLocation, transparent bg) when this is true.
	DataKeyASTCompiled = "ui.compile.ast_compiled"

	// DataKeyCacheVersions holds the CacheVersions struct injected at startup.
	// Set by NewUIPipeline into opCtx before Run() — CacheLookupStage reads it
	// to build the generation-aware cache key.
	DataKeyCacheVersions = "ui.cache.versions"

	// DataKeySessionCtx is the UISessionContext set by AuthzStage.
	// Stored as ui.UISessionContext (value, not pointer — immutable).
	DataKeySessionCtx = "ui.authz.session_ctx"

	// DataKeySchema is the compiled Schema set by CompileStage and read by ResponseStage.
	DataKeySchema = "ui.compile.schema"

	// DataKeyResponse is the final API envelope set by ResponseStage.
	DataKeyResponse = "ui.response"
)

// UISchemaInput is the typed input placed in opCtx.Input before pipeline.Run().
type UISchemaInput struct {
	Route string // e.g. "/finance/invoices"
}

// UISchemaOutput is read from opCtx.Data[DataKeyResponse] after pipeline.Run().
type UISchemaOutput struct {
	Schema   Schema
	CacheHit bool
	Route    string
}
