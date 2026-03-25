package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents a machine-to-machine (M2M) API credential belonging to a tenant.
//
// Security invariants:
//   - The raw bearer token is returned ONCE at creation and never stored.
//   - Only the SHA-256 hex hash (KeyHash) is persisted in the DB.
//   - Scopes are the permission ceiling — a key can never exceed the
//     creating user's own permissions at the time of creation.
type APIKey struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Scopes     []string   // permission ceiling; checked against resolved permissions
	CreatedBy  uuid.UUID  // user who created the key
	ExpiresAt  *time.Time // nil = never expires
	RevokedAt  *time.Time // nil = active
	LastUsedAt *time.Time // async-updated on each validated request
	CreatedAt  time.Time
}

// IsActive reports whether the key is usable: not revoked and not expired.
func (k *APIKey) IsActive() bool {
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

// CreateAPIKeyRequest is the input for creating a new API key.
// TenantID is derived from ctx (current_tenant_id) — do not pass it explicitly.
type CreateAPIKeyRequest struct {
	Name      string
	Scopes    []string   // must be a subset of CreatedBy user's own permissions
	ExpiresAt *time.Time // nil = no expiry
	CreatedBy uuid.UUID  // authenticated user creating the key
}
