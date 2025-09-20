package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// TenantMiddleware creates native HTTP middleware for tenant context injection
// This replaces the Gin-based middleware with a standard http.Handler approach
func TenantMiddleware(
	tenantService tenant.Service,
	store db.Store,
	whitelist *EndpointWhitelist,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Check if this is a public endpoint that doesn't require tenant context
			if whitelist != nil && whitelist.IsPublicEndpoint(r.Method, r.URL.Path) {
				logger.Debug("Public endpoint, bypassing tenant middleware", logger.Fields{
					"method": r.Method,
					"path":   r.URL.Path,
				})
				next.ServeHTTP(w, r)
				return
			}
			
			// 2. Special case: Tenant creation endpoints don't require existing tenant context
			if (r.Method == "POST" && r.URL.Path == "/api/v1/tenants") ||
			   (r.Method == "POST" && r.URL.Path == "/api/v1/tenant/onboard") {
				logger.Debug("Tenant creation endpoint, bypassing tenant middleware", logger.Fields{
					"method": r.Method,
					"path":   r.URL.Path,
				})
				next.ServeHTTP(w, r)
				return
			}

			// 3. Extract tenant ID from request (header priority, subdomain fallback)
			tenantIDStr, err := extractTenantID(r)
			if err != nil {
				logger.Warn("Failed to extract tenant ID", logger.Fields{
					"error":  err.Error(),
					"method": r.Method,
					"path":   r.URL.Path,
					"host":   r.Host,
				})
				writeErrorResponse(w, "Tenant resolution failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			// 3. Parse tenant ID as UUID
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				logger.Warn("Invalid tenant ID format", logger.Fields{
					"tenant_id": tenantIDStr,
					"error":     err.Error(),
				})
				writeErrorResponse(w, "Invalid tenant ID format", http.StatusBadRequest)
				return
			}

			// 4. Validate tenant exists using tenant service
			tenantEntity, err := tenantService.GetTenantByID(ctx, tenantID)
			if err != nil {
				if errors.Is(err, sharedErrors.ErrNotFound) {
					logger.Warn("Tenant not found", logger.Fields{
						"tenant_id": tenantID.String(),
					})
					writeErrorResponse(w, "Tenant not found", http.StatusNotFound)
					return
				}

				logger.Error("Failed to validate tenant", logger.Fields{
					"tenant_id": tenantID.String(),
					"error":     err.Error(),
				})
				writeErrorResponse(w, "Tenant validation failed", http.StatusInternalServerError)
				return
			}

			// 5. Set tenant context using shared utilities
			ctx = shared.WithTenantID(ctx, tenantID)

			// 6. Set database session variable for Row-Level Security (RLS)
			if err := store.SetTenantContextFromCtx(ctx); err != nil {
				logger.Error("Failed to set database tenant context", logger.Fields{
					"tenant_id": tenantID.String(),
					"error":     err.Error(),
				})
				writeErrorResponse(w, "Database context setup failed", http.StatusInternalServerError)
				return
			}

			// 7. Log successful tenant context setup
			logger.Debug("Tenant context established", logger.Fields{
				"tenant_id":   tenantID.String(),
				"tenant_name": tenantEntity.Name,
				"method":      r.Method,
				"path":        r.URL.Path,
			})

			// 8. Continue to next handler with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTenantID extracts tenant ID from HTTP request
// Priority: 1) X-Tenant-ID header, 2) Subdomain extraction
func extractTenantID(r *http.Request) (string, error) {
	// Priority 1: Check X-Tenant-ID header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			logger.Debug("Tenant ID extracted from header", logger.Fields{
				"tenant_id": tenantID,
			})
			return tenantID, nil
		}
	}

	// Priority 2: Extract from subdomain
	tenantID, err := extractFromSubdomain(r.Host)
	if err != nil {
		return "", fmt.Errorf("no valid tenant found in header or subdomain: %w", err)
	}

	logger.Debug("Tenant ID extracted from subdomain", logger.Fields{
		"tenant_id": tenantID,
		"host":      r.Host,
	})
	return tenantID, nil
}

// extractFromSubdomain extracts tenant ID from subdomain
// Supports patterns like:
// - bo.tenant1.domain.com → tenant1
// - portal.tenant2.domain.com → tenant2
// - tenant3.domain.com → tenant3
func extractFromSubdomain(host string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("empty host")
	}

	// Remove port if present (e.g., localhost:8080)
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Split by dots to analyze subdomain structure
	parts := strings.Split(strings.ToLower(host), ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid subdomain format: need at least 3 parts, got %d", len(parts))
	}

	// Handle known prefixes (bo, portal)
	if parts[0] == "bo" || parts[0] == "portal" {
		if len(parts) < 4 {
			return "", fmt.Errorf("missing tenant in %s subdomain format", parts[0])
		}
		// For bo.tenant1.domain.com, tenant is parts[1]
		return parts[1], nil
	}

	// Handle direct tenant subdomain (tenant1.domain.com)
	// First part is the tenant ID
	tenantCandidate := parts[0]

	// Basic validation: tenant should not be common service subdomains
	commonSubdomains := map[string]bool{
		"www": true, "api": true, "admin": true, "app": true,
		"mail": true, "ftp": true, "test": true, "dev": true,
		"staging": true, "prod": true, "production": true,
	}

	if commonSubdomains[tenantCandidate] {
		return "", fmt.Errorf("invalid tenant subdomain: %s is a reserved subdomain", tenantCandidate)
	}

	return tenantCandidate, nil
}

// writeErrorResponse writes a standardized JSON error response
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write simple JSON error response
	errorResponse := fmt.Sprintf(`{"error": "%s", "status": %d}`,
		strings.ReplaceAll(message, `"`, `\"`), statusCode)
	w.Write([]byte(errorResponse))
}

// Legacy context key constants for backward compatibility
// These match the original Gin middleware implementation
const (
	TenantIDKey TenantContextKey = "tenant_id"
	TenantKey   TenantContextKey = "tenant"
)

type TenantContextKey string

// Legacy helper functions for backward compatibility
// These are kept to ensure existing code continues to work

// GetTenantFromContext retrieves tenant entity from context (if available)
// Note: This requires the tenant entity to be stored in context, which
// the new middleware doesn't do by default to avoid unnecessary database calls
func GetTenantFromContext(ctx context.Context) (*tenant.Tenant, bool) {
	t, ok := ctx.Value(TenantKey).(*tenant.Tenant)
	return t, ok
}

// GetTenantIDFromContext retrieves tenant ID from context (legacy format)
// For new code, use shared.GetTenantID() instead
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	// First try the new shared context approach
	if tenantID, ok := shared.GetTenantID(ctx); ok {
		return tenantID.String(), true
	}

	// Fall back to legacy format
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}
