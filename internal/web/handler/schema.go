// Package handler provides the HTTP handler that serves AMIS page schemas.
// One handler, zero boilerplate per new page. Register your schema function
// in the registry package and it's automatically served.
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/iam/contract"
	"awo.so/internal/pipeline"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// PipelineRunner is the minimal interface SchemaHandler needs from the pipeline.
// Both *web.UIPipeline and *pipeline.PipelineBuilder satisfy this interface.
type PipelineRunner interface {
	Run(opCtx *pipeline.OperationContext) error
}

// SchemaHandler serves AMIS page schemas via the UI pipeline.
// Mount it at /schema/* — any sub-path resolves via the page registry.
//
//	app.Get("/schema/*", schemaHandler.Handle)
//
// Route chain MUST include, in order:
//  1. iam/middleware.Authenticate(cfg)    — validates token, sets Fiber Locals
//  2. contract.InjectSessionContext()     — bridges Fiber Locals → Go context
//  3. SchemaHandler.Handle               — builds OperationContext, runs pipeline
//
// SchemaHandler never reads Fiber Locals directly. All identity data arrives
// via contract.FromContext(c.UserContext()).
type SchemaHandler struct {
	pipeline PipelineRunner
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
	log      logger.Logger
}

// NewSchemaHandler constructs a SchemaHandler. pb must not be nil.
// tracer, metrics, and log are optional — nil providers are skipped.
func NewSchemaHandler(
	pb PipelineRunner,
	tracer tracing.Service,
	mp metrics.MetricsProvider,
	log logger.Logger,
) *SchemaHandler {
	return &SchemaHandler{
		pipeline: pb,
		tracer:   tracer,
		metrics:  mp,
		log:      log,
	}
}

// Handle serves the schema for the requested path.
// Path: everything after /schema, e.g. /schema/finance/invoices → /finance/invoices
func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
	// ── Contract boundary: session via Go context, not Fiber Locals ──────────
	// contract.InjectSessionContext() middleware must run before this handler.
	sc, ok := contract.FromContext(c.UserContext())
	if !ok || sc.IsZero() {
		return c.Status(fiber.StatusUnauthorized).JSON(unauthenticatedEnvelope())
	}

	route := "/" + c.Params("*")

	// ── Request-level trace span ─────────────────────────────────────────────
	ctx := c.UserContext()
	var requestSpan tracing.Span
	if h.tracer != nil {
		ctx, requestSpan = h.tracer.StartSpan(ctx, "ui.schema_handler",
			tracing.WithSpanKind(tracing.SpanKindServer),
		)
		defer requestSpan.End()
		requestSpan.SetAttributes(
			attribute.String("ui.route", route),
			attribute.String("ui.tenant_id", sc.TenantID().String()),
			attribute.String("ui.user_id", sc.UserID().String()),
		)
		c.SetUserContext(ctx)
	}

	// ── Build OperationContext from pool ─────────────────────────────────────
	opCtx := pipeline.AcquireOperationContext()
	defer pipeline.ReleaseOperationContext(opCtx)

	// opCtx.Ctx carries the contract.SessionContext — pipeline stages use
	// contract.FromContext(opCtx.Ctx) to read it. This is the canonical pattern.
	opCtx.Ctx          = ctx
	opCtx.TenantID     = sc.TenantID()
	opCtx.UserID       = sc.UserID()
	opCtx.OperationKey = ui.OperationKey
	opCtx.Input        = ui.UISchemaInput{Route: route}
	// Note: opCtx.Session is intentionally NOT set here.
	// UI stages MUST use contract.FromContext(opCtx.Ctx), not opCtx.Session.
	// Setting opCtx.Session would require importing internal IAM domain types
	// from this package, violating the contract boundary.

	// ── Run pipeline ─────────────────────────────────────────────────────────
	if err := h.pipeline.Run(opCtx); err != nil {
		return h.handlePipelineError(c, route, err, requestSpan)
	}

	// ── Read response from pipeline output ───────────────────────────────────
	out, ok := opCtx.Data[ui.DataKeyResponse].(ui.UISchemaOutput)
	if !ok {
		h.logError(c, route, "DataKeyResponse missing after successful pipeline.Run()")
		return c.Status(fiber.StatusInternalServerError).JSON(internalErrorEnvelope())
	}

	// ── Emit request-level metrics ───────────────────────────────────────────
	if h.metrics != nil {
		h.metrics.IncrementCounter("ui_schema_requests_total", metrics.Fields{
			"route":     route,
			"tenant_id": sc.TenantID().String(),
			"cache_hit": out.CacheHit,
		})
	}

	if requestSpan != nil {
		requestSpan.SetAttributes(
			attribute.Bool("ui.cache_hit", out.CacheHit),
		)
	}

	// ── AMIS envelope: { status: 0, data: <schema> } ─────────────────────────
	return c.JSON(fiber.Map{
		"status": 0,
		"data":   out.Schema,
	})
}

// handlePipelineError maps pipeline errors to HTTP responses.
func (h *SchemaHandler) handlePipelineError(c *fiber.Ctx, route string, err error, span tracing.Span) error {
	if span != nil {
		span.RecordError(err)
	}

	if errors.Is(err, ui.ErrPageNotFound) {
		h.logWarn(c, route, "page not found", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": 404,
			"msg":    "schema not found: " + route,
		})
	}

	if errors.Is(err, ui.ErrUnauthenticated) {
		return c.Status(fiber.StatusUnauthorized).JSON(unauthenticatedEnvelope())
	}

	if errors.Is(err, ui.ErrSchemaInvalid) {
		// Schema validation failure = programmer error. Log details, return 500.
		h.logError(c, route, "schema validation failed: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(internalErrorEnvelope())
	}

	if errors.Is(err, ui.ErrPermissionResolution) {
		h.logError(c, route, "IAM permission resolution failed: "+err.Error())
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": 503,
			"msg":    "service temporarily unavailable",
		})
	}

	// Unclassified error — internal server error.
	h.logError(c, route, "unclassified pipeline error: "+err.Error())
	return c.Status(fiber.StatusInternalServerError).JSON(internalErrorEnvelope())
}

func (h *SchemaHandler) logError(c *fiber.Ctx, route, msg string) {
	if h.log != nil {
		h.log.ErrorContext(c.UserContext(), msg, logger.Fields{
			"route":     route,
			"tenant_id": c.Get("X-Tenant-ID"),
		})
	}
}

func (h *SchemaHandler) logWarn(c *fiber.Ctx, route, msg string, err error) {
	if h.log != nil {
		h.log.WarnContext(c.UserContext(), msg, logger.Fields{
			"route": route,
			"error": err.Error(),
		})
	}
}

// unauthenticatedEnvelope returns an AMIS-compatible 401 response.
// Returning an AMIS envelope (not a plain error) means the browser renders
// a "Session Expired" page with a Login button instead of a blank screen.
//
// Schema is built via typed ast.PageNode + ast.CompileTree so it passes
// ValidateStage and obeys the no-raw-map rule in the handler layer.
func unauthenticatedEnvelope() fiber.Map {
	page := ast.PageNode{
		Title:    "Session Expired",
		SubTitle: "Your session has expired. Please log in to continue.",
		Toolbar: []ast.Node{
			ast.ActionNode{
				Label:      "Log In",
				ActionType: "link",
				Target:     "/ui/login",
				Level:      "primary",
				Icon:       "fa fa-sign-in-alt",
			},
		},
	}
	schema, err := ast.CompileTree(page)
	if err != nil {
		// Hardcoded schema — CompileTree should never fail here.
		// Fallback keeps the 401 semantics without a schema body.
		return fiber.Map{
			"status": 401,
			"msg":    "Session expired. Please log in.",
		}
	}
	return fiber.Map{
		"status": 401,
		"msg":    "Session expired.",
		"data":   schema,
	}
}

func internalErrorEnvelope() fiber.Map {
	return fiber.Map{
		"status": 500,
		"msg":    "An internal error occurred. Please try again.",
	}
}
