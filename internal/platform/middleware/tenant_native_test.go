package middleware

import (
	"net/http/httptest"
	"testing"
)

// TestExtractTenantID tests the tenant ID extraction logic
func TestExtractTenantID(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		host        string
		expected    string
		expectError bool
	}{
		// Header-based extraction (priority 1)
		{
			name:     "Valid UUID in header",
			headers:  map[string]string{"X-Tenant-ID": "550e8400-e29b-41d4-a716-446655440000"},
			host:     "bo.tenant1.awoerp.com",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Header with whitespace",
			headers:  map[string]string{"X-Tenant-ID": "  550e8400-e29b-41d4-a716-446655440000  "},
			host:     "bo.tenant1.awoerp.com",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Empty header value",
			headers:  map[string]string{"X-Tenant-ID": ""},
			host:     "bo.tenant1.awoerp.com",
			expected: "tenant1", // Should fall back to subdomain
		},
		{
			name:     "Whitespace-only header",
			headers:  map[string]string{"X-Tenant-ID": "   "},
			host:     "bo.tenant1.awoerp.com",
			expected: "tenant1", // Should fall back to subdomain
		},

		// Subdomain-based extraction (fallback)
		{
			name:     "BO subdomain with tenant",
			host:     "bo.tenant1.awoerp.com",
			expected: "tenant1",
		},
		{
			name:     "Portal subdomain with tenant",
			host:     "portal.tenant2.awoerp.com",
			expected: "tenant2",
		},
		{
			name:     "Direct tenant subdomain",
			host:     "tenant3.awoerp.com",
			expected: "tenant3",
		},
		{
			name:     "Subdomain with port",
			host:     "bo.tenant4.awoerp.com:8080",
			expected: "tenant4",
		},
		{
			name:     "Case insensitive subdomain",
			host:     "BO.TENANT5.AWOERP.COM",
			expected: "tenant5",
		},

		// Error cases
		{
			name:        "Missing tenant in BO subdomain",
			host:        "bo.awoerp.com",
			expectError: true,
		},
		{
			name:        "Invalid subdomain format",
			host:        "awoerp.com",
			expectError: true,
		},
		{
			name:        "Reserved subdomain",
			host:        "www.awoerp.com",
			expectError: true,
		},
		{
			name:        "Empty host",
			host:        "",
			expectError: true,
		},
		{
			name:        "Localhost without tenant",
			host:        "localhost:8080",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Host = tt.host

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result, err := extractTenantID(req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected tenant %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestExtractFromSubdomain tests subdomain extraction specifically
func TestExtractFromSubdomain(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		expected    string
		expectError bool
	}{
		// Valid patterns
		{
			name:     "BO subdomain pattern",
			host:     "bo.tenant1.awoerp.com",
			expected: "tenant1",
		},
		{
			name:     "Portal subdomain pattern",
			host:     "portal.tenant2.awoerp.com",
			expected: "tenant2",
		},
		{
			name:     "Direct tenant pattern",
			host:     "tenant3.awoerp.com",
			expected: "tenant3",
		},
		{
			name:     "With port number",
			host:     "tenant4.awoerp.com:8080",
			expected: "tenant4",
		},
		{
			name:     "Case insensitive",
			host:     "TENANT5.AWOERP.COM",
			expected: "tenant5",
		},
		{
			name:     "Complex domain",
			host:     "bo.tenant6.staging.awoerp.com",
			expected: "tenant6",
		},

		// Error cases
		{
			name:        "Empty host",
			host:        "",
			expectError: true,
		},
		{
			name:        "Too few parts",
			host:        "awoerp.com",
			expectError: true,
		},
		{
			name:        "Missing tenant in BO pattern",
			host:        "bo.awoerp.com",
			expectError: true,
		},
		{
			name:        "Missing tenant in portal pattern",
			host:        "portal.awoerp.com",
			expectError: true,
		},
		{
			name:        "Reserved subdomain www",
			host:        "www.awoerp.com",
			expectError: true,
		},
		{
			name:        "Reserved subdomain api",
			host:        "api.awoerp.com",
			expectError: true,
		},
		{
			name:        "Reserved subdomain admin",
			host:        "admin.awoerp.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractFromSubdomain(tt.host)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected tenant %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestWriteErrorResponse tests error response formatting
func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		statusCode     int
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Bad Request error",
			message:        "Invalid tenant ID",
			statusCode:     400,
			expectedStatus: 400,
			expectedBody:   `{"error": "Invalid tenant ID", "status": 400}`,
		},
		{
			name:           "Not Found error",
			message:        "Tenant not found",
			statusCode:     404,
			expectedStatus: 404,
			expectedBody:   `{"error": "Tenant not found", "status": 404}`,
		},
		{
			name:           "Internal Server Error",
			message:        "Database connection failed",
			statusCode:     500,
			expectedStatus: 500,
			expectedBody:   `{"error": "Database connection failed", "status": 500}`,
		},
		{
			name:           "Message with quotes",
			message:        `Invalid "tenant" format`,
			statusCode:     400,
			expectedStatus: 400,
			expectedBody:   `{"error": "Invalid \"tenant\" format", "status": 400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			writeErrorResponse(w, tt.message, tt.statusCode)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, body)
			}
		})
	}
}

