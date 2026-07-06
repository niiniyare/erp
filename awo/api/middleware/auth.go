package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/awo/api/response"
	"awo.so/awo/platform/iam"
	"awo.so/awo/runtime"
)

// RequireAuth validates the Bearer token and populates c.Locals("user_id")
// and c.Locals("session_id"). Must run after TenantResolver.
//
// Redis failure causes 503 (correct — cannot authenticate without session store).
func RequireAuth(svc *iam.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Wrap(&runtime.BusinessError{
				Code:    "iam.missing_token",
				Message: "Authentication token is required",
				Status:  401,
			}))
		}

		claims, err := svc.ValidateToken(c.UserContext(), token)
		if err != nil {
			return c.Status(response.HTTPStatus(err)).JSON(response.Wrap(err))
		}

		c.Locals("user_id", claims.UserID.String())
		c.Locals("session_id", claims.SessionID.String())
		return c.Next()
	}
}

func extractBearerToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}
