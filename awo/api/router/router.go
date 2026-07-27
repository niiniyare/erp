// Package router registers all auto-generated CRUD and action routes derived
// from the CompiledSchema onto a Fiber app.
//
// Route pattern (per entity, using EntitySchema.RoutePrefix):
//   - GET    /api/v1/{module}/{resource}          → List
//   - GET    /api/v1/{module}/{resource}/:id      → Get
//   - POST   /api/v1/{module}/{resource}          → Create
//   - PATCH  /api/v1/{module}/{resource}/:id      → Update
//   - DELETE /api/v1/{module}/{resource}/:id      → Delete
//   - POST   /api/v1/{module}/{resource}/:id/:action → Action handler
//
// Permission checks use EntitySchema.PermissionNamespace as the Casbin object.
// No entity name or path is constructed at runtime — all values come from
// the compiled EntitySchema.
package router

import (
	goredis "github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	pgxlib "github.com/jackc/pgx/v5/pgxpool"
	temporalclient "go.temporal.io/sdk/client"

	"awo.so/awo/api/authz"
	"awo.so/awo/api/handler"
	"awo.so/awo/api/middleware"
	"awo.so/awo/api/service"
	"awo.so/awo/audit"
	"awo.so/awo/auth"
	"awo.so/awo/cache"
	"awo.so/awo/compiler"
	contrib "awo.so/awo/contrib/pgx"
	contribredis "awo.so/awo/contrib/redis"
	"awo.so/awo/def"
	"awo.so/awo/driver"
	"awo.so/awo/runtime"
	"awo.so/awo/sdui"
)

// RegisterOptions carries dependencies needed to build handlers.
type RegisterOptions struct {
	Pool  *pgxlib.Pool
	Redis *goredis.Client
	// IAM provides session and API token validation. The interface type keeps
	// the router decoupled from the concrete *iam.AuthService implementation,
	// which is important for framework extraction readiness.
	IAM      middleware.SessionValidator                // required for RequireAuth
	Tenants  driver.EntityRepository[*def.EntityRecord] // required for TenantResolver
	Authz    auth.PolicyEvaluator                       // nil = RBAC disabled (dev/test)
	Temporal temporalclient.Client                      // nil = degraded mode (no workflow starts)
	// AuditWriter is the production audit implementation. When nil, auditing is
	// disabled and audit.NoopAuditWriter{} is used automatically. In production,
	// pass audit.NewTransactionalWriter(contrib.NewPoolQuerier(pool)).
	AuditWriter audit.AuditWriter

	// AuditSigningSecret is the HMAC-SHA256 key used to derive a non-reversible
	// SessionID for audit records. Set from the AUDIT_SIGNING_SECRET env var.
	// An empty string disables session correlation (acceptable in dev/test).
	AuditSigningSecret string

	// SDUIGenerator generates amis JSON page schemas from EntitySchema +
	// ViewerContext. When nil, the /api/sdui/* endpoints are not registered.
	SDUIGenerator *sdui.Generator
}

// Register mounts the full auto-generated API onto app under /api/v1/entities/.
// The API group has the full middleware pipeline applied (tenant→auth→ratelimit).
func Register(app *fiber.App, schema *compiler.CompiledSchema, opts RegisterOptions) {
	// Build a Redis-backed rate-limit counter.
	var counter cache.Counter
	if opts.Redis != nil {
		counter = contribredis.New(opts.Redis)
	} else {
		counter = cache.NoopCounter{}
	}

	// Shared middleware applied once at the /api/v1 group level.
	// Each entity group inherits these; individual route groups are mounted
	// directly at their RoutePrefix paths.
	api := app.Group("/api/v1")
	if opts.Tenants != nil {
		api.Use(middleware.TenantResolver(opts.Tenants))
	}
	if opts.IAM != nil {
		api.Use(middleware.RequireAuth(opts.IAM, opts.AuditSigningSecret))
	}
	api.Use(middleware.RateLimit(counter, middleware.DefaultRateLimit))

	// One pipeline shared across all entity handlers — it holds the compiled
	// schema and performs O(1) entity lookup by QualifiedName.
	// NewPipeline panics on nil auditWriter; guard here so callers don't need to.
	aw := opts.AuditWriter
	if aw == nil {
		aw = audit.NoopAuditWriter{}
	}
	pipeline := runtime.NewPipeline(schema, aw)

	for _, es := range schema.Entities {
		repo := contrib.NewRepository(opts.Pool, es)
		svc := service.NewEntityService(es, repo, pipeline, opts.Temporal)
		h := handler.NewEntityHandler(es, svc)

		// RoutePrefix is already the full path "/api/v1/{module}/{resource}".
		// Fiber groups interpret the path relative to the app, not the parent group,
		// when an absolute path is given. Use the compiled prefix directly.
		perm := es.PermissionNamespace // Casbin object — equals QualifiedName
		entity := api.Group("/" + es.Module + "/" + es.APIResource)

		// Apply RBAC per HTTP method if an evaluator is wired.
		if opts.Authz != nil {
			entity.Get("/", authz.RequirePermission(opts.Authz, perm, "read"), h.List)
			entity.Post("/", authz.RequirePermission(opts.Authz, perm, "create"), h.Create)
			entity.Get("/:id", authz.RequirePermission(opts.Authz, perm, "read"), h.Get)
			entity.Patch("/:id", authz.RequirePermission(opts.Authz, perm, "write"), h.Update)
			entity.Delete("/:id", authz.RequirePermission(opts.Authz, perm, "delete"), h.Delete)
		} else {
			entity.Get("/", h.List)
			entity.Post("/", h.Create)
			entity.Get("/:id", h.Get)
			entity.Patch("/:id", h.Update)
			entity.Delete("/:id", h.Delete)
		}

		// Register action routes from the compiled Actions slice.
		for _, action := range es.Actions {
			actionName := action.Name // capture loop var
			if opts.Authz != nil {
				entity.Post("/:id/"+actionName, authz.RequirePermission(opts.Authz, perm, actionName), func(c *fiber.Ctx) error {
					return h.Action(c)
				})
			} else {
				entity.Post("/:id/"+actionName, func(c *fiber.Ctx) error {
					return h.Action(c)
				})
			}
		}
	}

	// SDUI: /api/sdui/page/:entity/:view — returns amis JSON schema.
	// Registered only when SDUIGenerator is provided.
	// Auth middleware from /api/v1 does NOT apply here — mount a separate group.
	if opts.SDUIGenerator != nil {
		registerSDUI(app, opts)
	}
}

// registerSDUI mounts the SDUI schema endpoint group.
// Applies the same tenant + auth middleware as the REST API.
func registerSDUI(app *fiber.App, opts RegisterOptions) {
	var counter cache.Counter
	if opts.Redis != nil {
		counter = contribredis.New(opts.Redis)
	} else {
		counter = cache.NoopCounter{}
	}

	sdg := app.Group("/api/sdui")
	if opts.Tenants != nil {
		sdg.Use(middleware.TenantResolver(opts.Tenants))
	}
	if opts.IAM != nil {
		sdg.Use(middleware.RequireAuth(opts.IAM, opts.AuditSigningSecret))
	}
	sdg.Use(middleware.RateLimit(counter, middleware.DefaultRateLimit))

	gen := opts.SDUIGenerator
	sdg.Get("/page/:entity/:view", func(c *fiber.Ctx) error {
		viewer := auth.ViewerFromContext(c.UserContext())
		entityName := c.Params("entity")
		view := def.PageKind(c.Params("view"))

		schema, err := gen.GetPage(c.UserContext(), entityName, view, viewer)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return c.JSON(schema)
	})

	// Navigation: returns per-module menu entries derived from CompiledSchema.
	// GET /api/sdui/nav
	sdg.Get("/nav", handler.SDUINav(opts.SDUIGenerator))
}

// RegisterMiddleware applies the global (pre-auth) middleware pipeline to app.
// Must be called before Register. CORSConfig is applied at the global level so
// preflight OPTIONS requests are answered before the auth middleware runs.
func RegisterMiddleware(app *fiber.App, corsCfg middleware.CORSConfig) {
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS(corsCfg))
}
