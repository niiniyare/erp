package main

import (
	"context"
	"net/http"

	"goa.design/clue/debug"
	clueLog "goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"

	"github.com/niiniyare/erp/internal/api/handlers"
	financeHandler "github.com/niiniyare/erp/internal/api/handlers/finance"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/logger"

	// GOA generated packages
	abacGen "github.com/niiniyare/erp/internal/api/gen/abac"
	accessrequest "github.com/niiniyare/erp/internal/api/gen/access_request"
	adminfeatureflag "github.com/niiniyare/erp/internal/api/gen/admin_featureflag"
	auth "github.com/niiniyare/erp/internal/api/gen/auth"
	featureflag "github.com/niiniyare/erp/internal/api/gen/featureflag"
	finance "github.com/niiniyare/erp/internal/api/gen/finance"
	health "github.com/niiniyare/erp/internal/api/gen/health"
	openapi "github.com/niiniyare/erp/internal/api/gen/openapi"
	organization "github.com/niiniyare/erp/internal/api/gen/organization"
	goaTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	goaUser "github.com/niiniyare/erp/internal/api/gen/user"

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
)

// newRouter creates and configures the application's HTTP router and middleware.
func newRouter(app *application) http.Handler {
	// 1. Initialize Goa services, passing in the application's core services.
	abacSvc := handlers.NewABACGoaHandler(app.services.ABACService, app.metrics, app.tracer, app.logger)
	accessRequestSvc := handlers.NewAccessRequestGoaHandler(app.services.AccessRequestService, app.services.ConditionalAccessService, app.services.AnalyticsService, app.tracer, app.metrics)
	adminFeatureFlagSvc := handlers.NewAdminFeatureFlagService(app.services.AdminFeatureFlagService, app.services.ABACService, app.logger, app.metrics, app.tracer)
	authSvc := handlers.NewAuthHandlerWithIdentity(app.services.IdentityService, app.services.TenantService, app.tracer, app.metrics)
	featureFlagSvc := handlers.NewFeatureFlagService(app.services.FeatureFlagService, app.logger, app.metrics, app.tracer)
	financeSvc := financeHandler.NewFinanceHandler(*app.financeServices, app.tracer, app.metrics)
	healthChecker := handlers.NewHealthChecker(app.store, nil, app.logger, app.metrics, app.tracer) // FIXME: Pass a proper redis client
	healthSvc := handlers.NewHealthGoaHandler(healthChecker, app.tracer, app.metrics)
	organizationSvc := handlers.NewOrganizationGoaHandler(app.services.EntityService, app.tracer, app.metrics)
	tenantSvc := handlers.NewTenantGoaHandler(app.services.TenantService, app.tracer, app.metrics)
	userSvc := handlers.NewUserGoaHandler(app.services.IdentityService, app.services.AccessRequestService, app.services.ConditionalAccessService, app.services.AnalyticsService, app.tracer, app.metrics)
	openapiSvc := handlers.NewOpenapiHandler()

	// 2. Create Goa endpoints from the services.
	abacEndpoints := abacGen.NewEndpoints(abacSvc)
	accessRequestEndpoints := accessrequest.NewEndpoints(accessRequestSvc)
	adminFeatureFlagEndpoints := adminfeatureflag.NewEndpoints(adminFeatureFlagSvc)
	authEndpoints := auth.NewEndpoints(authSvc)
	featureFlagEndpoints := featureflag.NewEndpoints(featureFlagSvc)
	financeEndpoints := finance.NewEndpoints(financeSvc)
	healthEndpoints := health.NewEndpoints(healthSvc)
	organizationEndpoints := organization.NewEndpoints(organizationSvc)
	tenantEndpoints := goaTenant.NewEndpoints(tenantSvc)
	userEndpoints := goaUser.NewEndpoints(userSvc)
	openapiEndpoints := openapi.NewEndpoints(openapiSvc)

	// 3. Create a Goa HTTP muxer.
	mux := goahttp.NewMuxer()

	// 4. Create a production-ready error handler.
	eh := createProductionErrorHandler(app.logger, app.config.App.IsProduction())

	// 5. Instantiate the Goa server handlers.
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	abacServer := abacsvr.New(abacEndpoints, mux, dec, enc, eh, nil)
	accessRequestServer := accessrequestsvr.New(accessRequestEndpoints, mux, dec, enc, eh, nil)
	adminFeatureFlagServer := adminfeatureflagsvr.New(adminFeatureFlagEndpoints, mux, dec, enc, eh, nil)
	authServer := authsvr.New(authEndpoints, mux, dec, enc, eh, nil)
	featureFlagServer := featureflagsvr.New(featureFlagEndpoints, mux, dec, enc, eh, nil)
	financeServer := financesvr.New(financeEndpoints, mux, dec, enc, eh, nil)
	healthServer := healthsvr.New(healthEndpoints, mux, dec, enc, eh, nil)
	organizationServer := organizationsvr.New(organizationEndpoints, mux, dec, enc, eh, nil)
	tenantServer := tenantsvr.New(tenantEndpoints, mux, dec, enc, eh, nil)
	userServer := usersvr.New(userEndpoints, mux, dec, enc, eh, nil)
	openapiServer := openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)

	// 6. Mount the server handlers on the muxer.
	abacsvr.Mount(mux, abacServer)
	accessrequestsvr.Mount(mux, accessRequestServer)
	adminfeatureflagsvr.Mount(mux, adminFeatureFlagServer)
	authsvr.Mount(mux, authServer)
	featureflagsvr.Mount(mux, featureFlagServer)
	financesvr.Mount(mux, financeServer)
	healthsvr.Mount(mux, healthServer)
	organizationsvr.Mount(mux, organizationServer)
	tenantsvr.Mount(mux, tenantServer)
	usersvr.Mount(mux, userServer)
	openapisvr.Mount(mux, openapiServer)

	// 7. Define and apply the middleware stack.
	var handler http.Handler = mux

	// Tenant middleware for RLS.
	whitelist, _ := middleware.NewEndpointWhitelist(
		[]string{"GET /health*", "GET /api/v1/health*", "GET /swagger-ui/*"},
		[]string{"POST /api/v1/auth/login", "GET /api/v1/version"},
	)
	handler = middleware.TenantMiddleware(app.services.TenantService, app.store, whitelist)(handler)

	// Goa-specific middleware for logging and debugging.
	format := clueLog.FormatJSON
	if clueLog.IsTerminal() {
		format = clueLog.FormatTerminal
	}
	ctx := clueLog.Context(context.Background(), clueLog.WithFormat(format))
	handler = clueLog.HTTP(ctx)(handler)
	if app.config.App.ShouldEnableDevelopmentMode() {
		handler = debug.HTTP()(handler)
		debug.MountPprofHandlers(debug.Adapt(mux))
		debug.MountDebugLogEnabler(debug.Adapt(mux))
		app.logger.Info("Debug handlers enabled", logger.Fields{"mode": "development"})
	}

	app.logger.Info("HTTP router initialized with middleware")
	return handler
}

// createProductionErrorHandler creates an error handler optimized for production.
func createProductionErrorHandler(log logger.Logger, isProd bool) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		// Extract request information for better error tracking.
		requestID := ctx.Value("request-id")
		path := ctx.Value("path")

		fields := logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
			"path":       path,
		}

		// Determine error severity and log level.
		var statusCode int
		if gerr, ok := err.(goahttp.Statuser); ok {
			statusCode = gerr.StatusCode()
		} else {
			statusCode = http.StatusInternalServerError
		}

		if statusCode >= 500 {
			log.Error("Server error occurred", fields)
		} else {
			log.Warn("Client error occurred", fields)
		}

		// Don't expose internal errors in production.
		if isProd && statusCode >= 500 {
			http.Error(w, `{"error":"Internal server error","code":"INTERNAL_ERROR"}`, statusCode)
			return
		}

		// Use the default Goa encoder for other errors.
		enc := goahttp.ResponseEncoder(ctx, w)
		enc.Encode(err)
	}
}
