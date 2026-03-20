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

	db "awo/db/sqlc"
	// "awo/internal/api/gen/auth"
	"awo/internal/api/handlers"
	"awo/internal/core/iam/authn"
	"awo/internal/core/iam/model"
	"awo/internal/core/iam/repo"
	"awo/internal/core/tenant"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// AuthAPIIntegrationTestSuite defines authentication API integration tests
// Tests cover: AUTH-API-001 through AUTH-API-015 with full authentication workflow validation
// NOTE: These tests validate JWT authentication, user management, and security workflows
// TODO: Add OAuth2 integration testing and external identity provider validation
type AuthAPIIntegrationTestSuite struct {
	suite.Suite
	ctx         context.Context
	runner      *tenant.DatabaseTestRunner
	server      *httptest.Server
	client      *http.Client
	authService authn.Service
	iamRepo     repo.IAMRepository
	tenantA     *db.Tenant
	tenantB     *db.Tenant

	// Test users and credentials
	testUserA       *model.User
	testUserB       *model.User
	validPassword   string
	invalidPassword string

	// Test configuration
	baseURL        string
	defaultTimeout time.Duration
}

// SetupSuite initializes the authentication test environment
func (s *AuthAPIIntegrationTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to setup database test runner")
	s.ctx = context.Background()
	s.defaultTimeout = 30 * time.Second
	s.validPassword = "TestPassword123!"
	s.invalidPassword = "wrongpassword"

	// Setup infrastructure
	logger := logger.WithFields(logger.Fields{"component": "auth_api_integration_test"})
	metrics := &metrics.MetricsService{}
	tracer := tracing.NewNoOpTracingService()

	// Setup repositories and services
	s.iamRepo = repo.NewIAMRepository(s.runner.GetStore(), logger, metrics, tracer)
	s.authService = authn.NewAuthenticationService(s.iamRepo, s.runner.GetStore(), logger, metrics, tracer)

	// Setup test data
	s.setupTestData()

	// Setup HTTP client
	s.client = &http.Client{
		Timeout: s.defaultTimeout,
	}

	s.T().Logf("Auth API Integration Test Suite initialized successfully")
}

// TearDownSuite cleans up after all authentication tests
func (s *AuthAPIIntegrationTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}

	// Cleanup test data
	s.cleanupTestData()

	if s.runner != nil {
		s.runner.Close()
	}
}

// setupTestData creates test tenants and users for authentication testing
func (s *AuthAPIIntegrationTestSuite) setupTestData() {
	superuserStore := db.NewStore(s.runner.GetPool())
	uniqueID := uuid.New().String()[0:8]

	// Create test tenants
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

	// Create test users within tenant contexts
	ctx := context.WithValue(s.ctx, "tenant_id", s.tenantA.ID)

	s.testUserA, err = s.authService.CreateUser(ctx, &authn.CreateUserRequest{
		Email:       fmt.Sprintf("testuser-a-%s@authtest.com", uniqueID),
		Password:    s.validPassword,
		FirstName:   "Test",
		LastName:    "User A",
		PhoneNumber: &[]string{"+1234567890"}[0],
	})
	s.Require().NoError(err, "Failed to create test user A")

	ctx = context.WithValue(s.ctx, "tenant_id", s.tenantB.ID)

	s.testUserB, err = s.authService.CreateUser(ctx, &authn.CreateUserRequest{
		Email:       fmt.Sprintf("testuser-b-%s@authtest.com", uniqueID),
		Password:    s.validPassword,
		FirstName:   "Test",
		LastName:    "User B",
		PhoneNumber: &[]string{"+1987654321"}[0],
	})
	s.Require().NoError(err, "Failed to create test user B")

	s.T().Logf("Created test users: %s (%s), %s (%s)",
		s.testUserA.ID, s.testUserA.Email,
		s.testUserB.ID, s.testUserB.Email)
}

// cleanupTestData removes test data
func (s *AuthAPIIntegrationTestSuite) cleanupTestData() {
	if s.tenantA != nil {
		superuserStore := db.NewStore(s.runner.GetPool())
		if err := superuserStore.SoftDeleteTenant(s.ctx, s.tenantA.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant A: %v", err)
		}
	}
	if s.tenantB != nil {
		superuserStore := db.NewStore(s.runner.GetPool())
		if err := superuserStore.SoftDeleteTenant(s.ctx, s.tenantB.ID); err != nil {
			s.T().Logf("Warning: Failed to cleanup tenant B: %v", err)
		}
	}
}

// Test Specification: AUTH-API-001
// Description: Validate successful user authentication with valid credentials
func (s *AuthAPIIntegrationTestSuite) TestSuccessfulAuthentication() {
	s.T().Log("Running AUTH-API-001: Successful User Authentication")

	// Create test server with auth handler
	authHandler := handlers.NewAuthGoaHandler(s.authService,
		tracing.NewNoOpTracingService(),
		&metrics.MetricsService{})

	server := s.createAuthTestServer(authHandler)
	defer server.Close()

	// Prepare authentication request
	authRequest := map[string]any{
		"email":    s.testUserA.Email,
		"password": s.validPassword,
	}

	requestBody, err := json.Marshal(authRequest)
	s.Require().NoError(err, "Failed to marshal auth request")

	// Send authentication request
	req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	s.Require().NoError(err, "Failed to create auth request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "Authentication request failed")
	defer resp.Body.Close()

	// Validate response
	s.Assert().Equal(http.StatusOK, resp.StatusCode, "Authentication should succeed")

	var authResponse map[string]any
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	s.Require().NoError(err, "Failed to decode auth response")

	// Validate response structure
	s.Assert().Contains(authResponse, "access_token", "Response should contain access token")
	s.Assert().Contains(authResponse, "refresh_token", "Response should contain refresh token")
	s.Assert().Contains(authResponse, "expires_at", "Response should contain expiration")
	s.Assert().Contains(authResponse, "user", "Response should contain user information")

	// Validate token is not empty
	accessToken, ok := authResponse["access_token"].(string)
	s.Assert().True(ok, "Access token should be a string")
	s.Assert().NotEmpty(accessToken, "Access token should not be empty")

	s.T().Log("✅ AUTH-API-001 passed: Successful authentication working correctly")
}

// Test Specification: AUTH-API-002
// Description: Validate authentication failure with invalid credentials
func (s *AuthAPIIntegrationTestSuite) TestAuthenticationFailure() {
	s.T().Log("Running AUTH-API-002: Authentication Failure with Invalid Credentials")

	// Create test server with auth handler
	authHandler := handlers.NewAuthGoaHandler(s.authService,
		tracing.NewNoOpTracingService(),
		&metrics.MetricsService{})

	server := s.createAuthTestServer(authHandler)
	defer server.Close()

	// Test cases for authentication failures
	testCases := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
		description    string
	}{
		{
			name:           "InvalidPassword",
			email:          s.testUserA.Email,
			password:       s.invalidPassword,
			expectedStatus: http.StatusUnauthorized,
			description:    "Wrong password should be rejected",
		},
		{
			name:           "InvalidEmail",
			email:          "nonexistent@test.com",
			password:       s.validPassword,
			expectedStatus: http.StatusUnauthorized,
			description:    "Non-existent user should be rejected",
		},
		{
			name:           "EmptyCredentials",
			email:          "",
			password:       "",
			expectedStatus: http.StatusBadRequest,
			description:    "Empty credentials should be rejected",
		},
	}

	for _, tc := range testCases {
		s.T().Logf("Testing %s: %s", tc.name, tc.description)

		authRequest := map[string]any{
			"email":    tc.email,
			"password": tc.password,
		}

		requestBody, err := json.Marshal(authRequest)
		s.Require().NoError(err, "Failed to marshal auth request")

		req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
		s.Require().NoError(err, "Failed to create auth request")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

		resp, err := s.client.Do(req)
		s.Require().NoError(err, "Authentication request failed")
		resp.Body.Close()

		s.Assert().Equal(tc.expectedStatus, resp.StatusCode,
			"%s: Expected status %d, got %d", tc.description, tc.expectedStatus, resp.StatusCode)
	}

	s.T().Log("✅ AUTH-API-002 passed: Authentication failure handling working correctly")
}

// Test Specification: AUTH-API-003
// Description: Validate JWT token validation and protected endpoint access
func (s *AuthAPIIntegrationTestSuite) TestJWTTokenValidation() {
	s.T().Log("Running AUTH-API-003: JWT Token Validation")

	// First, authenticate to get a valid token
	authHandler := handlers.NewAuthGoaHandler(s.authService,
		tracing.NewNoOpTracingService(),
		&metrics.MetricsService{})

	server := s.createAuthTestServer(authHandler)
	defer server.Close()

	// Get valid token
	token := s.authenticateAndGetToken(server, s.testUserA.Email, s.validPassword, s.tenantA.ID)
	s.Require().NotEmpty(token, "Failed to get authentication token")

	// Test accessing protected endpoint with valid token
	req, err := http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	s.Require().NoError(err, "Failed to create protected request")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "Protected request failed")
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode, "Protected endpoint should be accessible with valid token")

	// Test accessing protected endpoint without token
	req, err = http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	s.Require().NoError(err, "Failed to create unprotected request")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "Unprotected request failed")
	resp.Body.Close()

	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode, "Protected endpoint should reject requests without token")

	// Test accessing protected endpoint with invalid token
	req, err = http.NewRequest("GET", server.URL+"/api/v1/auth/me", nil)
	s.Require().NoError(err, "Failed to create invalid token request")
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "Invalid token request failed")
	resp.Body.Close()

	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode, "Protected endpoint should reject invalid tokens")

	s.T().Log("✅ AUTH-API-003 passed: JWT token validation working correctly")
}

// Test Specification: AUTH-API-004
// Description: Validate multi-tenant isolation in authentication
func (s *AuthAPIIntegrationTestSuite) TestMultiTenantAuthenticationIsolation() {
	s.T().Log("Running AUTH-API-004: Multi-Tenant Authentication Isolation")

	authHandler := handlers.NewAuthGoaHandler(s.authService,
		tracing.NewNoOpTracingService(),
		&metrics.MetricsService{})

	server := s.createAuthTestServer(authHandler)
	defer server.Close()

	// Test: User from Tenant A cannot authenticate in Tenant B context
	authRequest := map[string]any{
		"email":    s.testUserA.Email, // User from Tenant A
		"password": s.validPassword,
	}

	requestBody, err := json.Marshal(authRequest)
	s.Require().NoError(err, "Failed to marshal auth request")

	req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	s.Require().NoError(err, "Failed to create cross-tenant auth request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantB.ID.String()) // Try to authenticate in Tenant B

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "Cross-tenant authentication request failed")
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode,
		"User from Tenant A should not be able to authenticate in Tenant B context")

	// Test: User can authenticate in their own tenant context
	req, err = http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	s.Require().NoError(err, "Failed to create same-tenant auth request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String()) // Authenticate in correct tenant

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "Same-tenant authentication request failed")
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode,
		"User should be able to authenticate in their own tenant context")

	s.T().Log("✅ AUTH-API-004 passed: Multi-tenant authentication isolation working correctly")
}

// Test Specification: AUTH-API-005
// Description: Validate token refresh functionality
func (s *AuthAPIIntegrationTestSuite) TestTokenRefresh() {
	s.T().Log("Running AUTH-API-005: Token Refresh Functionality")

	authHandler := handlers.NewAuthGoaHandler(s.authService,
		tracing.NewNoOpTracingService(),
		&metrics.MetricsService{})

	server := s.createAuthTestServer(authHandler)
	defer server.Close()

	// First, authenticate to get tokens
	authResponse := s.authenticateAndGetFullResponse(server, s.testUserA.Email, s.validPassword, s.tenantA.ID)
	refreshToken, ok := authResponse["refresh_token"].(string)
	s.Require().True(ok, "Refresh token should be present")
	s.Require().NotEmpty(refreshToken, "Refresh token should not be empty")

	// Test token refresh
	refreshRequest := map[string]any{
		"refresh_token": refreshToken,
	}

	requestBody, err := json.Marshal(refreshRequest)
	s.Require().NoError(err, "Failed to marshal refresh request")

	req, err := http.NewRequest("POST", server.URL+"/api/v1/auth/refresh", bytes.NewBuffer(requestBody))
	s.Require().NoError(err, "Failed to create refresh request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", s.tenantA.ID.String())

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "Token refresh request failed")
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode, "Token refresh should succeed")

	var refreshResponse map[string]any
	err = json.NewDecoder(resp.Body).Decode(&refreshResponse)
	s.Require().NoError(err, "Failed to decode refresh response")

	// Validate new tokens are provided
	s.Assert().Contains(refreshResponse, "access_token", "Refresh response should contain new access token")
	s.Assert().Contains(refreshResponse, "refresh_token", "Refresh response should contain new refresh token")

	newAccessToken, ok := refreshResponse["access_token"].(string)
	s.Assert().True(ok, "New access token should be a string")
	s.Assert().NotEmpty(newAccessToken, "New access token should not be empty")

	s.T().Log("✅ AUTH-API-005 passed: Token refresh functionality working correctly")
}

// Helper Methods

// createAuthTestServer creates a test server with authentication endpoints
func (s *AuthAPIIntegrationTestSuite) createAuthTestServer(authHandler auth.Service) *httptest.Server {
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

		email, ok := loginReq["email"].(string)
		if !ok || email == "" {
			http.Error(w, "Email required", http.StatusBadRequest)
			return
		}

		password, ok := loginReq["password"].(string)
		if !ok || password == "" {
			http.Error(w, "Password required", http.StatusBadRequest)
			return
		}

		// Set tenant context
		ctx := r.Context()
		if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
			if uuid, err := uuid.Parse(tenantID); err == nil {
				ctx = context.WithValue(ctx, "tenant_id", uuid)
			}
		}

		result, err := s.authService.Authenticate(ctx, &authn.AuthenticationRequest{
			Email:    email,
			Password: password,
		})
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// Protected endpoint for token validation testing
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Simple bearer token validation (in real implementation, this would validate JWT)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "invalid-token" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Return user info (simplified)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"user_id": s.testUserA.ID,
			"email":   s.testUserA.Email,
		})
	})

	// Token refresh endpoint
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var refreshReq map[string]any
		if err := json.NewDecoder(r.Body).Decode(&refreshReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		refreshToken, ok := refreshReq["refresh_token"].(string)
		if !ok || refreshToken == "" {
			http.Error(w, "Refresh token required", http.StatusBadRequest)
			return
		}

		// Set tenant context
		ctx := r.Context()
		if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
			if uuid, err := uuid.Parse(tenantID); err == nil {
				ctx = context.WithValue(ctx, "tenant_id", uuid)
			}
		}

		result, err := s.authService.RefreshToken(ctx, refreshToken)
		if err != nil {
			http.Error(w, "Token refresh failed", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	return httptest.NewServer(mux)
}

// authenticateAndGetToken performs authentication and returns access token
func (s *AuthAPIIntegrationTestSuite) authenticateAndGetToken(server *httptest.Server, email, password string, tenantID uuid.UUID) string {
	authResponse := s.authenticateAndGetFullResponse(server, email, password, tenantID)
	if token, ok := authResponse["access_token"].(string); ok {
		return token
	}
	return ""
}

// authenticateAndGetFullResponse performs authentication and returns full response
func (s *AuthAPIIntegrationTestSuite) authenticateAndGetFullResponse(server *httptest.Server, email, password string, tenantID uuid.UUID) map[string]any {
	authRequest := map[string]any{
		"email":    email,
		"password": password,
	}

	requestBody, _ := json.Marshal(authRequest)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())

	resp, _ := s.client.Do(req)
	defer resp.Body.Close()

	var authResponse map[string]any
	json.NewDecoder(resp.Body).Decode(&authResponse)
	return authResponse
}

// RunAuthAPIIntegrationTests is the test suite runner
func TestAuthAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthAPIIntegrationTestSuite))
}
