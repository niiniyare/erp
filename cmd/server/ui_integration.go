package main

//
// import (
// 	"net/http"
//
// 	"github.com/niiniyare/erp/internal/application"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/ui/router"
// )
//
// // UIIntegration handles the integration of UI services with the main server
// type UIIntegration struct {
// 	uiRouter *router.UIRouter
// 	logger   logger.Logger
// }
//
// // NewUIIntegration creates a new UI integration
// func NewUIIntegration(app *application.Core, services *Services, logger logger.Logger) (*UIIntegration, error) {
// 	// Get infrastructure services
// 	appServices := app.GetServices()
//
// 	// Create UI router with existing services
// 	uiConfig := router.DefaultUIRouterConfig()
// 	// Configure based on environment or config file
// 	// uiConfig.CookieSecure = app.Config.Production
// 	// uiConfig.CookieDomain = app.Config.Domain
//
// 	uiRouter := router.NewUIRouter(
// 		services.IAMService,     // Using unified IAM service
// 		services.ABACService,    // Using existing ABAC service
// 		services.TenantService,  // Using existing tenant service
// 		services.AuditService,   // Using existing audit service
// 		appServices.RedisClient, // Using existing cache service
// 		logger,
// 		uiConfig,
// 	)
//
// 	return &UIIntegration{
// 		uiRouter: uiRouter,
// 		logger:   logger,
// 	}, nil
// }
//
// // MountUIRoutes mounts UI routes on the provided mux
// func (ui *UIIntegration) MountUIRoutes(mux *http.ServeMux) {
// 	// Mount the UI router on the main server mux
// 	// This integrates UI routes with the existing GOA routes
//
// 	// Option 1: Mount UI on specific paths
// 	mux.Handle("/console/", ui.uiRouter)
// 	mux.Handle("/workspace/", ui.uiRouter)
// 	mux.Handle("/portal/", ui.uiRouter)
// 	mux.Handle("/static/", ui.uiRouter)
// 	mux.Handle("/health", ui.uiRouter)
//
// 	// Option 2: Mount root handler with fallback to GOA
// 	// This would require more sophisticated routing logic
//
// 	ui.logger.Info("UI routes mounted successfully", logger.Fields{
// 		"console_path":   "/console/",
// 		"workspace_path": "/workspace/",
// 		"portal_path":    "/portal/",
// 		"static_path":    "/static/",
// 	})
// }
//
// // GetUIHandler returns the UI router handler
// func (ui *UIIntegration) GetUIHandler() http.Handler {
// 	return ui.uiRouter
// }
//
// // CreateCombinedHandler creates a handler that combines GOA API and UI routes
// func (ui *UIIntegration) CreateCombinedHandler(goaHandler http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		path := r.URL.Path
//
// 		// Route to UI handler for UI paths
// 		if isUIPath(path) {
// 			ui.uiRouter.ServeHTTP(w, r)
// 			return
// 		}
//
// 		// Route to GOA handler for API paths
// 		goaHandler.ServeHTTP(w, r)
// 	})
// }
//
// // isUIPath determines if a path should be handled by the UI router
// func isUIPath(path string) bool {
// 	uiPaths := []string{
// 		"/console",
// 		"/workspace",
// 		"/portal",
// 		"/static",
// 		"/health", // UI health check
// 	}
//
// 	for _, uiPath := range uiPaths {
// 		if path == uiPath || (len(path) > len(uiPath) && path[:len(uiPath)+1] == uiPath+"/") {
// 			return true
// 		}
// 	}
//
// 	// Root path goes to UI for redirection
// 	if path == "/" {
// 		return true
// 	}
//
// 	return false
// }
