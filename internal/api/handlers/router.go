package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// NewRouter creates a new router with all handlers
func NewRouter(tenantService tenant.Service, entityService entity.Service, tracing *tracing.TracingService, metrics *metrics.MetricsService) *gin.Engine {
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// Initialize handlers
	tenantHandler := NewTenantHandler(tenantService)
	entityHandler := NewEntityHandler(entityService, tracing, metrics)
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

		// Entity routes
		entities := v1.Group("/entities")
		{
			entities.POST("/", entityHandler.CreateEntity)
			entities.GET("/", entityHandler.ListEntities)
			entities.GET("/tree", entityHandler.GetEntityTree)
			entities.POST("/sequence/next", entityHandler.GetNextSequence)
			entities.POST("/sequence/reset", entityHandler.ResetSequence)
			entities.GET("/:id", entityHandler.GetEntity)
			entities.PUT("/:id", entityHandler.UpdateEntity)
			entities.DELETE("/:id", entityHandler.DeleteEntity)
			entities.POST("/:id/restore", entityHandler.RestoreEntity)
			entities.GET("/:id/children", entityHandler.GetEntityChildren)
			entities.GET("/:id/ancestors", entityHandler.GetEntityAncestors)
			entities.GET("/:id/hierarchy", entityHandler.GetEntityWithHierarchy)
		}
	}

	return r
}
