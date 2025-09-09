package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
)

// SecurityHeadersConfig defines the configuration for security headers
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSPPolicy string `json:"csp_policy"`
	
	// HTTP Strict Transport Security
	HSTSMaxAge            int  `json:"hsts_max_age"`
	HSTSIncludeSubDomains bool `json:"hsts_include_subdomains"`
	HSTSPreload           bool `json:"hsts_preload"`
	
	// X-Frame-Options
	FrameOptions string `json:"frame_options"` // DENY, SAMEORIGIN, or ALLOW-FROM uri
	
	// X-Content-Type-Options
	ContentTypeOptions bool `json:"content_type_options"`
	
	// X-XSS-Protection
	XSSProtection bool `json:"xss_protection"`
	
	// Referrer Policy
	ReferrerPolicy string `json:"referrer_policy"`
	
	// Permission Policy (formerly Feature Policy)
	PermissionsPolicy string `json:"permissions_policy"`
	
	// Cross-Origin policies
	CrossOriginEmbedderPolicy string `json:"cross_origin_embedder_policy"`
	CrossOriginOpenerPolicy   string `json:"cross_origin_opener_policy"`
	CrossOriginResourcePolicy string `json:"cross_origin_resource_policy"`
	
	// Custom headers
	CustomHeaders map[string]string `json:"custom_headers"`
	
	// Environment-specific settings
	Development bool `json:"development"`
}

// SecurityHeadersMiddleware provides comprehensive security headers middleware
type SecurityHeadersMiddleware struct {
	config  SecurityHeadersConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(
	config SecurityHeadersConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
) *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{
		config:  config,
		logger:  logger,
		metrics: metrics,
	}
}

// DefaultSecurityHeadersConfig returns a secure default configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		// Strict CSP for ERP application
		CSPPolicy: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' https: wss:; " +
			"media-src 'none'; " +
			"object-src 'none'; " +
			"frame-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests",

		// HSTS for 1 year with subdomains and preload
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubDomains: true,
		HSTSPreload:           true,

		// Frame options
		FrameOptions: "DENY",

		// Enable content type options
		ContentTypeOptions: true,

		// Enable XSS protection
		XSSProtection: true,

		// Strict referrer policy
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Restrictive permissions policy
		PermissionsPolicy: "camera=(), microphone=(), geolocation=(), " +
			"payment=(), usb=(), magnetometer=(), gyroscope=(), " +
			"accelerometer=(), fullscreen=(self)",

		// Cross-origin policies
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "cross-origin",

		// Custom security headers
		CustomHeaders: map[string]string{
			"X-Robots-Tag":           "noindex, nofollow, nosnippet, noarchive",
			"X-Permitted-Cross-Domain-Policies": "none",
			"X-DNS-Prefetch-Control": "off",
		},

		Development: false,
	}
}

// DevelopmentSecurityHeadersConfig returns a more relaxed configuration for development
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	config := DefaultSecurityHeadersConfig()
	
	// More relaxed CSP for development
	config.CSPPolicy = "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; " +
		"style-src 'self' 'unsafe-inline' https:; " +
		"img-src 'self' data: https: http:; " +
		"connect-src 'self' https: http: ws: wss:; " +
		"font-src 'self' https: data:; " +
		"media-src 'self' https: data:; " +
		"object-src 'self'; " +
		"frame-src 'self';"

	// Less strict HSTS for development
	config.HSTSMaxAge = 0 // Disable HSTS in development
	config.HSTSIncludeSubDomains = false
	config.HSTSPreload = false

	// Allow framing in development
	config.FrameOptions = "SAMEORIGIN"

	// Less restrictive cross-origin policies
	config.CrossOriginEmbedderPolicy = "unsafe-none"
	config.CrossOriginOpenerPolicy = "unsafe-none"

	config.Development = true
	
	return config
}

// HTTPMiddleware creates an HTTP middleware that adds security headers
func (m *SecurityHeadersMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Add security headers before processing the request
			m.addSecurityHeaders(w, r)

			// Record metrics
			m.recordSecurityHeadersMetrics(r.Method, r.URL.Path, time.Since(start))

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// addSecurityHeaders adds all configured security headers to the response
func (m *SecurityHeadersMiddleware) addSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()

	// Content Security Policy
	if m.config.CSPPolicy != "" {
		headers.Set("Content-Security-Policy", m.config.CSPPolicy)
	}

	// HTTP Strict Transport Security
	if m.config.HSTSMaxAge > 0 && r.TLS != nil {
		hstsValue := fmt.Sprintf("max-age=%d", m.config.HSTSMaxAge)
		if m.config.HSTSIncludeSubDomains {
			hstsValue += "; includeSubDomains"
		}
		if m.config.HSTSPreload {
			hstsValue += "; preload"
		}
		headers.Set("Strict-Transport-Security", hstsValue)
	}

	// X-Frame-Options
	if m.config.FrameOptions != "" {
		headers.Set("X-Frame-Options", m.config.FrameOptions)
	}

	// X-Content-Type-Options
	if m.config.ContentTypeOptions {
		headers.Set("X-Content-Type-Options", "nosniff")
	}

	// X-XSS-Protection
	if m.config.XSSProtection {
		headers.Set("X-XSS-Protection", "1; mode=block")
	}

	// Referrer Policy
	if m.config.ReferrerPolicy != "" {
		headers.Set("Referrer-Policy", m.config.ReferrerPolicy)
	}

	// Permissions Policy
	if m.config.PermissionsPolicy != "" {
		headers.Set("Permissions-Policy", m.config.PermissionsPolicy)
	}

	// Cross-Origin policies
	if m.config.CrossOriginEmbedderPolicy != "" {
		headers.Set("Cross-Origin-Embedder-Policy", m.config.CrossOriginEmbedderPolicy)
	}

	if m.config.CrossOriginOpenerPolicy != "" {
		headers.Set("Cross-Origin-Opener-Policy", m.config.CrossOriginOpenerPolicy)
	}

	if m.config.CrossOriginResourcePolicy != "" {
		headers.Set("Cross-Origin-Resource-Policy", m.config.CrossOriginResourcePolicy)
	}

	// Custom headers
	for name, value := range m.config.CustomHeaders {
		headers.Set(name, value)
	}

	// Remove server information
	headers.Set("Server", "")

	// Cache control for sensitive pages
	if m.isSensitivePath(r.URL.Path) {
		headers.Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")
	}
}

// isSensitivePath determines if a path contains sensitive information
func (m *SecurityHeadersMiddleware) isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/auth/",
		"/admin/",
		"/api/v1/auth/",
		"/api/v1/user/",
		"/api/v1/finance/",
		"/api/v1/admin/",
		"/swagger-ui/",
	}

	pathLower := strings.ToLower(path)
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(pathLower, sensitive) {
			return true
		}
	}

	return false
}

// recordSecurityHeadersMetrics records metrics for security headers middleware
func (m *SecurityHeadersMiddleware) recordSecurityHeadersMetrics(method, path string, duration time.Duration) {
	labels := metrics.Fields{
		"method": method,
		"sensitive": func() string {
			if m.isSensitivePath(path) {
				return "true"
			}
			return "false"
		}(),
	}

	m.metrics.IncrementCounter("security_headers_requests_total", labels)
	m.metrics.ObserveHistogram("security_headers_middleware_duration", duration.Seconds(), labels)
}

// SecurityHeadersReport contains information about applied security headers
type SecurityHeadersReport struct {
	Applied   []string `json:"applied"`
	Skipped   []string `json:"skipped"`
	CustomSet int      `json:"custom_set"`
}

// GetSecurityHeadersReport returns a report of which security headers were applied
func (m *SecurityHeadersMiddleware) GetSecurityHeadersReport() SecurityHeadersReport {
	report := SecurityHeadersReport{
		Applied:   []string{},
		Skipped:   []string{},
		CustomSet: len(m.config.CustomHeaders),
	}

	// Check which headers would be applied
	if m.config.CSPPolicy != "" {
		report.Applied = append(report.Applied, "Content-Security-Policy")
	} else {
		report.Skipped = append(report.Skipped, "Content-Security-Policy")
	}

	if m.config.HSTSMaxAge > 0 {
		report.Applied = append(report.Applied, "Strict-Transport-Security")
	} else {
		report.Skipped = append(report.Skipped, "Strict-Transport-Security")
	}

	if m.config.FrameOptions != "" {
		report.Applied = append(report.Applied, "X-Frame-Options")
	} else {
		report.Skipped = append(report.Skipped, "X-Frame-Options")
	}

	if m.config.ContentTypeOptions {
		report.Applied = append(report.Applied, "X-Content-Type-Options")
	} else {
		report.Skipped = append(report.Skipped, "X-Content-Type-Options")
	}

	if m.config.XSSProtection {
		report.Applied = append(report.Applied, "X-XSS-Protection")
	} else {
		report.Skipped = append(report.Skipped, "X-XSS-Protection")
	}

	if m.config.ReferrerPolicy != "" {
		report.Applied = append(report.Applied, "Referrer-Policy")
	} else {
		report.Skipped = append(report.Skipped, "Referrer-Policy")
	}

	if m.config.PermissionsPolicy != "" {
		report.Applied = append(report.Applied, "Permissions-Policy")
	} else {
		report.Skipped = append(report.Skipped, "Permissions-Policy")
	}

	// Cross-origin policies
	crossOriginPolicies := []string{
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy", 
		"Cross-Origin-Resource-Policy",
	}

	values := []string{
		m.config.CrossOriginEmbedderPolicy,
		m.config.CrossOriginOpenerPolicy,
		m.config.CrossOriginResourcePolicy,
	}

	for i, policy := range crossOriginPolicies {
		if values[i] != "" {
			report.Applied = append(report.Applied, policy)
		} else {
			report.Skipped = append(report.Skipped, policy)
		}
	}

	return report
}

// ValidateSecurityConfig validates the security headers configuration
func ValidateSecurityConfig(config SecurityHeadersConfig) error {
	// Validate Frame Options
	if config.FrameOptions != "" {
		validFrameOptions := []string{"DENY", "SAMEORIGIN"}
		valid := false
		for _, option := range validFrameOptions {
			if config.FrameOptions == option {
				valid = true
				break
			}
		}
		if !valid && !strings.HasPrefix(config.FrameOptions, "ALLOW-FROM ") {
			return fmt.Errorf("invalid frame options: %s", config.FrameOptions)
		}
	}

	// Validate HSTS settings
	if config.HSTSMaxAge < 0 {
		return fmt.Errorf("HSTS max age cannot be negative")
	}

	// Validate Referrer Policy
	if config.ReferrerPolicy != "" {
		validPolicies := []string{
			"no-referrer",
			"no-referrer-when-downgrade",
			"origin",
			"origin-when-cross-origin",
			"same-origin",
			"strict-origin",
			"strict-origin-when-cross-origin",
			"unsafe-url",
		}
		valid := false
		for _, policy := range validPolicies {
			if config.ReferrerPolicy == policy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid referrer policy: %s", config.ReferrerPolicy)
		}
	}

	// Validate Cross-Origin policies
	if config.CrossOriginEmbedderPolicy != "" {
		validCOEP := []string{"unsafe-none", "require-corp"}
		valid := false
		for _, policy := range validCOEP {
			if config.CrossOriginEmbedderPolicy == policy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid Cross-Origin-Embedder-Policy: %s", config.CrossOriginEmbedderPolicy)
		}
	}

	if config.CrossOriginOpenerPolicy != "" {
		validCOOP := []string{"unsafe-none", "same-origin-allow-popups", "same-origin"}
		valid := false
		for _, policy := range validCOOP {
			if config.CrossOriginOpenerPolicy == policy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid Cross-Origin-Opener-Policy: %s", config.CrossOriginOpenerPolicy)
		}
	}

	if config.CrossOriginResourcePolicy != "" {
		validCORP := []string{"same-site", "same-origin", "cross-origin"}
		valid := false
		for _, policy := range validCORP {
			if config.CrossOriginResourcePolicy == policy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid Cross-Origin-Resource-Policy: %s", config.CrossOriginResourcePolicy)
		}
	}

	return nil
}