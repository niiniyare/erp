package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthProvider identifies the OAuth/OIDC identity provider.
type OAuthProvider string

const (
	OAuthProviderGoogle    OAuthProvider = "google"
	OAuthProviderMicrosoft OAuthProvider = "microsoft"
)

// SSOProvider is the per-tenant OAuth/OIDC provider configuration stored in DB.
// ClientSecretEnc is AES-256-GCM encrypted; the plaintext secret is never serialised.
type SSOProvider struct {
	ID               uuid.UUID         `json:"id"`
	TenantID         uuid.UUID         `json:"tenant_id"`
	Provider         OAuthProvider     `json:"provider"`
	ClientID         string            `json:"client_id"`
	ClientSecretEnc  string            `json:"-"` // encrypted; never exposed in JSON
	Scopes           []string          `json:"scopes"`
	RedirectURI      string            `json:"redirect_uri"`
	ExtraParams      map[string]string `json:"extra_params,omitempty"`
	AutoProvision    bool              `json:"auto_provision"`
	// DefaultEntityID is the entity assigned to JIT-provisioned SSO users.
	// uuid.Nil means no default is configured and JIT provisioning will fail.
	DefaultEntityID  uuid.UUID         `json:"default_entity_id,omitempty"`
	IsActive         bool              `json:"is_active"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// SSOUserInfo holds the normalised user attributes returned by the identity provider
// after a successful OAuth exchange.
type SSOUserInfo struct {
	Sub      string        // provider-specific user ID (stable identifier)
	Email    string
	Name     string
	Provider OAuthProvider
}

// CreateSSOProviderRequest is the input for registering or updating an SSO provider.
// ClientSecret is the plaintext secret; the service layer encrypts it before storage.
// The tenant is derived from the request context (cache.TenantIDKey) — no TenantID field needed.
type CreateSSOProviderRequest struct {
	Provider        OAuthProvider
	ClientID        string
	ClientSecret    string // plaintext; encrypted before storage
	Scopes          []string
	RedirectURI     string
	ExtraParams     map[string]string
	AutoProvision   bool
	// DefaultEntityID is the entity assigned to JIT-provisioned SSO users.
	// Must be non-nil when AutoProvision is true.
	DefaultEntityID uuid.UUID
}
