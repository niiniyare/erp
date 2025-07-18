package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/core/user"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// NewRouter creates a new router with all handlers
func NewRouter(
	tenantService tenant.Service,
	entityService entity.Service,
	userService user.Service,
	accessRequestService user.AccessRequestService,
	conditionalAccessService user.ConditionalAccessService,
	analyticsService user.UserAnalyticsService,
	tracing *tracing.TracingService,
	metrics *metrics.MetricsService,
) *gin.Engine {
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.TracingMiddleware(tracing))
	r.Use(middleware.MetricsMiddleware(metrics))

	// Initialize handlers
	tenantHandler := NewTenantHandler(tenantService)
	entityHandler := NewEntityHandler(entityService, tracing, metrics)
	userHandler := NewUserHandler(userService, tracing, metrics)
	accessRequestHandler := NewAccessRequestHandler(accessRequestService, conditionalAccessService, analyticsService, tracing, metrics)
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

		// Person Management routes (temporarily disabled - implementation pending)
		// persons := v1.Group("/persons")
		// {
		// 	persons.GET("/", userManagementHandler.ListPersons)
		// 	persons.POST("/", userManagementHandler.CreatePerson)
		// 	persons.GET("/:person_id", userManagementHandler.GetPerson)
		// }

		// Employee Management routes (temporarily disabled - implementation pending)
		// employees := v1.Group("/employees")
		// {
		// 	employees.GET("/", userManagementHandler.ListEmployees)
		// 	employees.POST("/", userManagementHandler.CreateEmployee)
		// 	employees.GET("/:employee_id", userManagementHandler.GetEmployee)
		// }

		// User Management routes
		users := v1.Group("/users")
		{
			// Basic user management (working)
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/auth", userHandler.AuthenticateUser)
			users.GET("/search", userHandler.SearchUsers)
			users.PUT("/:id/password", userHandler.UpdateUserPassword)
			users.GET("/:id/roles", userHandler.GetUserRoles)
		}

		// Access Request Workflow routes
		accessRequests := v1.Group("/access-requests")
		{
			accessRequests.POST("/", accessRequestHandler.CreateAccessRequest)
			accessRequests.GET("/:id", accessRequestHandler.GetAccessRequest)
			accessRequests.POST("/:id/process", accessRequestHandler.ProcessAccessRequest)
			accessRequests.DELETE("/:id", accessRequestHandler.RevokeAccessRequest)
			accessRequests.GET("/", accessRequestHandler.ListAccessRequests)
			accessRequests.GET("/stats", accessRequestHandler.GetAccessRequestStats)
		}

		// Conditional Access routes
		conditionalAccess := v1.Group("/conditional-access")
		{
			conditionalAccess.POST("/evaluate", accessRequestHandler.EvaluateConditionalAccess)
			conditionalAccess.POST("/rules", accessRequestHandler.CreateConditionalAccessRule)
		}

		// User Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/users/:user_id/behavior", accessRequestHandler.GetUserBehaviorAnalytics)
			analytics.GET("/users/:user_id/risk", accessRequestHandler.GetUserRiskAssessment)
			analytics.GET("/users/:user_id/insights", accessRequestHandler.GetUserPersonalizedInsights)
			analytics.POST("/users/:user_id/detect-anomalies", accessRequestHandler.DetectUserAnomalies)
		}
	}

	// Add CORS middleware for API testing
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Entity-ID, X-User-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	return r
}
