package contract

import (
	"context"

	"github.com/google/uuid"

	"awo.so/internal/core/iam"
)

// contextKey is an unexported type for the Go context key used by WithContext
// and FromContext. Using a private type prevents key collisions with other
// packages that use context.WithValue.
type contextKey struct{}

// SessionContext is the IAM-provided read-only identity and runtime context
// that non-IAM modules receive. It wraps a validated [iam.ResolvedSession]
// and intentionally exposes only the consumer-relevant fields.
//
// Value semantics: copying is safe (the pointer inside is immutable after
// construction). There are no setters.
//
// What it does NOT expose:
//   - permission maps or Can()/CanDo() (removed — all authz via Casbin)
//   - ToPrincipal() — IAM-internal; needed only for Enforce() calls
//   - raw UserType string — use IsPlatform()/IsPortal() for context checks
type SessionContext struct {
	resolved *iam.ResolvedSession
}

// newSessionContext wraps a validated ResolvedSession into a SessionContext.
// This is the only way to construct a non-zero SessionContext; consumers
// receive it from FromContext or from the AuthService contract.
func newSessionContext(s *iam.ResolvedSession) SessionContext {
	return SessionContext{resolved: s}
}

// IsZero reports whether sc holds no underlying session. A zero SessionContext
// is returned by FromContext when the context does not carry a session.
func (sc SessionContext) IsZero() bool {
	return sc.resolved == nil
}

// ── Identity ─────────────────────────────────────────────────────────────────

// UserID returns the authenticated user's UUID.
// Returns uuid.Nil for a zero SessionContext.
func (sc SessionContext) UserID() uuid.UUID {
	if sc.resolved == nil {
		return uuid.Nil
	}
	return sc.resolved.UserID
}

// TenantID returns the tenant UUID that acts as the RLS boundary.
// Returns uuid.Nil for a zero SessionContext.
func (sc SessionContext) TenantID() uuid.UUID {
	if sc.resolved == nil {
		return uuid.Nil
	}
	return sc.resolved.TenantID
}

// DisplayName returns the user's display name as stored in the session.
func (sc SessionContext) DisplayName() string {
	if sc.resolved == nil {
		return ""
	}
	return sc.resolved.DisplayName
}

// IsPlatform reports whether this session belongs to a platform actor.
// Safe to use for UI context decisions (e.g. "show platform admin menu").
// MUST NOT be used to skip Casbin enforcement.
func (sc SessionContext) IsPlatform() bool {
	if sc.resolved == nil {
		return false
	}
	return sc.resolved.IsPlatform()
}

// IsPortal reports whether this session belongs to an external portal user.
func (sc SessionContext) IsPortal() bool {
	if sc.resolved == nil {
		return false
	}
	return sc.resolved.IsPortal()
}

// ── Entity Scope ─────────────────────────────────────────────────────────────

// EntityScope returns the pre-computed entity visibility scope.
// Service methods use this to add WHERE clauses before querying entity-
// partitioned data (Layer 2 isolation — see docs/reference/modules/iam/09b).
func (sc SessionContext) EntityScope() iam.EntityScope {
	if sc.resolved == nil {
		return iam.EntityScope{}
	}
	return sc.resolved.EntityScope
}

// ── Runtime Metadata (non-security) ─────────────────────────────────────────

// FeatureEnabled reports whether the named feature flag is enabled in
// the session's pre-computed Configuration. Returns false when absent.
//
// This is a SOFT context check — not an authorization decision.
// The result comes from the flag snapshot taken at login time.
func (sc SessionContext) FeatureEnabled(key string) bool {
	if sc.resolved == nil {
		return false
	}
	return sc.resolved.FeatureEnabled(key)
}

// Setting returns the tenant setting value for key, or fallback if absent.
// Settings are pre-computed at login and stored as strings.
func (sc SessionContext) Setting(key, fallback string) string {
	if sc.resolved == nil {
		return fallback
	}
	return sc.resolved.SettingString(key, fallback)
}

// SettingBool parses a tenant setting as a boolean.
// Returns fallback when absent or unparseable.
func (sc SessionContext) SettingBool(key string, fallback bool) bool {
	if sc.resolved == nil {
		return fallback
	}
	return sc.resolved.SettingBool(key, fallback)
}

// SettingInt parses a tenant setting as an integer.
// Returns fallback when absent or unparseable.
func (sc SessionContext) SettingInt(key string, fallback int) int {
	if sc.resolved == nil {
		return fallback
	}
	return sc.resolved.SettingInt(key, fallback)
}

// Preference returns a user preference value for key, or fallback if absent.
// Preferences are per-user (e.g. "ui.theme", "locale") and stored in
// Configuration.Prefs at login time.
func (sc SessionContext) Preference(key, fallback string) string {
	if sc.resolved == nil {
		return fallback
	}
	if sc.resolved.Configuration.Prefs == nil {
		return fallback
	}
	if v, ok := sc.resolved.Configuration.Prefs[key]; ok {
		return v
	}
	return fallback
}

// ── Go Context Helpers ────────────────────────────────────────────────────────

// WithContext stores sc in ctx so that downstream service-layer code can
// retrieve it via FromContext without depending on *fiber.Ctx.
//
// Called by [InjectSessionContext] Fiber middleware — callers outside of
// middleware orchestration should not need this directly.
func WithContext(ctx context.Context, sc SessionContext) context.Context {
	return context.WithValue(ctx, contextKey{}, sc)
}

// FromContext retrieves the SessionContext injected by IAM middleware.
// Returns (zero, false) when ctx does not carry a session.
//
// Usage in any service method:
//
//	sc, ok := contract.FromContext(ctx)
//	if !ok {
//	    return errors.New("unauthenticated")
//	}
//	tenantID := sc.TenantID()
func FromContext(ctx context.Context) (SessionContext, bool) {
	sc, ok := ctx.Value(contextKey{}).(SessionContext)
	if !ok || sc.IsZero() {
		return SessionContext{}, false
	}
	return sc, true
}
