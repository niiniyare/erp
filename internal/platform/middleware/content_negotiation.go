package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ResponseMode represents the type of response expected
type ResponseMode string

const (
	ResponseModePage     ResponseMode = "page"     // Full HTML page with layout
	ResponseModeFragment ResponseMode = "fragment" // HTML fragment for HTMX
	ResponseModeJSON     ResponseMode = "json"     // JSON API response
)

// RequestType represents the type of incoming request
type RequestType string

const (
	RequestTypeBrowser RequestType = "browser" // Browser navigation
	RequestTypeHTMX    RequestType = "htmx"    // HTMX request
	RequestTypeAPI     RequestType = "api"     // API consumer
)

// ContentNegotiationMiddleware intelligently detects request type and sets response mode
func ContentNegotiationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestType, responseMode := determineRequestTypeAndMode(c)

		// Set context values for handlers to use
		c.Locals("requestType", requestType)
		c.Locals("responseMode", responseMode)

		// Set response headers based on mode
		setResponseHeaders(c, responseMode)

		return c.Next()
	}
}

// determineRequestTypeAndMode analyzes the request to determine appropriate response
func determineRequestTypeAndMode(c *fiber.Ctx) (RequestType, ResponseMode) {
	path := c.Path()
	method := c.Method()

	// 1. Check for explicit API routes
	if strings.HasPrefix(path, "/api/") {
		return RequestTypeAPI, ResponseModeJSON
	}

	// 2. Check for HTMX requests
	if isHTMXRequest(c) {
		// HTMX requests usually want fragments, but some may want JSON
		if wantsJSON(c) {
			return RequestTypeHTMX, ResponseModeJSON
		}
		return RequestTypeHTMX, ResponseModeFragment
	}

	// 3. Check for explicit JSON requests
	if wantsJSON(c) {
		return RequestTypeAPI, ResponseModeJSON
	}

	// 4. Check for form submissions that might want different responses
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		// Form submissions from HTMX
		if hasHTMXHeaders(c) {
			if wantsJSON(c) {
				return RequestTypeHTMX, ResponseModeJSON
			}
			return RequestTypeHTMX, ResponseModeFragment
		}

		// Regular form submissions - redirect or show page
		return RequestTypeBrowser, ResponseModePage
	}

	// 5. Default to full page for browser navigation
	return RequestTypeBrowser, ResponseModePage
}

// isHTMXRequest checks if the request is from HTMX
func isHTMXRequest(c *fiber.Ctx) bool {
	return c.Get("HX-Request") == "true"
}

// hasHTMXHeaders checks for any HTMX-related headers
func hasHTMXHeaders(c *fiber.Ctx) bool {
	htmxHeaders := []string{
		"HX-Request",
		"HX-Target",
		"HX-Trigger",
		"HX-Current-URL",
		"HX-Boosted",
	}

	for _, header := range htmxHeaders {
		if c.Get(header) != "" {
			return true
		}
	}
	return false
}

// wantsJSON checks if the client expects JSON response
func wantsJSON(c *fiber.Ctx) bool {
	accept := c.Get("Accept")
	contentType := c.Get("Content-Type")

	// Check Accept header
	if strings.Contains(accept, "application/json") {
		return true
	}

	// Check if it's a JSON content type
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// Check for API-like patterns in URL
	if strings.Contains(c.Path(), "/api/") {
		return true
	}

	// Check for AJAX requests that want JSON
	if c.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}

	return false
}

// setResponseHeaders sets appropriate headers based on response mode
func setResponseHeaders(c *fiber.Ctx, mode ResponseMode) {
	switch mode {
	case ResponseModeJSON:
		c.Set("Content-Type", "application/json")
	case ResponseModeFragment:
		c.Set("Content-Type", "text/html; charset=utf-8")
		// Set cache headers for fragments
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	case ResponseModePage:
		c.Set("Content-Type", "text/html; charset=utf-8")
	}
}

// Helper functions for handlers

// GetResponseMode returns the response mode from context
func GetResponseMode(c *fiber.Ctx) ResponseMode {
	if mode, ok := c.Locals("responseMode").(ResponseMode); ok {
		return mode
	}
	return ResponseModePage // default
}

// GetRequestType returns the request type from context
func GetRequestType(c *fiber.Ctx) RequestType {
	if reqType, ok := c.Locals("requestType").(RequestType); ok {
		return reqType
	}
	return RequestTypeBrowser // default
}

// IsHTMXRequest checks if current request is from HTMX
func IsHTMXRequest(c *fiber.Ctx) bool {
	return GetRequestType(c) == RequestTypeHTMX
}

// IsAPIRequest checks if current request is an API request
func IsAPIRequest(c *fiber.Ctx) bool {
	return GetRequestType(c) == RequestTypeAPI
}

// WantsJSON checks if the client wants JSON response
func WantsJSON(c *fiber.Ctx) bool {
	return GetResponseMode(c) == ResponseModeJSON
}

// WantsFragment checks if the client wants HTML fragment
func WantsFragment(c *fiber.Ctx) bool {
	return GetResponseMode(c) == ResponseModeFragment
}

// WantsFullPage checks if the client wants full HTML page
func WantsFullPage(c *fiber.Ctx) bool {
	return GetResponseMode(c) == ResponseModePage
}

// HTMXResponseHeaders contains HTMX-specific response headers
type HTMXResponseHeaders struct {
	Trigger       string // HX-Trigger
	TriggerAfter  string // HX-Trigger-After-Swap
	TriggerSettle string // HX-Trigger-After-Settle
	Redirect      string // HX-Redirect
	Refresh       bool   // HX-Refresh
	Location      string // HX-Location
	PushURL       string // HX-Push-Url
	ReplaceURL    string // HX-Replace-Url
	Reswap        string // HX-Reswap
	Retarget      string // HX-Retarget
}

// SetHTMXHeaders sets HTMX-specific response headers
func SetHTMXHeaders(c *fiber.Ctx, headers HTMXResponseHeaders) {
	if headers.Trigger != "" {
		c.Set("HX-Trigger", headers.Trigger)
	}
	if headers.TriggerAfter != "" {
		c.Set("HX-Trigger-After-Swap", headers.TriggerAfter)
	}
	if headers.TriggerSettle != "" {
		c.Set("HX-Trigger-After-Settle", headers.TriggerSettle)
	}
	if headers.Redirect != "" {
		c.Set("HX-Redirect", headers.Redirect)
	}
	if headers.Refresh {
		c.Set("HX-Refresh", "true")
	}
	if headers.Location != "" {
		c.Set("HX-Location", headers.Location)
	}
	if headers.PushURL != "" {
		c.Set("HX-Push-Url", headers.PushURL)
	}
	if headers.ReplaceURL != "" {
		c.Set("HX-Replace-Url", headers.ReplaceURL)
	}
	if headers.Reswap != "" {
		c.Set("HX-Reswap", headers.Reswap)
	}
	if headers.Retarget != "" {
		c.Set("HX-Retarget", headers.Retarget)
	}
}
