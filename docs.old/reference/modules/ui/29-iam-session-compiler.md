[<-- Back to Index](README.md)

> **🚫 SUPERSEDED.** This document describes the design specification for IAM integration. The
> implemented system differs in key ways: `SessionContext` is `contract.SessionContext` (not
> `iam/session/session.go`), `UIContext` does not exist (use `UISessionContext`), and the
> `PolicyEngine`/`FeatureEngine` interfaces are internal to the pipeline stages.
>
> The directory structure shown here does not match `internal/web/`.
>
> **Read instead:**
> - [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md) — `AuthzStage` and `SessionStage`
> - [IAM Integration](../02-architecture/03-iam-integration.md) — correct type hierarchy
> - [Glossary](../appendices/A-glossary.md) — `contract.SessionContext` vs `UISessionContext`
>
> This file is kept as an **Architecture Decision Record** showing the original design intent.

# IAM/Session-Aware UI Compiler — Production Architecture

Extends [§28 Go-First UI Compiler](./28-go-ui-compiler.md).

The UI is not just a function of tenant and user.

```
UI = f(Session, IAM State, Tenant Context, Feature Flags, Preferences)
```

Where:
- **Session** defines WHO the user is and whether their identity is valid RIGHT NOW
- **IAM** defines WHAT they are permitted to do, via which policy
- **Tenant** defines WHERE they operate (isolated namespace)
- **Feature Flags** define WHAT capabilities are enabled for this session
- **Preferences** define HOW the UI is personalized

The schema handler may not execute without a valid session. There is no guest UI, no anonymous path, no default rendering. Session is the only key.

---

## What Changes From §28

§28 introduced `UIContext` — a plain struct carrying resolved state. This document replaces it with:

1. `SessionContext` — the authoritative session object, the only input to the compiler
2. `PolicyEngine` — a typed interface for permission evaluation (replaces raw map access)
3. `FeatureEngine` — a typed interface for flag evaluation (session-scoped, not global)
4. `SchemaValidator` — validates compiled schemas before they leave the server
5. `UIBlock` system — reusable session-aware schema fragments for ERP-scale composition
6. Full IAM pipeline with explicit layers
7. Caching strategy for IAM-at-scale
8. Security threat model

---

## Full Directory Structure

```
internal/
├── iam/
│   ├── session/
│   │   ├── session.go       — SessionContext struct (the single entry point)
│   │   ├── store.go         — SessionStore interface (Redis impl)
│   │   └── builder.go       — SessionBuilder: constructs SessionContext from auth data
│   ├── policy/
│   │   ├── engine.go        — PolicyEngine interface + CasbinEngine impl
│   │   ├── permissions.go   — Permission key constants ("invoice.approve" etc.)
│   │   └── resolver.go      — PermissionResolver: bulk-resolves UI permissions
│   └── flags/
│       ├── engine.go        — FeatureEngine interface
│       └── resolver.go      — FlagResolver: session-scoped flag evaluation
├── web/
│   ├── compiler/
│   │   ├── compiler.go      — Compile(path, session) + CompileApp(session)
│   │   ├── validator.go     — SchemaValidator: structural + IAM rule enforcement
│   │   └── cache.go         — SchemaCache: fingerprint-keyed compiled schema cache
│   ├── blocks/
│   │   ├── blocks.go        — UIBlock type + block registry
│   │   ├── finance.go       — Finance-domain blocks
│   │   ├── approvals.go     — Approval inbox block
│   │   └── inventory.go     — Inventory summary block
│   ├── registry/
│   │   ├── pages.go         — page path → PageFn
│   │   └── nav.go           — nav contributors
│   ├── amis/                — builder package (unchanged from §28)
│   └── pages/               — page schema functions
└── middleware/
    ├── auth.go              — AuthMiddleware: session extraction + validation
    ├── iam.go               — IAMMiddleware: permission + flag resolution
    └── session.go           — SessionFromFiber() + SetSession()
```

---

## Layer 1 — SessionContext

The single entry point for all UI compilation. Every page function, nav function, and block receives this and nothing else.

```go
// internal/iam/session/session.go
package session

import "time"

// SessionContext is the authoritative identity + authorization state
// for a single authenticated request.
//
// It is constructed once per request by AuthMiddleware and IAMMiddleware.
// It is immutable after construction — no field may be written after build.
// It is passed by value into all UI functions to enforce immutability.
//
// RULE: No UI function may execute without a valid SessionContext.
// RULE: No UI function may mutate SessionContext.
// RULE: No UI function may access SessionContext.Permissions directly —
//       use the PolicyEngine interface.
type SessionContext struct {
    // Identity
    SessionID string
    UserID    string
    TenantID  string

    // Authentication state
    Authenticated bool
    AuthMethod    string    // "password", "sso", "api_key"
    IssuedAt      time.Time
    ExpiresAt     time.Time

    // Authorization (populated by IAMMiddleware via PolicyEngine)
    // INTERNAL — do not access directly. Use policy.Can(session, perm).
    // Exposed for serialization and caching only.
    permissions map[string]bool

    // Identity attributes
    Roles  []string
    Scopes []string // OAuth-style scopes for API access

    // Feature flags resolved for this tenant+user combination
    // INTERNAL — do not access directly. Use flags.Enabled(session, flag).
    featureFlags map[string]bool

    // User preferences
    Preferences map[string]any

    // Tenant configuration
    Tenant TenantConfig

    // Request metadata (read-only after construction)
    Metadata SessionMetadata
}

type TenantConfig struct {
    ID       string
    Name     string
    Currency string
    Timezone string
    LogoURL  string
    Plan     string // "starter", "professional", "enterprise"
}

type SessionMetadata struct {
    IPAddress string
    UserAgent string
    DeviceID  string
    LoginAt   time.Time
    RequestID string // trace correlation
}

// IsExpired reports whether the session has passed its expiry time.
func (s SessionContext) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}

// HasRole reports whether this session includes the given role.
func (s SessionContext) HasRole(role string) bool {
    for _, r := range s.Roles {
        if r == role {
            return true
        }
    }
    return false
}

// HasScope reports whether this session includes an OAuth-style scope.
func (s SessionContext) HasScope(scope string) bool {
    for _, sc := range s.Scopes {
        if sc == scope {
            return true
        }
    }
    return false
}

// resolvedPermissions returns the internal permission map.
// Only used by PolicyEngine implementations — not for direct use.
func (s SessionContext) resolvedPermissions() map[string]bool {
    return s.permissions
}

// resolvedFlags returns the internal feature flag map.
// Only used by FeatureEngine implementations — not for direct use.
func (s SessionContext) resolvedFlags() map[string]bool {
    return s.featureFlags
}
```

```go
// internal/iam/session/builder.go
package session

import "time"

// Builder constructs a SessionContext in a controlled sequence.
// The compiler layer MUST use this — never construct SessionContext literals.
type Builder struct {
    s SessionContext
}

func NewBuilder(sessionID, userID, tenantID string) *Builder {
    return &Builder{s: SessionContext{
        SessionID:     sessionID,
        UserID:        userID,
        TenantID:      tenantID,
        Authenticated: false,
        IssuedAt:      time.Now(),
    }}
}

func (b *Builder) Authenticated(method string, expiresAt time.Time) *Builder {
    b.s.Authenticated = true
    b.s.AuthMethod    = method
    b.s.ExpiresAt     = expiresAt
    return b
}

func (b *Builder) WithRoles(roles []string) *Builder {
    b.s.Roles = roles
    return b
}

func (b *Builder) WithScopes(scopes []string) *Builder {
    b.s.Scopes = scopes
    return b
}

func (b *Builder) WithPermissions(perms map[string]bool) *Builder {
    b.s.permissions = perms
    return b
}

func (b *Builder) WithFeatureFlags(flags map[string]bool) *Builder {
    b.s.featureFlags = flags
    return b
}

func (b *Builder) WithPreferences(prefs map[string]any) *Builder {
    b.s.Preferences = prefs
    return b
}

func (b *Builder) WithTenant(cfg TenantConfig) *Builder {
    b.s.Tenant = cfg
    return b
}

func (b *Builder) WithMetadata(meta SessionMetadata) *Builder {
    b.s.Metadata = meta
    return b
}

// Build finalizes and returns the immutable SessionContext.
// Returns an error if the session is not authenticated.
func (b *Builder) Build() (SessionContext, error) {
    if !b.s.Authenticated {
        return SessionContext{}, ErrUnauthenticated
    }
    if b.s.UserID == "" || b.s.TenantID == "" {
        return SessionContext{}, ErrMissingIdentity
    }
    if b.s.IsExpired() {
        return SessionContext{}, ErrSessionExpired
    }
    return b.s, nil
}

var (
    ErrUnauthenticated = sessionErr("session is not authenticated")
    ErrMissingIdentity = sessionErr("userID and tenantID are required")
    ErrSessionExpired  = sessionErr("session has expired")
)

type sessionErr string
func (e sessionErr) Error() string { return string(e) }
```

---

## Layer 2 — Policy Engine

Page functions NEVER access `session.permissions` directly. They always call `policy.Can(session, perm)`. This is a hard architectural rule.

```go
// internal/iam/policy/engine.go
package policy

import (
    "context"
    "awo.so/internal/iam/session"
)

// PolicyEngine is the single interface for permission evaluation.
// All UI functions use this — never session.permissions directly.
//
// Rationale: Direct map access cannot be audited, cannot add ABAC conditions,
// cannot be mocked cleanly, and does not enforce the "policy is the authority"
// principle. The engine is the authority.
type PolicyEngine interface {
    // Can returns true if the session is permitted to perform the action.
    // permission format: "resource.action" (e.g. "invoice.approve")
    Can(ctx context.Context, s session.SessionContext, permission string) bool

    // CanAll returns true only if ALL permissions are granted.
    CanAll(ctx context.Context, s session.SessionContext, permissions ...string) bool

    // CanAny returns true if ANY permission is granted.
    CanAny(ctx context.Context, s session.SessionContext, permissions ...string) bool

    // Resolve returns all granted permissions for a session.
    // Used by IAMMiddleware to pre-populate the session.
    Resolve(ctx context.Context, tenantID, userID string, roles []string) (map[string]bool, error)
}

// CasbinEngine implements PolicyEngine using Casbin.
// The fast path reads the pre-resolved permission map.
// The slow path re-evaluates via Casbin for ABAC rules not encodable in the map.
type CasbinEngine struct {
    enforcer  CasbinEnforcer // interface to casbin.Enforcer
    audit     AuditLogger
}

// Can checks the pre-resolved permission map first (O(1)).
// Falls back to Casbin evaluation for ABAC conditions.
func (e *CasbinEngine) Can(ctx context.Context, s session.SessionContext, permission string) bool {
    // Fast path: pre-resolved permissions from IAMMiddleware
    perms := s.ResolvedPermissions()
    if granted, exists := perms[permission]; exists {
        e.audit.Log(ctx, s, permission, granted)
        return granted
    }

    // Slow path: evaluate via Casbin (ABAC conditions, time-based rules)
    // This executes when a permission was not pre-resolved (rare).
    granted, err := e.enforcer.Enforce(s.TenantID, s.UserID, permission)
    if err != nil {
        // Fail closed on policy engine error — deny access
        e.audit.LogError(ctx, s, permission, err)
        return false
    }

    e.audit.Log(ctx, s, permission, granted)
    return granted
}

func (e *CasbinEngine) CanAll(ctx context.Context, s session.SessionContext, permissions ...string) bool {
    for _, p := range permissions {
        if !e.Can(ctx, s, p) {
            return false
        }
    }
    return true
}

func (e *CasbinEngine) CanAny(ctx context.Context, s session.SessionContext, permissions ...string) bool {
    for _, p := range permissions {
        if e.Can(ctx, s, p) {
            return true
        }
    }
    return false
}

func (e *CasbinEngine) Resolve(ctx context.Context, tenantID, userID string, roles []string) (map[string]bool, error) {
    // Bulk-resolve all UI-relevant permissions for this user.
    // Called once per request by IAMMiddleware, result stored in session.
    result := make(map[string]bool, len(UIPermissions))
    for _, perm := range UIPermissions {
        granted, err := e.enforcer.Enforce(tenantID, userID, perm)
        if err != nil {
            return nil, err
        }
        result[perm] = granted
    }
    return result, nil
}
```

```go
// internal/iam/policy/permissions.go
package policy

// UIPermissions is the authoritative list of all permissions the UI checks.
// Add here when introducing a new resource or action.
// Format: "resource.action"
//
// This list drives both:
// 1. IAMMiddleware bulk-resolution (one Casbin call per entry)
// 2. Schema validator checks (catches permission keys not in this list)
var UIPermissions = []string{
    // Accounts
    "account.read",
    "account.create",
    "account.update",
    "account.delete",

    // Invoices
    "invoice.read",
    "invoice.create",
    "invoice.update",
    "invoice.delete",
    "invoice.approve",
    "invoice.void",
    "invoice.export",

    // Journal Entries
    "journal.read",
    "journal.create",
    "journal.post",
    "journal.reverse",

    // Purchase Orders
    "purchase_order.read",
    "purchase_order.create",
    "purchase_order.update",
    "purchase_order.approve",
    "purchase_order.cancel",

    // Payroll
    "payroll.read",
    "payroll.run",
    "payroll.approve",

    // Inventory
    "inventory.read",
    "inventory.adjust",

    // Administration
    "user.read",
    "user.create",
    "user.update",
    "user.deactivate",
    "tenant.read",
    "tenant.update",
    "role.read",
    "role.assign",

    // Reporting
    "report.financial",
    "report.payroll",
    "report.inventory",
    "report.audit",
}

// Perm is a type-safe permission constant helper.
// Usage: policy.Perm("invoice", "approve") → "invoice.approve"
func Perm(resource, action string) string {
    return resource + "." + action
}
```

---

## Layer 3 — Feature Engine

```go
// internal/iam/flags/engine.go
package flags

import (
    "context"
    "awo.so/internal/iam/session"
)

// FeatureEngine evaluates feature flags in session scope.
// Flags are NOT global — they are evaluated per tenant+user combination.
//
// A flag can be:
// - Enabled globally for a tenant (plan-based: "payroll" requires Enterprise)
// - Enabled for specific users (beta program)
// - Enabled by percentage rollout (hash(userID) < threshold)
// - Enabled by time window (gradual rollout)
type FeatureEngine interface {
    // Enabled returns true if the flag is active for this session.
    Enabled(ctx context.Context, s session.SessionContext, flag string) bool

    // EnabledAll returns true if ALL flags are active.
    EnabledAll(ctx context.Context, s session.SessionContext, flags ...string) bool

    // Resolve bulk-resolves all flags for a session.
    // Called once per request by IAMMiddleware.
    Resolve(ctx context.Context, tenantID, userID string) (map[string]bool, error)
}

// RedisFeatureEngine resolves flags from Redis with tenant+user override support.
type RedisFeatureEngine struct {
    store FlagStore
}

func (e *RedisFeatureEngine) Enabled(ctx context.Context, s session.SessionContext, flag string) bool {
    // Fast path: use pre-resolved flags stored in session
    flags := s.ResolvedFlags()
    return flags[flag]
}

func (e *RedisFeatureEngine) EnabledAll(ctx context.Context, s session.SessionContext, flags ...string) bool {
    for _, f := range flags {
        if !e.Enabled(ctx, s, f) {
            return false
        }
    }
    return true
}

func (e *RedisFeatureEngine) Resolve(ctx context.Context, tenantID, userID string) (map[string]bool, error) {
    // 1. Load tenant-level flags (plan-based, admin-configured)
    tenantFlags, err := e.store.TenantFlags(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // 2. Load user-level overrides (beta program, individual grants)
    userOverrides, err := e.store.UserFlagOverrides(ctx, tenantID, userID)
    if err != nil {
        return nil, err
    }

    // User overrides win over tenant defaults
    result := make(map[string]bool, len(tenantFlags))
    for k, v := range tenantFlags   { result[k] = v }
    for k, v := range userOverrides { result[k] = v }

    return result, nil
}
```

```go
// internal/iam/flags/flags.go — canonical flag names
package flags

const (
    // Module flags — controlled by subscription plan
    FlagGLModule        = "gl.module"
    FlagPayrollModule   = "payroll.module"
    FlagInventoryModule = "inventory.module"
    FlagMultiCurrency   = "finance.multi_currency"
    FlagAdvancedReports = "reports.advanced"
    FlagAuditExport     = "audit.export"
    FlagAPIAccess       = "api.external_access"

    // Feature flags — gradual rollouts
    FlagNewInvoiceForm  = "ui.new_invoice_form"
    FlagBulkImport      = "ui.bulk_import"
    FlagConditionBuilder = "ui.condition_builder_v2"
)
```

---

## Layer 4 — IAM Middleware Pipeline

```go
// internal/middleware/auth.go
package middleware

import (
    "strings"
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/iam/session"
    "awo.so/internal/shared/response"
)

// AuthMiddleware validates the session token and sets a partial SessionContext.
// It does NOT resolve permissions or flags — that is IAMMiddleware's job.
//
// On failure: returns 401 AMIS envelope immediately.
// AMIS fetcher handles 401 → redirects to /login. Frontend has no auth logic.
func AuthMiddleware(store session.Store) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            return c.Status(401).JSON(response.Err("Session expired. Please log in."))
        }

        sess, err := store.Load(c.Context(), token)
        if err != nil || sess.IsExpired() {
            // Distinguish expired from invalid for audit
            if err == nil && sess.IsExpired() {
                return c.Status(401).JSON(response.Err("Session expired. Please log in."))
            }
            return c.Status(401).JSON(response.Err("Invalid session."))
        }

        // Attach partial session to request. IAMMiddleware completes it.
        setPartialSession(c, sess)
        return c.Next()
    }
}

// extractToken reads session token from cookie (primary) or Authorization header (API clients).
func extractToken(c *fiber.Ctx) string {
    if cookie := c.Cookies("awo_session"); cookie != "" {
        return cookie
    }
    auth := c.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    return ""
}
```

```go
// internal/middleware/iam.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/iam/policy"
    "awo.so/internal/iam/flags"
    "awo.so/internal/iam/session"
    "awo.so/internal/shared/response"
)

// IAMMiddleware resolves permissions and feature flags for the authenticated session.
// It completes the SessionContext begun by AuthMiddleware.
//
// MUST run after AuthMiddleware.
// MUST run before any schema handler.
//
// This is where ALL IAM resolution happens — once per request, results cached in session.
// Page functions never call the policy engine or flag engine directly.
// They receive a fully-resolved SessionContext and use typed accessors.
func IAMMiddleware(policyEngine policy.PolicyEngine, featureEngine flags.FeatureEngine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        partial := partialSessionFromLocals(c)

        // 1. Bulk-resolve all UI permissions
        perms, err := policyEngine.Resolve(
            c.Context(),
            partial.TenantID,
            partial.UserID,
            partial.Roles,
        )
        if err != nil {
            // Policy engine failure → deny all, log error
            c.Context().Logger().Printf("[IAM] permission resolution failed: %v", err)
            return c.Status(500).JSON(response.Err("Authorization service unavailable."))
        }

        // 2. Bulk-resolve all feature flags
        resolvedFlags, err := featureEngine.Resolve(
            c.Context(),
            partial.TenantID,
            partial.UserID,
        )
        if err != nil {
            // Flag resolution failure is non-fatal: default all flags to false (locked down)
            c.Context().Logger().Printf("[IAM] flag resolution failed: %v", err)
            resolvedFlags = map[string]bool{}
        }

        // 3. Rebuild SessionContext with resolved data (immutable from here)
        sess, err := session.NewBuilder(partial.SessionID, partial.UserID, partial.TenantID).
            Authenticated(partial.AuthMethod, partial.ExpiresAt).
            WithRoles(partial.Roles).
            WithScopes(partial.Scopes).
            WithPermissions(perms).
            WithFeatureFlags(resolvedFlags).
            WithPreferences(partial.Preferences).
            WithTenant(partial.Tenant).
            WithMetadata(partial.Metadata).
            Build()
        if err != nil {
            return c.Status(401).JSON(response.Err("Session invalid."))
        }

        // 4. Store completed session in Fiber locals
        SetSession(c, sess)
        return c.Next()
    }
}
```

```go
// internal/middleware/session.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/iam/session"
)

const localKeySession = "awo_session"
const localKeyPartial = "awo_session_partial"

func SetSession(c *fiber.Ctx, s session.SessionContext) {
    c.Locals(localKeySession, s)
}

func SessionFromFiber(c *fiber.Ctx) (session.SessionContext, bool) {
    s, ok := c.Locals(localKeySession).(session.SessionContext)
    return s, ok && s.Authenticated
}

func setPartialSession(c *fiber.Ctx, s session.SessionContext) {
    c.Locals(localKeyPartial, s)
}

func partialSessionFromLocals(c *fiber.Ctx) session.SessionContext {
    s, _ := c.Locals(localKeyPartial).(session.SessionContext)
    return s
}
```

---

## Layer 5 — UI Compiler

```go
// internal/web/compiler/compiler.go
package compiler

import (
    "context"
    "errors"
    "fmt"

    "awo.so/internal/iam/session"
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
)

// PageFn is the contract for all page schema functions.
// The session is the ONLY parameter. It is passed by value (immutable).
type PageFn func(session session.SessionContext) map[string]any

// NavFn contributes nav items to the compiled app shell.
type NavFn func(session session.SessionContext) []NavItem

// NavItem is a node in the compiled navigation tree.
type NavItem struct {
    Label    string
    Icon     string
    URL      string
    Children []NavItem
}

// Compiler compiles page schemas from session context.
type Compiler struct {
    policy  PolicyChecker   // thin interface: Can(session, perm) bool
    flags   FlagChecker     // thin interface: Enabled(session, flag) bool
    cache   SchemaCache
    validator *SchemaValidator
}

func NewCompiler(policy PolicyChecker, flags FlagChecker, cache SchemaCache) *Compiler {
    return &Compiler{
        policy:    policy,
        flags:     flags,
        cache:     cache,
        validator: NewSchemaValidator(),
    }
}

// Compile resolves and executes the page function for the given path.
// Returns the validated AMIS schema ready for serialization.
func (c *Compiler) Compile(ctx context.Context, path string, sess session.SessionContext) (map[string]any, error) {
    // Guard: session must be authenticated and unexpired
    if !sess.Authenticated || sess.IsExpired() {
        return nil, ErrSessionInvalid
    }

    // Cache lookup: keyed by path + session fingerprint
    key := cacheKey(path, sess)
    if cached, ok := c.cache.Get(ctx, key); ok {
        return cached, nil
    }

    // Registry lookup
    fn := registry.LookupPage(path)
    if fn == nil {
        return nil, ErrNotFound{Path: path}
    }

    // Execute the page function — this is the compilation step
    schema := fn(sess)

    // Validate the compiled schema
    if err := c.validator.Validate(schema); err != nil {
        return nil, fmt.Errorf("schema validation failed for %q: %w", path, err)
    }

    // Cache the result
    c.cache.Set(ctx, key, schema)

    return schema, nil
}

// CompileApp builds the full app shell schema with session-filtered navigation.
func (c *Compiler) CompileApp(ctx context.Context, sess session.SessionContext) map[string]any {
    if !sess.Authenticated || sess.IsExpired() {
        return unauthenticatedShell()
    }

    var items []NavItem
    for _, fn := range registry.AllNavFns() {
        contributed := fn(sess)
        items = append(items, contributed...)
    }

    return amis.BuildAppSchema(sess, items)
}

// ErrNotFound is returned when no page is registered for a path.
type ErrNotFound struct{ Path string }
func (e ErrNotFound) Error() string {
    return fmt.Sprintf("no page registered for path %q", e.Path)
}

var ErrSessionInvalid = errors.New("session is not authenticated or has expired")

// unauthenticatedShell returns a minimal AMIS page redirecting to login.
// This should rarely appear — AuthMiddleware handles 401 before compilation.
func unauthenticatedShell() map[string]any {
    return map[string]any{
        "type": "page",
        "body": map[string]any{
            "type":  "alert",
            "level": "danger",
            "body":  "Session expired. Please log in again.",
        },
    }
}
```

### Schema Cache

```go
// internal/web/compiler/cache.go
package compiler

import (
    "context"
    "crypto/sha256"
    "fmt"
    "sort"
    "strings"
    "time"

    "awo.so/internal/iam/session"
)

// SchemaCache caches compiled schemas.
// Cache key = path + permission fingerprint + flag fingerprint.
//
// Why NOT keyed by sessionID:
// Two users with identical roles and flags produce identical schemas.
// Keying by sessionID would miss the cache for every new login.
// Keying by permission+flag fingerprint collapses equivalent sessions.
//
// Cache invalidation triggers:
// - Role assignment change for a user → permission fingerprint changes
// - Feature flag toggle for a tenant → flag fingerprint changes
// - Schema code deployment → server restart clears in-memory cache
type SchemaCache interface {
    Get(ctx context.Context, key string) (map[string]any, bool)
    Set(ctx context.Context, key string, schema map[string]any)
    Invalidate(ctx context.Context, pattern string)
}

// cacheKey generates a deterministic cache key for a path + session combination.
// Two sessions with identical permissions and flags produce the same key.
func cacheKey(path string, sess session.SessionContext) string {
    pf := permFingerprint(sess.ResolvedPermissions())
    ff := flagFingerprint(sess.ResolvedFlags())
    return fmt.Sprintf("%s|%s|%s|%s", sess.TenantID, path, pf, ff)
}

func permFingerprint(perms map[string]bool) string {
    // Collect only granted permissions — denied ones don't affect schema
    var granted []string
    for k, v := range perms {
        if v {
            granted = append(granted, k)
        }
    }
    sort.Strings(granted)
    h := sha256.Sum256([]byte(strings.Join(granted, ",")))
    return fmt.Sprintf("%x", h[:8]) // 8 bytes = 16 hex chars, enough for cache key
}

func flagFingerprint(flags map[string]bool) string {
    var enabled []string
    for k, v := range flags {
        if v {
            enabled = append(enabled, k)
        }
    }
    sort.Strings(enabled)
    h := sha256.Sum256([]byte(strings.Join(enabled, ",")))
    return fmt.Sprintf("%x", h[:8])
}

// InMemorySchemaCache is a simple TTL cache for development and single-instance deployments.
// Replace with Redis-backed cache for multi-instance production.
type InMemorySchemaCache struct {
    mu      sync.RWMutex
    entries map[string]cacheEntry
    ttl     time.Duration
}

type cacheEntry struct {
    schema    map[string]any
    expiresAt time.Time
}

func NewInMemoryCache(ttl time.Duration) *InMemorySchemaCache {
    return &InMemorySchemaCache{
        entries: make(map[string]cacheEntry),
        ttl:     ttl,
    }
}

func (c *InMemorySchemaCache) Get(ctx context.Context, key string) (map[string]any, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    entry, ok := c.entries[key]
    if !ok || time.Now().After(entry.expiresAt) {
        return nil, false
    }
    return entry.schema, true
}

func (c *InMemorySchemaCache) Set(ctx context.Context, key string, schema map[string]any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.entries[key] = cacheEntry{
        schema:    schema,
        expiresAt: time.Now().Add(c.ttl),
    }
}

func (c *InMemorySchemaCache) Invalidate(ctx context.Context, pattern string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    for k := range c.entries {
        if strings.Contains(k, pattern) {
            delete(c.entries, k)
        }
    }
}
```

### Schema Validator

```go
// internal/web/compiler/validator.go
package compiler

import (
    "fmt"
    "regexp"
    "strings"
)

// SchemaValidator enforces structural and IAM correctness rules on compiled schemas.
// Runs in development (always) and in production (sampled at 1% to catch regressions).
//
// What it catches:
// - API URLs missing the method prefix ("get:", "post:", etc.)
// - visibleOn expressions containing business logic instead of boolean keys
// - Hardcoded permission expressions in AMIS schema
// - Missing syncLocation on crud components
// - Missing backgroundColor:transparent on chart components
type SchemaValidator struct {
    rules []validationRule
}

type validationRule struct {
    name  string
    check func(node map[string]any, path string) []ValidationError
}

type ValidationError struct {
    Path    string
    Rule    string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Rule, e.Path, e.Message)
}

var (
    // Detects business logic in AMIS expressions:
    // role checks, permission key access, direct comparisons
    // Allowed: ${can_create}, ${is_admin}, ${show_export} (pure boolean keys)
    // Forbidden: ${user.role === 'ADMIN'}, ${permissions.invoice_approve}
    dangerousExprPattern = regexp.MustCompile(
        `\$\{[^}]*(\.role|\.permission|\.roles|Permission|\.scope)[^}]*\}`,
    )

    // API must have method prefix
    apiMethodPattern = regexp.MustCompile(`^(get|post|put|patch|delete|GET|POST|PUT|PATCH|DELETE):`)
)

func NewSchemaValidator() *SchemaValidator {
    v := &SchemaValidator{}
    v.rules = []validationRule{
        {
            name: "no-iam-in-expressions",
            check: func(node map[string]any, path string) []ValidationError {
                var errs []ValidationError
                for _, key := range []string{"visibleOn", "disabledOn", "hiddenOn"} {
                    if expr, ok := node[key].(string); ok {
                        if dangerousExprPattern.MatchString(expr) {
                            errs = append(errs, ValidationError{
                                Path: path + "." + key,
                                Rule: "no-iam-in-expressions",
                                Message: fmt.Sprintf(
                                    "expression %q contains IAM logic — resolve in Go, inject boolean key",
                                    expr,
                                ),
                            })
                        }
                    }
                }
                return errs
            },
        },
        {
            name: "api-method-prefix",
            check: func(node map[string]any, path string) []ValidationError {
                var errs []ValidationError
                if api, ok := node["api"].(string); ok {
                    if !apiMethodPattern.MatchString(api) {
                        errs = append(errs, ValidationError{
                            Path: path + ".api",
                            Rule: "api-method-prefix",
                            Message: fmt.Sprintf("api %q missing method prefix (use 'get:...', 'post:...')", api),
                        })
                    }
                }
                return errs
            },
        },
        {
            name: "crud-sync-location",
            check: func(node map[string]any, path string) []ValidationError {
                if node["type"] != "crud" {
                    return nil
                }
                if sync, ok := node["syncLocation"].(bool); !ok || !sync {
                    return []ValidationError{{
                        Path: path,
                        Rule: "crud-sync-location",
                        Message: "crud component missing syncLocation:true — filter state won't persist in URL",
                    }}
                }
                return nil
            },
        },
        {
            name: "chart-transparent-bg",
            check: func(node map[string]any, path string) []ValidationError {
                if node["type"] != "chart" {
                    return nil
                }
                cfg, _ := node["config"].(map[string]any)
                if cfg["backgroundColor"] != "transparent" {
                    return []ValidationError{{
                        Path: path + ".config",
                        Rule: "chart-transparent-bg",
                        Message: "chart missing backgroundColor:transparent — will flash white in dark mode",
                    }}
                }
                return nil
            },
        },
    }
    return v
}

// Validate walks the schema tree and returns all validation errors.
func (v *SchemaValidator) Validate(schema map[string]any) error {
    var errs []ValidationError
    v.walk(schema, "$", &errs)
    if len(errs) == 0 {
        return nil
    }
    msgs := make([]string, len(errs))
    for i, e := range errs {
        msgs[i] = e.Error()
    }
    return fmt.Errorf("schema validation errors:\n%s", strings.Join(msgs, "\n"))
}

func (v *SchemaValidator) walk(node map[string]any, path string, errs *[]ValidationError) {
    for _, rule := range v.rules {
        *errs = append(*errs, rule.check(node, path)...)
    }
    // Recurse into body, children, tabs, columns
    for _, key := range []string{"body", "children", "tabs", "columns", "buttons", "toolbar"} {
        switch child := node[key].(type) {
        case map[string]any:
            v.walk(child, path+"."+key, errs)
        case []any:
            for i, item := range child {
                if m, ok := item.(map[string]any); ok {
                    v.walk(m, fmt.Sprintf("%s.%s[%d]", path, key, i), errs)
                }
            }
        }
    }
}
```

---

## Layer 6 — UI Block System

UI Blocks are reusable, session-aware schema fragments. Pages compose them rather than duplicating schema logic.

```go
// internal/web/blocks/blocks.go
package blocks

import "awo.so/internal/iam/session"

// UIBlock is a reusable schema fragment that accepts session context.
// Blocks are NOT pages — they are fragments composed inside pages.
// A block returns nil if the session has no access to its content.
type UIBlock func(sess session.SessionContext) map[string]any

// Compose returns all non-nil blocks as a slice suitable for AMIS body arrays.
func Compose(sess session.SessionContext, blockFns ...UIBlock) []any {
    var out []any
    for _, fn := range blockFns {
        schema := fn(sess)
        if schema != nil {
            out = append(out, schema)
        }
    }
    return out
}
```

```go
// internal/web/blocks/approvals.go
package blocks

import (
    "awo.so/internal/iam/session"
    "awo.so/internal/web/amis"
)

// ApprovalInboxBlock renders the pending approvals count + quick-action table.
// Returns nil if the user has no approval responsibilities.
func ApprovalInboxBlock(sess session.SessionContext) map[string]any {
    // Gate: user must be able to approve at least one resource type
    canApproveAny := sess.Can("invoice.approve") ||
        sess.Can("purchase_order.approve") ||
        sess.Can("payroll.approve")

    if !canApproveAny {
        return nil // block invisible — not rendered, not in JSON
    }

    return amis.Service("get:/api/v1/approvals/summary").
        Body(
            amis.Panel("Pending Approvals").
                Body(amis.CRUD("get:/api/v1/approvals/pending").
                    ID("approval-inbox").
                    PerPage(10).
                    Columns(
                        amis.Column("record_type", "Type").Build(),
                        amis.Column("record_ref", "Reference").Build(),
                        amis.Column("requested_by", "Requested By").Build(),
                        amis.Column("requested_at", "Date").Type("date").Build(),
                        buildApprovalActions(),
                    ).
                    EmptyText(
                        "No pending approvals.",
                        "No approvals match current filters.",
                    ).
                    Build(),
                ).
                Build(),
        ).
        Build()
}

func buildApprovalActions() amis.M {
    return amis.Column("", "Actions").Buttons(
        amis.M{
            "label": "View", "type": "button", "level": "link",
            "actionType": "drawer",
            "drawer": map[string]any{
                "title": "Approve ${record_type} #${record_ref}",
                "body": map[string]any{
                    "type":      "service",
                    "schemaApi": "get:/schema/approvals/${record_type}/${record_id}",
                },
            },
        },
        amis.AjaxBtn("Approve", "success", "post:/api/v1/approvals/${request_id}/approve", ""),
        amis.AjaxBtn("Reject", "danger", "post:/api/v1/approvals/${request_id}/reject",
            "Reject this request?"),
    ).Build()
}
```

```go
// internal/web/blocks/finance.go
package blocks

import (
    "awo.so/internal/iam/session"
    "awo.so/internal/iam/flags"
    "awo.so/internal/web/amis"
)

// ReceivablesSummaryBlock renders AR aging KPIs.
// Returns nil if user cannot read invoices.
func ReceivablesSummaryBlock(sess session.SessionContext) map[string]any {
    if !sess.Can("invoice.read") {
        return nil
    }
    return amis.Service("get:/api/v1/finance/receivables/summary").
        Body(
            amis.Grid(
                amis.Col("col-sm-4", amis.Stat("${total}", "Total Receivables")),
                amis.Col("col-sm-4", amis.Stat("${overdue_30}", "Overdue 30d")),
                amis.Col("col-sm-4", amis.Stat("${overdue_90}", "Overdue 90d+")),
            ),
        ).
        Build()
}

// RevenueChartBlock renders monthly revenue chart.
// Requires advanced reports flag.
func RevenueChartBlock(sess session.SessionContext) map[string]any {
    if !sess.Can("report.financial") || !sess.Flag(flags.FlagAdvancedReports) {
        return nil
    }
    return amis.Chart("get:/api/v1/analytics/revenue/monthly").
        Config(amis.M{
            "backgroundColor": "transparent",
            "xAxis":  amis.M{"type": "category", "data": "${months}"},
            "yAxis":  amis.M{"type": "value"},
            "series": []any{amis.M{"type": "bar", "data": "${values}", "name": "Revenue"}},
        }).
        Height(300).
        Build()
}
```

---

## Layer 7 — Session-Aware Page Function

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/iam/session"
    "awo.so/internal/web/amis"
    "awo.so/internal/web/compiler"
    "awo.so/internal/web/registry"
)

func init() {
    registry.RegisterPage("/finance/invoices", Schema)
}

// Schema compiles the Invoices list page for the given session.
//
// ALL conditional logic is resolved here.
// The compiled JSON has no branching — it is the final, flat schema
// for this specific session's permission and flag state.
func Schema(sess session.SessionContext) map[string]any {
    // Fail fast: this page requires read access
    if !sess.Can("invoice.read") {
        return compiler.AccessDeniedPage("Invoices", "invoice.read")
    }

    // Resolve all UI-relevant permissions upfront
    // These become boolean keys in page data — AMIS reads, never evaluates
    canCreate := sess.Can("invoice.create")
    canApprove := sess.Can("invoice.approve")
    canExport  := sess.Can("invoice.export")
    canDelete  := sess.Can("invoice.delete")
    currency   := sess.Tenant.Currency

    // Build toolbar — only permitted actions included
    var toolbar []any
    if canCreate {
        toolbar = append(toolbar,
            amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices",
                buildInvoiceFields(currency)...))
    }
    if canExport {
        toolbar = append(toolbar, amis.M{
            "type":       "button",
            "label":      "Export CSV",
            "level":      "default",
            "actionType": "ajax",
            "api":        "post:/api/v1/finance/invoices/export",
        })
    }

    crud := amis.CRUD("get:/api/v1/finance/invoices").
        ID("invoices-list").
        PerPage(sess.Pref("table_page_size", 20).(int)).
        DefaultSort("due_date", "asc").
        Columns(buildColumns(sess.Tenant.Currency, canApprove, canDelete)...).
        Filter(buildFilters()...).
        EmptyText(
            "No invoices yet. "+ifStr(canCreate, "Create your first invoice.", ""),
            "No invoices match the current filters.",
        )

    if len(toolbar) > 0 {
        crud = crud.Toolbar(toolbar...)
    }

    return amis.Page("Invoices").
        Body(crud.Build()).
        Build()
}

func buildColumns(currency string, canApprove, canDelete bool) []amis.M {
    cols := []amis.M{
        amis.Column("number", "Invoice #").Sortable().Fixed("left").Build(),
        amis.Column("supplier_name", "Supplier").Sortable().Build(),
        amis.Column("amount", "Amount").
            Tpl(currency+" ${amount | number:2}").
            Align("right").Sortable().Build(),
        amis.Column("status", "Status").
            ColorMap(amis.M{
                "DRAFT":   "default",
                "OPEN":    "processing",
                "PAID":    "success",
                "OVERDUE": "error",
                "VOIDED":  "warning",
            }).Build(),
        amis.Column("due_date", "Due Date").Type("date").Sortable().Build(),
    }

    // Approval column: only included if user can approve
    // Not hidden via visibleOn — simply absent from the schema
    var actionBtns []amis.M
    actionBtns = append(actionBtns, amis.ViewBtn(invoiceDetailDrawer()))
    actionBtns = append(actionBtns,
        amis.EditBtn("put:/api/v1/finance/invoices/${id}", buildInvoiceFields(currency)...))

    if canApprove {
        actionBtns = append(actionBtns,
            amis.AjaxBtn("Approve", "success",
                "post:/api/v1/finance/invoices/${id}/approve", "Approve this invoice?"))
    }
    if canDelete {
        actionBtns = append(actionBtns,
            amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}"))
    }

    cols = append(cols, amis.Column("", "Actions").Buttons(actionBtns...).Fixed("right").Build())
    return cols
}

func buildFilters() []amis.M {
    return []amis.M{
        amis.TextField("keywords", "Search"),
        amis.SelectField("status", "Status",
            amis.Opt("Draft", "DRAFT"),
            amis.Opt("Open", "OPEN"),
            amis.Opt("Overdue", "OVERDUE"),
            amis.Opt("Paid", "PAID"),
        ),
        amis.DateRangeField("date_range", "Due Date Range"),
    }
}

func buildInvoiceFields(currency string) []amis.M {
    return []amis.M{
        amis.Required(amis.SelectAPIField("supplier_id", "Supplier",
            "get:/api/v1/suppliers?perPage=200")),
        amis.Required(amis.DateField("invoice_date", "Invoice Date")),
        amis.Required(amis.DateField("due_date", "Due Date")),
        amis.Required(amis.AmountField("amount", "Amount", currency)),
        amis.Optional(amis.TextareaField("notes", "Notes")),
    }
}

func invoiceDetailDrawer() amis.M {
    return amis.M{
        "type": "service",
        "api":  "get:/api/v1/finance/invoices/${id}",
        "body": amis.Descriptions("Invoice").
            Item("Invoice #", "number").
            Item("Supplier", "supplier_name").
            Item("Amount", "amount").
            Item("Status", "status").
            Item("Due Date", "due_date").
            Columns(2).Build(),
    }
}

func ifStr(cond bool, t, f string) string {
    if cond { return t }
    return f
}
```

---

## Layer 8 — Schema Handler (Session-Gated)

```go
// internal/web/handler/schema.go
package handler

import (
    "errors"
    "strings"

    "github.com/gofiber/fiber/v2"
    "awo.so/internal/middleware"
    "awo.so/internal/web/compiler"
    "awo.so/internal/shared/response"
)

type SchemaHandler struct {
    compiler *compiler.Compiler
}

func NewSchemaHandler(c *compiler.Compiler) *SchemaHandler {
    return &SchemaHandler{compiler: c}
}

// Handle is the single entry point for all /schema/* requests.
// Session MUST be present — enforced by AuthMiddleware + IAMMiddleware upstream.
func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
    // Session is the single entry point. No session → no schema.
    sess, ok := middleware.SessionFromFiber(c)
    if !ok {
        // Should never reach here — AuthMiddleware blocks unauthenticated requests.
        // Defensive check: if session somehow missing, refuse with 401.
        return c.Status(401).JSON(response.Err("Session required."))
    }

    path := strings.TrimPrefix(c.Path(), "/schema")
    if path == "" || path == "/app" {
        schema := h.compiler.CompileApp(c.Context(), sess)
        return c.JSON(schema)
    }

    schema, err := h.compiler.Compile(c.Context(), path, sess)
    if err != nil {
        var notFound compiler.ErrNotFound
        if errors.As(err, &notFound) {
            return c.Status(404).JSON(notFoundPage(notFound.Path))
        }
        if errors.Is(err, compiler.ErrSessionInvalid) {
            return c.Status(401).JSON(response.Err("Session expired."))
        }
        c.Context().Logger().Printf("[schema] compilation error for %q: %v", path, err)
        return c.Status(500).JSON(response.Err("Schema compilation failed."))
    }

    return c.JSON(schema)
}

func notFoundPage(path string) map[string]any {
    return map[string]any{
        "type": "page", "title": "Not Found",
        "body": map[string]any{
            "type": "alert", "level": "warning",
            "body": "Page not found: " + path,
        },
    }
}
```

---

## Compiled AMIS JSON Output Example

Input: session for tenant `acme`, user `finance-manager`, permissions include `invoice.read`, `invoice.create`, `invoice.approve`, `invoice.export` but NOT `invoice.delete`. Currency = `KES`.

```json
{
  "type": "page",
  "title": "Invoices",
  "body": {
    "type": "crud",
    "id": "invoices-list",
    "api": "get:/api/v1/finance/invoices",
    "syncLocation": true,
    "perPage": 20,
    "orderBy": "due_date",
    "orderDir": "asc",
    "headerToolbar": [
      {
        "type": "button",
        "label": "New Invoice",
        "level": "primary",
        "actionType": "dialog",
        "dialog": {
          "title": "New Invoice",
          "body": {
            "type": "form",
            "api": "post:/api/v1/finance/invoices",
            "mode": "horizontal",
            "reload": "list",
            "body": [
              { "type": "select", "name": "supplier_id", "label": "Supplier",
                "source": "get:/api/v1/suppliers?perPage=200",
                "required": true },
              { "type": "input-date", "name": "invoice_date",
                "label": "Invoice Date", "required": true },
              { "type": "input-date", "name": "due_date",
                "label": "Due Date", "required": true },
              { "type": "input-number", "name": "amount", "label": "Amount",
                "prefix": "KES ", "precision": 2, "required": true },
              { "type": "textarea", "name": "notes",
                "label": "Notes", "remark": "(optional)" }
            ]
          }
        }
      },
      {
        "type": "button", "label": "Export CSV", "level": "default",
        "actionType": "ajax", "api": "post:/api/v1/finance/invoices/export"
      },
      "search-box",
      "bulkActions",
      { "type": "columns-toggler" },
      { "type": "reload" },
      { "type": "export-csv" }
    ],
    "footerToolbar": ["statistics", "pagination"],
    "filter": {
      "body": [
        { "type": "input-text", "name": "keywords", "label": "Search" },
        { "type": "select", "name": "status", "label": "Status",
          "clearable": true,
          "options": [
            { "label": "Draft",   "value": "DRAFT"   },
            { "label": "Open",    "value": "OPEN"    },
            { "label": "Overdue", "value": "OVERDUE" },
            { "label": "Paid",    "value": "PAID"    }
          ]
        },
        { "type": "input-date-range", "name": "date_range", "label": "Due Date Range" }
      ]
    },
    "columns": [
      { "name": "number",        "label": "Invoice #",  "sortable": true, "fixed": "left" },
      { "name": "supplier_name", "label": "Supplier",   "sortable": true },
      { "name": "amount",        "label": "Amount",     "type": "tpl",
        "tpl": "KES ${amount | number:2}", "align": "right", "sortable": true },
      { "name": "status",        "label": "Status",     "type": "tag",
        "colorMap": { "DRAFT":"default","OPEN":"processing","PAID":"success",
                      "OVERDUE":"error","VOIDED":"warning" } },
      { "name": "due_date",      "label": "Due Date",   "type": "date", "sortable": true },
      {
        "type": "operation", "label": "Actions", "fixed": "right",
        "buttons": [
          { "type": "button", "label": "View", "level": "link",
            "actionType": "drawer",
            "drawer": { "size": "lg", "body": {
              "type": "service", "api": "get:/api/v1/finance/invoices/${id}",
              "body": { "type": "descriptions", "title": "Invoice",
                "columns": 2,
                "items": [
                  { "label": "Invoice #",  "name": "number" },
                  { "label": "Supplier",   "name": "supplier_name" },
                  { "label": "Amount",     "name": "amount" },
                  { "label": "Status",     "name": "status" },
                  { "label": "Due Date",   "name": "due_date" }
                ]
              }
            }}
          },
          { "type": "button", "label": "Edit", "level": "link",
            "actionType": "dialog",
            "dialog": { "title": "Edit", "body": { "...": "edit form" } }
          },
          { "type": "button", "label": "Approve", "level": "success",
            "actionType": "ajax",
            "api": "post:/api/v1/finance/invoices/${id}/approve",
            "confirmText": "Approve this invoice?"
          }
        ]
      }
    ],
    "placeholder": { "empty": "No invoices yet. Create your first invoice." },
    "filterEmptyText": "No invoices match the current filters."
  }
}
```

**What is NOT in this JSON:**
- No `visibleOn` with role or permission checks
- No `canDelete` button (user lacks `invoice.delete` — button is absent, not hidden)
- No feature flag references
- No user ID, session ID, or sensitive data
- No business logic of any kind

The schema is inert data. The browser renders it. That is all.

---

## Security Threat Model

### T1: Session Token Theft

**Attack:** Attacker steals `awo_session` cookie via XSS or network interception.

**Mitigations:**
- Cookie: `HttpOnly; Secure; SameSite=Strict` — not readable by JS, not sent cross-site
- Session store: Redis with short TTL (8h) + absolute TTL (24h max)
- Revocation: `store.Revoke(sessionID)` on logout — stateless JWT cannot do this
- Device fingerprint check: `session.Metadata.DeviceID` mismatch → flag for review
- Concurrent session detection: alert on same session from two different IPs

**Schema handler risk:** Schema endpoints are authenticated — stolen session allows schema access for that user's permission set only. No privilege escalation possible via schema endpoint alone.

---

### T2: Privilege Escalation via Schema Manipulation

**Attack:** Attacker modifies the AMIS JSON returned by `/schema/*` to add UI elements that call unauthorized APIs.

**Why it fails:** The UI is not the enforcement layer. Every API call to `/api/v1/*` goes through:
```
AuthMiddleware → IAMMiddleware → RequirePermission(action, resource)
```
A fabricated "Approve" button that calls `POST /api/v1/invoices/${id}/approve` fails at the API layer with 403. The schema controls UX, not security.

**IMPORTANT:** This only holds if API handlers enforce permissions independently of what the schema says. This must be verified. An API handler that trusts the frontend to have shown the right buttons is broken.

---

### T3: Tenant Isolation Violation

**Attack:** Tenant A's session is used to request Tenant B's schemas or data.

**Schema layer:** `sess.TenantID` is embedded in the session (not from the request). The schema function uses `sess.TenantID` for all API URLs. The compiled schema's API endpoints include `tenantID` only via session-resolved paths — never from query parameters.

**API layer:** Every handler reads `tenantID` from session context (middleware-injected), not from request body. SQL: `WHERE tenant_id = $1` with session-resolved `tenantID`.

**Cache layer:** Cache key includes `sess.TenantID`. Schema cache for `acme` never collides with `globocorp`.

**Risk:** If `ContextFromFiber` (§28) ever reads `tenantID` from a request header instead of session, this layer fails. `ResolveTenant` middleware must read from session-bound identity, not user-supplied input.

---

### T4: Session Fixation

**Attack:** Attacker pre-sets a known session ID, victim authenticates, attacker now owns the session.

**Mitigation:** On successful authentication, always generate a new session ID. Never reuse pre-authentication session IDs. The session store's `Create` method generates the ID server-side.

---

### T5: Schema Injection via Compiled Data

**Attack:** User preference or tenant config value contains AMIS expression syntax (`${...}`) that gets interpolated into the schema.

**Example:** Tenant name is `Acme ${system.env.DATABASE_URL} Corp`.

**Mitigation:** The schema builder never interpolates Go string values into AMIS expression positions. Tenant name appears as a literal data value in `page.data`, not in a template string. The AMIS `tpl` component would only interpolate it in controlled template positions. Validate all tenant-config strings against an expression-syntax allowlist before storage.

---

### T6: Permission Map Staleness

**Attack (accidental):** Admin revokes a user's `invoice.approve` permission. User's cached session still shows `invoice.approve: true`. For the schema cache TTL duration, the user sees the Approve button.

**Why it's not a security failure:** The API enforces permissions at request time. The Approve button appears, but the API call returns 403. UX is wrong; security holds.

**To fix UX staleness:** On permission change, publish an invalidation event. IAMMiddleware subscribes and forces session refresh on next request. Schema cache entry for that permission fingerprint becomes unreachable (key changes when permission set changes).

---

## Scalability Analysis

### At 10 Users / 10 Pages

No issues. In-memory schema cache. Casbin evaluates synchronously. IAM resolution < 2ms. Schema compilation < 1ms.

---

### At 1,000 Users / 100 Modules

**IAM load:**
- 30 permission checks × 1,000 concurrent users = 30,000 Casbin evaluations/sec
- Casbin with policy in-process: ~10µs/check → 300ms cumulative BUT
- Pre-resolution in `IAMMiddleware` batches all 30 checks per request once
- After pre-resolution, all page function calls are O(1) map lookups — Casbin not called again

**Nav compilation:**
- 100 modules × nav filter per request = 100 NavFn calls per schema/app request
- Each NavFn is a pure function with 3–5 permission checks (map lookups)
- 100 nav functions × 5 permission lookups = 500 map lookups = ~50µs
- Cache `/schema/app` by nav fingerprint — most users share the same nav variant

**Schema cache hit rate:**
- 1,000 users likely map to ≤ 20 distinct permission+flag fingerprints (role combinations)
- Cache hit rate: ~95%+ after warm-up
- 5% misses → compile from scratch → typically < 5ms

**Verdict:** System handles 1,000 users comfortably without caching schema endpoint.

---

### At 100,000 Users / Multi-Tenant ERP Scale

**Session store:** Redis cluster. Session lookup: ~1ms RTT. Acceptable.

**IAM pre-resolution bottleneck:**
- 30 Casbin evaluations × 100k users → Casbin with PostgreSQL adapter becomes the bottleneck
- Solution: Cache the permission map in Redis keyed by `(tenantID, userID, roleFingerprint)` with 5-minute TTL
- Permission cache miss (cold start, role change): 30 Casbin calls → write cache → all subsequent requests for this user hit Redis
- Redis lookup: ~0.5ms vs Casbin PostgreSQL: ~30ms

**Feature flag resolution:**
- Tenant flags change rarely (admin action). Cache in Redis at tenant level, 1-minute TTL.
- User overrides: cache in Redis at user level, 30-second TTL.
- Combined: 1 Redis call per request (hash-get on tenant flags + user overrides)

**Schema compilation:**
- Schema cache backed by Redis with `tenantID+path+permFingerprint+flagFingerprint` key
- At 100 modules × 20 role variants × 100 tenants = 200,000 max cache entries
- Each compiled schema: ~5–50KB JSON. 200,000 × 25KB = 5GB Redis memory — too large.
- Mitigation: cap cache to 10,000 entries (LRU). Hot schemas (dashboard, list pages) stay cached. Cold schemas compile on first access. Acceptable.

**Nav compilation:**
- `/schema/app` is highest traffic endpoint — every page load
- Cache by `(tenantID, navFingerprint)` where navFingerprint = hash of granted module permissions
- Typical tenant: 3–5 distinct nav variants (admin, finance-manager, viewer, etc.)
- 100 tenants × 5 nav variants = 500 cache entries. Near-perfect hit rate.

**Tenant isolation at scale:**
- All cache keys prefixed with `tenantID` — no cross-tenant collision possible
- Redis keyspace notification on tenant flag change: invalidate all `tenantID|*` schema entries
- Role change event: invalidate `tenantID|userID|*` permission cache entries

**Bottleneck ranking at 100k users:**
1. Session validation (Redis lookup) — mitigated by Redis cluster
2. IAM permission resolution — mitigated by permission cache in Redis
3. Schema compilation (cache miss) — mitigated by schema cache + LRU
4. Database I/O from API handlers — not a schema layer concern

**Where it breaks first at 100k:** Redis becomes the single point of failure. Mitigation: Redis Sentinel / Redis Cluster with read replicas for session and cache reads.

---

## Final Verdict

### Is This Architecture Production-Grade?

**Yes, with conditions.**

The architecture is sound. The session-as-single-entry-point principle is correctly enforced. The policy engine abstraction prevents the permission map from being accessed directly. The schema validator catches IAM leakage into AMIS expressions at compile time. The cache design correctly keys by permission+flag fingerprint rather than sessionID, enabling high hit rates without cross-user data leakage.

---

### Where Will It Break First?

**The `UIPermissions` list in `policy/permissions.go`.**

This list drives IAM bulk-resolution and schema validation. Every new resource requires a developer to remember to add entries here. When they forget (and they will), the permission is not pre-resolved, the policy engine falls back to Casbin's slow path on every request for that permission, and the schema validator does not catch undeclared permission keys.

Fix: add a compile-time or startup check that every `sess.Can(perm)` call in every page function uses a permission string that appears in `UIPermissions`. A linter rule or a test that iterates all registered page functions with a mock session would catch this.

---

### What Must Be Enforced Immediately?

1. **`policy.Can()` is the ONLY way to check permissions.** Direct `session.permissions[key]` access is forbidden. Enforce with a linter rule (grep for `.permissions[` in page files → CI failure).

2. **`SchemaValidator` runs in CI.** Every build compiles all registered schemas with a minimal session and validates them. IAM expressions in `visibleOn` are caught before merge.

3. **API handlers enforce permissions independently.** The UI layer is UX, not security. Every mutation endpoint calls `access.Can()` regardless of what the schema says. Audit this on every new handler.

4. **Session cookies are `HttpOnly; Secure; SameSite=Strict`.** Non-negotiable.

5. **Permission changes invalidate cached schemas.** Without this, revoked permissions show buttons that fail silently at the API layer. Poor UX erodes user trust even when security holds.

---

### What Design Decisions Are Dangerous at Scale?

**`UIPermissions` as a static list:** Scalable to 100 modules with discipline. Dangerous because it is a shared mutable list with no ownership. Modules should declare their own permissions and register them in `init()`, similar to page registration. A central aggregated list is a merge conflict magnet at 50+ modules.

**In-memory schema cache:** Fine for single-instance deployments. Fatal for multi-instance. Replace with Redis before horizontal scaling. The cache interface (`SchemaCache`) is already abstracted — the swap is mechanical.

**Casbin as policy engine:** Casbin with a database adapter has known performance issues at high concurrency. At 100k users, the permission cache in Redis is non-optional. If Redis is unavailable, the fallback must be fail-closed (deny all), not fail-open (allow all).

**Session passed by value:** Correct for immutability. But at 30+ fields, copying `SessionContext` on every PageFn call is not free. Profile and consider passing `*SessionContext` (read-only pointer convention) if copy cost appears in profiling.
