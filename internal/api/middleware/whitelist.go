package middleware

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"awo.so/internal/shared/logger"
	"gopkg.in/yaml.v3"
)

// EndpointWhitelist manages public endpoints that don't require tenant context
type EndpointWhitelist struct {
	patterns     []string
	exactMatches map[string]bool
	compiled     []*regexp.Regexp // Pre-compiled patterns for performance
}

// PublicEndpointsConfig represents the YAML configuration structure
type PublicEndpointsConfig struct {
	PublicEndpoints struct {
		Patterns     []string `yaml:"patterns"`
		ExactMatches []string `yaml:"exact_matches"`
	} `yaml:"public_endpoints"`
	Validation struct {
		RequireApproval bool `yaml:"require_approval"`
		ChangeTracking  bool `yaml:"change_tracking"`
		SecurityReview  bool `yaml:"security_review"`
	} `yaml:"validation"`
}

// NewEndpointWhitelist creates a new whitelist from configuration
func NewEndpointWhitelist(patterns []string, exactMatches []string) (*EndpointWhitelist, error) {
	whitelist := &EndpointWhitelist{
		patterns:     patterns,
		exactMatches: make(map[string]bool),
		compiled:     make([]*regexp.Regexp, 0, len(patterns)),
	}

	// Convert exact matches to map for O(1) lookup
	for _, endpoint := range exactMatches {
		whitelist.exactMatches[endpoint] = true
	}

	// Pre-compile regex patterns for performance
	for _, pattern := range patterns {
		// Convert filepath patterns to regex
		regexPattern := convertToRegex(pattern)
		compiled, err := regexp.Compile(regexPattern)
		if err != nil {
			logger.Warn("Invalid whitelist pattern, skipping", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
			continue
		}
		whitelist.compiled = append(whitelist.compiled, compiled)
	}

	logger.Info("Endpoint whitelist initialized", logger.Fields{
		"exact_matches":     len(whitelist.exactMatches),
		"patterns":          len(patterns),
		"compiled_patterns": len(whitelist.compiled),
	})

	return whitelist, nil
}

// LoadWhitelistFromYAML loads whitelist configuration from YAML data
func LoadWhitelistFromYAML(yamlData []byte) (*EndpointWhitelist, error) {
	var config PublicEndpointsConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse whitelist YAML: %w", err)
	}

	return NewEndpointWhitelist(
		config.PublicEndpoints.Patterns,
		config.PublicEndpoints.ExactMatches,
	)
}

// IsPublicEndpoint checks if an endpoint should bypass tenant validation
// Returns true if the endpoint is whitelisted (public)
func (w *EndpointWhitelist) IsPublicEndpoint(method, path string) bool {
	if w == nil {
		// If no whitelist is configured, all endpoints require tenant context
		return false
	}

	// Normalize the path
	path = strings.TrimSpace(path)
	method = strings.ToUpper(strings.TrimSpace(method))

	// Create full endpoint string for pattern matching
	fullEndpoint := method + " " + path

	// Check exact matches first (fastest lookup)
	if w.exactMatches[fullEndpoint] {
		logger.Debug("Endpoint matched exact whitelist entry", logger.Fields{
			"endpoint": fullEndpoint,
		})
		return true
	}

	// Also check path-only exact matches (for backwards compatibility)
	if w.exactMatches[path] {
		logger.Debug("Path matched exact whitelist entry", logger.Fields{
			"path": path,
		})
		return true
	}

	// Check compiled regex patterns
	for _, regex := range w.compiled {
		if regex.MatchString(fullEndpoint) {
			logger.Debug("Endpoint matched whitelist pattern", logger.Fields{
				"endpoint": fullEndpoint,
				"pattern":  regex.String(),
			})
			return true
		}

		// Also check path-only pattern matches
		if regex.MatchString(path) {
			logger.Debug("Path matched whitelist pattern", logger.Fields{
				"path":    path,
				"pattern": regex.String(),
			})
			return true
		}
	}

	// Check simple glob patterns using filepath.Match as fallback
	for _, pattern := range w.patterns {
		if matched, _ := filepath.Match(pattern, fullEndpoint); matched {
			logger.Debug("Endpoint matched glob pattern", logger.Fields{
				"endpoint": fullEndpoint,
				"pattern":  pattern,
			})
			return true
		}

		if matched, _ := filepath.Match(pattern, path); matched {
			logger.Debug("Path matched glob pattern", logger.Fields{
				"path":    path,
				"pattern": pattern,
			})
			return true
		}
	}

	// Not whitelisted - requires tenant context
	return false
}

// GetStats returns statistics about the whitelist
func (w *EndpointWhitelist) GetStats() map[string]int {
	if w == nil {
		return map[string]int{
			"exact_matches":     0,
			"patterns":          0,
			"compiled_patterns": 0,
		}
	}

	return map[string]int{
		"exact_matches":     len(w.exactMatches),
		"patterns":          len(w.patterns),
		"compiled_patterns": len(w.compiled),
	}
}

// ValidateConfiguration validates the whitelist configuration
func (w *EndpointWhitelist) ValidateConfiguration() []string {
	var issues []string

	if w == nil {
		issues = append(issues, "whitelist is nil")
		return issues
	}

	// Check for common security issues
	for endpoint := range w.exactMatches {
		if strings.Contains(endpoint, "*") {
			issues = append(issues, fmt.Sprintf("exact match contains wildcard: %s", endpoint))
		}

		// Check for overly permissive patterns
		if strings.HasPrefix(endpoint, "GET /") && strings.Contains(endpoint, "admin") {
			issues = append(issues, fmt.Sprintf("potentially insecure admin endpoint: %s", endpoint))
		}
	}

	for _, pattern := range w.patterns {
		// Check for overly broad patterns
		if pattern == "*" || pattern == "**" {
			issues = append(issues, fmt.Sprintf("overly broad pattern: %s", pattern))
		}

		// if strings.Contains(pattern, "/admin/*") {
		// 	issues = append(issues, fmt.Sprintf("potentially insecure admin pattern: %s", pattern))
		// }
	}

	return issues
}

// convertToRegex converts filepath-style patterns to regex
// This handles patterns like "GET /api/v1/health*" → "^GET /api/v1/health.*$"
func convertToRegex(pattern string) string {
	// Escape special regex characters except * and ?
	escaped := regexp.QuoteMeta(pattern)

	// Convert escaped wildcards back to regex equivalents
	escaped = strings.ReplaceAll(escaped, `\*`, `.*`)
	escaped = strings.ReplaceAll(escaped, `\?`, `.`)

	// Anchor the pattern
	return "^" + escaped + "$"
}

// DefaultWhitelist creates a whitelist with common public endpoints
func DefaultWhitelist() *EndpointWhitelist {
	patterns := []string{
		"GET /health*",
		"GET /api/v1/health*",
		"GET /openapi*",
		"GET /swagger-ui/*",
		"GET /debug/*",
		"GET /api/v1/auth/oauth/*", // OAuth begin + callback
	}

	exactMatches := []string{
		"POST /api/v1/auth/mfa/complete",
		"POST /api/v1/auth/forgot-password",
		"POST /api/v1/auth/reset-password",
		"POST /api/v1/auth/refresh",
		"GET /api/v1/auth/validate",
		"POST /api/v1/tenant/onboard",
		"GET /api/v1/version",
		"GET /health",
		"GET /ready",
	}

	whitelist, err := NewEndpointWhitelist(patterns, exactMatches)
	if err != nil {
		logger.Error("Failed to create default whitelist", logger.Fields{
			"error": err.Error(),
		})
		return nil
	}

	return whitelist
}
