package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/api/handlers"
	financeHandler "github.com/niiniyare/erp/internal/api/handlers/finance"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"

	// GOA generated packages
	abacGen "github.com/niiniyare/erp/internal/api/gen/abac"
	accessrequest "github.com/niiniyare/erp/internal/api/gen/access_request"
	adminfeatureflag "github.com/niiniyare/erp/internal/api/gen/admin_featureflag"
	auth "github.com/niiniyare/erp/internal/api/gen/auth"
	featureflag "github.com/niiniyare/erp/internal/api/gen/featureflag"
	finance "github.com/niiniyare/erp/internal/api/gen/finance"
	health "github.com/niiniyare/erp/internal/api/gen/health"
	abacsvr "github.com/niiniyare/erp/internal/api/gen/http/abac/server"
	accessrequestsvr "github.com/niiniyare/erp/internal/api/gen/http/access_request/server"
	adminfeatureflagsvr "github.com/niiniyare/erp/internal/api/gen/http/admin_featureflag/server"
	authsvr "github.com/niiniyare/erp/internal/api/gen/http/auth/server"
	featureflagsvr "github.com/niiniyare/erp/internal/api/gen/http/featureflag/server"
	financesvr "github.com/niiniyare/erp/internal/api/gen/http/finance/server"
	healthsvr "github.com/niiniyare/erp/internal/api/gen/http/health/server"
	openapisvr "github.com/niiniyare/erp/internal/api/gen/http/openapi/server"
	organizationsvr "github.com/niiniyare/erp/internal/api/gen/http/organization/server"
	tenantsvr "github.com/niiniyare/erp/internal/api/gen/http/tenant/server"
	usersvr "github.com/niiniyare/erp/internal/api/gen/http/user/server"
	openapi "github.com/niiniyare/erp/internal/api/gen/openapi"
	organization "github.com/niiniyare/erp/internal/api/gen/organization"
	goaTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	goaUser "github.com/niiniyare/erp/internal/api/gen/user"
	"goa.design/clue/debug"
	clueLog "goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

type GOAServer struct {
	Handler http.Handler
	Mux     goahttp.Muxer
}

func InitializeGOAServer(services *Services, financeServices *financeService.Services, store db.Store, cacheService cache.Service, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*GOAServer, error) {
	// Initialize GOA services
	var (
		abacSvc             abacGen.Service
		accessRequestSvc    accessrequest.Service
		adminFeatureFlagSvc adminfeatureflag.Service
		authSvc             auth.Service
		featureFlagSvc      featureflag.Service
		financeSvc          finance.Service
		healthSvc           health.Service
		organizationSvc     organization.Service
		tenantSvc           goaTenant.Service
		userSvc             goaUser.Service
		openapiSvc          openapi.Service
	)

	abacSvc = handlers.NewABACGoaHandler(services.ABACService, metricsService, tracingService, logger.WithFields(logger.Fields{}))
	accessRequestSvc = handlers.NewAccessRequestGoaHandler(services.AccessRequestService, services.ConditionalAccessService, services.AnalyticsService, tracingService, metricsService)
	adminFeatureFlagSvc = handlers.NewAdminFeatureFlagService(services.AdminFeatureFlagService, services.ABACService, logger.WithFields(logger.Fields{}), metricsService, tracingService)
	// For now, create a simple adapter for IAM service
	// TODO: Replace with proper IAM service when fully implemented
	authSvc = handlers.NewAuthHandlerWithIdentity(services.IdentityService, services.TenantService, tracingService, metricsService)
	featureFlagSvc = handlers.NewFeatureFlagService(services.FeatureFlagService, logger.WithFields(logger.Fields{}), metricsService, tracingService)
	// Initialize finance service only if finance services are available
	if financeServices != nil {
		financeSvc = financeHandler.NewFinanceHandler(*financeServices, tracingService, metricsService)
	} else {
		// Set to nil for now - finance endpoints won't be available
		financeSvc = nil
		logger.Warn("Finance services not available, finance endpoints disabled", logger.Fields{
			"service": "finance",
			"status": "disabled",
		})
	}
	// Create a simple health checker instance (nil store for now - needs proper initialization)
	healthChecker := handlers.NewHealthChecker(nil, nil, logger.WithFields(logger.Fields{}), metricsService, tracingService)
	healthSvc = handlers.NewHealthGoaHandler(healthChecker, tracingService, metricsService)
	organizationSvc = handlers.NewOrganizationGoaHandler(services.EntityService, tracingService, metricsService)
	tenantSvc = handlers.NewUnifiedTenantHandler(
		services.TenantService,
		services.TenantProvisioningService,
		tracingService,
		metricsService,
	)
	userSvc = handlers.NewUserGoaHandler(services.IdentityService, services.AccessRequestService, services.ConditionalAccessService, services.AnalyticsService, tracingService, metricsService)
	openapiSvc = handlers.NewOpenapiHandler()

	// Create GOA endpoints
	var (
		abacEndpoints             *abacGen.Endpoints
		accessRequestEndpoints    *accessrequest.Endpoints
		adminFeatureFlagEndpoints *adminfeatureflag.Endpoints
		authEndpoints             *auth.Endpoints
		featureFlagEndpoints      *featureflag.Endpoints
		financeEndpoints          *finance.Endpoints
		healthEndpoints           *health.Endpoints
		organizationEndpoints     *organization.Endpoints
		tenantEndpoints           *goaTenant.Endpoints
		userEndpoints             *goaUser.Endpoints
		openapiEndpoints          *openapi.Endpoints
	)

	abacEndpoints = abacGen.NewEndpoints(abacSvc)
	abacEndpoints.Use(debug.LogPayloads())
	abacEndpoints.Use(clueLog.Endpoint)

	accessRequestEndpoints = accessrequest.NewEndpoints(accessRequestSvc)
	accessRequestEndpoints.Use(debug.LogPayloads())
	accessRequestEndpoints.Use(clueLog.Endpoint)

	adminFeatureFlagEndpoints = adminfeatureflag.NewEndpoints(adminFeatureFlagSvc)
	adminFeatureFlagEndpoints.Use(debug.LogPayloads())
	adminFeatureFlagEndpoints.Use(clueLog.Endpoint)

	authEndpoints = auth.NewEndpoints(authSvc)
	authEndpoints.Use(debug.LogPayloads())
	authEndpoints.Use(clueLog.Endpoint)

	featureFlagEndpoints = featureflag.NewEndpoints(featureFlagSvc)
	featureFlagEndpoints.Use(debug.LogPayloads())
	featureFlagEndpoints.Use(clueLog.Endpoint)

	// Only initialize finance endpoints if finance service is available
	if financeSvc != nil {
		financeEndpoints = finance.NewEndpoints(financeSvc)
		financeEndpoints.Use(debug.LogPayloads())
		financeEndpoints.Use(clueLog.Endpoint)
	}

	healthEndpoints = health.NewEndpoints(healthSvc)
	healthEndpoints.Use(debug.LogPayloads())
	healthEndpoints.Use(clueLog.Endpoint)

	organizationEndpoints = organization.NewEndpoints(organizationSvc)
	organizationEndpoints.Use(debug.LogPayloads())
	organizationEndpoints.Use(clueLog.Endpoint)

	tenantEndpoints = goaTenant.NewEndpoints(tenantSvc)
	tenantEndpoints.Use(debug.LogPayloads())
	tenantEndpoints.Use(clueLog.Endpoint)

	userEndpoints = goaUser.NewEndpoints(userSvc)
	userEndpoints.Use(debug.LogPayloads())
	userEndpoints.Use(clueLog.Endpoint)

	openapiEndpoints = openapi.NewEndpoints(openapiSvc)
	openapiEndpoints.Use(debug.LogPayloads())
	openapiEndpoints.Use(clueLog.Endpoint)

	// Create GOA HTTP mux with production-ready configuration
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	mux := goahttp.NewMuxer()

	// Add debug handlers only in development
	if isDevelopmentMode() {
		debug.MountPprofHandlers(debug.Adapt(mux))
		debug.MountDebugLogEnabler(debug.Adapt(mux))
		logger.Info("Debug handlers enabled", logger.Fields{"mode": "development"})
	}

	// Create production-ready error handler
	eh := createProductionErrorHandler()

	abacServer := abacsvr.New(abacEndpoints, mux, dec, enc, eh, nil)
	accessRequestServer := accessrequestsvr.New(accessRequestEndpoints, mux, dec, enc, eh, nil)
	adminFeatureFlagServer := adminfeatureflagsvr.New(adminFeatureFlagEndpoints, mux, dec, enc, eh, nil)
	authServer := authsvr.New(authEndpoints, mux, dec, enc, eh, nil)
	featureFlagServer := featureflagsvr.New(featureFlagEndpoints, mux, dec, enc, eh, nil)
	// Only create finance server if endpoints are available
	var financeServer *financesvr.Server
	if financeEndpoints != nil {
		financeServer = financesvr.New(financeEndpoints, mux, dec, enc, eh, nil)
	}
	healthServer := healthsvr.New(healthEndpoints, mux, dec, enc, eh, nil)
	organizationServer := organizationsvr.New(organizationEndpoints, mux, dec, enc, eh, nil)
	tenantServer := tenantsvr.New(tenantEndpoints, mux, dec, enc, eh, nil)
	userServer := usersvr.New(userEndpoints, mux, dec, enc, eh, nil)
	openapiServer := openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)

	// Mount GOA HTTP servers
	abacsvr.Mount(mux, abacServer)
	accessrequestsvr.Mount(mux, accessRequestServer)
	adminfeatureflagsvr.Mount(mux, adminFeatureFlagServer)
	authsvr.Mount(mux, authServer)
	featureflagsvr.Mount(mux, featureFlagServer)
	// Only mount finance server if it exists
	if financeServer != nil {
		financesvr.Mount(mux, financeServer)
	}
	healthsvr.Mount(mux, healthServer)
	organizationsvr.Mount(mux, organizationServer)
	tenantsvr.Mount(mux, tenantServer)
	usersvr.Mount(mux, userServer)
	openapisvr.Mount(mux, openapiServer)

	// Apply middleware to the muxer with IAM adapter
	var finalHandler http.Handler = mux

	// Create IAM service adapter for middleware integration
	iamAdapter := NewIAMServiceAdapter(services, store, cacheService, metricsService, tracingService)
	middlewareSetup, err := initializeMiddleware(iamAdapter, cacheService, metricsService, tracingService)
	if err != nil {
		logger.Warn("Failed to initialize middleware, continuing without it", logger.Fields{
			"error": err.Error(),
			"mode":  "fallback",
		})
	} else {
		finalHandler = middlewareSetup.ConfigureHTTPMuxer(mux)
		logger.Info("Middleware stack successfully integrated", logger.Fields{
			"status": "enabled",
			"mode":   getEnvironment(),
		})
	}

	// Create a custom wrapper that bypasses GOA for Swagger UI
	originalHandler := finalHandler

	// Create GOA-compatible context with logger
	format := clueLog.FormatJSON
	if clueLog.IsTerminal() {
		format = clueLog.FormatTerminal
	}
	ctx := clueLog.Context(context.Background(), clueLog.WithFormat(format))

	// Add logging and debugging to the GOA handler
	var handler http.Handler = originalHandler
	handler = clueLog.HTTP(ctx)(handler)
	handler = debug.HTTP()(handler)

	// Log mounted endpoints
	logMountedEndpoints(abacServer.Mounts, "ABAC")
	logMountedEndpoints(accessRequestServer.Mounts, "AccessRequest")
	logMountedEndpoints(adminFeatureFlagServer.Mounts, "AdminFeatureFlag")
	logMountedEndpoints(authServer.Mounts, "Auth")
	logMountedEndpoints(featureFlagServer.Mounts, "FeatureFlag")
	// Only log finance endpoints if server exists
	if financeServer != nil {
		logMountedEndpoints(financeServer.Mounts, "Finance")
	}
	logMountedEndpoints(organizationServer.Mounts, "Organization")
	logMountedEndpoints(tenantServer.Mounts, "Tenant")
	logMountedEndpoints(userServer.Mounts, "User")
	logMountedEndpoints(openapiServer.Mounts, "OpenAPI")

	return &GOAServer{
		Handler: handler,
		Mux:     mux,
	}, nil
}

func logMountedEndpoints(mounts any, serviceName string) {
	// Use reflection to handle different mount types
	v := reflect.ValueOf(mounts)
	if v.Kind() != reflect.Slice {
		return
	}

	count := v.Len()
	if count == 0 {
		return
	}

	logger.Info("🔗 API endpoints mounted", logger.Fields{
		"service": serviceName,
		"count":   count,
	})

	for i := 0; i < count; i++ {
		mount := v.Index(i)
		if mount.Kind() == reflect.Ptr {
			mount = mount.Elem()
		}

		// Extract fields using reflection
		method := getFieldValue(mount, "Method")
		verb := getFieldValue(mount, "Verb")
		pattern := getFieldValue(mount, "Pattern")

		logger.Info("  → "+serviceName, logger.Fields{
			"method": method,
			"verb":   verb,
			"path":   pattern,
		})
	}
}

func getFieldValue(v reflect.Value, fieldName string) string {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return ""
	}
	return field.String()
}

// InitializeGinRouter is deprecated - use GOA-only mode instead
func InitializeGinRouter(services *Services, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (http.Handler, error) {
	logger.Warn("InitializeGinRouter is deprecated - migration mode no longer supported", logger.Fields{
		"service":        "gin-router",
		"status":         "deprecated",
		"recommendation": "Use GOA-only mode (SERVER_MODE=goa-only)",
	})

	// Return a simple handler that redirects to GOA endpoints
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone)
		w.Write([]byte(`{"error": "Migration mode deprecated. Use GOA-only mode.", "migration_guide": "/api/openapi.json"}`))
	}), nil
}

// CombinedHandler is deprecated - use GOA-only mode instead
type CombinedHandler struct {
	goaHandler http.Handler
	ginHandler http.Handler // Deprecated: no longer used
}

// ServeHTTP implements the http.Handler interface (deprecated)
func (c *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Always route to GOA handler since migration mode is deprecated
	logger.Warn("CombinedHandler is deprecated - use GOA-only mode", logger.Fields{
		"path":           r.URL.Path,
		"method":         r.Method,
		"recommendation": "Set SERVER_MODE=goa-only",
	})

	c.goaHandler.ServeHTTP(w, r)
}

// isDevelopmentMode checks if we're running in development mode
func isDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}
	return env == "development" || env == "dev" || env == ""
}

// createProductionErrorHandler creates an error handler optimized for production
func createProductionErrorHandler() func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		// Extract request information for better error tracking
		requestID := ctx.Value("request-id")
		userAgent := ctx.Value("user-agent")
		path := ctx.Value("path")

		// Log error with context
		fields := logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
			"user_agent": userAgent,
			"path":       path,
		}

		// Determine error severity and log level
		statusCode := getErrorStatusCode(err)
		if statusCode >= 500 {
			logger.Error("Server error occurred", fields)
		} else if statusCode >= 400 {
			logger.Warn("Client error occurred", fields)
		} else {
			logger.Info("Request processed with error", fields)
		}

		// Don't expose internal errors in production
		if !isDevelopmentMode() && statusCode >= 500 {
			// Replace internal server errors with generic message
			http.Error(w, `{"error":"Internal server error","code":"INTERNAL_ERROR"}`, statusCode)
			return
		}

		// Fallback for development mode or non-500 errors
		// Use the default encoder to write the error to the response
		enc := goahttp.ResponseEncoder(ctx, w)
		enc.Encode(err)
	}
}

// getErrorStatusCode extracts HTTP status code from error
func getErrorStatusCode(err error) int {
	// This is a simplified version - in production you'd have more sophisticated error type checking
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return 404
	}
	if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
		return 401
	}
	if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
		return 403
	}
	if strings.Contains(strings.ToLower(err.Error()), "bad request") {
		return 400
	}
	return 500 // Default to internal server error
}

// initializeMiddleware creates and configures the middleware stack
func initializeMiddleware(iamAdapter *IAMServiceAdapter, cacheService cache.Service, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*MiddlewareSetup, error) {
	// Use default whitelist for critical endpoints
	whitelist, err := middleware.NewEndpointWhitelist(
		[]string{"GET /health*", "GET /api/v1/health*", "GET /swagger-ui/*", "POST /api/v1/tenants"},
		[]string{"POST /api/v1/auth/login", "GET /api/v1/version"},
	)
	if err != nil {
		logger.Warn("Failed to create endpoint whitelist, using default", logger.Fields{"error": err.Error()})
		// Fallback to default whitelist
		whitelist = middleware.DefaultWhitelist()
	}

	return &MiddlewareSetup{
		logger:    logger.WithFields(logger.Fields{"component": "middleware"}),
		cache:     cacheService,
		metrics:   metricsService,
		tracing:   tracingService,
		whitelist: whitelist,
		services:  iamAdapter.services, // Store services for tenant middleware
		store:     iamAdapter.store,    // Store database for RLS
	}, nil
}

// MiddlewareSetup provides native HTTP middleware integration for GOA server
type MiddlewareSetup struct {
	logger    logger.Logger
	cache     cache.Service
	metrics   *metrics.MetricsService
	tracing   tracing.TracingService
	whitelist *middleware.EndpointWhitelist
	config    *config.Config
	services  *Services
	store     db.Store
}

// ConfigureHTTPMuxer wraps the muxer with native tenant middleware
func (m *MiddlewareSetup) ConfigureHTTPMuxer(mux http.Handler) http.Handler {
	// Apply native tenant middleware chain
	// Order: Tenant Middleware → Request Logging → GOA Handler

	// 1. Create tenant middleware
	tenantMiddleware := middleware.TenantMiddleware(
		m.services.TenantService,
		m.store,
		m.whitelist,
	)

	// 2. Create request logging middleware
	requestLogger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logger.Debug("Request received", logger.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			next.ServeHTTP(w, r)
		})
	}

	// 3. Chain middlewares: Tenant → Logging → GOA Handler
	handler := tenantMiddleware(requestLogger(mux))

	m.logger.Info("Native middleware chain configured", logger.Fields{
		"middlewares": []string{"tenant", "request_logging"},
		"mode":        "goa-native",
	})

	return handler
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

// IAMServiceAdapter bridges existing services with IAM interface for middleware
type IAMServiceAdapter struct {
	identityService identity.Service
	abacService     abac.Service
	auditService    audit.Service
	tenantService   tenant.Service
	entityService   entity.Service
	cache           cache.Service
	logger          logger.Logger
	metrics         *metrics.MetricsService
	tracing         tracing.TracingService
	services        *Services // Reference to all services
	store           db.Store  // Database store for RLS operations
}

// NewIAMServiceAdapter creates a new IAM service adapter
func NewIAMServiceAdapter(services *Services, store db.Store, cacheService cache.Service, metricsService *metrics.MetricsService, tracingService tracing.TracingService) *IAMServiceAdapter {
	return &IAMServiceAdapter{
		identityService: services.IdentityService,
		abacService:     services.ABACService,
		auditService:    services.AuditService,
		tenantService:   services.TenantService,
		entityService:   services.EntityService,
		cache:           cacheService,
		logger:          logger.WithFields(logger.Fields{"component": "iam_adapter"}),
		metrics:         metricsService,
		tracing:         tracingService,
		services:        services,
		store:           store,
	}
}

// Authentication returns a minimal authentication service adapter
func (a *IAMServiceAdapter) Authentication() any {
	// Return a simple struct that implements basic auth methods needed by middleware
	return &AuthnAdapter{
		identityService: a.identityService,
		logger:          a.logger,
	}
}

// Authorization returns a minimal authorization service adapter
func (a *IAMServiceAdapter) Authorization() any {
	// Return a simple struct that implements basic authz methods needed by middleware
	return &AuthzAdapter{
		abacService: a.abacService,
		logger:      a.logger,
	}
}

// Policy returns a minimal policy service adapter
func (a *IAMServiceAdapter) Policy() any {
	// Return a simple struct that implements basic policy methods needed by middleware
	return &PolicyAdapter{
		abacService: a.abacService,
		logger:      a.logger,
	}
}

// AuthnAdapter provides minimal authentication functionality for middleware
type AuthnAdapter struct {
	identityService identity.Service
	logger          logger.Logger
}

// ValidateToken validates JWT tokens (simplified for middleware integration)
func (a *AuthnAdapter) ValidateToken(ctx context.Context, token string) (bool, error) {
	// TODO: Implement proper JWT validation using existing identity service
	// For now, return true to allow middleware testing
	a.logger.Debug("Token validation called", logger.Fields{"token_present": token != ""})
	return token != "", nil
}

// GetUser retrieves user by ID (bridge to existing identity service)
func (a *AuthnAdapter) GetUser(ctx context.Context, userID string) (any, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	return a.identityService.GetUserByID(ctx, id)
}

// AuthzAdapter provides minimal authorization functionality for middleware
type AuthzAdapter struct {
	abacService abac.Service
	logger      logger.Logger
}

// EvaluatePermission evaluates user permissions (bridge to existing ABAC service)
func (a *AuthzAdapter) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// TODO: Bridge to ABAC service with proper request structure
	// For now, return true to allow middleware testing
	a.logger.Debug("Permission evaluation called", logger.Fields{
		"user_id":  userID,
		"resource": resource,
		"action":   action,
	})
	return true, nil
}

// PolicyAdapter provides minimal policy functionality for middleware
type PolicyAdapter struct {
	abacService abac.Service
	logger      logger.Logger
}

// GetPolicy retrieves policy information (bridge to existing ABAC service)
func (a *PolicyAdapter) GetPolicy(ctx context.Context, policyID string) (any, error) {
	// TODO: Bridge to ABAC service policy operations
	// For now, return empty policy to allow middleware testing
	a.logger.Debug("Policy retrieval called", logger.Fields{"policy_id": policyID})
	return map[string]any{"id": policyID, "status": "active"}, nil
}
