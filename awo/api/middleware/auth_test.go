package middleware_test

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/awo/api/middleware"
	"awo.so/awo/auth"
	"awo.so/awo/runtime"
)

// ── stub SessionValidator ─────────────────────────────────────────────────────

type stubValidator struct {
	session   *auth.Session
	sessionErr error
	apiSession   *auth.Session
	apiErr     error
}

func (s *stubValidator) ValidateToken(_ context.Context, _ string) (*auth.Session, error) {
	return s.session, s.sessionErr
}

func (s *stubValidator) ValidateAPIToken(_ context.Context, _ string, _ uuid.UUID) (*auth.Session, error) {
	return s.apiSession, s.apiErr
}

// ── helpers ───────────────────────────────────────────────────────────────────

func appWith(svc middleware.SessionValidator) *fiber.App {
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(svc), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func request(t *testing.T, app *fiber.App, token, tenantID string) int {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode
}

func validSession() *auth.Session {
	return &auth.Session{
		Token:     "tok",
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Roles:     []string{"role:finance.viewer"},
		IssuedAt:  time.Now().Add(-time.Minute),
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestRequireAuth_MissingToken_Returns401(t *testing.T) {
	svc := &stubValidator{}
	status := request(t, appWith(svc), "", "")
	assert.Equal(t, 401, status)
}

func TestRequireAuth_ValidToken_Returns200(t *testing.T) {
	svc := &stubValidator{session: validSession()}
	status := request(t, appWith(svc), "valid-token", "")
	assert.Equal(t, 200, status)
}

func TestRequireAuth_ExpiredToken_Falls_Through_To401(t *testing.T) {
	svc := &stubValidator{
		sessionErr: &runtime.BusinessError{Code: "iam.session.not_found", Status: 401, Message: "not found"},
		apiErr:     &runtime.BusinessError{Code: "iam.session.not_found", Status: 401, Message: "not found"},
	}
	status := request(t, appWith(svc), "expired-token", uuid.New().String())
	assert.Equal(t, 401, status)
}

func TestRequireAuth_RedisDown_Returns503(t *testing.T) {
	svc := &stubValidator{
		sessionErr: &runtime.BusinessError{Code: "iam.service_unavailable", Status: 503, Message: "redis down"},
	}
	status := request(t, appWith(svc), "any-token", "")
	assert.Equal(t, 503, status)
}

func TestRequireAuth_APITokenFallback_Returns200(t *testing.T) {
	tenantID := uuid.New()
	svc := &stubValidator{
		// human session lookup returns 401
		sessionErr: &runtime.BusinessError{Code: "iam.session.not_found", Status: 401, Message: "not found"},
		// API token lookup succeeds
		apiSession: &auth.Session{
			Token:            "api-key",
			ServiceAccountID: uuid.New(),
			TenantID:         tenantID,
			Roles:            []string{"role:api-client"},
			IssuedAt:         time.Now().Add(-time.Minute),
			ExpiresAt:        time.Now().Add(time.Hour),
		},
	}
	status := request(t, appWith(svc), "api-key", tenantID.String())
	assert.Equal(t, 200, status)
}

func TestRequireAuth_APITokenFallback_NoTenantHeader_Returns401(t *testing.T) {
	svc := &stubValidator{
		sessionErr: &runtime.BusinessError{Code: "iam.session.not_found", Status: 401, Message: "not found"},
	}
	// no X-Tenant-ID header → cannot attempt API token
	status := request(t, appWith(svc), "api-key", "")
	assert.Equal(t, 401, status)
}
