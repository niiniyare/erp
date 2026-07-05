// Package auth defines the session management interface used by API middleware.
// Concrete implementations (Redis-backed, in-memory for tests) live in
// internal packages and are injected at bootstrap time.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// ErrSessionExpired is returned when the session token exists but has expired.
var ErrSessionExpired = errors.New("session expired")

// ErrSessionNotFound is returned when the session token is not found.
var ErrSessionNotFound = errors.New("session not found")

// Session is the validated session record retrieved from the session store.
type Session struct {
	// ID is the opaque session token (e.g. UUID or signed JWT ID).
	ID string

	// UserID is the authenticated user's UUID.
	UserID uuid.UUID

	// TenantID is the tenant this session is scoped to.
	TenantID uuid.UUID

	// Roles are the role names granted to the user at session creation time.
	// Refreshed on role change via cache invalidation.
	Roles []string

	// ExpiresAt is the session expiry time in UTC.
	ExpiresAt time.Time

	// IsSystem marks sessions created by machine tokens or background workers.
	IsSystem bool
}

// IsExpired reports whether the session has passed its expiry time.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// SessionManager validates session tokens and resolves them to ViewerContext.
// The framework API middleware calls Validate on every authenticated request.
//
// Implementations must be safe for concurrent use.
type SessionManager interface {
	// Validate looks up the session token and returns the Session.
	// Returns ErrSessionNotFound when the token is unknown.
	// Returns ErrSessionExpired when the token is found but past ExpiresAt.
	Validate(ctx context.Context, token string) (*Session, error)

	// ViewerFor constructs a ViewerContext from a validated Session.
	// This is a separate step so the middleware can cache the viewer
	// in c.Locals without re-querying on every sub-request.
	ViewerFor(ctx context.Context, s *Session) (def.ViewerContext, error)

	// Invalidate removes the session from the backing store (logout).
	Invalidate(ctx context.Context, token string) error
}

// SessionViewer is a minimal ViewerContext backed by a Session.
// Use in tests or when a full IAM integration is not yet wired.
type SessionViewer struct {
	session *Session
}

// NewSessionViewer wraps s in a ViewerContext.
func NewSessionViewer(s *Session) *SessionViewer {
	return &SessionViewer{session: s}
}

func (v *SessionViewer) ActorID() string      { return v.session.UserID.String() }
func (v *SessionViewer) TenantID() string     { return v.session.TenantID.String() }
func (v *SessionViewer) OrgUnitID() uuid.UUID { return uuid.Nil }
func (v *SessionViewer) IsSystem() bool       { return v.session.IsSystem }
func (v *SessionViewer) HasRole(role string) bool {
	for _, r := range v.session.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// compile-time check
var _ def.ViewerContext = (*SessionViewer)(nil)
