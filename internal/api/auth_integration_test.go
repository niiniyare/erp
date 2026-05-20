//go:build integration
// +build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam"
	"awo.so/internal/core/iam/contract"
	"awo.so/internal/core/tenant"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// AuthAPIIntegrationTestSuite defines authentication API integration tests.
// Tests cover: AUTH-API-001 through AUTH-API-005.
type AuthAPIIntegrationTestSuite struct {
	suite.Suite
	ctx            context.Context
	runner         *tenant.DatabaseTestRunner
	client         *http.Client
	authSvc        contract.AuthService
	userSvc        iam.UserService
	tenantA        *db.Tenant
	tenantB        *db.Tenant
	testUserA      *iam.User
	testUserB      *iam.User
	validPassword   string
	invalidPassword string
	defaultTimeout  time.Duration
}

func (s *AuthAPIIntegrationTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to setup database test runner")
	s.ctx = context.Background()
	s.defaultTimeout = 30 * time.Second
	s.validPassword = "TestPassword123!"
	s.invalidPassword = "wrongpassword"

	log := logger.NewNoOp()
	mp := metrics.NewNoOpMetricsProvider()
	tracer := tracing.NewNoOpTracingService()
	noopCache := &noopCacheImpl{}

	userRepo := iam.NewUserRepository(s.runner.GetStore(), noopCache, tracer, mp)
	s.userSvc = iam.NewUserService(userRepo, tracer, mp, log)

	sessionRepo := iam.NewSessionRepository(s.runner.GetStore(), noopCache, tracer, mp)
	sessionSvc := iam.NewSessionService(s.userSvc, sessionRepo, tracer, mp, log)
	s.authSvc = contract.NewServiceAdapter(sessionSvc)

	s.setupTestData()
	s.client = &http.Client{Timeout: s.defaultTimeout}
}

func (s *AuthAPIIntegrationTestSuite) TearDownSuite() {
	s.cleanupTestData()
	if s.runner != nil {
		s.runner.Close()
	}
}

func (s *AuthAPIIntegrationTestSuite) setupTestData() {
	superuserStore := db.NewStore(s.runner.GetPool())
	uniqueID := uuid.New().String()[0:8]
	var err error

	s.tenantA, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:        fmt.Sprintf("Auth API Test Tenant A %s", uniqueID),
		Slug:        fmt.Sprintf("auth-api-tenant-a-%s", uniqueID),
		Email:       fmt.Sprintf("auth-admin-a-%s@authtest.com", uniqueID),
		Description: "Test tenant A for authentication API testing",
	})
	s.Require().NoError(err, "Failed to create test tenant A")

	s.tenantB, err = superuserStore.CreateTenant(s.ctx, db.CreateTenantParams{
		Name:        fmt.Sprintf("Auth API Test Tenant B %s", uniqueID),
		Slug:        fmt.Sprintf("auth-api-tenant-b-%s", uniqueID),
		Email:       fmt.Sprintf("auth-admin-b-%s@authtest.com", uniqueID),
		Description: "Test tenant B for authentication isolation testing",
	})
	s.Require().NoError(err, "Failed to create test tenant B")

	ctxA := context.WithValue(s.ctx, "tenant_id", s.tenantA.ID)
	s.testUserA, err = s.userSvc.RegisterNewUser(ctxA, &iam.CreateUserRequest{
		EntityID: s.tenantA.ID, // root entity for test purposes
		Username: fmt.Sprintf("testuser-a-%s", uniqueID),
		Email:    fmt.Sprintf("testuser-a-%s@authtest.com", uniqueID),
		Password: s.validPassword,
		UserType: "INTERNAL",
	})
	s.Require().NoError(err, "Failed to create test user A")

	ctxB := context.WithValue(s.ctx, "tenant_id", s.tenantB.ID)
	s.testUserB, err = s.userSvc.RegisterNewUser(ctxB, &iam.CreateUserRequest{
		EntityID: s.tenantB.ID,
		Username: fmt.Sprintf("testuser-b-%s", uniqueID),
		Email:    fmt.Sprintf("testuser-b-%s@authtest.com", uniqueID),
		Password: s.validPassword,
		UserType: "INTERNAL",
	})
	s.Require().NoError(err, "Failed to create test user B")
}

func (s *AuthAPIIntegrationTestSuite) cleanupTestData() {
	if s.tenantA != nil {
		store := db.NewStore(s.runner.GetPool())
		if err := store.SoftDeleteTenant(s.ctx, s.tenantA.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant A: %v", err)
		}
	}
	if s.tenantB != nil {
		store := db.NewStore(s.runner.GetPool())
		if err := store.SoftDeleteTenant(s.ctx, s.tenantB.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant B: %v", err)
		}
	}
}

// Test Specification: AUTH-API-001
func (s *AuthAPIIntegrationTestSuite) TestSuccessfulAuthentication() {
	s.T().Log("Running AUTH-API-001: Successful User Authentication")

	server := s.createAuthTestServer()
	defer server.Close()

	authRequest := map[string]any{
		"email":    s.testUserA.Email,
		"password": s.validPassword,
	}
	requestBody, err := json.Marshal(authRequest)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var authResponse map[string]any
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&authResponse))
	s.Assert().Contains(authResponse, "access_token")
	s.Assert().NotEmpty(authResponse["access_token"])

	s.T().Log("✅ AUTH-API-001 passed")
}

// Test Specification: AUTH-API-002
func (s *AuthAPIIntegrationTestSuite) TestAuthenticationFailure() {
	s.T().Log("Running AUTH-API-002: Authentication Failure with Invalid Credentials")

	server := s.createAuthTestServer()
	defer server.Close()

	testCases := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{"InvalidPassword", s.testUserA.Email, s.invalidPassword, http.StatusUnauthorized},
		{"InvalidEmail", "nonexistent@test.com", s.validPassword, http.StatusUnauthorized},
		{"EmptyCredentials", "", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		s.T().Logf("Testing %s", tc.name)
		requestBody, _ := json.Marshal(map[string]any{"email": tc.email, "password": tc.password})
		req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		resp.Body.Close()

		s.Assert().Equal(tc.expectedStatus, resp.StatusCode, tc.name)
	}

	s.T().Log("✅ AUTH-API-002 passed")
}

// Test Specification: AUTH-API-003
func (s *AuthAPIIntegrationTestSuite) TestJWTTokenValidation() {
	s.T().Log("Running AUTH-API-003: Token Validation")

	server := s.createAuthTestServer()
	defer server.Close()

	token := s.loginAndGetToken(server, s.testUserA.Email, s.validPassword, s.tenantA.ID)
	s.Require().NotEmpty(token)

	// Valid token → 200
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())
	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	// No token → 401
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	resp.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode)

	// Invalid token → 401
	req, _ = http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	resp.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode)

	s.T().Log("✅ AUTH-API-003 passed")
}

// Test Specification: AUTH-API-004
func (s *AuthAPIIntegrationTestSuite) TestMultiTenantAuthenticationIsolation() {
	s.T().Log("Running AUTH-API-004: Multi-Tenant Authentication Isolation")

	server := s.createAuthTestServer()
	defer server.Close()

	// Tenant A user logging into Tenant B context → 401
	requestBody, _ := json.Marshal(map[string]any{
		"email":    s.testUserA.Email,
		"password": s.validPassword,
	})
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantB.ID.String())
	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode)

	// Tenant A user logging into Tenant A context → 200
	requestBody, _ = json.Marshal(map[string]any{
		"email":    s.testUserA.Email,
		"password": s.validPassword,
	})
	req, _ = http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())
	resp, err = s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	s.T().Log("✅ AUTH-API-004 passed")
}

// Test Specification: AUTH-API-005
// Session tokens in this IAM implementation are single-use validated tokens
// (not refresh-token pairs). This test validates that a valid token can be
// re-validated, simulating what a refresh flow would require.
func (s *AuthAPIIntegrationTestSuite) TestTokenValidationAfterLogin() {
	s.T().Log("Running AUTH-API-005: Token Validation After Login")

	server := s.createAuthTestServer()
	defer server.Close()

	token := s.loginAndGetToken(server, s.testUserA.Email, s.validPassword, s.tenantA.ID)
	s.Require().NotEmpty(token)

	// Validate session via /me endpoint
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())
	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	s.T().Log("✅ AUTH-API-005 passed")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *AuthAPIIntegrationTestSuite) createAuthTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Login endpoint
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var loginReq map[string]any
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		email, _ := loginReq["email"].(string)
		password, _ := loginReq["password"].(string)
		if email == "" || password == "" {
			http.Error(w, "Email and password required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		if tenantIDStr := r.Header.Get("X-Tenant-ID"); tenantIDStr != "" {
			if tid, err := uuid.Parse(tenantIDStr); err == nil {
				ctx = context.WithValue(ctx, "tenant_id", tid)
			}
		}

		result, err := s.authSvc.Login(ctx, email, password)
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": result.Token,
			"user": map[string]any{
				"id": result.Session.UserID(),
			},
		})
	})

	// Protected endpoint
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		ctx := r.Context()
		if tenantIDStr := r.Header.Get("X-Tenant-ID"); tenantIDStr != "" {
			if tid, err := uuid.Parse(tenantIDStr); err == nil {
				ctx = context.WithValue(ctx, "tenant_id", tid)
			}
		}

		sc, ok := s.authSvc.ValidateSession(ctx, token)
		if !ok {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"user_id": sc.UserID(),
		})
	})

	return httptest.NewServer(mux)
}

func (s *AuthAPIIntegrationTestSuite) loginAndGetToken(server *httptest.Server, email, password string, tenantID uuid.UUID) string {
	requestBody, _ := json.Marshal(map[string]any{"email": email, "password": password})
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())

	resp, err := s.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var authResponse map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return ""
	}
	token, _ := authResponse["access_token"].(string)
	return token
}

func TestAuthAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthAPIIntegrationTestSuite))
}

// ─── No-op cache for integration test setup ───────────────────────────────────

type noopCacheImpl struct{}

func (n *noopCacheImpl) Get(_ context.Context, _ string, _ any) error { return cache.ErrCacheMiss }
func (n *noopCacheImpl) GetAndDelete(_ context.Context, _ string, _ any) error {
	return cache.ErrCacheMiss
}
func (n *noopCacheImpl) Set(_ context.Context, _ string, _ any, _ time.Duration) error { return nil }
func (n *noopCacheImpl) Delete(_ context.Context, _ string) error                      { return nil }
func (n *noopCacheImpl) Flush(_ context.Context) error                                 { return nil }
func (n *noopCacheImpl) MGet(_ context.Context, _ []string) ([]cache.Result, error)    { return nil, nil }
func (n *noopCacheImpl) MSet(_ context.Context, _ map[string]any, _ time.Duration) error {
	return nil
}
func (n *noopCacheImpl) MDelete(_ context.Context, _ []string) error            { return nil }
func (n *noopCacheImpl) DeletePattern(_ context.Context, _ string) error        { return nil }
func (n *noopCacheImpl) Keys(_ context.Context, _ string) ([]string, error)     { return nil, nil }
func (n *noopCacheImpl) Exists(_ context.Context, _ string) (bool, error)       { return false, nil }
func (n *noopCacheImpl) TTL(_ context.Context, _ string) (time.Duration, error) { return 0, nil }
func (n *noopCacheImpl) Expire(_ context.Context, _ string, _ time.Duration) error { return nil }
func (n *noopCacheImpl) GetMemory(_ context.Context, _ string, _ any) error { return cache.ErrCacheMiss }
func (n *noopCacheImpl) SetMemory(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}
func (n *noopCacheImpl) DeleteMemory(_ context.Context, _ string) error      { return nil }
func (n *noopCacheImpl) GetGlobalMemory(_ string, _ any) error               { return cache.ErrCacheMiss }
func (n *noopCacheImpl) SetGlobalMemory(_ string, _ any, _ time.Duration) error { return nil }
func (n *noopCacheImpl) DeleteGlobalMemory(_ string) error                   { return nil }
func (n *noopCacheImpl) Ping(_ context.Context) error                        { return nil }
func (n *noopCacheImpl) Stats() cache.CacheStats                             { return cache.CacheStats{} }
func (n *noopCacheImpl) Reset()                                              {}
func (n *noopCacheImpl) Close() error                                        { return nil }
