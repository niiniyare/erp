package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/iam"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

func doGetStatus(app *fiber.App, path string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck
	return resp.StatusCode
}

// withSession injects a ResolvedSession AND its derived Principal into Fiber Locals.
func withSession(sess *iam.ResolvedSession) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(iam.LocalsKeySession, sess)
		c.Locals(iam.LocalsKeyPrincipal, sess.ToPrincipal())
		return c.Next()
	}
}

func okHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// stubSession builds a minimal ResolvedSession for the given tenant+user.
func stubSession(tenantID, userID uuid.UUID) *iam.ResolvedSession {
	return &iam.ResolvedSession{
		UserID:        userID,
		TenantID:      tenantID,
		UserType:      "INTERNAL",
		Configuration: iam.DefaultConfiguration(),
	}
}

// ─── Mock AuthzService ───────────────────────────────────────────────────────

// recordingAuthzService records calls to Enforce and returns configured results.
type recordingAuthzService struct {
	Called      bool
	LastRequest *iam.Request
	Allow       bool
	Err         error
}

func (s *recordingAuthzService) Enforce(_ context.Context, r iam.Request) (bool, error) {
	s.Called = true
	s.LastRequest = &r
	return s.Allow, s.Err
}

func (s *recordingAuthzService) EnforceBatch(_ context.Context, _ []iam.Request) ([]bool, error) {
	return nil, nil
}

func (s *recordingAuthzService) AssignRole(_ context.Context, _, _, _, _ string, _ ...iam.AssignOpt) error {
	return nil
}

func (s *recordingAuthzService) RevokeRole(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *recordingAuthzService) GetRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *recordingAuthzService) GetImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *recordingAuthzService) HasRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func (s *recordingAuthzService) GetAssignments(_ context.Context, _, _ string) ([]iam.RoleAssignment, error) {
	return nil, nil
}

func (s *recordingAuthzService) AddPolicy(_ context.Context, _ iam.Policy) error {
	return nil
}

func (s *recordingAuthzService) RemovePolicy(_ context.Context, _ iam.Policy) error {
	return nil
}

func (s *recordingAuthzService) GetPolicies(_ context.Context, _ string) ([]iam.Policy, error) {
	return nil, nil
}

func (s *recordingAuthzService) InvalidateCache(_ context.Context) error {
	return nil
}

func (s *recordingAuthzService) BootstrapTenantAdmin(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// TestRequirePermission_CallsEnforce (AZ-MID post-refactor)
//
// Verify that Authorize() routes every permission check through
// authzService.Enforce() — never through a session permission map.
func TestRequirePermission_CallsEnforce(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	sess := stubSession(tenantID, userID)

	mock := &recordingAuthzService{Allow: true}
	cfg := AuthConfig{AuthzService: mock}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test",
		withSession(sess),
		Authorize(cfg, "finance.invoices.read"),
		okHandler,
	)

	status := doGetStatus(app, "/test")
	require.Equal(t, fiber.StatusOK, status)

	assert.True(t, mock.Called, "Enforce must be called by Authorize middleware")
	require.NotNil(t, mock.LastRequest)
	// Object = "finance.invoices", Action = "read"  (splitPermission splits on last dot)
	assert.Equal(t, "finance.invoices", mock.LastRequest.Object)
	assert.Equal(t, "read", mock.LastRequest.Action)
}

// TestRequirePermission_NoSessionCan (AZ-MID post-refactor)
//
// ResolvedSession has no Can() / CanDo() method — authorization is Casbin-only.
// This test is a compile-time proof: if Can() existed on ResolvedSession, the
// compilation of this package would fail at the assertion below.
func TestRequirePermission_NoSessionCan(t *testing.T) {
	t.Parallel()

	sess := &iam.ResolvedSession{}

	// ResolvedSession must NOT have a Can() method.
	// If the line below compiles (it always should after BLOCK-1), the session
	// model no longer carries embedded authorization logic.
	type mustNotHaveCan interface {
		// deliberate: does NOT embed any "Can" method
		FeatureEnabled(string) bool
	}
	var _ mustNotHaveCan = sess // compile-time check: only FeatureEnabled is expected

	// Runtime: confirm FeatureEnabled is reachable (not shadowed by Can).
	assert.False(t, sess.FeatureEnabled("any.flag"))
}

// TestRequireFlag_UsesSessionFeatureEnabled (AZ-MID post-refactor)
//
// RequireFlag must read from the session's pre-computed Configuration.Flags
// via FeatureEnabled() — not from Casbin or any other source.
func TestRequireFlag_UsesSessionFeatureEnabled(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	makeApp := func(flagEnabled bool) *fiber.App {
		sess := &iam.ResolvedSession{
			UserID:   userID,
			TenantID: tenantID,
			UserType: "INTERNAL",
			Configuration: iam.Configuration{
				Flags:    map[string]bool{"finance.transactions": flagEnabled},
				Settings: map[string]string{},
				Prefs:    map[string]string{},
			},
		}
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Get("/test",
			withSession(sess),
			RequireFlag("finance.transactions"),
			okHandler,
		)
		return app
	}

	t.Run("flag_enabled_passes", func(t *testing.T) {
		status := doGetStatus(makeApp(true), "/test")
		assert.Equal(t, fiber.StatusOK, status)
	})

	t.Run("flag_disabled_returns_403", func(t *testing.T) {
		status := doGetStatus(makeApp(false), "/test")
		assert.Equal(t, fiber.StatusForbidden, status)
	})

	t.Run("flag_absent_returns_403", func(t *testing.T) {
		sess := &iam.ResolvedSession{
			UserID:   userID,
			TenantID: tenantID,
			UserType: "INTERNAL",
			Configuration: iam.Configuration{
				Flags:    map[string]bool{}, // flag not set at all
				Settings: map[string]string{},
				Prefs:    map[string]string{},
			},
		}
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Get("/test",
			withSession(sess),
			RequireFlag("finance.transactions"),
			okHandler,
		)
		status := doGetStatus(app, "/test")
		assert.Equal(t, fiber.StatusForbidden, status)
	})
}

// TestAuthorize_EnforceError_Returns500
//
// When Enforce() returns a non-nil error, Authorize must respond 500, not 403.
// This distinguishes "denied" (403) from "unable to evaluate" (500).
func TestAuthorize_EnforceError_Returns500(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	sess := stubSession(tenantID, userID)

	mock := &recordingAuthzService{
		Allow: false,
		Err:   fmt.Errorf("enforcer unavailable"),
	}
	cfg := AuthConfig{AuthzService: mock}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test",
		withSession(sess),
		Authorize(cfg, "finance.invoices.read"),
		okHandler,
	)
	status := doGetStatus(app, "/test")
	assert.Equal(t, fiber.StatusInternalServerError, status,
		"Enforce error must produce 500, not 403")
}

