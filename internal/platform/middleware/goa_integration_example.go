package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	goaHTTP "goa.design/goa/v3/http"
	"goa.design/goa/v3/security"
)

// GOAMiddlewareSetup provides complete GOA middleware integration
type GOAMiddlewareSetup struct {
	MiddlewareStack *MiddlewareStack
	logger          logger.Logger
}

// NewGOAMiddlewareSetup creates a new GOA middleware setup
func NewGOAMiddlewareSetup(
	iamService iam.Service,
	tenantService tenant.Service,
	cacheService cache.Service,
	environment string,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) (*GOAMiddlewareSetup, error) {
	// Create middleware configuration for environment
	config := DefaultMiddlewareConfig(environment)

	// Validate configuration
	if err := ValidateConfiguration(config); err != nil {
		return nil, fmt.Errorf("invalid middleware configuration: %w", err)
	}

	// Create middleware stack
	stack, err := NewMiddlewareStack(
		iamService,
		tenantService,
		cacheService,
		config,
		logger,
		metrics,
		tracer,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware stack: %w", err)
	}

	return &GOAMiddlewareSetup{
		MiddlewareStack: stack,
		logger:          logger,
	}, nil
}

// ConfigureHTTPMuxer configures a GOA HTTP muxer with all middleware
func (g *GOAMiddlewareSetup) ConfigureHTTPMuxer(mux goaHTTP.Muxer) http.Handler {
	// Apply HTTP middleware in order
	middlewares := g.MiddlewareStack.HTTPMiddlewareChain()

	// Start with the muxer as the base handler
	var handler http.Handler = mux

	// Apply custom middleware in reverse order (since they wrap)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	g.logger.Info("GOA HTTP muxer configured with middleware", logger.Fields{
		"middleware_count": len(middlewares),
		"environment":      g.MiddlewareStack.Config.Environment,
	})

	return handler
}

// SecurityMiddleware returns GOA security middleware for different authentication schemes
// Note: This is currently not used since GOA uses a different signature for security functions
// Use JWTSecurityFunc() instead for actual GOA integration
func (g *GOAMiddlewareSetup) SecurityMiddleware() map[string]interface{} {
	return map[string]interface{}{
		"jwt": g.MiddlewareStack.JWTAuth.JWTAuth,
		// Note: Basic auth and API key are not currently used in the GOA design
		// Only JWT authentication is implemented in the current API design
	}
}

// JWTSecurityFunc returns the JWT authentication function for GOA endpoints
func (g *GOAMiddlewareSetup) JWTSecurityFunc() func(context.Context, string, *security.JWTScheme) (context.Context, error) {
	return g.MiddlewareStack.JWTAuth.JWTAuth
}

// AuthorizationMiddleware returns specific authorization middleware for GOA services
func (g *GOAMiddlewareSetup) AuthorizationMiddleware() *AuthorizationGOAAdapter {
	return &AuthorizationGOAAdapter{
		authz: g.MiddlewareStack.Authorization,
	}
}

// AuthorizationGOAAdapter provides GOA-specific authorization methods
type AuthorizationGOAAdapter struct {
	authz *AuthorizationMiddleware
}

// RequireFinancePermission creates authorization middleware for finance operations
func (a *AuthorizationGOAAdapter) RequireFinancePermission(action string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return a.authz.RequirePermission("finance", action)
}

// RequireUserPermission creates authorization middleware for user operations
func (a *AuthorizationGOAAdapter) RequireUserPermission(action string) func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return a.authz.RequirePermission("user", action)
}

// RequireAdminRole creates authorization middleware for admin-only operations
func (a *AuthorizationGOAAdapter) RequireAdminRole() func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return a.authz.RequireRole("admin")
}

// RequireManagerRole creates authorization middleware for manager operations
func (a *AuthorizationGOAAdapter) RequireManagerRole() func(context.Context, any, *security.JWTScheme) (context.Context, error) {
	return a.authz.RequireRole("manager", "admin")
}

// EXAMPLE USAGE IN GOA SERVICE

/*
// Example: Finance Service with Complete Security
package finance

import (
	"context"
	"github.com/niiniyare/erp/internal/api/gen/finance"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

// FinanceService implements the finance service with security
type FinanceService struct {
	middlewareSetup *middleware.GOAMiddlewareSetup
	// other dependencies...
}

// NewFinanceService creates a new finance service
func NewFinanceService(middlewareSetup *middleware.GOAMiddlewareSetup) *FinanceService {
	return &FinanceService{
		middlewareSetup: middlewareSetup,
	}
}

// CreateTransaction implements transaction creation with authorization
func (s *FinanceService) CreateTransaction(ctx context.Context, p *finance.CreateTransactionPayload) (*finance.Transaction, error) {
	// The authorization middleware has already validated permissions
	// Context contains user information from authentication

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Business logic here...
	// The user is authenticated and authorized to create transactions

	return &finance.Transaction{
		ID:     "txn_123",
		Amount: p.Amount,
		UserID: userID,
	}, nil
}

// In your main server setup:
func setupFinanceService(middlewareSetup *middleware.GOAMiddlewareSetup) finance.Service {
	svc := NewFinanceService(middlewareSetup)

	// Get authorization adapter
	authz := middlewareSetup.AuthorizationMiddleware()

	// Apply security to endpoints
	finance.NewEndpoints(svc, authz.RequireFinancePermission("create"))

	return svc
}

// Complete server setup example:
func setupServer() {
	// Initialize services
	iamService := // ... initialize IAM service
	tenantService := // ... initialize tenant service
	cacheService := // ... initialize cache service

	// Initialize middleware
	middlewareSetup, err := middleware.NewGOAMiddlewareSetup(
		iamService,
		tenantService,
		cacheService,
		"production", // environment
		logger,
		metrics,
		tracer,
	)
	if err != nil {
		log.Fatal("Failed to setup middleware:", err)
	}

	// Create GOA services with security
	financeService := setupFinanceService(middlewareSetup)

	// Create HTTP muxer (following actual project pattern)
	mux := goaHTTP.NewMuxer()

	// Create finance server with security
	financeServer := financeHTTP.New(
		finance.NewEndpoints(financeService),
		mux,
		goaHTTP.RequestDecoder,
		goaHTTP.ResponseEncoder,
		errorHandler,
		// Add JWT security middleware
		middlewareSetup.JWTSecurityFunc(),
	)

	// Mount services
	financeHTTP.Mount(mux, financeServer)

	// Configure middleware and get final handler
	handler := middlewareSetup.ConfigureHTTPMuxer(mux)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", handler))
}
*/

// GetMiddlewareReport returns a report of applied middleware
func (g *GOAMiddlewareSetup) GetMiddlewareReport() MiddlewareReport {
	report := MiddlewareReport{
		Environment: g.MiddlewareStack.Config.Environment,
		Security:    g.MiddlewareStack.GetSecuritySummary(),
		Health:      g.MiddlewareStack.HealthCheck(),
	}

	// Add GOA-specific information
	report.GOAIntegration = map[string]interface{}{
		"security_schemes":    []string{"jwt"}, // Only JWT is currently implemented
		"authorization_types": []string{"permission", "role"},
		"middleware_count":    len(g.MiddlewareStack.HTTPMiddlewareChain()),
	}

	return report
}

// MiddlewareReport provides middleware status
type MiddlewareReport struct {
	Environment    string                 `json:"environment"`
	Security       SecuritySummary        `json:"security"`
	Health         map[string]interface{} `json:"health"`
	GOAIntegration map[string]interface{} `json:"goa_integration"`
}
