[<-- Back to Index](README.md)

## How Other Modules Consume IAM

> **Updated 2026-05-18**: Reflects current single-path Casbin enforcement and
> the `iam/contract` consumer interface layer. All references to
> `buildPermissions()`, `session.Can()`, `session.CanDo()`, permission JSONB,
> and `svc.Middleware()` are removed — those no longer exist.

---

### Architectural principle

IAM is the **only** identity and authorization authority. Every other module is
a **dumb consumer**: it reads context, calls IAM APIs, and relies on middleware
to enforce permissions. Modules must never evaluate permissions themselves.

```
┌─────────────────────────────────────────────────────────────┐
│  HTTP Request                                               │
│                                                             │
│  middleware.Authenticate    ← validates token, sets session │
│  contract.InjectSessionContext ← bridges to Go context      │
│  middleware.Authorize(cfg, "finance.invoices.read")         │
│       └─► authzService.Enforce(Casbin) ← SOLE authz path   │
│                                                             │
│  Handler / Service ← receives trusted SessionContext        │
│       └─► contract.FromContext(ctx).TenantID()              │
│       └─► contract.FromContext(ctx).FeatureEnabled("...")   │
└─────────────────────────────────────────────────────────────┘
```

---

### The one import: `iam/contract`

Non-IAM modules import exactly **one** IAM package:

```go
import "awo.so/internal/core/iam/contract"
```

They must **never** import:
- `awo.so/internal/core/iam/service` — internal implementation
- `awo.so/internal/core/iam/repository` — database layer
- `awo.so/internal/core/iam` directly for authz (use middleware instead)

---

### 1. Route registration — `middleware.Authorize`

Authorization is declared at the route level. Handlers receive a pre-authorized
request; they never call `Enforce()`.

```go
// internal/api/routes/finance.go

import (
    "awo.so/internal/api/middleware"
    "awo.so/internal/core/iam/contract"
)

func RegisterFinanceRoutes(app *fiber.App, authCfg middleware.AuthConfig) {
    g := app.Group("/api/v1/finance",
        middleware.Authenticate(authCfg),
        contract.InjectSessionContext(), // bridges Fiber locals → Go context
    )

    // Permission string: "module.resource.action" → split on last dot
    // "finance.invoices.read" → Object="finance.invoices", Action="read"
    g.Get("/invoices",
        middleware.Authorize(authCfg, "finance.invoices.read"),
        handler.ListInvoices,
    )
    g.Post("/invoices",
        middleware.Authorize(authCfg, "finance.invoices.create"),
        handler.CreateInvoice,
    )
    g.Post("/invoices/:id/approve",
        middleware.Authorize(authCfg, "finance.invoices.approve"),
        handler.ApproveInvoice,
    )
}
```

**What `middleware.Authorize` does:**
1. Reads `Principal` from Fiber locals (set by `Authenticate`)
2. Calls `authzService.Enforce(ctx, Request{Subject, Domain, Object, Action})`
3. Returns **500** if Enforce errors (enforcer failure, not denial)
4. Returns **403** if denied
5. Calls `Next()` if allowed

Handlers are only reached when the request is already authorized.

---

### 2. Reading session context in handlers

Handlers access identity and runtime metadata via `contract.FromContext`:

```go
// internal/api/handlers/finance/invoice.go

import "awo.so/internal/core/iam/contract"

func (h *InvoiceHandler) ListInvoices(c *fiber.Ctx) error {
    sc, ok := contract.FromContext(c.UserContext())
    if !ok {
        return fiber.ErrUnauthorized // should not happen after Authenticate
    }

    // Identity
    tenantID := sc.TenantID()  // for RLS / query scoping
    userID   := sc.UserID()    // for audit trail / ownership

    // Entity scope (application-layer data visibility)
    scope := sc.EntityScope()  // used in repo WHERE clauses

    // Feature flags
    if !sc.FeatureEnabled("finance.autopay") {
        return fiber.NewError(fiber.StatusForbidden, "feature not enabled")
    }

    // Tenant settings
    pageSize := sc.SettingInt("finance.default_page_size", 50)

    return h.svc.ListInvoices(c.UserContext(), tenantID, userID, scope, pageSize)
}
```

---

### 3. Reading session context in service-layer code

Service methods receive `context.Context` — no Fiber dependency needed:

```go
// internal/core/finance/invoice_service.go

import "awo.so/internal/core/iam/contract"

func (s *InvoiceService) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
    sc, ok := contract.FromContext(ctx)
    if !ok {
        return nil, errors.New("unauthenticated")
    }

    tenantID := sc.TenantID() // RLS boundary for all queries
    userID   := sc.UserID()   // audit: created_by

    // Services never call Enforce — the request is already authorized
    // by middleware.Authorize before it reaches here.
    return s.repo.Create(ctx, tenantID, userID, req)
}
```

---

### 4. Feature-gated operations

Use `contract.SessionContext.FeatureEnabled` for soft feature gates.
This is **not** an authorization decision — it is a configuration check.

```go
func (s *BillingService) ProcessAutopay(ctx context.Context) error {
    sc, ok := contract.FromContext(ctx)
    if !ok {
        return errors.New("unauthenticated")
    }

    // Feature flag check (not authz — no Casbin call)
    if !sc.FeatureEnabled("billing.autopay") {
        return ErrFeatureNotEnabled
    }

    threshold := sc.SettingDecimal("billing.autopay_threshold", 10_000.00)
    return s.repo.ProcessPending(ctx, sc.TenantID(), threshold)
}
```

---

### 5. Auth entrypoints (login / logout)

Modules that need to trigger auth flows (e.g. an auth handler) depend on
`contract.AuthService` — not on `iam.SessionService` directly:

```go
// internal/api/handlers/auth/handler.go

import "awo.so/internal/core/iam/contract"

type AuthHandler struct {
    auth contract.AuthService  // ← narrow contract; no MFA/SSO internals
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.ErrBadRequest
    }

    result, err := h.auth.Login(c.Context(), req.Email, req.Password)
    if errors.Is(err, iam.ErrMFARequired) {
        // result.Token is the pending MFA token — send to client for completion
        return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
            "mfa_required": true,
            "pending_token": result.Token,
        })
    }
    if err != nil {
        return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
    }

    // Set HttpOnly session cookie
    c.Cookie(&fiber.Cookie{
        Name:     "session",
        Value:    result.Token,
        HTTPOnly: true,
        Secure:   true,
    })
    return c.JSON(fiber.Map{"user_id": result.Session.UserID()})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
    token := c.Cookies("session")
    _ = h.auth.Logout(c.Context(), token)
    c.ClearCookie("session")
    return c.SendStatus(fiber.StatusNoContent)
}
```

Wire up in main:

```go
sessions := iam.NewSessionService(identityRepo, sessionRepo, tracer, metrics, log)
authSvc   := contract.NewServiceAdapter(sessions)
authHandler := auth.NewHandler(authSvc)
```

---

### 6. Platform administration routes

Platform routes follow the same pattern — `middleware.Authorize` uses the
`_platform_` domain derived from the authenticated principal:

```go
func RegisterPlatformRoutes(app *fiber.App, authCfg middleware.AuthConfig) {
    g := app.Group("/platform",
        middleware.Authenticate(authCfg),
        contract.InjectSessionContext(),
    )

    g.Get("/tenants",
        middleware.Authorize(authCfg, "platform.tenants.read"),
        platformHandler.ListTenants,
    )
    g.Post("/tenants/:id/suspend",
        middleware.Authorize(authCfg, "platform.tenants.suspend"),
        platformHandler.SuspendTenant,
    )
}
```

The `Principal` derived from a platform session automatically carries
`Domain = "_platform_"` — no special handling needed in the handler.

---

### 7. What modules must NOT do

```go
// ❌ WRONG: calling Enforce directly in a service
func (s *InvoiceService) CreateInvoice(ctx context.Context, ...) error {
    ok, _ := s.authz.Enforce(ctx, iam.Request{...}) // bypass middleware
    if !ok { return iam.ErrForbidden }
    ...
}

// ❌ WRONG: checking UserType for access decisions
sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
if sess.UserType == "SYSADMIN" { // bypass Casbin
    return specialAdminAction()
}

// ❌ WRONG: importing session repository
import "awo.so/internal/core/iam/repository"
repo.GetSessionByToken(ctx, token) // direct DB access

// ❌ WRONG: importing service internals
import "awo.so/internal/core/iam/service"
svc.NewAuthzService(...) // bypasses the contract

// ✔ CORRECT: read context, let middleware handle authz
sc, _ := contract.FromContext(ctx)
tenantID := sc.TenantID()
```

---

### Contract package API summary

| Symbol | Type | Purpose |
|---|---|---|
| `SessionContext` | struct (value) | Read-only identity + metadata carrier |
| `SessionContext.UserID()` | `uuid.UUID` | Authenticated user |
| `SessionContext.TenantID()` | `uuid.UUID` | RLS boundary |
| `SessionContext.EntityScope()` | `iam.EntityScope` | Data visibility scope |
| `SessionContext.FeatureEnabled(key)` | `bool` | Feature flag check |
| `SessionContext.Setting(key, fallback)` | `string` | Tenant setting |
| `SessionContext.SettingBool/Int(key, fb)` | typed | Typed setting access |
| `SessionContext.Preference(key, fallback)` | `string` | User preference |
| `SessionContext.IsPlatform()` | `bool` | Context helper (not authz gate) |
| `WithContext(ctx, sc)` | `context.Context` | Inject session into Go context |
| `FromContext(ctx)` | `(SessionContext, bool)` | Read session from Go context |
| `InjectSessionContext()` | `fiber.Handler` | Bridge Fiber locals → Go context |
| `AuthService` | interface | Narrow auth entrypoints |
| `AuthService.Login(ctx, email, pw)` | `(LoginResult, error)` | Authenticate user |
| `AuthService.Logout(ctx, token)` | `error` | Invalidate session |
| `AuthService.ValidateSession(ctx, tok)` | `(SessionContext, bool)` | Check live session |
| `NewServiceAdapter(sessions)` | `AuthService` | Production implementation |

---

### What SessionContext does NOT expose

- No `Can()` / `CanDo()` — removed (BLOCK-1)
- No permissions map — removed (SES-1)
- No role list — use `AuthzService.GetRoles()` for audit only, never for access gates
- No `ToPrincipal()` — IAM-internal; needed only for `Enforce()` calls
- No raw `UserType` string — use `IsPlatform()` / `IsPortal()` for context checks
- No `Enforce()` or Casbin-level methods
