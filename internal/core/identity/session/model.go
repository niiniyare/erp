package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"awo/internal/core/authz"
)

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

// Session maps to the user_sessions table row.
// session_token stores sha256hex(raw_token) — the raw token is never persisted.
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

// marshalJSON marshals a value to JSON bytes, returning nil on error.
func marshalJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// ResolvedSession is the runtime view of a validated session.
// It is set in Fiber Locals under LocalsKeySession by the authn middleware
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
//	sess := c.Locals(session.LocalsKeySession).(*session.ResolvedSession)
const LocalsKeySession = "resolved_session"

// Can reports whether the session holds the given permission.
// This is an O(1) map lookup — it never hits the DB or Casbin engine.
func (s *ResolvedSession) Can(permission string) bool {
	if s == nil || s.Permissions == nil {
		return false
	}
	return s.Permissions[permission]
}

// ToPrincipal converts the session into an authz.Principal suitable for
// Casbin-backed management operations (role assignment, policy evaluation).
func (s *ResolvedSession) ToPrincipal() authz.Principal {
	userID := s.UserID.String()
	switch s.UserType {
	case string(authz.ActorPlatform):
		return authz.Principal{
			Subject: authz.PlatformSubject(userID),
			Domain:  authz.DomainPlatform,
		}
	case string(authz.ActorPortal):
		return authz.Principal{
			Subject: authz.PortalSubject(userID),
			Domain:  authz.PortalDomain(s.TenantID.String()),
		}
	default: // "tenant"
		return authz.Principal{
			Subject: authz.TenantSubject(userID),
			Domain:  authz.TenantDomain(s.TenantID.String()),
		}
	}
}

// IsPortal reports whether the session belongs to a portal (external) user.
func (s *ResolvedSession) IsPortal() bool {
	return s.UserType == string(authz.ActorPortal)
}
