# AuthN / AuthZ Completion Guide — AWO ERP
**Goal**: Basic, working Authentication + Authorization using Casbin RBAC, then expand.
**Approach**: No ABAC, no over-engineering. Wire what exists, fill the gaps.
**Constraint**: `internal/core/{iam,abac,access}` — do NOT modify. Copy patterns only.

---

## 1. What You Have vs What's Missing

### Done ✓
| Component | Location | State |
|-----------|----------|-------|
| Casbin RBAC engine | `internal/core/authz/` | Fully working, integration-tested |
| Tenant context (RLS) | `db/migration/000055` | Production-ready with ADRs |
| RLS policies on all tables | `db/migration/000056+` | Fail-closed, working |
| User + session DB schema | `000303`, `000304` | Tables exist, all columns correct |
| Password hashing | `identity/service.go` | bcrypt, working |
| Tenant CRUD | `internal/core/tenant/` | Fully working |
| Casbin role assignment | `authz/roles.go` | Working, lazy expiry |
| Tenant middleware | `internal/api/middleware/tenant.go` | Fully working — the pattern to follow |

### Missing ✗ — the three gaps
| What | Why it blocks everything |
|------|--------------------------|
| **Session service** — `Login()`, `ValidateSession()`, `Logout()` | Nothing creates or reads `user_sessions`; no token is ever issued |
| **Auth middleware** — reads token → sets `authz_principal` | `authz.Middleware()` depends on this; `jwt_auth.go` and `authorization.go` import dead `iam` packages |
| **Login post-processing** — failed-attempt counter, `last_login_at` | Brute-force protection and audit trail; columns exist, service never writes them |

### Already in the right place, wrong imports
| File | Current imports | Fix |
|------|----------------|-----|
| `internal/api/middleware/authorization.go` | `internal/core/iam`, `internal/core/iam/authz` | Rewrite to use `internal/core/authz` |
| `internal/api/middleware/jwt_auth.go` | `internal/core/iam`, `internal/core/iam/authn` | Rewrite to use `internal/core/identity` + `internal/core/authz` |

Do **not** move these files. They're in the right place. Just fix their internals.

---

## 2. Architecture Constraints

### Package Rules
```
internal/core/authz/      ← Casbin engine + Middleware() factory (service-pure)
internal/core/identity/   ← User CRUD, Authenticate() (service-pure)
internal/core/identity/session/  ← NEW: Login, ValidateSession, Logout (service-pure)
internal/api/middleware/  ← All Fiber middleware: auth, authz, tenant
```

- **No middleware in `internal/core/`** — these are pure service layers
- `authz.Middleware()` is the exception: it's a factory that *returns* a fiber.Handler but lives in authz because it's tightly coupled to the Casbin enforcer. This is acceptable.
- **Do not modify** `internal/core/{iam,abac,access}` — they can be deleted later or left as dead code. Copy patterns from them if useful.

### sys-desing.md Middleware Chain
The middleware order matters. Follow this exactly:
```
Logger → Recovery → CORS → RateLimiter → Authenticate → ResolveTenant → SetTenantContext → Authorize
```
`Authenticate` = `internal/api/middleware/jwt_auth.go` (rewritten)
`Authorize`    = `internal/api/middleware/authorization.go` (rewritten)

### What to Copy from `internal/core/iam`
Don't throw away their ideas, just re-implement cleanly:
| Concept | Where it exists in iam | Where to implement in your stack |
|---------|------------------------|----------------------------------|
| `ResolvedSession` type | `iam/session.go` or similar | `internal/core/identity/session/model.go` |
| Permission map building | `iam/authn` or similar | `identity/session/service.go::buildPermissions()` |
| `Can(permission)` helper | iam session type | Same as above, on `ResolvedSession` |
| `AssignableTo` guard | iam role model | `authz/roles.go::AssignRole()` (4-line addition) |

---

## 3. Key Design Decisions (from module.md)

### 3.1 Permission Format
```
{module}.{resource}.{action}

finance.ledger.accounts.read
finance.receivables.invoices.approve
portal.self.invoices.read          ← self.* scoped by RLS to principal_id
platform.tenants.provision         ← platform.* only for platform users
```
Map to Casbin: `object = "finance.ledger.accounts"`, `action = "read"`.
`keyMatch2` already handles `finance.ledger.*` wildcards.

### 3.2 ResolvedSession — the central type
```go
// internal/core/identity/session/model.go

type ResolvedSession struct {
    UserID      uuid.UUID
    UserType    string          // "platform" | "tenant" | "portal"
    TenantID    uuid.UUID       // zero for platform
    PrincipalID uuid.UUID       // portal users: their contact/employee ID
    DisplayName string
    Permissions map[string]bool // pre-computed at login; session.permissions JSONB
}

const LocalsKeySession = "session"

func (s *ResolvedSession) Can(permission string) bool {
    return s.Permissions[permission]
}

// Builds the authz.Principal that authz.Middleware() reads from Fiber context
func (s *ResolvedSession) ToPrincipal() authz.Principal {
    return authz.Principal{Subject: subjectFor(s), Domain: domainFor(s)}
}

func subjectFor(s *ResolvedSession) string {
    switch s.UserType {
    case "platform": return "platform:" + s.UserID.String()
    case "portal":   return "portal:" + s.UserID.String()
    default:         return "tenant:" + s.UserID.String()
    }
}

func domainFor(s *ResolvedSession) string {
    if s.UserType == "platform" {
        return "_platform_"
    }
    return s.TenantID.String()
}
```

### 3.3 Two-layer permission check
```
Login time:   Casbin.GetPolicies(domain) → filter for user's roles → build map → store in sessions row
Request time: sessions.permissions["finance.ledger.accounts.read"] → true/false  (O(1), zero DB)
Casbin:       still used for management API (add role, check effective permissions)
```

---

## 4. Completion Plan — Step by Step

Do these in order. Each step is independently testable.

### Step 1: Fix identity.Authenticate() — wire the TODOs

**File**: `internal/core/identity/service.go` (~line 204)

```go
// At the START of Authenticate() — check lockout:
if user.LockoutUntil != nil && user.LockoutUntil.After(time.Now()) {
    return nil, ErrAccountLocked
}

// On auth failure:
repo.IncrementFailedAttempts(ctx, user.ID)
if user.FailedLoginAttempts+1 >= 5 {
    repo.LockAccount(ctx, user.ID, time.Now().Add(15*time.Minute))
}
return nil, ErrInvalidCredentials

// On auth success:
repo.ResetFailedAttempts(ctx, user.ID)
repo.UpdateLastLogin(ctx, user.ID, time.Now())
```

Add to `internal/core/identity/repo.go`:
- `IncrementFailedAttempts(ctx, userID uuid.UUID) error`
- `ResetFailedAttempts(ctx, userID uuid.UUID) error`
- `LockAccount(ctx, userID uuid.UUID, until time.Time) error`
- `UpdateLastLogin(ctx, userID uuid.UUID, at time.Time) error`

All columns already exist in `users` table.

### Step 2: Build the session service

```
internal/core/identity/session/
├── model.go     ← Session, ResolvedSession types
├── service.go   ← Login, ValidateSession, Logout
└── repo.go      ← CreateSession, GetByTokenHash, Invalidate, UpdateLastSeen
```

**`session.Service` interface:**
```go
type Service interface {
    Login(ctx context.Context, email, password string) (*ResolvedSession, string, error)
    ValidateSession(ctx context.Context, token string) (*ResolvedSession, error)
    Logout(ctx context.Context, token string) error
}
```

**Login flow:**
```go
func (s *service) Login(ctx, email, password string) (*ResolvedSession, string, error) {
    // 1. identity.Authenticate() handles brute-force check (Step 1)
    user, err := s.identity.Authenticate(ctx, email, password)
    if err != nil { return nil, "", err }

    // 2. Compute permission map via Casbin
    domain := authz.TenantDomain(user.TenantID.String())
    perms, err := s.buildPermissions(ctx, user, domain)

    // 3. Generate opaque token (crypto/rand, 32 bytes, hex-encoded)
    token := generateToken()

    // 4. Insert into user_sessions — store sha256(token), never raw
    err = s.repo.CreateSession(ctx, Session{
        UserID:      user.ID,
        TenantID:    user.TenantID,
        TokenHash:   sha256hex(token),
        Permissions: perms,
        ExpiresAt:   time.Now().Add(sessionTTL(s.cfg)),
        IPAddress:   extractIP(ctx),
    })

    // 5. Return raw token to caller (goes into cookie/header)
    resolved := &ResolvedSession{
        UserID:      user.ID,
        TenantID:    user.TenantID,
        UserType:    user.UserType,
        Permissions: perms,
    }
    return resolved, token, nil
}
```

**buildPermissions — uses Casbin:**
```go
func (s *service) buildPermissions(ctx context.Context, user *identity.User, domain string) (map[string]bool, error) {
    subject := authz.TenantSubject(user.ID.String())
    policies, err := s.authz.GetPolicies(ctx, domain)
    if err != nil { return nil, err }

    roles, err := s.authz.GetRoles(ctx, subject, domain)
    if err != nil { return nil, err }

    roleSet := make(map[string]bool, len(roles))
    for _, r := range roles { roleSet[r] = true }

    perms := make(map[string]bool)
    for _, p := range policies {
        if roleSet[p.Subject] && p.Effect == "allow" {
            // permission key = object + "." + action
            perms[p.Object+"."+p.Action] = true
        }
    }
    return perms, nil
}
```

### Step 3: Rewrite `internal/api/middleware/jwt_auth.go`

Replace the broken iam imports. The new version:

```go
package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "your/module/internal/core/authz"
    "your/module/internal/core/identity/session"
)

type AuthConfig struct {
    SessionSvc session.Service
    // Feature flags
    RequireHTTPS bool
    CookieName   string // default: "session"
}

func Authenticate(cfg AuthConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c, cfg.CookieName)
        if token == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }

        resolved, err := cfg.SessionSvc.ValidateSession(c.Context(), token)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired session")
        }

        // Set BOTH locals — session for handlers, principal for authz.Middleware()
        c.Locals(session.LocalsKeySession, resolved)
        c.Locals(authz.LocalsKeyPrincipal, resolved.ToPrincipal())
        return c.Next()
    }
}

func extractToken(c *fiber.Ctx, cookieName string) string {
    // 1. Authorization: Bearer <token>
    if h := c.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
        return strings.TrimPrefix(h, "Bearer ")
    }
    // 2. Cookie
    if cookieName == "" { cookieName = "session" }
    return c.Cookies(cookieName)
}
```

### Step 4: Rewrite `internal/api/middleware/authorization.go`

Replace the broken iam imports. The new version wraps `authz.Service.Middleware()`:

```go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "your/module/internal/core/authz"
    "your/module/internal/core/identity/session"
)

// Authorize returns a middleware that checks if the resolved session's
// pre-computed permission map allows the given permission.
// Falls back to Casbin for any permission not in the map (shouldn't happen normally).
func Authorize(object, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess, ok := c.Locals(session.LocalsKeySession).(*session.ResolvedSession)
        if !ok || sess == nil {
            return fiber.NewError(fiber.StatusUnauthorized, "no session")
        }

        permission := object + "." + action
        if sess.Can(permission) {
            return c.Next()
        }
        return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
    }
}

// AuthorizeWithCasbin is the Casbin-backed version for management operations
// where the pre-computed map may be stale (e.g., after role changes).
func AuthorizeWithCasbin(svc authz.Service, object, action string) fiber.Handler {
    return svc.Middleware(object, action)
}
```

### Step 5: Wire routes (follow sys-desing.md middleware chain)

```go
// Logger → Recovery → CORS → RateLimiter applied at app level (existing)

app := fiber.New()
app.Use(logger, recovery, cors, rateLimiter) // existing

api := app.Group("/api")

// Public routes — before auth middleware
api.Post("/auth/login",  loginHandler(sessionSvc))

// Auth middleware chain: Authenticate → ResolveTenant → SetTenantContext → Authorize
authenticated := api.Group("",
    middleware.Authenticate(authCfg),
    middleware.ResolveTenant(tenantCfg),   // existing tenant.go middleware
)

// Protected routes with authz
invoices := authenticated.Group("/invoices")
invoices.Get("/",   middleware.Authorize("finance.receivables.invoices", "read"),   listInvoicesHandler)
invoices.Post("/",  middleware.Authorize("finance.receivables.invoices", "create"), createInvoiceHandler)

// Logout — needs auth but no extra authz
authenticated.Post("/auth/logout", logoutHandler(sessionSvc))
```

### Step 6: Login and Logout HTTP handlers

```go
// internal/api/handler/auth.go

func LoginHandler(svc session.Service, cfg LoginConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req struct {
            Email    string `json:"email"`
            Password string `json:"password"`
        }
        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "invalid request")
        }

        resolved, token, err := svc.Login(c.Context(), req.Email, req.Password)
        if err != nil {
            // map identity errors to HTTP
            return mapAuthError(err)
        }

        c.Cookie(&fiber.Cookie{
            Name:     "session",
            Value:    token,
            HTTPOnly: true,
            Secure:   cfg.RequireHTTPS,
            SameSite: "Lax",
            MaxAge:   int(8 * time.Hour / time.Second),
        })
        return c.JSON(resolved)
    }
}

func LogoutHandler(svc session.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c, "session")
        if token != "" {
            _ = svc.Logout(c.Context(), token) // best-effort
        }
        c.ClearCookie("session")
        return c.SendStatus(fiber.StatusNoContent)
    }
}
```

### Step 7: Seed one working role end-to-end

```go
func seedDefaultRoles(ctx context.Context, authzSvc authz.Service, tenantID string) error {
    domain := authz.TenantDomain(tenantID)
    policies := []authz.Policy{
        {Subject: "role:tenant.admin", Domain: domain, Object: "finance.*",  Action: "*", Effect: "allow"},
        {Subject: "role:tenant.admin", Domain: domain, Object: "people.*",   Action: "*", Effect: "allow"},
        {Subject: "role:tenant.admin", Domain: domain, Object: "settings.*", Action: "*", Effect: "allow"},
    }
    for _, p := range policies {
        if err := authzSvc.AddPolicy(ctx, p); err != nil && err != authz.ErrPolicyConflict {
            return err
        }
    }
    return nil
}

func assignAdminRole(ctx context.Context, authzSvc authz.Service, userID, tenantID string) {
    subject := authz.TenantSubject(userID)
    domain  := authz.TenantDomain(tenantID)
    authzSvc.AssignRole(ctx, tenantID, subject, "role:tenant.admin", domain)
}
```

---

## 5. Feature Flags & Configuration

Gate auth features via the settings system. Don't hardcode security thresholds.

```go
// internal/core/identity/session/config.go

type Config struct {
    // Session
    SessionTTL         time.Duration // default: 8h
    RefreshEnabled     bool          // future: sliding sessions

    // Brute-force protection
    MaxFailedAttempts  int           // default: 5
    LockoutDuration    time.Duration // default: 15m

    // Feature flags
    MFAEnabled         bool          // default: false (future)
    RequireHTTPS       bool          // default: true in prod
    RiskScoringEnabled bool          // future: populate risk_score

    // Cookie
    CookieName         string        // default: "session"
    CookieDomain       string        // for multi-subdomain setups
}

func sessionTTL(cfg Config) time.Duration {
    if cfg.SessionTTL == 0 { return 8 * time.Hour }
    return cfg.SessionTTL
}
```

Load from the app's config/settings system (not hardcoded). Pass `Config` into `session.New(db, identitySvc, authzSvc, cfg)`.

---

## 6. Security Gaps to Close

| # | Gap | Where | Effort |
|---|-----|-------|--------|
| G1 | Store `sha256(token)` not raw | `identity/session/repo.go:CreateSession` | 1 line — **do this from the start** |
| G2 | Brute-force counter | `identity/service.go` | Step 1 above |
| G3 | `validate_and_set_tenant_context()` allows PENDING | `000106` migration | New migration to fix |
| G4 | `policy_evaluations` RLS unsafe `current_setting()` | `000703` migration | 2-line fix in new migration |
| G5 | `AssignableTo` guard in `authz/roles.go` | Add 4 lines to `AssignRole()` | After things work |

---

## 7. Minimum Viable Completion Checklist

Work through this in order. Stop after each item and verify.

- [ ] **S1** `identity/repo.go` — add `IncrementFailedAttempts`, `ResetFailedAttempts`, `LockAccount`, `UpdateLastLogin`
- [ ] **S2** `identity/service.go` — wire the four calls in `Authenticate()`; lockout check at top
- [ ] **S3** `identity/session/model.go` — `Session`, `ResolvedSession`, `Can()`, `ToPrincipal()`
- [ ] **S4** `identity/session/repo.go` — `CreateSession`, `GetByTokenHash`, `Invalidate`, `UpdateLastSeen`
- [ ] **S5** `identity/session/service.go` — `Login`, `ValidateSession`, `Logout` with `buildPermissions`
- [ ] **S6** Rewrite `api/middleware/jwt_auth.go` — `Authenticate()` calling `ValidateSession`, setting locals
- [ ] **S7** Rewrite `api/middleware/authorization.go` — `Authorize()` using `ResolvedSession.Can()`
- [ ] **S8** `api/handler/auth.go` — `LoginHandler` (sets cookie), `LogoutHandler` (clears cookie)
- [ ] **S9** Seed one tenant + one admin user + `role:tenant.admin` policies via Casbin
- [ ] **S10** One protected route: `Authenticate → Authorize("finance.receivables.invoices", "read")` — verify end-to-end

After S10: working AuthN + AuthZ. Everything after is expansion: more roles, MFA, portal users, API keys.

---

## 8. What NOT to Do

| Temptation | Why not |
|-----------|---------|
| Modify `internal/core/{iam,abac,access}` | Dead packages. Copy patterns, don't touch. |
| Put middleware logic in `internal/core/` | Core = service-pure. Fiber-specific code belongs in `api/middleware/`. |
| JWT tokens | Session tokens in DB are simpler, revocable, and auditable. |
| Store raw session tokens | Always `sha256(token)`. Token leak = instant account takeover otherwise. |
| Separate identity stores per user type | One `users` table. `user_type` column differentiates. |
| Compute permissions on every request | Compute once at login, store in `sessions.permissions` JSONB. |
| Restructure `authz/` before it works | Don't break working code to make it prettier. |
| Hardcode security thresholds | Use `session.Config` — max attempts, lockout duration, TTL all configurable. |
