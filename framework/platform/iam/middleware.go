package iam

import (
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"

	"awo.so/framework/def"
)

const (
	// viewerLocalKey is the c.Locals key where *SessionViewer is stored.
	viewerLocalKey = "iam:viewer"

	bearerPrefix = "Bearer "
)

// AuthMiddleware returns a Fiber middleware that:
//  1. Extracts the Bearer token from the Authorization header.
//  2. Looks up the session in Redis.
//  3. Sets c.Locals(viewerLocalKey, *SessionViewer) on success.
//
// Requests with no token or an invalid/expired token are allowed through
// without a viewer. Pair with RequireAuth to enforce authentication on
// specific route groups.
func AuthMiddleware(redis redis.Cmdable) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawToken := extractBearer(c)
		if rawToken == "" {
			return c.Next()
		}
		sess, err := loadSession(c.Context(), redis, rawToken)
		if err != nil {
			// Invalid / expired token — let the handler decide whether to reject.
			return c.Next()
		}
		c.Locals(viewerLocalKey, newSessionViewer(sess))
		return c.Next()
	}
}

// RequireAuth returns a Fiber middleware that rejects unauthenticated requests
// with 401. Mount it after AuthMiddleware on protected route groups.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := c.Locals(viewerLocalKey).(*SessionViewer); !ok {
			return fiber.ErrUnauthorized
		}
		return c.Next()
	}
}

// ViewerFromCtx is the api.ViewerFromCtx implementation for IAM-backed sessions.
// Register this as bootstrap.Options.ViewerFn to replace the anonymous viewer.
//
// Returns 401 when no session is present. If the route should be publicly
// accessible, do not use RequireAuth and handle nil viewer in the handler.
func ViewerFromCtx(c *fiber.Ctx) (def.ViewerContext, error) {
	v, ok := c.Locals(viewerLocalKey).(*SessionViewer)
	if !ok || v == nil {
		return nil, fiber.ErrUnauthorized
	}
	return v, nil
}

// extractBearer extracts the raw token from "Authorization: Bearer <token>".
// Returns "" when the header is absent or malformed.
func extractBearer(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, bearerPrefix) {
		return strings.TrimSpace(auth[len(bearerPrefix):])
	}
	return ""
}
