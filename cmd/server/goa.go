package main

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"

	// GOA generated packages
	auth "github.com/niiniyare/erp/gen/auth"
	authsvr "github.com/niiniyare/erp/gen/http/auth/server"
	openapisvr "github.com/niiniyare/erp/gen/http/openapi/server"
	organizationsvr "github.com/niiniyare/erp/gen/http/organization/server"
	tenantsvr "github.com/niiniyare/erp/gen/http/tenant/server"
	usersvr "github.com/niiniyare/erp/gen/http/user/server"
	openapi "github.com/niiniyare/erp/gen/openapi"
	organization "github.com/niiniyare/erp/gen/organization"
	goaTenant "github.com/niiniyare/erp/gen/tenant"
	goaUser "github.com/niiniyare/erp/gen/user"
	abacGen "github.com/niiniyare/erp/internal/gen/gen/abac"
	abacsvr "github.com/niiniyare/erp/internal/gen/gen/http/abac/server"
	"goa.design/clue/debug"
	clueLog "goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

type GOAServer struct {
	Handler http.Handler
	Mux     goahttp.Muxer
}

func InitializeGOAServer(services *Services, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*GOAServer, error) {
	// Initialize GOA services
	var (
		abacSvc         abacGen.Service
		authSvc         auth.Service
		organizationSvc organization.Service
		tenantSvc       goaTenant.Service
		userSvc         goaUser.Service
		openapiSvc      openapi.Service
	)

	abacSvc = handlers.NewABACGoaHandler(services.ABACService, metricsService, tracingService, logger.WithFields(logger.Fields{}))
	authSvc = handlers.NewAuthHandler(services.IdentityService, tracingService, metricsService)
	organizationSvc = handlers.NewOrganizationGoaHandler(services.EntityService, tracingService, metricsService)
	tenantSvc = handlers.NewTenantGoaHandler(services.TenantService, tracingService, metricsService)
	userSvc = handlers.NewUserGoaHandler(services.IdentityService, services.AccessRequestService, services.ConditionalAccessService, services.AnalyticsService, tracingService, metricsService)
	openapiSvc = handlers.NewOpenapiHandler()

	// Create GOA endpoints
	var (
		abacEndpoints         *abacGen.Endpoints
		authEndpoints         *auth.Endpoints
		organizationEndpoints *organization.Endpoints
		tenantEndpoints       *goaTenant.Endpoints
		userEndpoints         *goaUser.Endpoints
		openapiEndpoints      *openapi.Endpoints
	)

	abacEndpoints = abacGen.NewEndpoints(abacSvc)
	abacEndpoints.Use(debug.LogPayloads())
	abacEndpoints.Use(clueLog.Endpoint)

	authEndpoints = auth.NewEndpoints(authSvc)
	authEndpoints.Use(debug.LogPayloads())
	authEndpoints.Use(clueLog.Endpoint)

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

	// Create GOA HTTP mux
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	mux := goahttp.NewMuxer()
	debug.MountPprofHandlers(debug.Adapt(mux))
	debug.MountDebugLogEnabler(debug.Adapt(mux))

	// Create GOA HTTP servers
	eh := func(ctx context.Context, w http.ResponseWriter, err error) {
		logger.Error("HTTP Error", logger.Fields{"error": err.Error()})
	}

	abacServer := abacsvr.New(abacEndpoints, mux, dec, enc, eh, nil)
	authServer := authsvr.New(authEndpoints, mux, dec, enc, eh, nil)
	organizationServer := organizationsvr.New(organizationEndpoints, mux, dec, enc, eh, nil)
	tenantServer := tenantsvr.New(tenantEndpoints, mux, dec, enc, eh, nil)
	userServer := usersvr.New(userEndpoints, mux, dec, enc, eh, nil)
	openapiServer := openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)

	// Mount GOA HTTP servers
	abacsvr.Mount(mux, abacServer)
	authsvr.Mount(mux, authServer)
	organizationsvr.Mount(mux, organizationServer)
	tenantsvr.Mount(mux, tenantServer)
	usersvr.Mount(mux, userServer)
	openapisvr.Mount(mux, openapiServer)

	// Create a custom wrapper that bypasses GOA for Swagger UI
	originalHandler := mux

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
	logMountedEndpoints(authServer.Mounts, "Auth")
	logMountedEndpoints(organizationServer.Mounts, "Organization")
	logMountedEndpoints(tenantServer.Mounts, "Tenant")
	logMountedEndpoints(userServer.Mounts, "User")
	logMountedEndpoints(openapiServer.Mounts, "OpenAPI")

	return &GOAServer{
		Handler: handler,
		Mux:     originalHandler,
	}, nil
}

func logMountedEndpoints(mounts interface{}, serviceName string) {
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

// InitializeGinRouter creates and configures the Gin router for additional routes
func InitializeGinRouter(services *Services, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*gin.Engine, error) {
	// Create Gin router using the existing handlers.NewRouter function
	ginRouter := handlers.NewRouter(
		services.TenantService.(tenant.Service),
		services.EntityService.(entity.Service),
		services.IdentityService.(identity.Service),
		services.AccessRequestService.(request.AccessRequestService),
		services.ConditionalAccessService.(conditional.ConditionalAccessService),
		services.AnalyticsService.(analytics.UserAnalyticsService),
		tracingService,
		metricsService,
	)

	logger.Info("Gin router initialized for Swagger UI", logger.Fields{
		"service": "gin-router",
		"status":  "ready",
	})

	return ginRouter, nil
}

// CombinedHandler routes requests between GOA and Gin handlers
type CombinedHandler struct {
	goaHandler http.Handler
	ginHandler http.Handler
}

// ServeHTTP implements the http.Handler interface
func (c *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route specific paths to Gin handler
	if strings.HasPrefix(r.URL.Path, "/swagger-ui/") ||
		strings.HasPrefix(r.URL.Path, "/health") ||
		strings.HasPrefix(r.URL.Path, "/ready") ||
		strings.HasPrefix(r.URL.Path, "/api/v1/") {
		c.ginHandler.ServeHTTP(w, r)
		return
	}

	// Everything else goes to GOA handler
	c.goaHandler.ServeHTTP(w, r)
}
