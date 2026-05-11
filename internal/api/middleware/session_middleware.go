package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared"
)

// validSubjectPrefixes are the only prefixes a resolved Principal subject may
// carry.  Any other value indicates a corrupted session or injection attempt.
// Rule 1 — docs/reference/modules/iam/20-business-rules-and-validation.md.
var validSubjectPrefixes = []string{"platform:", "tenant:", "portal:", "api:"}

// isValidSubject returns true if subject has one of the four recognised
// prefixes and a non-empty suffix after that prefix.
func isValidSubject(subject string) bool {
	for _, prefix := range validSubjectPrefixes {
		if strings.HasPrefix(subject, prefix) {
			return len(subject) > len(prefix) // must have a non-empty id part
		}
	}
	return false
}

// splitPermission splits a dotted permission string into (object, action).
// "finance.accounts.read" → ("finance.accounts", "read")
// "finance.accounts"      → ("finance", "accounts")
func splitPermission(perm string) (object, action string) {
	i := strings.LastIndex(perm, ".")
	if i < 0 {
		return perm, "*"
	}
	return perm[:i], perm[i+1:]
}

// AuthConfig holds configuration for session-based authentication middleware.
type AuthConfig struct {
	// SessionService validates tokens and returns ResolvedSessions.
	SessionService iam.SessionService
	// AuthzService enforces permission checks in Authorize middleware.
	// Required when using Authorize; optional otherwise.
	AuthzService iam.AuthzService
	// APIKeyService validates "eak_" prefixed bearer tokens.
	// Optional — API key auth is skipped when nil.
	APIKeyService iam.APIKeyService
	// CookieName is the HttpOnly cookie that carries the raw session token.
	CookieName string
}

// DefaultAuthConfig returns an AuthConfig with safe defaults backed by svc.
func DefaultAuthConfig(svc iam.SessionService) AuthConfig {
	return AuthConfig{
		SessionService: svc,
		CookieName:     "session",
	}
}

// Authenticate is a Fiber middleware that validates the session token from
// the cookie (or Authorization: Bearer header) and stores the ResolvedSession
// in c.Locals(iam.LocalsKeySession).
//
// Returns 401 if the token is missing or invalid.
func Authenticate(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(cfg.CookieName)
		if token == "" {
			if h := c.Get(fiber.HeaderAuthorization); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		// API key path: "eak_" prefix identifies a machine-to-machine bearer token.
		if cfg.APIKeyService != nil && strings.HasPrefix(token, "eak_") {
			resolved, err := cfg.APIKeyService.ValidateAPIKey(c.Context(), token)
			if err != nil || resolved == nil {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid or revoked API key")
			}
			setSessionLocals(c, resolved)
			if !isValidSubject(resolved.ToPrincipal().Subject) {
				return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
			}
			return c.Next()
		}

		// Session cookie / bearer token path.
		resolved, err := cfg.SessionService.ValidateSession(c.Context(), token)
		if err != nil || resolved == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired session")
		}

		setSessionLocals(c, resolved)
		if !isValidSubject(resolved.ToPrincipal().Subject) {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		return c.Next()
	}
}

// Authorize returns a Fiber middleware that enforces the given permission via
// the Casbin authz service stored in cfg.
//
// Must run after Authenticate (requires LocalsKeySession and LocalsKeyPrincipal).
// Returns 403 if the permission is denied or cfg.AuthzService is nil.
func Authorize(cfg AuthConfig, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		if cfg.AuthzService == nil {
			return fiber.NewError(fiber.StatusForbidden, "permission denied")
		}
		principal, ok := c.Locals(iam.LocalsKeyPrincipal).(iam.Principal)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		obj, act := splitPermission(permission)
		allowed, err := cfg.AuthzService.Enforce(c.Context(), iam.Request{
			Subject: principal.Subject,
			Domain:  principal.Domain,
			Object:  obj,
			Action:  act,
		})
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "permission denied")
		}
		return c.Next()
	}
}

// RequireFlag returns a Fiber middleware that checks whether the authenticated
// session has the named feature flag enabled in its pre-computed Configuration.
//
// Must run after Authenticate. Returns 403 with "feature not enabled" if the
// flag is absent or false.
func RequireFlag(flagKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		if !sess.FeatureEnabled(flagKey) {
			return fiber.NewError(fiber.StatusForbidden, "feature not enabled")
		}
		return c.Next()
	}
}

// setSessionLocals populates all Fiber Locals and Go context values required
// by downstream handlers and authorization middleware:
//   - iam.LocalsKeySession   → *iam.ResolvedSession  (permissions, flags, settings)
//   - iam.LocalsKeyPrincipal → iam.Principal         (Casbin subject + domain)
//   - cache.TenantIDKey      → tenant UUID string     (RLS context for SQLC queries)
func setSessionLocals(c *fiber.Ctx, resolved *iam.ResolvedSession) {
	c.Locals(iam.LocalsKeySession, resolved)
	c.Locals(iam.LocalsKeyPrincipal, resolved.ToPrincipal())
	// Wrap the existing user context (from tenant middleware) preserving prior values.
	// Set shared.TenantIDKey (uuid.UUID) for WithTenantFromCtx / DB layer.
	// Set cache.TenantIDKey (string) for the cache service's tenant namespace lookup.
	ctx := shared.WithTenantID(c.UserContext(), resolved.TenantID)
	ctx = context.WithValue(ctx, cache.TenantIDKey, resolved.TenantID.String())
	c.SetUserContext(ctx)
}
