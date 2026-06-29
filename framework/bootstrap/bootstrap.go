// Package bootstrap wires the full Awo Framework onto a Fiber application.
//
// Usage:
//
//	pool, _ := pgxpool.New(ctx, dsn)
//	app := fiber.New()
//	bootstrap.Mount(app, bootstrap.Options{
//	    Pool:     pool,
//	    ViewerFn: myViewerExtractor,
//	})
package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tclient "go.temporal.io/sdk/client"

	"awo.so/framework/api"
	"awo.so/framework/def"
	"awo.so/framework/persistence/pgstore"
	platformorg "awo.so/framework/platform/org"
	platformorgpgorg "awo.so/framework/platform/org/pgorg"
	"awo.so/framework/sdui"
	"awo.so/framework/workflow"
)

// Options configures a framework Mount.
type Options struct {
	// Pool is the pgx connection pool. Required.
	Pool *pgxpool.Pool

	// APIPrefix is the base path for all entity and SDUI routes.
	// Defaults to "/api".
	APIPrefix string

	// ViewerFn extracts authentication context from each request.
	// Defaults to anonymous viewer when nil.
	ViewerFn api.ViewerFromCtx

	// TenantResolver resolves a tenant slug or non-UUID string to a UUID.
	// Required when ViewerFn may return a slug instead of a UUID from TenantID().
	// Example: look up tenants table by slug.
	TenantResolver api.TenantResolver

	// OrgTree is the org unit tree used by AllowWithinOrgScope policies.
	// When nil, a default PgTree backed by Pool is constructed automatically.
	// Supply a custom implementation to use caching or a test double.
	OrgTree platformorg.Tree

	// TemporalClient enables workflow triggers on entity mutations.
	// Pass nil to disable workflow integration.
	TemporalClient tclient.Client
}

// Mount registers all framework routes on app using opts.
//
// Routes registered:
//   - GET/POST   /<prefix>/<table>          — list / create
//   - GET/PUT/DELETE /<prefix>/<table>/:id  — read / update / delete
//   - GET        /<prefix>/sdui/nav         — navigation tree
//   - GET        /<prefix>/sdui/:entity     — CRUD page schema
//   - GET        /<prefix>/sdui/:entity/form — form schema
func Mount(app *fiber.App, opts Options) {
	if opts.APIPrefix == "" {
		opts.APIPrefix = "/api"
	}
	if opts.ViewerFn == nil {
		opts.ViewerFn = anonymousViewer
	}
	if opts.OrgTree == nil {
		opts.OrgTree = platformorgpgorg.New(opts.Pool)
	}

	tenantStore := pgstore.NewTenantStore(opts.Pool)

	var wfExec *workflow.Executor
	if opts.TemporalClient != nil {
		wfExec = workflow.NewExecutor(opts.TemporalClient)
	}
	_ = wfExec // used by modules that register workflow hooks via def.HookDef

	apiGroup := app.Group(opts.APIPrefix)
	handlerOpts := []api.HandlerOption{
		api.WithOrgTree(opts.OrgTree),
	}
	if opts.TenantResolver != nil {
		handlerOpts = append(handlerOpts, api.WithTenantResolver(opts.TenantResolver))
	}
	for _, def := range def.All() {
		h := api.NewHandler(def, tenantStore, opts.ViewerFn, handlerOpts...)
		api.Register(apiGroup, "/"+def.TableName(), h)
	}

	sdui.Register(app, opts.APIPrefix)
}

// anonymousViewer is the fallback ViewerFromCtx when none is provided.
// Returns an unauthenticated viewer — suitable only for dev/testing.
func anonymousViewer(c *fiber.Ctx) (def.ViewerContext, error) {
	tenantID := c.Get("X-Awo-Tenant")
	if tenantID == "" {
		tenantID, _ = c.Locals("tenant_slug").(string)
	}
	return &anonViewer{tenantID: tenantID}, nil
}

type anonViewer struct{ tenantID string }

func (v *anonViewer) ActorID() string       { return "anonymous" }
func (v *anonViewer) TenantID() string      { return v.tenantID }
func (v *anonViewer) OrgUnitID() uuid.UUID  { return uuid.Nil }
func (v *anonViewer) IsSystem() bool        { return false }
func (v *anonViewer) HasRole(_ string) bool { return false }
