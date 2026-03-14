# AuthN / AuthZ Completion Guide — AWO ERP
**Goal**: Basic, working Authentication + Authorization using Casbin RBAC, then expand.
**Approach**: No ABAC, no over-engineering. Wire what exists, fill the gaps.

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

### Missing ✗ — the three gaps
| What | Why it blocks everything |
|------|--------------------------|
| **Session service** — `Login()`, `ValidateSession()`, `Logout()` | Nothing creates or reads `user_sessions`; no token is ever issued |
| **Auth middleware** — reads token → sets `authz_principal` | `authz.Middleware()` depends on this being set; without it every route 401s |
| **Login post-processing** — failed-attempt counter, `last_login_at` | Brute-force protection and audit trail; columns exist, service never writes them |

Everything else (Casbin, RLS, bcrypt, session schema) is ready to be plugged together.

---

## 2. What module.md Tells You to Build

The module.md IAM section gives you the exact design. These are the key decisions
that are already made for you — don't reinvent them.

### 2.1 Permission Format

```
{module}.{resource}.{action}

finance.ledger.accounts.read
finance.receivables.invoices.approve
portal.self.invoices.read          ← self.* always scoped by RLS to principal_id
platform.tenants.provision         ← platform.* only for platform users
```

Every nav item, every route → exactly one permission string.
Map these to Casbin objects as-is: `object = "finance.ledger.accounts"`, `action = "read"`.
Casbin's `keyMatch2` already handles `finance.ledger.*` wildcards.

### 2.2 ResolvedSession — the central type

module.md defines this as what every middleware and handler gets:

```go
// Put this in internal/core/authz/session.go  (new file, doesn't break anything)

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

// Build the authz.Principal from a resolved session (for Casbin middleware)
func (s *ResolvedSession) ToPrincipal() Principal {
    return Principal{Subject: subjectFor(s), Domain: domainFor(s)}
}
```

This is the single type that flows through the entire request. Handlers read it as
`c.Locals("session").(*ResolvedSession)`. Casbin middleware reads the Principal it
embeds. No two separate context objects.

### 2.3 Two-layer permission check

module.md says: the `sessions.permissions` JSONB is the **fast path** — an in-memory
map lookup with zero DB hits. Casbin is the **source of truth** used to compute that
map at login and for management operations (add role, check effective permissions).

```
Login time:   Casbin.GetPermissions(subject, domain) → build map → store in sessions row
Request time: sessions.permissions["finance.ledger.accounts.read"] → true/false (O(1))
Casbin:       still available for management API (what roles does user X have?)
```

You **don't choose between them**. Casbin generates the map; the map serves requests.

---

## 3. Completion Plan — Step by Step

Do these in order. Each step is independently testable.

### Step 1: Fix identity.Authenticate() — wire the TODOs

**File**: `internal/core/identity/service.go` (~line 204)

The `Authenticate()` method needs three additions that the DB schema already supports:

```go
// On auth failure — add to repo:
repo.IncrementFailedAttempts(ctx, user.ID)
// After N failures (5) — add to repo:
repo.LockAccount(ctx, user.ID, time.Now().Add(15*time.Minute))
// On auth success — add to repo:
repo.ResetFailedAttempts(ctx, user.ID)
repo.UpdateLastLogin(ctx, user.ID, time.Now())
// At the START of Authenticate() — check lockout:
if user.LockoutUntil != nil && user.LockoutUntil.After(time.Now()) {
    return nil, ErrAccountLocked
}
```

Add these four repo methods. All columns already exist in the `users` table.

### Step 2: Build the session service

Create `internal/core/iam/session/service.go` (new package, follows tenant pattern).

```
internal/core/iam/
├── session/
│   ├── service.go     ← Login, ValidateSession, Logout, RefreshSession
│   ├── repo.go        ← CreateSession, GetByToken, Invalidate, UpdateLastSeen
│   └── model.go       ← Session, ResolvedSession types
```

**`session.Service` interface:**

```go
type Service interface {
    // Login validates credentials, computes permissions, creates session row.
    // Returns opaque token + resolved session.
    Login(ctx context.Context, email, password string) (*ResolvedSession, string, error)

    // ValidateSession reads the token, checks expiry, updates last_seen_at.
    // This is called on every request by the middleware.
    ValidateSession(ctx context.Context, token string) (*ResolvedSession, error)

    // Logout marks the session is_active=false.
    Logout(ctx context.Context, token string) error
}
```

**Login flow (pseudocode):**

```go
func (s *service) Login(ctx, email, password string) (*ResolvedSession, string, error) {
    // 1. Authenticate credentials (calls identity.Authenticate)
    user, err := s.identity.Authenticate(ctx, email, password)
    if err != nil { return nil, "", err }

    // 2. Compute permission map via Casbin
    //    subject = authz.TenantSubject(user.ID.String())
    //    domain  = authz.TenantDomain(user.TenantID.String())
    //    policies = authz.GetPolicies(ctx, domain)  → build map[permission]bool
    perms, err := s.buildPermissions(ctx, user)

    // 3. Generate opaque token (crypto/rand, 32 bytes, hex-encoded)
    token := generateToken()

    // 4. Insert into user_sessions (store sha256 hash of token, not raw)
    err = s.repo.CreateSession(ctx, Session{
        UserID:    user.ID,
        TenantID:  user.TenantID,
        TokenHash: sha256hex(token),  // never store raw token
        Permissions: perms,           // sessions.permissions JSONB
        ExpiresAt: time.Now().Add(8 * time.Hour),
        IPAddress: extractIP(ctx),
    })

    // 5. Return raw token to caller (goes into cookie/header)
    resolved := &ResolvedSession{
        UserID: user.ID, TenantID: user.TenantID,
        UserType: user.UserType, Permissions: perms,
    }
    return resolved, token, nil
}
```

**ValidateSession flow:**

```go
func (s *service) ValidateSession(ctx, token string) (*ResolvedSession, error) {
    hash := sha256hex(token)
    sess, err := s.repo.GetByTokenHash(ctx, hash)
    // check: is_active=true, expires_at > now
    // UPDATE last_seen_at = NOW()  (async, don't block)
    // Build ResolvedSession from sess.permissions JSONB
    return resolveSession(sess), nil
}
```

### Step 3: Build the auth middleware

Create `internal/core/iam/middleware.go`.

This is the **one middleware** that authenticates every request. It replaces the
current empty placeholder.

```go
// AuthMiddleware reads the session token from Cookie or Authorization header,
// calls session.ValidateSession(), and stores the resolved session in Fiber context.
func AuthMiddleware(sessions session.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)  // Bearer header or "session" cookie
        if token == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
        }

        resolved, err := sessions.ValidateSession(c.Context(), token)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired session")
        }

        // Set BOTH locals:
        c.Locals(authz.LocalsKeyPrincipal, resolved.ToPrincipal()) // for authz.Middleware()
        c.Locals("session", resolved)                               // for handlers
        return c.Next()
    }
}
```

### Step 4: Wire routes

```go
// In your Fiber app setup:

api := app.Group("/api")
api.Use(iam.AuthMiddleware(sessionSvc))     // Step 3 — all routes need authn

// IAM routes (no additional authz needed beyond being logged in for own data)
api.Post("/auth/login",  loginHandler)     // public — no AuthMiddleware
api.Post("/auth/logout", logoutHandler)

// Protected routes
invoices := api.Group("/invoices")
invoices.Use(authzSvc.Middleware("finance.receivables.invoices", "read"))
invoices.Get("/",   listInvoicesHandler)

invoices.Use(authzSvc.Middleware("finance.receivables.invoices", "create"))
invoices.Post("/", createInvoiceHandler)
```

### Step 5: Seed one working role end-to-end

Don't try to seed all roles at once. Pick `tenant.admin` and make it work fully
before adding others.

```go
// Call this after provisioning a tenant:
func seedDefaultRoles(ctx context.Context, authzSvc authz.Service, tenantID string) error {
    domain := authz.TenantDomain(tenantID)

    // Add the permission policies for tenant.admin role
    policies := []authz.Policy{
        {Subject: "role:tenant.admin", Domain: domain, Object: "finance.*",   Action: "*", Effect: "allow"},
        {Subject: "role:tenant.admin", Domain: domain, Object: "people.*",    Action: "*", Effect: "allow"},
        {Subject: "role:tenant.admin", Domain: domain, Object: "settings.*",  Action: "*", Effect: "allow"},
    }
    for _, p := range policies {
        authzSvc.AddPolicy(ctx, p)
    }
    return nil
}

// Then when creating the first admin user:
func assignAdminRole(ctx context.Context, authzSvc authz.Service, userID, tenantID string) {
    subject := authz.TenantSubject(userID)
    domain  := authz.TenantDomain(tenantID)
    authzSvc.AssignRole(ctx, tenantID, subject, "role:tenant.admin", domain)
}
```

---

## 4. Ideas from module.md — Add Without Breaking

These are additions, not changes. None touch existing authz or identity code.

### 4.1 `AssignableTo` guard on role assignment

This is the single most important security addition from module.md. Without it, a
tenant admin can assign platform roles to tenant users.

Add to `authz/roles.go::AssignRole()` before the DB write:

```go
// AssignableTo is a field you add to role definitions (seed_roles or roles table).
// Check it here before writing.
if !slices.Contains(role.AssignableTo, actorType) {
    return fmt.Errorf("authz: role %q is not assignable to %s users", role, actorType)
}
```

The `actorType` comes from the caller's `user.UserType`. This is a two-line change,
doesn't break the interface, and prevents privilege escalation.

### 4.2 `ResolvedSession.Can()` for handler-level checks

Once `ResolvedSession` exists (Step 2 above), handlers can do fine-grained checks
without calling Casbin:

```go
sess := c.Locals("session").(*iam.ResolvedSession)
if !sess.Can("finance.receivables.invoices.approve") {
    return c.Status(403).JSON(...)
}
```

This is O(1), zero DB hits. Works alongside the Casbin middleware which does
the coarser route-level gate.

### 4.3 `PrincipalID` for portal self-service

From module.md: a portal user's `PrincipalID` is their `contacts.id` or
`employee.id`. Every portal data handler uses this — never a request parameter.

```go
// Portal invoice list handler:
sess := c.Locals("session").(*iam.ResolvedSession)
if sess.IsPortal() {
    // RLS already filters by principal_id at DB level.
    // But also assert at app layer:
    if queryParams.CustomerID != sess.PrincipalID {
        return c.Status(403).JSON(...)
    }
}
```

The `user_sessions` table already has columns for this. Just populate `principal_id`
at login time and include it in `ResolvedSession`.

### 4.4 Session risk score (future, low effort to prepare)

`user_sessions.risk_score INT` already exists. When you create a session, store a
simple initial score. Start at 0. Later you can increment it for:
- Login from new IP
- Login from new user_agent
- Login outside business hours

Middleware can check `risk_score > threshold → require re-auth`. This is one column
and one integer comparison — trivially addable later if you populate it from the start.

---

## 5. Structure Alignment

Your `internal/core/authz/` is flat. Your `internal/core/tenant/` uses
`domain/`, `repository/`, `service/` subdirectories.

**Don't restructure `authz/` now.** It works and has tests. Moving files mid-feature
causes merge conflicts and broken imports for no immediate gain.

**Do** follow the tenant pattern for any **new** packages you create:

```
internal/core/iam/          ← new, follows the pattern
├── session/
│   ├── model.go            domain types: Session, ResolvedSession
│   ├── service.go          Login, ValidateSession, Logout
│   └── repo.go             DB operations on user_sessions
├── middleware.go           AuthMiddleware (uses session.Service)
└── iam.go                  facade: re-exports what callers need
```

The `authz/` package plugs into this from the outside — `iam` calls `authz` to
compute permissions; it doesn't merge into it.

**Future refactor** (after things work): move `authz/` to
`internal/core/authz/engine/` + `internal/core/authz/middleware/` etc. Not now.

---

## 6. What NOT to Do

These are traps — avoid them for now:

| Temptation | Why not |
|-----------|---------|
| Implement ABAC policy evaluation | You already decided against this. The `policies` table and `policy_evaluations` tables can stay dormant. |
| JWT tokens | Session tokens in a DB table are simpler, revocable, and auditable. JWT can't be invalidated on logout without a deny-list (which is a session table). |
| Separate identity store per user type | module.md explicitly says one `users` table. Don't split it. |
| Store raw session tokens | Always store `sha256(token)`. If the sessions table leaks, raw tokens mean instant account takeover. |
| Restructure `authz/` before it's wired | Breaking working code to make it look nicer is the enemy of shipping. |
| Computing permissions on every request | Compute once at login, store in `sessions.permissions` JSONB, read the map on each request. |

---

## 7. Minimum Viable Completion Checklist

Work through this in order. Stop after each item and verify it works.

- [ ] **S1** `identity/repo.go` — add `IncrementFailedAttempts`, `ResetFailedAttempts`, `LockAccount`, `UpdateLastLogin`
- [ ] **S2** `identity/service.go` — wire the four calls in `Authenticate()`; add lockout check at top
- [ ] **S3** `internal/core/iam/session/model.go` — define `Session`, `ResolvedSession`, `Can()`, `ToPrincipal()`
- [ ] **S4** `internal/core/iam/session/repo.go` — `CreateSession`, `GetByTokenHash`, `Invalidate`, `UpdateLastSeen`
- [ ] **S5** `internal/core/iam/session/service.go` — `Login`, `ValidateSession`, `Logout`
- [ ] **S6** `internal/core/iam/middleware.go` — `AuthMiddleware` reading token, calling `ValidateSession`, setting locals
- [ ] **S7** Login HTTP handler — `POST /auth/login` → calls `session.Login`, sets cookie with token
- [ ] **S8** Seed one tenant + one admin user + assign `role:tenant.admin` + add its policies via Casbin
- [ ] **S9** One protected route using `AuthMiddleware + authzSvc.Middleware(...)` — verify end-to-end
- [ ] **S10** Logout handler — `POST /auth/logout` → calls `session.Logout`, clears cookie

After S10, you have working AuthN + AuthZ. Everything after that is expansion:
more roles, MFA, portal users, API keys, `AssignableTo` guard.

---

## 8. Security Gaps to Close Alongside (not blockers, but do them soon)

| # | Gap | Where | Effort |
|---|-----|-------|--------|
| G1 | Store `sha256(token)` not raw | `iam/session/repo.go:CreateSession` | 1 line |
| G2 | Brute-force counter | `identity/service.go` | S1+S2 above |
| G3 | `validate_and_set_tenant_context()` allows PENDING | `000106` migration | Add a new DOWN/UP to fix |
| G4 | `policy_evaluations` RLS uses unsafe `current_setting()` | `000703` migration | 2-line fix in new migration |
| G5 | `AssignableTo` guard in `authz/roles.go` | Add 4 lines to `AssignRole()` | After S10 |

G1 is part of building the session service — do it in the initial implementation.
G2 is Step 1. G3 and G4 are standalone migrations that can be done any time.
