package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/cmd/server/services"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/web/pages/auth"
)

// UserResponse represents user information in responses
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AuthUnifiedHandler handles both HTML and JSON authentication requests
type AuthUnifiedHandler struct {
	coreServices   *services.CoreServices
	tracingService tracing.TracingService
	metricsService *metrics.MetricsService
	logger         logger.Logger
}

// NewAuthUnifiedHandler creates a new unified auth handler
func NewAuthUnifiedHandler(
	coreServices *services.CoreServices,
	tracingService tracing.TracingService,
	metricsService *metrics.MetricsService,
) *AuthUnifiedHandler {
	return &AuthUnifiedHandler{
		coreServices:   coreServices,
		tracingService: tracingService,
		metricsService: metricsService,
		logger:         logger.WithFields(logger.Fields{"handler": "auth_unified"}),
	}
}

// ShowLogin handles GET /login - serves login form or returns login page data
func (h *AuthUnifiedHandler) ShowLogin(c *fiber.Ctx) error {
	_, span := h.tracingService.StartSpan(c.Context(), "auth.show_login")
	defer span.End()

	// Get response mode from content negotiation middleware
	responseMode := middleware.GetResponseMode(c)

	// Prepare login page props
	loginProps := auth.LoginPageProps{
		Title:            "Sign In",
		Subtitle:         "Welcome back! Please sign in to your account.",
		ShowRememberMe:   true,
		ShowForgotLink:   true,
		ShowRegisterLink: false,
		OrganizationName: "ERP System",
		HxPost:           "/auth/login",
		HxTarget:         "#login-form",
	}

	switch responseMode {
	case middleware.ResponseModeJSON:
		// Return JSON data for API consumers
		return Success(c, fiber.Map{
			"title":              loginProps.Title,
			"subtitle":           loginProps.Subtitle,
			"show_remember_me":   loginProps.ShowRememberMe,
			"show_forgot_link":   loginProps.ShowForgotLink,
			"show_register_link": loginProps.ShowRegisterLink,
			"organization_name":  loginProps.OrganizationName,
		})

	case middleware.ResponseModeFragment:
		// Return HTML fragment for HTMX requests
		c.Set("Content-Type", "text/html; charset=utf-8")
		return renderTemplate(c, auth.LoginForm(loginProps))

	default: // ResponseModePage
		// Return full HTML page for browser navigation
		c.Set("Content-Type", "text/html; charset=utf-8")
		return renderTemplate(c, auth.LoginPage(loginProps))
	}
}

// Login handles POST /login - processes login and returns appropriate response
func (h *AuthUnifiedHandler) Login(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "auth.login")
	defer span.End()

	// Parse login request
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleLoginError(c, "Invalid request body", nil)
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		return h.handleLoginError(c, "Email and password are required", map[string]string{
			"email":    "Email is required",
			"password": "Password is required",
		})
	}

	h.logger.InfoContext(ctx, "Processing login request", logger.Fields{
		"email": req.Email,
	})

	// TODO: Implement actual authentication logic
	// For now, simulate a successful login
	authResult := LoginResponse{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: UserResponse{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Email: req.Email,
			Name:  "John Doe",
		},
	}

	// Get response mode from content negotiation middleware
	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		// Return JSON response for API consumers
		return Success(c, authResult)

	case middleware.ResponseModeFragment:
		// For HTMX requests, redirect using HX-Redirect header
		middleware.SetHTMXHeaders(c, middleware.HTMXResponseHeaders{
			Redirect: "/dashboard",
		})
		return c.SendStatus(200)

	default: // ResponseModePage
		// For browser requests, use traditional redirect
		return c.Redirect("/dashboard", 302)
	}
}

// Logout handles POST /logout
func (h *AuthUnifiedHandler) Logout(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "auth.logout")
	defer span.End()

	h.logger.InfoContext(ctx, "Processing logout request")

	// TODO: Implement actual logout logic (invalidate token, etc.)

	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		return Success(c, fiber.Map{"message": "Logged out successfully"})

	case middleware.ResponseModeFragment:
		// For HTMX requests, redirect to login
		middleware.SetHTMXHeaders(c, middleware.HTMXResponseHeaders{
			Redirect: "/login",
		})
		return c.SendStatus(200)

	default: // ResponseModePage
		return c.Redirect("/login", 302)
	}
}

// handleLoginError handles login errors with appropriate response format
func (h *AuthUnifiedHandler) handleLoginError(c *fiber.Ctx, message string, fieldErrors map[string]string) error {
	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		// Convert string map to interface map for ValidationError
		var interfaceErrors map[string]interface{}
		if fieldErrors != nil {
			interfaceErrors = make(map[string]interface{})
			for k, v := range fieldErrors {
				interfaceErrors[k] = v
			}
		}
		return ValidationError(c, message, interfaceErrors)

	case middleware.ResponseModeFragment:
		// Return login form with errors for HTMX requests
		loginProps := auth.LoginPageProps{
			Title:            "Sign In",
			Subtitle:         "Welcome back! Please sign in to your account.",
			ShowRememberMe:   true,
			ShowForgotLink:   true,
			ShowRegisterLink: false,
			OrganizationName: "ERP System",
			HxPost:           "/auth/login",
			HxTarget:         "#login-form",
			Errors:           fieldErrors,
		}
		if fieldErrors == nil {
			loginProps.Errors = map[string]string{"general": message}
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return renderTemplate(c, auth.LoginForm(loginProps))

	default: // ResponseModePage
		// Return full page with errors for browser requests
		loginProps := auth.LoginPageProps{
			Title:            "Sign In",
			Subtitle:         "Welcome back! Please sign in to your account.",
			ShowRememberMe:   true,
			ShowForgotLink:   true,
			ShowRegisterLink: false,
			OrganizationName: "ERP System",
			Errors:           fieldErrors,
		}
		if fieldErrors == nil {
			loginProps.Errors = map[string]string{"general": message}
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Status(400)
		return renderTemplate(c, auth.LoginPage(loginProps))
	}
}

// renderTemplate renders a templ component and returns it as response
func renderTemplate(c *fiber.Ctx, component interface{}) error {
	// TODO: Implement proper templ rendering
	// For now, return a placeholder HTML response
	html := `<!DOCTYPE html>
<html>
<head>
    <title>ERP System</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/htmx.org@1.9.8"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div class="max-w-md w-full space-y-8">
            <div>
                <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                    Sign in to your account
                </h2>
            </div>
            <form class="mt-8 space-y-6" hx-post="/auth/login" hx-target="#login-form" hx-swap="outerHTML">
                <div id="login-form">
                    <div class="rounded-md shadow-sm -space-y-px">
                        <div>
                            <input id="email" name="email" type="email" required 
                                   class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm" 
                                   placeholder="Email address">
                        </div>
                        <div>
                            <input id="password" name="password" type="password" required 
                                   class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm" 
                                   placeholder="Password">
                        </div>
                    </div>
                    <div class="mt-6">
                        <button type="submit" 
                                class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                            Sign in
                        </button>
                    </div>
                </div>
            </form>
        </div>
    </div>
</body>
</html>`

	return c.SendString(html)
}

// SetupAuthUnifiedRoutes sets up unified auth routes
func SetupAuthUnifiedRoutes(app fiber.Router, handler *AuthUnifiedHandler) {
	// Login routes
	app.Get("/login", handler.ShowLogin)
	app.Post("/login", handler.Login)
	app.Post("/logout", handler.Logout)

	// API routes with explicit /api prefix
	api := app.Group("/api")
	api.Get("/login", handler.ShowLogin)
	api.Post("/login", handler.Login)
	api.Post("/logout", handler.Logout)
}
