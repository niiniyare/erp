package iam

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// MiddlewareSuite — tests for service.Middleware(object, action).
// Uses the real service backed by a test DB and Fiber's app.Test() helper.
// Requires DATABASE_URL environment variable.
// ---------------------------------------------------------------------------

type MiddlewareSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	svc  Service
	ctx  context.Context
}

func TestMiddlewareSuite(t *testing.T) { suite.Run(t, new(MiddlewareSuite)) }

func (s *MiddlewareSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pool = testPool(s.T())
	s.svc = newTestService(s.T(), s.pool)
	seedTestTenant(s.T(), s.pool)
}

func (s *MiddlewareSuite) SetupTest() {
	cleanTables(s.T(), s.pool)
}

func (s *MiddlewareSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// fiberApp builds a minimal Fiber app with the given handler chain.
func fiberApp(handlers ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", handlers...)
	return app
}

// fiberAppWithID builds a minimal Fiber app with an :id route param.
func fiberAppWithID(handlers ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test/:id", handlers...)
	return app
}

// withPrincipal injects a Principal into Fiber Locals before the middleware.
func withPrincipal(p Principal) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(LocalsKeyPrincipal, p)
		return c.Next()
	}
}

// okHandler is the downstream handler that returns 200 OK.
func okHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// doGet performs a GET request against the Fiber app and returns the status.
func doGet(app *fiber.App, path string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck
	return resp.StatusCode
}

// ---------------------------------------------------------------------------
// Test cases
// ---------------------------------------------------------------------------

func (s *MiddlewareSuite) TestMiddleware_NoPrincipal_Returns401() {
	app := fiberApp(
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusUnauthorized, doGet(app, "/test"))
}

func (s *MiddlewareSuite) TestMiddleware_EmptySubject_Returns401() {
	app := fiberApp(
		withPrincipal(Principal{Subject: "", Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusUnauthorized, doGet(app, "/test"))
}

func (s *MiddlewareSuite) TestMiddleware_NoPolicyForUser_Returns403() {
	// No policies or role assignments — default deny.
	app := fiberApp(
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusForbidden, doGet(app, "/test"))
}

func (s *MiddlewareSuite) TestMiddleware_AllowPolicy_Returns200_AndCallsNext() {
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	app := fiberApp(
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusOK, doGet(app, "/test"))
}

func (s *MiddlewareSuite) TestMiddleware_ObjectExpansion_WithIDParam() {
	// Route has :id → object becomes "invoice/{id}"
	// Policy must cover "invoice/*" (wildcard) or exact "invoice/inv_123".
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	app := fiberAppWithID(
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusOK, doGet(app, "/test/inv_123"))
}

func (s *MiddlewareSuite) TestMiddleware_ObjectExpansion_IDParamUsedAsResourceID() {
	// Policy only covers invoice/specific_id, not invoice/*.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/specific_id",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	app := fiberAppWithID(
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)

	// Correct id: allowed.
	s.Equal(fiber.StatusOK, doGet(app, "/test/specific_id"))

	// Wrong id: denied.
	s.Equal(fiber.StatusForbidden, doGet(app, "/test/wrong_id"))
}

func (s *MiddlewareSuite) TestMiddleware_NoIDParam_UsesPlainObject() {
	// Route has no :id param → object is "invoice" as-is.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice",
		Action:  "read",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))

	app := fiberApp( // no :id route
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusOK, doGet(app, "/test"))
}

func (s *MiddlewareSuite) TestMiddleware_DenyRule_Returns403() {
	// Even with an allow policy via role, a blanket deny on the subject wins.
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testRole,
		Domain:  testDomain,
		Object:  "invoice/*",
		Action:  "*",
		Effect:  "allow",
	}))
	s.Require().NoError(s.svc.AssignRole(s.ctx, testTenantID, testSubject, testRole, testDomain))
	s.Require().NoError(s.svc.AddPolicy(s.ctx, Policy{
		Subject: testSubject,
		Domain:  testDomain,
		Object:  "*",
		Action:  "*",
		Effect:  "deny",
	}))

	app := fiberAppWithID(
		withPrincipal(Principal{Subject: testSubject, Domain: testDomain}),
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusForbidden, doGet(app, "/test/inv_001"))
}

func (s *MiddlewareSuite) TestMiddleware_LocalsKey_IsCaseSensitive() {
	// Storing Principal under the wrong key → 401.
	wrongKey := fiber.Handler(func(c *fiber.Ctx) error {
		c.Locals("WRONG_KEY", Principal{Subject: testSubject, Domain: testDomain})
		return c.Next()
	})

	app := fiberApp(
		wrongKey,
		testMiddlewareHandler(s.svc,"invoice", "read"),
		okHandler,
	)
	s.Equal(fiber.StatusUnauthorized, doGet(app, "/test"))
}
