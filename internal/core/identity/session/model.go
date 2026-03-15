package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/authz"
)

// Session maps to the user_sessions table row.
// session_token stores sha256hex(raw_token) — the raw token is never persisted.
type Session struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TenantID      uuid.UUID
	TokenHash     string          // sha256hex(raw_token) — stored in session_token column
	Permissions   map[string]bool // pre-computed at login, stored as JSONB
	PrincipalID   uuid.UUID       // portal users: the contact/employee UUID they represent
	IsActive      bool
	ExpiresAt     time.Time
	LastSeenAt    time.Time
	IPAddress     string
	UserAgent     string
	RiskScore     int
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
