package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/niiniyare/erp/internal/domain/tenant"
    "github.com/niiniyare/erp/internal/shared/errors"
)

// TenantContextKey is the key for tenant context
type TenantContextKey string

const (
    TenantIDKey TenantContextKey = "tenant_id"
    TenantKey   TenantContextKey = "tenant"
)

// TenantMiddleware extracts tenant from subdomain and sets context
func TenantMiddleware(tenantService tenant.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract subdomain from request
        subdomain := extractSubdomain(c.Request.Host)
        if subdomain == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subdomain"})
            c.Abort()
            return
        }
        
        // Get tenant by subdomain
        tenant, err := tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
        if err != nil {
            if errors.Is(err, errors.ErrTenantNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            }
            c.Abort()
            return
        }
        
        // Set tenant context for RLS
        ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenant.ID.String())
        ctx = context.WithValue(ctx, TenantKey, tenant)
        c.Request = c.Request.WithContext(ctx)
        
        // Set database session variable for RLS
        if err := setTenantRLS(ctx, tenant.ID.String()); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set tenant context"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// extractSubdomain extracts subdomain from host
func extractSubdomain(host string) string {
    parts := strings.Split(host, ".")
    if len(parts) >= 3 {
        return parts[0]
    }
    return ""
}

// setTenantRLS sets the PostgreSQL session variable for RLS
func setTenantRLS(ctx context.Context, tenantID string) error {
    // This would execute: SET app.tenant_id = 'tenant-uuid'
    // Implementation depends on your database connection
    return nil
}

// GetTenantFromContext retrieves tenant from context
func GetTenantFromContext(ctx context.Context) (*tenant.Tenant, bool) {
    t, ok := ctx.Value(TenantKey).(*tenant.Tenant)
    return t, ok
}

// GetTenantIDFromContext retrieves tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
    tenantID, ok := ctx.Value(TenantIDKey).(string)
    return tenantID, ok
}
