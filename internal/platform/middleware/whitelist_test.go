package middleware

import (
	"reflect"
	"testing"
)

// TestNewEndpointWhitelist tests whitelist creation
func TestNewEndpointWhitelist(t *testing.T) {
	patterns := []string{
		"GET /health*",
		"GET /api/v1/health*",
		"GET /swagger-ui/*",
	}

	exactMatches := []string{
		"POST /api/v1/auth/login",
		"GET /api/v1/version",
		"GET /health",
	}

	whitelist, err := NewEndpointWhitelist(patterns, exactMatches)
	if err != nil {
		t.Fatalf("Failed to create whitelist: %v", err)
	}

	// Check exact matches
	if len(whitelist.exactMatches) != len(exactMatches) {
		t.Errorf("Expected %d exact matches, got %d", len(exactMatches), len(whitelist.exactMatches))
	}

	for _, endpoint := range exactMatches {
		if !whitelist.exactMatches[endpoint] {
			t.Errorf("Expected exact match %s not found", endpoint)
		}
	}

	// Check patterns
	if len(whitelist.patterns) != len(patterns) {
		t.Errorf("Expected %d patterns, got %d", len(patterns), len(whitelist.patterns))
	}

	// Check compiled patterns (should match valid patterns)
	if len(whitelist.compiled) != len(patterns) {
		t.Errorf("Expected %d compiled patterns, got %d", len(patterns), len(whitelist.compiled))
	}
}

// TestIsPublicEndpoint tests endpoint whitelist matching
func TestIsPublicEndpoint(t *testing.T) {
	patterns := []string{
		"GET /health*",
		"GET /api/v1/health*",
		"GET /swagger-ui/*",
		"GET /debug/*",
	}

	exactMatches := []string{
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/refresh",
		"GET /api/v1/version",
		"GET /health",
		"POST /api/v1/tenant/onboard",
	}

	whitelist, err := NewEndpointWhitelist(patterns, exactMatches)
	if err != nil {
		t.Fatalf("Failed to create whitelist: %v", err)
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		// Exact matches
		{
			name:     "Login endpoint",
			method:   "POST",
			path:     "/api/v1/auth/login",
			expected: true,
		},
		{
			name:     "Refresh endpoint",
			method:   "POST",
			path:     "/api/v1/auth/refresh",
			expected: true,
		},
		{
			name:     "Version endpoint",
			method:   "GET",
			path:     "/api/v1/version",
			expected: true,
		},
		{
			name:     "Health endpoint",
			method:   "GET",
			path:     "/health",
			expected: true,
		},
		{
			name:     "Tenant onboard",
			method:   "POST",
			path:     "/api/v1/tenant/onboard",
			expected: true,
		},

		// Pattern matches
		{
			name:     "Health check with path",
			method:   "GET",
			path:     "/health/check",
			expected: true,
		},
		{
			name:     "API health endpoint",
			method:   "GET",
			path:     "/api/v1/health",
			expected: true,
		},
		{
			name:     "API health with sub-path",
			method:   "GET",
			path:     "/api/v1/health/detailed",
			expected: true,
		},
		{
			name:     "Swagger UI main page",
			method:   "GET",
			path:     "/swagger-ui/",
			expected: true,
		},
		{
			name:     "Swagger UI asset",
			method:   "GET",
			path:     "/swagger-ui/index.html",
			expected: true,
		},
		{
			name:     "Debug endpoint",
			method:   "GET",
			path:     "/debug/pprof",
			expected: true,
		},

		// Protected endpoints (should require tenant)
		{
			name:     "Finance accounts",
			method:   "GET",
			path:     "/api/v1/finance/accounts",
			expected: false,
		},
		{
			name:     "User management",
			method:   "GET",
			path:     "/api/v1/users",
			expected: false,
		},
		{
			name:     "Transaction creation",
			method:   "POST",
			path:     "/api/v1/finance/transactions",
			expected: false,
		},
		{
			name:     "Logout endpoint",
			method:   "POST",
			path:     "/api/v1/auth/logout",
			expected: false,
		},
		{
			name:     "Admin endpoints",
			method:   "GET",
			path:     "/api/v1/admin/users",
			expected: false,
		},

		// Case sensitivity tests
		{
			name:     "Lowercase method",
			method:   "get",
			path:     "/health",
			expected: true,
		},
		{
			name:     "Mixed case method",
			method:   "Get",
			path:     "/health",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := whitelist.IsPublicEndpoint(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("IsPublicEndpoint(%s, %s) = %v, expected %v",
					tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsPublicEndpoint_NilWhitelist tests behavior with nil whitelist
func TestIsPublicEndpoint_NilWhitelist(t *testing.T) {
	var whitelist *EndpointWhitelist = nil

	// With nil whitelist, all endpoints should require tenant context
	result := whitelist.IsPublicEndpoint("GET", "/health")
	if result != false {
		t.Errorf("Expected false for nil whitelist, got %v", result)
	}
}

// TestLoadWhitelistFromYAML tests YAML configuration loading
func TestLoadWhitelistFromYAML(t *testing.T) {
	yamlConfig := `
public_endpoints:
  patterns:
    - "GET /health*"
    - "GET /api/v1/health*"
    - "GET /swagger-ui/*"
  exact_matches:
    - "POST /api/v1/auth/login"
    - "GET /api/v1/version"
    - "GET /health"

validation:
  require_approval: true
  change_tracking: true
  security_review: true
`

	whitelist, err := LoadWhitelistFromYAML([]byte(yamlConfig))
	if err != nil {
		t.Fatalf("Failed to load whitelist from YAML: %v", err)
	}

	// Test that it loaded correctly
	if !whitelist.IsPublicEndpoint("POST", "/api/v1/auth/login") {
		t.Error("Expected login endpoint to be public")
	}

	if !whitelist.IsPublicEndpoint("GET", "/health/check") {
		t.Error("Expected health check to match pattern")
	}

	if whitelist.IsPublicEndpoint("GET", "/api/v1/finance/accounts") {
		t.Error("Expected finance endpoint to be protected")
	}
}

// TestDefaultWhitelist tests the default whitelist creation
func TestDefaultWhitelist(t *testing.T) {
	whitelist := DefaultWhitelist()
	if whitelist == nil {
		t.Fatal("DefaultWhitelist returned nil")
	}

	// Test some expected default endpoints
	expectedPublic := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/api/v1/health"},
		{"POST", "/api/v1/auth/login"},
		{"GET", "/api/v1/version"},
		{"GET", "/swagger-ui/index.html"},
	}

	for _, endpoint := range expectedPublic {
		if !whitelist.IsPublicEndpoint(endpoint.method, endpoint.path) {
			t.Errorf("Expected default endpoint %s %s to be public",
				endpoint.method, endpoint.path)
		}
	}

	// Test that protected endpoints are not public
	expectedProtected := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/finance/accounts"},
		{"POST", "/api/v1/finance/transactions"},
		{"GET", "/api/v1/users"},
	}

	for _, endpoint := range expectedProtected {
		if whitelist.IsPublicEndpoint(endpoint.method, endpoint.path) {
			t.Errorf("Expected endpoint %s %s to be protected",
				endpoint.method, endpoint.path)
		}
	}
}

// TestGetStats tests whitelist statistics
func TestGetStats(t *testing.T) {
	patterns := []string{"GET /health*", "GET /api/*"}
	exactMatches := []string{"POST /login", "GET /version"}

	whitelist, err := NewEndpointWhitelist(patterns, exactMatches)
	if err != nil {
		t.Fatalf("Failed to create whitelist: %v", err)
	}

	stats := whitelist.GetStats()

	expected := map[string]int{
		"exact_matches":     2,
		"patterns":          2,
		"compiled_patterns": 2,
	}

	if !reflect.DeepEqual(stats, expected) {
		t.Errorf("Expected stats %v, got %v", expected, stats)
	}

	// Test nil whitelist
	var nilWhitelist *EndpointWhitelist = nil
	nilStats := nilWhitelist.GetStats()

	expectedNil := map[string]int{
		"exact_matches":     0,
		"patterns":          0,
		"compiled_patterns": 0,
	}

	if !reflect.DeepEqual(nilStats, expectedNil) {
		t.Errorf("Expected nil stats %v, got %v", expectedNil, nilStats)
	}
}

// TestValidateConfiguration tests security validation
func TestValidateConfiguration(t *testing.T) {
	// Test with potentially insecure configuration
	patterns := []string{
		"*",            // Overly broad
		"GET /admin/*", // Potentially insecure admin access
	}

	exactMatches := []string{
		"GET /admin*",      // Contains wildcard in exact match
		"GET /admin/users", // Admin endpoint
	}

	whitelist, err := NewEndpointWhitelist(patterns, exactMatches)
	if err != nil {
		t.Fatalf("Failed to create whitelist: %v", err)
	}

	issues := whitelist.ValidateConfiguration()

	if len(issues) == 0 {
		t.Error("Expected validation issues but got none")
	}

	// Check that specific issues are detected
	hasWildcardIssue := false
	hasBroadPatternIssue := false
	hasAdminIssue := false

	for _, issue := range issues {
		if containsString(issue, "wildcard") {
			hasWildcardIssue = true
		}
		if containsString(issue, "broad pattern") {
			hasBroadPatternIssue = true
		}
		if containsString(issue, "admin") {
			hasAdminIssue = true
		}
	}

	if !hasWildcardIssue {
		t.Error("Expected wildcard issue to be detected")
	}
	if !hasBroadPatternIssue {
		t.Error("Expected broad pattern issue to be detected")
	}
	if !hasAdminIssue {
		t.Error("Expected admin security issue to be detected")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		s[0:len(substr)] == substr[:len(substr)] ||
		(len(s) > len(substr) && containsString(s[1:], substr))
}
