package middleware

import (
	"context"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam/contract"
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
	SessionService contract.SessionService
	// AuthzService enforces permission checks in Authorize middleware.
	// Required when using Authorize; optional otherwise.
	AuthzService contract.AuthzService
	// APIKeyService validates "eak_" prefixed bearer tokens.
	// Optional — API key auth is skipped when nil.
	APIKeyService contract.APIKeyService
	// CookieName is the HttpOnly cookie that carries the raw session token.
	CookieName string
}

// DefaultAuthConfig returns an AuthConfig with safe defaults backed by svc.
func DefaultAuthConfig(svc contract.SessionService) AuthConfig {
	return AuthConfig{
		SessionService: svc,
		CookieName:     "session",
	}
}

// Authenticate is a Fiber middleware that validates the session token from
// the cookie (or Authorization: Bearer header) and stores the ResolvedSession
// in c.Locals(contract.LocalsKeySession).
//
// On failure:
//   - Browser requests (Accept: text/html) → 302 redirect to /ui/login?redirect=<path>
//   - API / machine requests              → 401 JSON
func Authenticate(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(cfg.CookieName)
		if token == "" {
			if h := c.Get(fiber.HeaderAuthorization); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}
		if token == "" {
			return unauthenticated(c)
		}

		// API key path: "eak_" prefix identifies a machine-to-machine bearer token.
		if cfg.APIKeyService != nil && strings.HasPrefix(token, "eak_") {
			resolved, err := cfg.APIKeyService.ValidateAPIKey(c.Context(), token)
			if err != nil || resolved == nil {
				return unauthenticated(c)
			}
			if !isValidSubject(resolved.ToPrincipal().Subject) {
				return unauthenticated(c)
			}
			setSessionLocals(c, resolved)
			return c.Next()
		}

		// Session cookie / bearer token path.
		resolved, err := cfg.SessionService.ValidateSession(c.Context(), token)
		if err != nil || resolved == nil {
			return unauthenticated(c)
		}
		if !isValidSubject(resolved.ToPrincipal().Subject) {
			return unauthenticated(c)
		}
		setSessionLocals(c, resolved)
		return c.Next()
	}
}

// unauthenticated sends a 302 redirect for browser navigation or a 401 JSON
// response for API / machine clients.
// Browser detection: Accept header contains "text/html".
// Open-redirect prevention: only relative paths (start with "/", not "//") are echoed.
func unauthenticated(c *fiber.Ctx) error {
	if strings.Contains(c.Get(fiber.HeaderAccept), "text/html") {
		target := "/ui/login"
		if orig := c.OriginalURL(); strings.HasPrefix(orig, "/") && !strings.HasPrefix(orig, "//") && orig != "/ui/login" {
			target += "?redirect=" + url.QueryEscape(orig)
		}
		return c.Redirect(target, fiber.StatusFound)
	}
	return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
}

// Authorize returns a Fiber middleware that enforces the given permission via
// the Casbin authz service stored in cfg.
//
// Must run after Authenticate (requires LocalsKeySession and LocalsKeyPrincipal).
// Returns 403 if the permission is denied or cfg.AuthzService is nil.
func Authorize(cfg AuthConfig, permission string) fiber.Handler {
	obj, act := splitPermission(permission) // computed once at registration, not per-request
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(contract.LocalsKeySession).(*contract.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		if cfg.AuthzService == nil {
			// No authz service configured — pass through (dev/test mode).
			return c.Next()
		}
		principal, ok := c.Locals(contract.LocalsKeyPrincipal).(contract.Principal)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		allowed, err := cfg.AuthzService.Enforce(c.Context(), contract.Request{
			Subject: principal.Subject,
			Domain:  principal.Domain,
			Object:  obj,
			Action:  act,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "authorization check failed")
		}
		if !allowed {
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
		sess, ok := c.Locals(contract.LocalsKeySession).(*contract.ResolvedSession)
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
//   - contract.LocalsKeySession   → *contract.ResolvedSession  (permissions, flags, settings)
//   - contract.LocalsKeyPrincipal → contract.Principal         (Casbin subject + domain)
//   - cache.TenantIDKey           → tenant UUID string          (RLS context for SQLC queries)
func setSessionLocals(c *fiber.Ctx, resolved *contract.ResolvedSession) {
	c.Locals(contract.LocalsKeySession, resolved)
	c.Locals(contract.LocalsKeyPrincipal, resolved.ToPrincipal())
	// Build enriched Go context preserving any prior values (e.g. from TenantMiddleware).
	// shared.TenantIDKey  (uuid.UUID) — DB layer / WithTenantFromCtx
	// cache.TenantIDKey   (string)    — cache namespace lookup
	// contract.contextKey (SessionContext) — SchemaHandler pipeline
	ctx := shared.WithTenantID(c.UserContext(), resolved.TenantID)
	ctx = context.WithValue(ctx, cache.TenantIDKey, resolved.TenantID.String())
	ctx = contract.WithContext(ctx, contract.NewSessionContext(resolved))
	c.SetUserContext(ctx)
}
