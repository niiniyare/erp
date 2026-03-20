package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	db "awo/db/sqlc"
	"awo/internal/core/tenant"
	"awo/internal/shared"
	sharedErrors "awo/internal/shared/errors"
	"awo/internal/shared/logger"
)

// Fiber context keys
const (
	TenantIDKey = "tenant_id"
	TenantKey   = "tenant"
	EntityKey   = "entity_id"
)

// ErrorResponse represents standardized API error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Status    int    `json:"status"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// TenantCacheEntry represents a cached tenant validation result
type TenantCacheEntry struct {
	Tenant    *tenant.Tenant
	ExpiresAt time.Time
}

// TenantCache provides thread-safe caching for tenant validation
type TenantCache struct {
	mu      sync.RWMutex
	entries map[string]*TenantCacheEntry
	ttl     time.Duration
}

// NewTenantCache creates a new tenant cache with specified TTL
func NewTenantCache(ttl time.Duration) *TenantCache {
	cache := &TenantCache{
		entries: make(map[string]*TenantCacheEntry),
		ttl:     ttl,
	}

	// Start cleanup goroutine to remove expired entries
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a tenant from cache if it exists and is not expired
func (tc *TenantCache) Get(tenantID string) (*tenant.Tenant, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	entry, exists := tc.entries[tenantID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Tenant, true
}

// Set stores a tenant in cache with expiration
func (tc *TenantCache) Set(tenantID string, t *tenant.Tenant) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.entries[tenantID] = &TenantCacheEntry{
		Tenant:    t,
		ExpiresAt: time.Now().Add(tc.ttl),
	}
}

// Invalidate removes a tenant from cache
func (tc *TenantCache) Invalidate(tenantID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	delete(tc.entries, tenantID)
}

// cleanupLoop periodically removes expired entries
func (tc *TenantCache) cleanupLoop() {
	ticker := time.NewTicker(tc.ttl)
	defer ticker.Stop()

	for range ticker.C {
		tc.cleanup()
	}
}

// cleanup removes expired entries from cache
func (tc *TenantCache) cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now()
	for id, entry := range tc.entries {
		if now.After(entry.ExpiresAt) {
			delete(tc.entries, id)
		}
	}
}

// TenantMiddlewareConfig holds configuration for tenant middleware
type TenantMiddlewareConfig struct {
	TenantService tenant.Service
	Store         db.Store
	Whitelist     *EndpointWhitelist
	CacheTTL      time.Duration // Default: 5 minutes
	EnableCache   bool          // Default: true

	// NOTE: SkipPaths allows bypassing tenant validation for specific paths
	// Add paths like ["/health", "/metrics"] that don't need tenant context
	SkipPaths []string
}

// TenantMiddleware creates Fiber middleware for tenant context injection
// This provides multi-tenant isolation with RLS (Row-Level Security) support
func TenantMiddleware(config TenantMiddlewareConfig) fiber.Handler {
	// Set defaults
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	var cache *TenantCache
	if config.EnableCache {
		cache = NewTenantCache(config.CacheTTL)
	}

	return func(c *fiber.Ctx) error {
		// Generate or retrieve request ID for tracing
		requestID := c.Get(fiber.HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set(fiber.HeaderXRequestID, requestID)
		}

		path := c.Path()
		method := c.Method()

		// 1. Bypass tenant validation for admin UI assets and static files
		if strings.HasPrefix(path, "/admin/") ||
			strings.HasPrefix(path, "/static/") ||
			path == "/favicon.ico" {
			return c.Next()
		}

		// 2. Check custom skip paths
		for _, skipPath := range config.SkipPaths {
			if path == skipPath || strings.HasPrefix(path, skipPath) {
				logger.Debug("Skipped path, bypassing tenant middleware", logger.Fields{
					"method":     method,
					"path":       path,
					"request_id": requestID,
				})
				return c.Next()
			}
		}

		// 3. Check if this is a public endpoint that doesn't require tenant context
		if config.Whitelist != nil && config.Whitelist.IsPublicEndpoint(method, path) {
			logger.Debug("Public endpoint, bypassing tenant middleware", logger.Fields{
				"method":     method,
				"path":       path,
				"request_id": requestID,
			})
			return c.Next()
		}

		// 4. Special case: Tenant creation/onboarding endpoints
		if (method == fiber.MethodPost && path == "/api/v1/tenants") ||
			(method == fiber.MethodPost && path == "/api/v1/tenant/onboard") {
			logger.Debug("Tenant creation endpoint, bypassing tenant middleware", logger.Fields{
				"method":     method,
				"path":       path,
				"request_id": requestID,
			})
			return c.Next()
		}

		// 5. Extract tenant ID from request (header priority, subdomain fallback)
		tenantIDStr, err := extractTenantID(c)
		if err != nil {
			logger.Warn("Failed to extract tenant ID", logger.Fields{
				"error":      err.Error(),
				"method":     method,
				"path":       path,
				"host":       c.Hostname(),
				"request_id": requestID,
			})
			return sendErrorResponse(c, "Tenant identification required", fiber.StatusBadRequest, requestID)
		}

		// 6. Parse tenant ID as UUID
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			logger.Warn("Invalid tenant ID format", logger.Fields{
				"tenant_id":  tenantIDStr,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return sendErrorResponse(c, "Invalid tenant ID format", fiber.StatusBadRequest, requestID)
		}

		// 7. Check cache first for tenant validation
		var tenantEntity *tenant.Tenant
		if cache != nil {
			if cached, found := cache.Get(tenantID.String()); found {
				tenantEntity = cached
				logger.Debug("Tenant retrieved from cache", logger.Fields{
					"tenant_id":  tenantID.String(),
					"request_id": requestID,
				})
			}
		}

		// 8. Validate tenant exists using tenant service (if not cached)
		if tenantEntity == nil {
			// NOTE: Circuit breaker implementation needed for production resilience
			// Recommended approach using existing infrastructure:
			// 1. Use cache.CircuitBreaker from internal/platform/cache/redis.go
			// 2. Implement fallback to cached tenant data when service is degraded
			// 3. Add health checks for tenant service availability
			// 4. Configure fail-open vs fail-closed policy based on security requirements
			tenantEntity, err = config.TenantService.GetTenantByID(c.Context(), tenantID)
			if err != nil {
				if errors.Is(err, sharedErrors.ErrNotFound) {
					logger.Warn("Tenant not found", logger.Fields{
						"tenant_id":  tenantID.String(),
						"request_id": requestID,
					})
					return sendErrorResponse(c, "Tenant not found", fiber.StatusNotFound, requestID)
				}

				logger.Error("Failed to validate tenant", logger.Fields{
					"tenant_id":  tenantID.String(),
					"error":      err.Error(),
					"request_id": requestID,
				})
				return sendErrorResponse(c, "Tenant validation failed", fiber.StatusInternalServerError, requestID)
			}

			// Cache the validated tenant
			if cache != nil {
				cache.Set(tenantID.String(), tenantEntity)
			}
		}

		// 9. Check tenant status before allowing access
		if tenantEntity.Status != tenant.StatusActive {
			logger.Warn("Inactive tenant attempted access", logger.Fields{
				"tenant_id":  tenantID.String(),
				"status":     string(tenantEntity.Status),
				"request_id": requestID,
			})

			var message string
			switch tenantEntity.Status {
			case tenant.StatusSuspended:
				message = "Tenant account is suspended"
			case tenant.StatusTrial:
				message = "Tenant account is pending activation"
			default:
				message = "Tenant account is not active"
			}

			return sendErrorResponse(c, message, fiber.StatusForbidden, requestID)
		}

		// 10. Store tenant context in Fiber locals for easy access in handlers
		c.Locals(TenantIDKey, tenantID)
		c.Locals(TenantKey, tenantIDStr)
		c.Locals(EntityKey, tenantEntity)

		// 11. Set tenant context in request context (for compatibility with shared utilities)
		ctx := shared.WithTenantID(c.Context(), tenantID)

		// NOTE: The following stores the enhanced context back to Fiber.
		// This ensures downstream handlers can access both Fiber locals and Go context.
		c.SetUserContext(ctx)

		// 12. Set database session variable for Row-Level Security (RLS)
		// NOTE: This sets the tenant context for PostgreSQL RLS policies.
		// The RLS policies should be configured as:
		// CREATE POLICY tenant_isolation ON your_table
		// USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
		//
		// Connection pooling considerations:
		// - SetTenantContextFromCtx uses session variables which persist for the connection
		// - Connection pools may reuse connections across requests
		// - The implementation handles this by setting context on each request
		if err := config.Store.SetTenantContextFromCtx(ctx); err != nil {
			logger.Error("Failed to set database tenant context", logger.Fields{
				"tenant_id":  tenantID.String(),
				"error":      err.Error(),
				"request_id": requestID,
			})
			return sendErrorResponse(c, "Database context setup failed", fiber.StatusInternalServerError, requestID)
		}

		// 13. Log successful tenant context establishment
		logger.Debug("Tenant context established", logger.Fields{
			"tenant_id":   tenantID.String(),
			"tenant_name": tenantEntity.Name,
			"method":      method,
			"path":        path,
			"request_id":  requestID,
		})

		// 14. Continue to next handler
		return c.Next()
	}
}

// extractTenantID extracts tenant ID from Fiber request context
// Priority: 1) X-Tenant-ID header, 2) Subdomain extraction, 3) Query parameter
func extractTenantID(c *fiber.Ctx) (string, error) {
	// Priority 1: Check X-Tenant-ID header
	if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			logger.Debug("Tenant ID extracted from header", logger.Fields{
				"tenant_id": tenantID,
				"source":    "header",
			})
			return tenantID, nil
		}
	}

	// Priority 2: Check query parameter (useful for webhooks, callbacks)
	// NOTE: Consider removing this if query-based tenant ID is a security concern
	if tenantID := c.Query("tenant_id"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			logger.Debug("Tenant ID extracted from query parameter", logger.Fields{
				"tenant_id": tenantID,
				"source":    "query",
			})
			return tenantID, nil
		}
	}

	// Priority 3: Extract from subdomain
	tenantID, err := extractFromSubdomain(c.Hostname())
	if err != nil {
		return "", fmt.Errorf("no valid tenant identifier found in header, query, or subdomain: %w", err)
	}

	logger.Debug("Tenant ID extracted from subdomain", logger.Fields{
		"tenant_id": tenantID,
		"host":      c.Hostname(),
		"source":    "subdomain",
	})
	return tenantID, nil
}

// extractFromSubdomain extracts tenant ID from subdomain with improved validation
// Supports patterns like:
// - bo.tenant1.domain.com → tenant1
// - portal.tenant2.domain.com → tenant2
// - tenant3.domain.com → tenant3
func extractFromSubdomain(host string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("empty hostname")
	}

	// Handle localhost and IP addresses
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return "", fmt.Errorf("localhost and IP addresses don't support subdomain-based tenant identification")
	}

	// Check if it's an IP address (IPv4 or IPv6)
	if net.ParseIP(host) != nil {
		return "", fmt.Errorf("IP addresses don't support subdomain-based tenant identification")
	}

	// Remove port if present (e.g., localhost:8080)
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Split by dots to analyze subdomain structure
	parts := strings.Split(strings.ToLower(host), ".")

	// Need at least 3 parts for subdomain extraction (subdomain.domain.tld)
	// FIXME: This logic doesn't handle country-code TLDs properly (e.g., .co.uk, .com.au)
	// TODO: Consider using a library like "golang.org/x/net/publicsuffix" for proper
	// domain parsing that handles all TLD variations correctly
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid subdomain format: need at least 3 parts (subdomain.domain.tld), got %d", len(parts))
	}

	// Handle known prefixes (bo, portal, app, etc.)
	// NOTE: Add more prefixes here as your system grows
	knownPrefixes := map[string]bool{
		"bo":     true, // Back office
		"portal": true, // Customer portal
		"app":    true, // Main application
		"api":    true, // API gateway (if using subdomain routing)
	}

	if knownPrefixes[parts[0]] {
		if len(parts) < 4 {
			return "", fmt.Errorf("missing tenant identifier after %s prefix", parts[0])
		}
		// For bo.tenant1.domain.com, tenant is parts[1]
		return parts[1], nil
	}

	// Handle direct tenant subdomain (tenant1.domain.com)
	tenantCandidate := parts[0]

	// Validate against reserved/common subdomains
	// NOTE: Keep this list updated with your infrastructure subdomains
	reservedSubdomains := map[string]bool{
		"www":        true,
		"admin":      true,
		"mail":       true,
		"ftp":        true,
		"smtp":       true,
		"pop":        true,
		"imap":       true,
		"test":       true,
		"dev":        true,
		"staging":    true,
		"prod":       true,
		"production": true,
		"demo":       true,
		"sandbox":    true,
		"cdn":        true,
		"static":     true,
		"assets":     true,
		"docs":       true,
		"blog":       true,
		"status":     true,
		"monitoring": true,
	}

	if reservedSubdomains[tenantCandidate] {
		return "", fmt.Errorf("cannot use reserved subdomain '%s' as tenant identifier", tenantCandidate)
	}

	// Additional validation: tenant ID should be alphanumeric with hyphens
	// NOTE: Adjust this regex pattern based on your tenant ID format requirements
	if !isValidTenantSubdomain(tenantCandidate) {
		return "", fmt.Errorf("invalid tenant subdomain format: '%s' (must be alphanumeric with hyphens)", tenantCandidate)
	}

	return tenantCandidate, nil
}

// isValidTenantSubdomain validates tenant subdomain format
func isValidTenantSubdomain(subdomain string) bool {
	if subdomain == "" || len(subdomain) > 63 {
		return false
	}

	// Must start and end with alphanumeric character
	if !isAlphanumeric(subdomain[0]) || !isAlphanumeric(subdomain[len(subdomain)-1]) {
		return false
	}

	// Can contain alphanumeric characters and hyphens
	for _, char := range subdomain {
		if !isAlphanumeric(byte(char)) && char != '-' {
			return false
		}
	}

	return true
}

// isAlphanumeric checks if a byte is alphanumeric
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// extractTenantIDFromHTTPRequest extracts tenant ID from standard HTTP request (for testing)
func extractTenantIDFromHTTPRequest(r *http.Request) (string, error) {
	// Priority 1: Check X-Tenant-ID header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			return tenantID, nil
		}
	}

	// Priority 2: Check query parameter
	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			return tenantID, nil
		}
	}

	// Priority 3: Extract from subdomain
	tenantID, err := extractFromSubdomain(r.Host)
	if err != nil {
		return "", fmt.Errorf("no valid tenant identifier found in header, query, or subdomain: %w", err)
	}

	return tenantID, nil
}

// writeErrorResponse writes an error response to an HTTP ResponseWriter (for testing)
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s", "status": %d}`, message, statusCode)))
}

// sendErrorResponse sends a standardized JSON error response
func sendErrorResponse(c *fiber.Ctx, message string, statusCode int, requestID string) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Error:     message,
		Status:    statusCode,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Fiber context helper functions for accessing tenant information

// GetTenantID retrieves tenant UUID from Fiber context
func GetTenantID(c *fiber.Ctx) (uuid.UUID, error) {
	tenantID, ok := c.Locals(TenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}

// GetTenantIDString retrieves tenant ID as string from Fiber context
func GetTenantIDString(c *fiber.Ctx) (string, error) {
	tenantID, ok := c.Locals(TenantKey).(string)
	if !ok {
		return "", fmt.Errorf("tenant ID string not found in context")
	}
	return tenantID, nil
}

// GetTenantEntity retrieves full tenant entity from Fiber context
// NOTE: This will return nil if caching is disabled or entity wasn't stored
func GetTenantEntity(c *fiber.Ctx) (*tenant.Tenant, error) {
	t, ok := c.Locals(EntityKey).(*tenant.Tenant)
	if !ok || t == nil {
		return nil, fmt.Errorf("tenant entity not found in context")
	}
	return t, nil
}

// MustGetTenantID retrieves tenant UUID from Fiber context or panics
// Use this only in handlers where tenant context is guaranteed by middleware
func MustGetTenantID(c *fiber.Ctx) uuid.UUID {
	tenantID, err := GetTenantID(c)
	if err != nil {
		panic("tenant ID must be present in context - ensure TenantMiddleware is applied")
	}
	return tenantID
}

// Legacy compatibility functions for gradual migration

// GetTenantFromContext retrieves tenant entity from Go context (legacy)
// TODO: Deprecated - Use GetTenantEntity(c) with Fiber context instead
func GetTenantFromContext(ctx context.Context) (*tenant.Tenant, bool) {
	t, ok := ctx.Value(TenantKey).(*tenant.Tenant)
	return t, ok
}

// GetTenantIDFromContext retrieves tenant ID from Go context (legacy)
// TODO: Deprecated - Use GetTenantID(c) with Fiber context instead
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	// Try the new shared context approach first
	if tenantID, ok := shared.GetTenantID(ctx); ok {
		return tenantID.String(), true
	}

	// Fall back to legacy format
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}
