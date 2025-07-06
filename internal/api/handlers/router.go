package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

// NewRouter creates a new router with all handlers
func NewRouter(tenantService tenant.Service) *gin.Engine {
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// Initialize handlers
	tenantHandler := NewTenantHandler(tenantService)
	healthHandler := NewHealthHandler()

	// Health check routes
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Tenant routes
		tenants := v1.Group("/tenants")
		{
			tenants.POST("/", tenantHandler.CreateTenant)
			tenants.GET("/:id", tenantHandler.GetTenant)
			tenants.PUT("/:id", tenantHandler.UpdateTenant)
			tenants.DELETE("/:id", tenantHandler.DeleteTenant)
			tenants.GET("/", tenantHandler.ListTenants)
			tenants.GET("/subdomain/:subdomain", tenantHandler.GetTenantBySubdomain)
		}
	}

	return r
}
