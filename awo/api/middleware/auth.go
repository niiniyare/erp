package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/awo/api/response"
	"awo.so/awo/auth"
	"awo.so/awo/platform/iam"
	"awo.so/awo/runtime"
)

// RequireAuth validates the Bearer token, loads the session from Redis, builds
// a ViewerContext, and stores both in Fiber locals for downstream handlers.
//
// Locals set on success:
//   - "session"  (*auth.Session)      — raw session, used by logout/me handlers
//   - "viewer"   (auth.ViewerContext) — authorization surface for hooks and policies
//
// Redis failure causes 503 (correct — cannot authenticate without session store).
func RequireAuth(svc *iam.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Wrap(&runtime.BusinessError{
				Code:    "iam.missing_token",
				Message: "Authentication token is required.",
				Status:  401,
			}))
		}

		session, err := svc.ValidateToken(c.UserContext(), token)
		if err != nil {
			return c.Status(response.HTTPStatus(err)).JSON(response.Wrap(err))
		}

		viewer := session.ToViewer()

		// Embed viewer in Go context so hooks and policies can call auth.ViewerFromContext.
		ctx := auth.WithViewer(c.UserContext(), viewer)
		c.SetUserContext(ctx)

		// Store session and viewer in Fiber locals for handlers that need them directly.
		c.Locals("session", session)
		c.Locals("viewer", viewer)

		return c.Next()
	}
}

func extractBearerToken(c *fiber.Ctx) string {
	h := c.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}
