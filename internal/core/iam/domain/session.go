package domain

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Session Configuration
// =============================================================================

// SessionConfig holds process-level session defaults.
//
// These values are used as fallbacks when no tenant-level override is present.
// At login time the session service should attempt to read per-tenant overrides
// from the settings store and embed the resolved values in Configuration.Settings
// (key "iam.session_ttl_hours") before falling back to these defaults.
//
// SessionTTL is the process-level default. The session service overrides this
// at login time by reading the tenant setting "iam.session_ttl_hours" from
// the pre-computed Configuration.Settings map. If the setting is absent the
// default is used.
type SessionConfig struct {
	SessionTTL time.Duration // default: 8h; overridden per-tenant via "iam.session_ttl_hours"
	CookieName string        // default: "session"
}

// DefaultSessionConfig returns safe defaults for development and testing.
// Production deployments should override SessionTTL from tenant settings.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		SessionTTL: 8 * time.Hour,
		CookieName: "session",
	}
}

// =============================================================================
// Entity Scope — Value Object
// =============================================================================

// EntityScopeType controls how broadly a session can see data across the
// entity hierarchy within a single tenant.
//
// Entity scope is an APPLICATION-LAYER concern (Layer 2 — see actor.go).
// PostgreSQL RLS already enforces the tenant boundary invisibly; EntityScope is
// enforced by service methods that add WHERE clauses or ltree path predicates
// based on the session's EntityScope before issuing any query.
//
// Scope breadth intentionally increases from bottom to top:
//
//	EntityScopeEntity  < EntityScopeSubtree < EntityScopeAll
type EntityScopeType string

const (
	// EntityScopeAll grants visibility across every entity in the tenant.
	// Reserved for platform administrators and tenant super-admins.
	EntityScopeAll EntityScopeType = "all"

	// EntityScopeSubtree grants visibility over the user's home entity and
	// all of its descendants in the entity ltree hierarchy.
	// Typical for branch managers and regional supervisors.
	EntityScopeSubtree EntityScopeType = "subtree"

	// EntityScopeEntity restricts visibility to the user's own entity only.
	// Default for regular employees and portal users.
	EntityScopeEntity EntityScopeType = "entity"
)

// EntityScope is the pre-computed entity access scope stored in
// user_sessions.entity_scope as JSONB.
//
// The session service resolves the user's entity membership and highest-
// privilege role at login time and produces this value once.  It is embedded
// in ResolvedSession so that every handler and repository can apply entity
// filtering without additional DB round-trips.
//
// Application of entity scope in repositories

//
//	switch sess.EntityScope.Type {
//	case domain.EntityScopeAll:
//	    // no extra WHERE clause; RLS already enforces tenant
//	case domain.EntityScopeSubtree:
//	    q = q.Where("entity_path <@ ?", sess.EntityScope.PathPrefix)
//	case domain.EntityScopeEntity:
//	    q = q.Where("entity_id = ?", sess.EntityScope.EntityID)
//	}
//
// JSON shape: {"type":"subtree","entity_id":"<uuid>","path_prefix":"/uuid/uuid/"}
type EntityScope struct {
	// Type is the breadth of entity visibility for this session.
	Type EntityScopeType `json:"type"`

	// EntityID is the user's home entity UUID.
	// Present for EntityScopeEntity and EntityScopeSubtree; empty for EntityScopeAll.
	EntityID string `json:"entity_id,omitempty"`

	// PathPrefix is the ltree path of the home entity, used to construct
	// "entity_path <@ $prefix" subtree queries.
	// Present for EntityScopeSubtree; empty otherwise.
	PathPrefix string `json:"path_prefix,omitempty"`
}

// =============================================================================
// Configuration Snapshot — Value Object
// =============================================================================

// Configuration is the pre-computed per-session snapshot of feature flags,
// tenant settings, and user preferences stored in user_sessions.configuration
// as JSONB.
//
// Why snapshot at login?

//
//	Resolving flags and settings once at login and embedding the result in the
//	session row avoids per-request lookups against the settings / feature-flag
//	stores.  Handlers simply read sess.Configuration.Flags["hr.payroll_v2"].
//
// Staleness trade-off

//
//	Long-lived sessions may see stale configuration if a flag or setting
//	changes after login.  Mitigate with short session TTLs for
//	flag-sensitive features, or by adding a "re-resolve configuration"
//	operation that refreshes the session row when a critical flag changes.
//
// JSON shape:
//
//	{
//	  "flags":    {"hr.payroll_v2.enabled": true},
//	  "settings": {"iam.session_ttl_hours": "8"},
//	  "prefs":    {"ui.theme": "dark"}
//	}
type Configuration struct {
	// Flags holds resolved feature-flag values keyed by flag name.
	// Consumers: if !sess.Configuration.Flags["feature.name"] { return ErrFeatureDisabled }
	Flags map[string]bool `json:"flags"`

	// Settings holds resolved tenant-level settings keyed by setting name.
	// Consumers: ttl := sess.Configuration.Settings["iam.session_ttl_hours"]
	Settings map[string]string `json:"settings"`

	// Prefs holds user-level UI preferences keyed by preference name.
	// Consumers: theme := sess.Configuration.Prefs["ui.theme"]
	Prefs map[string]string `json:"prefs"`
}

// DefaultConfiguration returns empty maps safe for tests and for the
// development login path before flag / settings services are wired in.
func DefaultConfiguration() Configuration {
	return Configuration{
		Flags:    map[string]bool{},
		Settings: map[string]string{},
		Prefs:    map[string]string{},
	}
}

// =============================================================================
// Session — Aggregate Root
// =============================================================================

// Session is the aggregate root for an authenticated user session.
// It maps 1:1 to a row in the user_sessions table.
//
// Security invariants

//
//   - TokenHash stores sha256hex(raw_token).  The raw token is returned to the
//     client exactly once at login and is NEVER persisted anywhere.  The hash
//     is stored in the DB column named session_token (the column name omits
//     "_hash" for brevity; always treat session_token values as hashes).
//
//   - Authorization decisions are made by the Casbin enforcer at request time,
//     NOT from a permission snapshot in the session.  The session carries
//     identity and context only.  This ensures role revocations take effect
//     on the next request without requiring session invalidation.
//
// TenantID vs EntityScope

//
//	TenantID is the RLS boundary (Layer 1).  Every DB transaction that uses
//	this session must execute:
//	  SET LOCAL app.tenant_id = '<TenantID>'
//	PostgreSQL's RLS policies then silently filter all tables to that tenant.
//
//	EntityScope is the application-layer boundary (Layer 2).  It encodes how
//	broadly within the tenant this session can see entity-partitioned data.
//	Service methods apply EntityScope filtering in Go before issuing queries.
//
// PrincipalID

//
//	Non-nil for ActorPortal sessions only.  It identifies the external contact
//	or party record (e.g. a supplier, customer, or employee on a self-service
//	portal) that this user is operating on behalf of.  Nil for all internal and
//	platform sessions.
type Session struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TenantID      uuid.UUID     // RLS key — must be set via app.tenant_id in every TX
	UserType      string        // persisted enum: "INTERNAL"|"SYSADMIN"|"CUSTOMER"|"PORTAL"|"API"
	TokenHash     string        // sha256hex(raw_token); DB column: session_token
	PrincipalID   *uuid.UUID    // non-nil for portal sessions; identifies the represented party
	EntityScope   EntityScope   // application-layer entity visibility scope
	Configuration Configuration // pre-computed flags, tenant settings, and user preferences
	IsActive      bool
	ExpiresAt     time.Time
	LastSeenAt    time.Time
	IPAddress     string
	UserAgent     string
}

// ActorType translates the persisted UserType into the canonical ActorType
// used by the authorization layer.  Delegates to ActorTypeFromUserType so
// this struct never contains raw string comparison logic.
func (s *Session) ActorType() ActorType {
	return ActorTypeFromUserType(s.UserType)
}

// IsExpired reports whether the session has passed its expiry time.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid reports whether the session is both active and not yet expired.
// This is the primary liveness check; always prefer this over checking
// IsActive and IsExpired separately.
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// =============================================================================
// ResolvedSession — Runtime Value Object
// =============================================================================

// ResolvedSession is the lightweight, request-scoped view of a validated
// session.  It is constructed by the authn middleware, stored in Fiber Locals
// under LocalsKeySession, and consumed by handlers and the authz middleware.
//
// Design goal: a handler or service method should be able to answer all three
// questions below by reading this struct alone — with no extra DB or cache
// round-trips:
//
//  1. "Can this user perform this action?"  → Can() / CanDo()
//  2. "Which entities can this user see?"   → EntityScope
//  3. "Is this feature enabled?"            → Configuration.Flags["name"]
//
// TenantID vs EntityScope (application layer)

//
//	TenantID is the RLS key.  The authn middleware (or a DB middleware it
//	delegates to) must execute "SET LOCAL app.tenant_id = '<TenantID>'" at
//	the start of every transaction so PostgreSQL RLS silently enforces the
//	tenant boundary.
//
//	EntityScope is the application-layer entity restriction.  Service methods
//	inspect EntityScope.Type and build WHERE clauses accordingly.  This
//	boundary is intentionally softer than RLS: an elevated-privilege
//	operation (e.g. a cross-entity consolidation report) can bypass entity
//	scope with an explicit override, but can never bypass RLS.
//
// PrincipalID

//
//	Mirrors Session.PrincipalID — non-nil for portal users only.
type ResolvedSession struct {
	UserID        uuid.UUID       `json:"user_id"`
	UserType      string          `json:"user_type"` // persisted enum; always use ActorTypeFromUserType() for authz logic
	TenantID      uuid.UUID       `json:"tenant_id"` // RLS key — set as app.tenant_id in every DB transaction
	PrincipalID   *uuid.UUID      `json:"principal_id,omitempty"` // non-nil for portal users; identifies the represented party
	DisplayName   string          `json:"display_name"`
	Permissions   map[string]bool `json:"permissions"` // pre-computed at login; O(1) permission checks via Can() / CanDo()
	EntityScope   EntityScope     `json:"entity_scope"`     // application-layer entity visibility; enforced in service methods
	Configuration Configuration   `json:"configuration"`    // feature flags, tenant settings, and user preferences
}

// LocalsKeySession is the Fiber Locals key for the authenticated ResolvedSession.
//
// Usage in a handler:
//
//	sess, ok := c.Locals(domain.LocalsKeySession).(*domain.ResolvedSession)
//	if !ok || sess == nil {
//	    return fiber.ErrUnauthorized
//	}
const LocalsKeySession = "resolved_session"

// Can reports whether the session holds the given permission key.
//
// The key format is "resource.action" — e.g. "invoice.read" or
// "hr.employee.create".  This is an O(1) map lookup; it never touches the DB
// or the Casbin engine.  Permission maps are computed once at login and
// embedded in the session row.
//
// Returns false for nil receivers and nil permission maps, so callers do not
// need a nil guard before calling Can.
func (s *ResolvedSession) Can(permission string) bool {
	if s == nil || s.Permissions == nil {
		return false
	}
	// "*" is a superuser sentinel set by buildPermissions when the session
	// holds a wildcard allow policy (e.g. tenant_admin with Object="*").
	if s.Permissions["*"] {
		return true
	}
	return s.Permissions[permission]
}

// CanDo is a convenience wrapper over Can for callers that hold resource and
// action as separate strings.
//
//	if !sess.CanDo("invoice", "approve") {
//	    return fiber.ErrForbidden
//	}
func (s *ResolvedSession) CanDo(resource, action string) bool {
	return s.Can(resource + "." + action)
}

// FeatureEnabled reports whether the named feature flag is enabled in this
// session's pre-computed Configuration.  Returns false for nil receivers,
// unknown flags, and flags explicitly set to false.
func (s *ResolvedSession) FeatureEnabled(flag string) bool {
	if s == nil || s.Configuration.Flags == nil {
		return false
	}
	return s.Configuration.Flags[flag]
}

// SettingString returns the tenant setting value for key, or def if the key is
// absent.  Settings are stored as strings; use SettingBool / SettingInt /
// SettingDecimal for typed access.
func (s *ResolvedSession) SettingString(key, def string) string {
	if s == nil || s.Configuration.Settings == nil {
		return def
	}
	if v, ok := s.Configuration.Settings[key]; ok {
		return v
	}
	return def
}

// SettingBool parses the setting value for key as a boolean (via strconv.ParseBool).
// Returns def on absence or parse error.
func (s *ResolvedSession) SettingBool(key string, def bool) bool {
	v := s.SettingString(key, "")
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// SettingInt parses the setting value for key as a base-10 integer.
// Returns def on absence or parse error.
func (s *ResolvedSession) SettingInt(key string, def int) int {
	v := s.SettingString(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// SettingDecimal parses the setting value for key as a 64-bit float.
// Returns def on absence or parse error.
func (s *ResolvedSession) SettingDecimal(key string, def float64) float64 {
	v := s.SettingString(key, "")
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

// ToPrincipal converts the session into a Principal value object suitable for
// Casbin-backed management operations (role assignment, policy evaluation, etc.).
//
// Actor type translation is delegated entirely to ActorTypeFromUserType — the
// single canonical mapping point — so this method never branches on raw
// UserType strings.
//
// NOTE: ActorAPI sessions are not yet fully supported in the login flow.
// API key authentication will need a dedicated path that populates a ClientID
// (distinct from UserID) in the session.  Until then, API sessions fall back
// to producing a Principal with APISubject(userID), which is functionally
// correct for policy lookups but semantically imprecise.
func (s *ResolvedSession) ToPrincipal() Principal {
	userID := s.UserID.String()
	tenantID := s.TenantID.String()

	switch ActorTypeFromUserType(s.UserType) {
	case ActorPlatform:
		return Principal{
			Subject: PlatformSubject(userID),
			Domain:  DomainPlatform,
		}
	case ActorPortal:
		return Principal{
			Subject: PortalSubject(userID),
			Domain:  PortalDomain(tenantID),
		}
	case ActorAPI:
		// TODO: replace userID with clientID once API key auth is implemented.
		return Principal{
			Subject: APISubject(userID),
			Domain:  APIDomain(tenantID),
		}
	default: // ActorTenant — "INTERNAL", "EMPLOYEE", and any unmapped types
		return Principal{
			Subject: TenantSubject(userID),
			Domain:  TenantDomain(tenantID),
		}
	}
}

// IsPortal reports whether the session belongs to a portal (external) user.
func (s *ResolvedSession) IsPortal() bool {
	return ActorTypeFromUserType(s.UserType) == ActorPortal
}

// IsPlatform reports whether the session belongs to a platform administrator.
func (s *ResolvedSession) IsPlatform() bool {
	return ActorTypeFromUserType(s.UserType) == ActorPlatform
}

// =============================================================================
// Helpers
// =============================================================================

// MarshalSessionJSON marshals v to JSON bytes, returning the safe fallback
// "{}" if marshalling fails.  Used when writing EntityScope, Configuration,
// or Permissions to JSONB columns in the session store.
func MarshalSessionJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
