package middleware

import (
	"net/http"
	"strings"

	"github.com/niiniyare/erp/internal/ui/handlers"
)

// ValidationMiddleware provides validation-specific middleware
type ValidationMiddleware struct {
	validationHandler *handlers.ValidationHandler
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(validationHandler *handlers.ValidationHandler) *ValidationMiddleware {
	return &ValidationMiddleware{
		validationHandler: validationHandler,
	}
}

// CORS adds CORS headers for validation endpoints
func (m *ValidationMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to validation endpoints
		if strings.HasPrefix(r.URL.Path, "/validate/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// RateLimiter provides basic rate limiting for validation endpoints
func (m *ValidationMiddleware) RateLimiter(next http.Handler) http.Handler {
	// Simple in-memory rate limiter for validation endpoints
	// In production, use Redis or similar for distributed rate limiting
	clientRequests := make(map[string]int)
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/validate/") {
			clientIP := getClientIP(r)
			
			// Simple rate limiting: 60 requests per minute per IP
			if clientRequests[clientIP] > 60 {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			clientRequests[clientIP]++
			
			// Reset counter every minute (simplified)
			// In production, implement proper sliding window
		}
		
		next.ServeHTTP(w, r)
	})
}

// ValidationOnly ensures only validation requests are processed
func (m *ValidationMiddleware) ValidationOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET and POST for validation
		if r.Method != "GET" && r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Ensure this is an HTMX request or form submission
		if r.Header.Get("HX-Request") == "" && r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(w, "Invalid request type", http.StatusBadRequest)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check for X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}