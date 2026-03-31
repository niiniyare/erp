package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/iam"
	"awo.so/internal/core/iam/domain"
)

// Request / Response types

type createAPIKeyRequest struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"` // ISO 8601; omit for no expiry
}

type createAPIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	// Token is the raw bearer token — shown ONCE and never returned again.
	Token string `json:"token"`
}

type apiKeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Handlers

// CreateAPIKeyHandler creates a new API key for the authenticated user's tenant.
// The plaintext bearer token is returned once in the response body.
func CreateAPIKeyHandler(svc iam.APIKeyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		var req createAPIKeyRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		if req.Name == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "name is required")
		}
		if len(req.Scopes) == 0 {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "at least one scope is required")
		}

		domReq := &domain.CreateAPIKeyRequest{
			Name:      req.Name,
			Scopes:    req.Scopes,
			ExpiresAt: req.ExpiresAt,
			CreatedBy: sess.UserID,
		}

		key, rawToken, err := svc.CreateAPIKey(c.UserContext(), domReq)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create API key")
		}

		return c.Status(fiber.StatusCreated).JSON(createAPIKeyResponse{
			ID:        key.ID,
			Name:      key.Name,
			Scopes:    key.Scopes,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			Token:     rawToken,
		})
	}
}

// ListAPIKeysHandler returns all API keys for the current tenant.
// The raw token is never returned in list responses.
func ListAPIKeysHandler(svc iam.APIKeyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		keys, err := svc.ListAPIKeys(c.UserContext())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to list API keys")
		}

		out := make([]apiKeyResponse, 0, len(keys))
		for _, k := range keys {
			out = append(out, toAPIKeyResponse(k))
		}
		return c.JSON(fiber.Map{"data": out})
	}
}

// RevokeAPIKeyHandler immediately revokes the specified API key.
func RevokeAPIKeyHandler(svc iam.APIKeyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		keyID, err := uuid.Parse(idStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid key id")
		}

		if err := svc.RevokeAPIKey(c.UserContext(), keyID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to revoke API key")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// Helpers

func toAPIKeyResponse(k *domain.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Scopes:     k.Scopes,
		ExpiresAt:  k.ExpiresAt,
		RevokedAt:  k.RevokedAt,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
	}
}
