package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Session Config ───────────────────────────────────────────────────────────

// SessionConfig holds session-related settings.
// TODO(settings): read SessionTTL from settings key "iam.session_ttl_hours" per tenant
// FIXME(feature-flags): MFAEnabled should be read from feature-flag "iam.mfa_enabled"
type SessionConfig struct {
	SessionTTL time.Duration // default: 8h
	CookieName string        // default: "session"
}

// DefaultSessionConfig returns safe defaults for development.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		SessionTTL: 8 * time.Hour,
		CookieName: "session",
	}
}

// ─── Entity Scope ─────────────────────────────────────────────────────────────

// EntityScopeType identifies how broadly an entity-scoped session can see data.
type EntityScopeType string

const (
	EntityScopeAll     EntityScopeType = "all"     // platform/sysadmin: all entities
	EntityScopeSubtree EntityScopeType = "subtree" // manager: entity + descendants
	EntityScopeEntity  EntityScopeType = "entity"  // regular user: own entity only
)

// EntityScope is the pre-computed access scope stored in user_sessions.entity_scope.
// {"type":"all"|"subtree"|"entity","entity_id":"uuid","path_prefix":"/uuid/uuid/"}
type EntityScope struct {
	Type       EntityScopeType `json:"type"`
	EntityID   string          `json:"entity_id,omitempty"`
	PathPrefix string          `json:"path_prefix,omitempty"`
}

// ─── Session Configuration ────────────────────────────────────────────────────

// Configuration is the pre-computed session configuration stored in user_sessions.configuration.
// {"flags":{"hr.enabled":true,...},"settings":{"iam.session_ttl_hours":"8"},"prefs":{}}
type Configuration struct {
	Flags    map[string]bool   `json:"flags"`
	Settings map[string]string `json:"settings"`
	Prefs    map[string]string `json:"prefs"`
}

// DefaultConfiguration returns an empty pre-computed configuration.
// Populated at login once flags/settings services are wired in.
func DefaultConfiguration() Configuration {
	return Configuration{
		Flags:    map[string]bool{},
		Settings: map[string]string{},
		Prefs:    map[string]string{},
	}
}

// ─── Session Aggregate ────────────────────────────────────────────────────────

// Session maps to the user_sessions table row.
// TokenHash stores sha256hex(raw_token) — the raw token is never persisted.
type Session struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TenantID      uuid.UUID
	UserType      string          // "INTERNAL" | "SYSADMIN" | "CUSTOMER" | "PORTAL"
	TokenHash     string          // sha256hex(raw_token) — stored in session_token column
	Permissions   map[string]bool // pre-computed at login, stored as JSONB
	PrincipalID   uuid.UUID       // portal users: the contact/employee UUID they represent
	EntityScope   EntityScope     // pre-computed entity access scope
	Configuration Configuration   // pre-computed flags + settings + prefs
	IsActive      bool
	ExpiresAt     time.Time
	LastSeenAt    time.Time
	IPAddress     string
	UserAgent     string
	RiskScore     int
}

// ─── Resolved Session ─────────────────────────────────────────────────────────

// ResolvedSession is the runtime view of a validated session.
// Set in Fiber Locals under LocalsKeySession by the authn middleware
// and consumed by handlers and authorization middleware.
type ResolvedSession struct {
	UserID      uuid.UUID
	UserType    string          // "platform" | "tenant" | "portal"
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	DisplayName string
	Permissions map[string]bool
}

// LocalsKeySession is the Fiber Locals key for the authenticated ResolvedSession.
//
//	sess := c.Locals(session.LocalsKeySession).(*domain.ResolvedSession)
const LocalsKeySession = "resolved_session"

// Can reports whether the session holds the given permission.
// This is an O(1) map lookup — it never hits the DB or Casbin engine.
func (s *ResolvedSession) Can(permission string) bool {
	if s == nil || s.Permissions == nil {
		return false
	}
	return s.Permissions[permission]
}

// ToPrincipal converts the session into a Principal suitable for
// Casbin-backed management operations (role assignment, policy evaluation).
func (s *ResolvedSession) ToPrincipal() Principal {
	userID := s.UserID.String()
	switch s.UserType {
	case string(ActorPlatform):
		return Principal{
			Subject: PlatformSubject(userID),
			Domain:  DomainPlatform,
		}
	case string(ActorPortal):
		return Principal{
			Subject: PortalSubject(userID),
			Domain:  PortalDomain(s.TenantID.String()),
		}
	default: // "tenant"
		return Principal{
			Subject: TenantSubject(userID),
			Domain:  TenantDomain(s.TenantID.String()),
		}
	}
}

// IsPortal reports whether the session belongs to a portal (external) user.
func (s *ResolvedSession) IsPortal() bool {
	return s.UserType == string(ActorPortal)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// MarshalSessionJSON marshals v to JSON bytes, returning "{}" on error.
func MarshalSessionJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
