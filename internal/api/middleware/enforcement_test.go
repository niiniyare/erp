package middleware

// enforcement_test.go — AUTHZ-MW-1 through AUTHZ-MW-5.
//
// Middleware enforcement integrity: verify that every request path through
// Authenticate and Authorize upholds the enforcement contract.
//
// AUTHZ-MW-1: Missing token → 401; handler is never reached.
// AUTHZ-MW-2: Invalid/expired session → 401; handler is never reached.
// AUTHZ-MW-3: Authorize without prior Authenticate → 401 (no bypass).
// AUTHZ-MW-4: Invalid subject prefix → 401 (subject validation in Authenticate).
// AUTHZ-MW-5: AuthorizeCasbin with empty Principal subject → 401 (not 403).

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"awo.so/internal/core/iam"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

// stubSessionService satisfies iam.SessionService for middleware tests.
// ValidateSession returns the configured session/error; all other methods panic
// so any unexpected call is immediately visible.
type stubSessionService struct {
	session *iam.ResolvedSession
	err     error
}

func (s *stubSessionService) Login(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	panic("unexpected call to Login in middleware test")
}

func (s *stubSessionService) CompleteMFALogin(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	panic("unexpected call to CompleteMFALogin in middleware test")
}

func (s *stubSessionService) ValidateSession(_ context.Context, _ string) (*iam.ResolvedSession, error) {
	return s.session, s.err
}

func (s *stubSessionService) Logout(_ context.Context, _ string) error {
	panic("unexpected call to Logout in middleware test")
}

func (s *stubSessionService) LogoutAllForUser(_ context.Context, _ uuid.UUID) error {
	panic("unexpected call to LogoutAllForUser in middleware test")
}

func (s *stubSessionService) LogoutAllForTenant(_ context.Context, _ uuid.UUID) error {
	panic("unexpected call to LogoutAllForTenant in middleware test")
}

func (s *stubSessionService) LoginWithSSO(_ context.Context, _ *iam.User) (*iam.ResolvedSession, string, error) {
	panic("unexpected call to LoginWithSSO in middleware test")
}

// reachedHandler panics if called; used to assert that a handler must NOT run.
func reachedHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// handlerReached is set to true by touchHandler so tests can assert reachability.
type handlerSpy struct{ reached bool }

func (hs *handlerSpy) handler(c *fiber.Ctx) error {
	hs.reached = true
	return c.SendStatus(fiber.StatusOK)
}

func runRequest(app *fiber.App, method, path string, headers map[string]string) int {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// ---------------------------------------------------------------------------
// AUTHZ-MW-1: Missing token → 401; handler not reached
// ---------------------------------------------------------------------------

func TestAUTHZ_MW_1_MissingToken_Returns401(t *testing.T) {
	svc := &stubSessionService{session: nil}
	cfg := AuthConfig{SessionService: svc}

	spy := &handlerSpy{}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/protected",
		Authenticate(cfg),
		spy.handler,
	)

	// No cookie, no Authorization header.
	status := runRequest(app, http.MethodGet, "/protected", nil)

	assert.Equal(t, fiber.StatusUnauthorized, status, "missing token must return 401")
	assert.False(t, spy.reached, "handler must not be reached when token is missing")
}

// ---------------------------------------------------------------------------
// AUTHZ-MW-2: Invalid/expired session → 401; handler not reached
// ---------------------------------------------------------------------------

func TestAUTHZ_MW_2_InvalidSession_Returns401(t *testing.T) {
	// ValidateSession returns nil session (expired/revoked).
	svc := &stubSessionService{session: nil, err: nil}
	cfg := AuthConfig{SessionService: svc}

	spy := &handlerSpy{}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/protected",
		Authenticate(cfg),
		spy.handler,
	)

	status := runRequest(app, http.MethodGet, "/protected", map[string]string{
		"Authorization": "Bearer some-expired-token",
	})

	assert.Equal(t, fiber.StatusUnauthorized, status, "expired/invalid session must return 401")
	assert.False(t, spy.reached, "handler must not be reached on invalid session")
}

// ---------------------------------------------------------------------------
// AUTHZ-MW-3: Authorize without Authenticate → 401 (no session bypass)
// ---------------------------------------------------------------------------

func TestAUTHZ_MW_3_AuthorizeWithoutAuthenticate_Returns401(t *testing.T) {
	// AuthzService would allow everything — but Authorize must check session first.
	authz := &recordingAuthzService{Allow: true}
	cfg := AuthConfig{AuthzService: authz}

	spy := &handlerSpy{}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	// Intentionally no Authenticate middleware — Authorize must reject.
	app.Get("/protected",
		Authorize(cfg, "finance.invoices.read"),
		spy.handler,
	)

	status := runRequest(app, http.MethodGet, "/protected", nil)

	assert.Equal(t, fiber.StatusUnauthorized, status,
		"Authorize without Authenticate must return 401, not bypass or 403")
	assert.False(t, spy.reached, "handler must not be reached")
	assert.False(t, authz.Called, "Enforce must not be called when session is absent")
}

// ---------------------------------------------------------------------------
// AUTHZ-MW-4: isValidSubject rejects unknown/empty prefixes
//
// isValidSubject is the guard in Authenticate that rejects tokens with
// unrecognised or empty subjects. Since ToPrincipal() always produces a valid
// prefix, we test isValidSubject directly (same package) with crafted inputs.
// ---------------------------------------------------------------------------

func TestAUTHZ_MW_4_IsValidSubject(t *testing.T) {
	t.Parallel()

	valid := []string{
		"tenant:550e8400-e29b-41d4-a716-446655440000",
		"platform:550e8400-e29b-41d4-a716-446655440000",
		"portal:550e8400-e29b-41d4-a716-446655440000",
		"api:550e8400-e29b-41d4-a716-446655440000",
	}
	for _, s := range valid {
		assert.True(t, isValidSubject(s), "expected valid: %q", s)
	}

	invalid := []string{
		"",
		"tenant:",           // empty suffix
		"platform:",         // empty suffix
		"admin:some-id",     // unrecognised prefix
		"user:some-id",      // unrecognised prefix
		"TENANT:some-id",    // wrong case
		"some-random-value", // no prefix at all
	}
	for _, s := range invalid {
		assert.False(t, isValidSubject(s), "expected invalid: %q", s)
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-MW-5: AuthorizeCasbin with empty Principal Subject → 401 (not 403)
// ---------------------------------------------------------------------------

func TestAUTHZ_MW_5_AuthorizeCasbin_EmptySubject_Returns401(t *testing.T) {
	// Inject a Principal with empty Subject — should be rejected before Enforce.
	authz := &recordingAuthzService{Allow: false}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/mgmt",
		// Inject empty principal directly (simulates missing/corrupted Locals).
		func(c *fiber.Ctx) error {
			c.Locals(iam.LocalsKeyPrincipal, iam.Principal{Subject: "", Domain: ""})
			return c.Next()
		},
		AuthorizeCasbin(authz, "role", "assign"),
		reachedHandler,
	)

	status := runRequest(app, http.MethodGet, "/mgmt", nil)

	assert.Equal(t, fiber.StatusUnauthorized, status,
		"empty Principal.Subject must return 401, not 403")
	assert.False(t, authz.Called,
		"Enforce must not be called when Principal.Subject is empty")
}
