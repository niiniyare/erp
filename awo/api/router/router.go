// Package router registers all auto-generated CRUD and action routes derived
// from the CompiledSchema onto a Fiber app.
//
// Route pattern (all under /api/v1/entities/):
//   - GET    /:entity          → List
//   - GET    /:entity/:id      → Get
//   - POST   /:entity          → Create
//   - PATCH  /:entity/:id      → Update
//   - DELETE /:entity/:id      → Delete
//   - POST   /:entity/:id/:action → Action handler
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
	"awo.so/awo/cache"
	"awo.so/awo/compiler"
	contrib "awo.so/awo/contrib/pgx"
	contribredis "awo.so/awo/contrib/redis"
	"awo.so/awo/def"
	"awo.so/awo/driver"
	"awo.so/awo/platform/iam"
	"awo.so/awo/runtime"
)

// RegisterOptions carries dependencies needed to build handlers.
type RegisterOptions struct {
	Pool     *pgxlib.Pool
	Redis    *goredis.Client
	IAM      *iam.Service                               // required for RequireAuth
	Tenants  driver.EntityRepository[*def.EntityRecord] // required for TenantResolver
	Authz    *authz.Enforcer                            // nil = RBAC disabled (dev/test)
	Temporal temporalclient.Client                      // nil = degraded mode (no workflow starts)
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

	// Protected API group — tenant, auth, rate-limit.
	api := app.Group("/api/v1/entities")
	if opts.Tenants != nil {
		api.Use(middleware.TenantResolver(opts.Tenants))
	}
	if opts.IAM != nil {
		api.Use(middleware.RequireAuth(opts.IAM))
	}
	api.Use(middleware.RateLimit(counter, middleware.DefaultRateLimit))

	for _, es := range schema.Entities {
		repo := contrib.NewRepository(opts.Pool, es)
		pipeline := runtime.NewPipeline(schema)
		svc := service.NewEntityService(es, repo, pipeline, opts.Temporal)
		h := handler.NewEntityHandler(es, svc)
		entityName := es.TableName

		entity := api.Group("/" + entityName)

		// Apply RBAC per HTTP method if an enforcer is wired.
		if opts.Authz != nil {
			entity.Get("/", opts.Authz.RequirePermission(entityName, "read"), h.List)
			entity.Post("/", opts.Authz.RequirePermission(entityName, "create"), h.Create)
			entity.Get("/:id", opts.Authz.RequirePermission(entityName, "read"), h.Get)
			entity.Patch("/:id", opts.Authz.RequirePermission(entityName, "write"), h.Update)
			entity.Delete("/:id", opts.Authz.RequirePermission(entityName, "delete"), h.Delete)
		} else {
			entity.Get("/", h.List)
			entity.Post("/", h.Create)
			entity.Get("/:id", h.Get)
			entity.Patch("/:id", h.Update)
			entity.Delete("/:id", h.Delete)
		}

		// Register action routes.
		for actionName := range es.ActionsByName {
			name := actionName // capture loop var
			if opts.Authz != nil {
				entity.Post("/:id/"+name, opts.Authz.RequirePermission(entityName, name), func(c *fiber.Ctx) error {
					return h.Action(c)
				})
			} else {
				entity.Post("/:id/"+name, func(c *fiber.Ctx) error {
					return h.Action(c)
				})
			}
		}
	}
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
