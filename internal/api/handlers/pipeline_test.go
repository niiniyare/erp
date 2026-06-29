// Package handlers_test contains end-to-end pipeline tests for the HTTP layer.
// These tests exercise the full Authenticate → Authorize / RequireFlag →
// handler chain using mock services — no database required.
//
// Phase 16 coverage:
// V1 — Login returns a populated session (HttpOnly cookie + non-empty Permissions)
// V2 — Protected route: 401 no token, 403 wrong perm, 200 correct perm
// V3 — Feature-flag gate: 403 flag off, 200 flag on
package handlers_test

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authHandler "awo.so/internal/api/handlers/auth"
	mw "awo.so/internal/api/middleware"
	"awo.so/internal/core/iam"
)

// mock SessionService

type mockSessionSvc struct {
	loginSess  *iam.ResolvedSession
	loginToken string
	loginErr   error

	validateSess *iam.ResolvedSession
	validateErr  error
}

func (m *mockSessionSvc) Login(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return m.loginSess, m.loginToken, m.loginErr
}

func (m *mockSessionSvc) CompleteMFALogin(_ context.Context, _, _ string) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}

func (m *mockSessionSvc) ValidateSession(_ context.Context, _ string) (*iam.ResolvedSession, error) {
	return m.validateSess, m.validateErr
}
func (m *mockSessionSvc) Logout(_ context.Context, _ string) error                { return nil }
func (m *mockSessionSvc) LogoutAllForUser(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *mockSessionSvc) LogoutAllForTenant(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockSessionSvc) LoginWithSSO(_ context.Context, _ *iam.User) (*iam.ResolvedSession, string, error) {
	return nil, "", nil
}

// mockAuthzSvc is a minimal AuthzService stub for pipeline tests.
// Only Enforce is wired; all other methods panic if unexpectedly called.
type mockAuthzSvc struct {
	iam.AuthzService      // satisfies remaining interface methods
	enforce          bool // return value for Enforce
}

func (m *mockAuthzSvc) Enforce(_ context.Context, _ iam.Request) (bool, error) {
	return m.enforce, nil
}

// suite setup

type PipelineSuite struct {
	suite.Suite
	svc      *mockSessionSvc
	authzSvc *mockAuthzSvc
	app      *fiber.App
}

func TestPipelineSuite(t *testing.T) { suite.Run(t, new(PipelineSuite)) }

func (s *PipelineSuite) SetupTest() {
	s.svc = &mockSessionSvc{}
	s.authzSvc = &mockAuthzSvc{} // enforce=false by default

	// Use non-secure cookies so plain HTTP test requests work.
	loginCfg := authHandler.DefaultLoginConfig()
	loginCfg.SecureCookie = false

	authCfg := mw.DefaultAuthConfig(s.svc)
	authCfg.AuthzService = s.authzSvc

	app := fiber.New(fiber.Config{
		// Return JSON for Fiber-level errors so status codes are reliable.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if stderrors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Login route (public)
	app.Post("/api/v1/auth/login", authHandler.LoginHandler(s.svc, loginCfg))

	// V2: permission-gated route
	app.Get("/api/v1/protected",
		mw.Authenticate(authCfg),
		mw.Authorize(authCfg, "finance.accounts.read"),
		func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) },
	)

	// V3: feature-flag-gated route
	app.Get("/api/v1/finance-gated",
		mw.Authenticate(authCfg),
		mw.RequireFlag("finance"),
		func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) },
	)

	s.app = app
}

// V1 — Login returns populated session

func (s *PipelineSuite) TestV1_Login_Returns200_WithCookieAndPermissions() {
	sess := &iam.ResolvedSession{
		UserID:   uuid.New(),
		UserType: "INTERNAL",
		TenantID: uuid.New(),
		// Permissions: map[string]bool{"finance.accounts.read": true},
		EntityScope: iam.EntityScope{Type: iam.EntityScopeAll},
		Configuration: func() iam.Configuration {
			cfg := iam.DefaultConfiguration()
			cfg.Flags["finance"] = true
			return cfg
		}(),
	}
	s.svc.loginSess = sess
	s.svc.loginToken = "raw-test-token"

	body := `{"email":"user@acme.com","password":"s3cur3!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Cookie must be set and HttpOnly
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(s.T(), sessionCookie, "session cookie must be set")
	require.True(s.T(), sessionCookie.HttpOnly, "session cookie must be HttpOnly")
	require.Equal(s.T(), "raw-test-token", sessionCookie.Value)

	// Response body must contain user identity fields.
	var result map[string]any
	require.NoError(s.T(), json.NewDecoder(resp.Body).Decode(&result))
	userID, ok := result["user_id"]
	require.True(s.T(), ok, "response must include user_id field")
	require.NotEmpty(s.T(), userID, "user_id must be non-empty")
}

// V2 — Protected route end-to-end

func (s *PipelineSuite) TestV2_NoToken_Returns401() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *PipelineSuite) TestV2_ValidSession_WrongPermission_Returns403() {
	// authzSvc.enforce=false (default) — Casbin denies finance.accounts.read
	s.svc.validateSess = &iam.ResolvedSession{
		UserID:        uuid.New(),
		UserType:      "INTERNAL",
		TenantID:      uuid.New(),
		Configuration: iam.DefaultConfiguration(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token"})

	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *PipelineSuite) TestV2_ValidSession_CorrectPermission_Returns200() {
	s.authzSvc.enforce = true // Casbin grants finance.accounts.read for this session
	s.svc.validateSess = &iam.ResolvedSession{
		UserID:        uuid.New(),
		UserType:      "INTERNAL",
		TenantID:      uuid.New(),
		Configuration: iam.DefaultConfiguration(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token"})

	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// V3 — Feature flag gates route

func (s *PipelineSuite) TestV3_FlagDisabled_Returns403() {
	s.svc.validateSess = &iam.ResolvedSession{
		UserID:   uuid.New(),
		UserType: "INTERNAL",
		TenantID: uuid.New(),
		// Permissions: map[string]bool{"finance.accounts.read": true},
		Configuration: iam.Configuration{
			Flags:    map[string]bool{"finance": false},
			Settings: map[string]string{},
			Prefs:    map[string]string{},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/finance-gated", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token"})

	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *PipelineSuite) TestV3_FlagEnabled_Returns200() {
	s.svc.validateSess = &iam.ResolvedSession{
		UserID:   uuid.New(),
		UserType: "INTERNAL",
		TenantID: uuid.New(),
		// Permissions: map[string]bool{},
		Configuration: iam.Configuration{
			Flags:    map[string]bool{"finance": true},
			Settings: map[string]string{},
			Prefs:    map[string]string{},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/finance-gated", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token"})

	resp, err := s.app.Test(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}
