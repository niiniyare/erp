package iam

import (
	"github.com/google/uuid"

	"awo.so/framework/def"
)

// SessionViewer implements def.ViewerContext from a baked Session.
// It is populated by AuthMiddleware and injected into every request context.
type SessionViewer struct {
	sess        *Session
	permissions map[string]bool
	roles       map[string]bool
}

func newSessionViewer(sess *Session) *SessionViewer {
	perms := make(map[string]bool, len(sess.Permissions))
	for _, p := range sess.Permissions {
		perms[p] = true
	}
	roles := make(map[string]bool, len(sess.Roles))
	for _, r := range sess.Roles {
		roles[r] = true
	}
	return &SessionViewer{sess: sess, permissions: perms, roles: roles}
}

// ActorID returns the authenticated user UUID as a string.
func (v *SessionViewer) ActorID() string { return v.sess.UserID.String() }

// TenantID returns the tenant UUID as a string (used by pgstore for RLS).
func (v *SessionViewer) TenantID() string { return v.sess.TenantID.String() }

// OrgUnitID returns the viewer's primary org unit.
// uuid.Nil means tenant-wide access (no OU restriction).
func (v *SessionViewer) OrgUnitID() uuid.UUID { return v.sess.OrgUnitID }

// IsSystem returns false for all session-backed viewers.
// System callers use a separate identity path (not yet implemented).
func (v *SessionViewer) IsSystem() bool { return false }

// HasRole reports whether the viewer holds the named role slug.
// Checks the baked snapshot; no DB round-trip.
func (v *SessionViewer) HasRole(role string) bool { return v.roles[role] }

// HasPermission reports whether the baked snapshot includes perm.
// Checks exact match only — wildcard evaluation handled by auth middleware.
// Not part of def.ViewerContext; call from application-level handlers.
func (v *SessionViewer) HasPermission(perm string) bool {
	return v.permissions[perm] || v.permissions["*.*.*"]
}

// Session returns the underlying Session for inspection.
func (v *SessionViewer) Session() *Session { return v.sess }

var _ def.ViewerContext = (*SessionViewer)(nil)
