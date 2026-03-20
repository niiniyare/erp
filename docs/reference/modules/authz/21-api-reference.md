[<-- Back to Index](README.md)

## API Reference

### Go Service Interface

```go
// Package path: awo/internal/core/authz

type Service interface {
    // ENFORCEMENT
    Enforce(ctx context.Context, r Request) (bool, error)
    EnforceBatch(ctx context.Context, reqs []Request) ([]bool, error)

    // ROLE MANAGEMENT
    AssignRole(ctx context.Context, tenantID, subject, role, domain string, opts ...AssignOpt) error
    RevokeRole(ctx context.Context, subject, role, domain string) error
    GetRoles(ctx context.Context, subject, domain string) ([]string, error)
    HasRole(ctx context.Context, subject, role, domain string) (bool, error)
    GetAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error)

    // POLICY MANAGEMENT
    AddPolicy(ctx context.Context, p Policy) error
    RemovePolicy(ctx context.Context, p Policy) error
    GetPolicies(ctx context.Context, domain string) ([]Policy, error)

    // HTTP
    Middleware(object, action string) fiber.Handler

    // CACHE
    InvalidateCache(ctx context.Context) error
}
```

---

### Constructor

```go
func New(cfg Config) (Service, error)

type Config struct {
    Pool    *pgxpool.Pool          // required
    Logger  logger.Logger          // required
    Metrics metrics.MetricsProvider // optional
    Tracer  tracing.Service        // optional
}
```

**Errors:**
- `"authz.New: pool is required"` — Pool is nil
- `"authz.New: logger is required"` — Logger is nil
- `"authz.New: build model: ..."` — Casbin model string parse error
- `"authz.New: create enforcer: ..."` — Casbin enforcer init error (usually DB connectivity)

---

### Enforce

```go
func (s *service) Enforce(ctx context.Context, r Request) (bool, error)

type Request struct {
    Subject string
    Domain  string
    Object  string
    Action  string
}
```

**Returns:**
- `(true, nil)` — access granted
- `(false, nil)` — access denied (no error — this is normal)
- `(false, ErrInvalidRequest)` — any of Subject/Domain/Object/Action is empty
- `(false, err)` — Casbin evaluation error (rare)

**Side effects:** Calls `revokeExpiredRoles(ctx, r.Subject, r.Domain)` before evaluation. If expiry check fails (non-fatal), a warning is logged and enforcement proceeds.

---

### EnforceBatch

```go
func (s *service) EnforceBatch(ctx context.Context, reqs []Request) ([]bool, error)
```

**Returns:**
- `([]bool, nil)` — slice parallel to input; `true` = allowed, `false` = denied
- `(nil, nil)` — if reqs is empty
- `(nil, ErrInvalidRequest)` — if any request has empty fields
- `(nil, err)` — Casbin batch evaluation error

**Note:** Does NOT call `revokeExpiredRoles` per-item (not on hot path for batch). If temporal role cleanup is needed before batch, call Enforce once first or call revokeExpiredRoles manually (not exported — future improvement).

---

### AssignRole

```go
func (s *service) AssignRole(
    ctx context.Context,
    tenantID, subject, role, domain string,
    opts ...AssignOpt,
) error
```

**Functional Options:**
```go
authz.WithExpiry(t time.Time) AssignOpt
authz.WithAssignedBy(sub string) AssignOpt
authz.WithDelegatedBy(sub string) AssignOpt
```

**Behavior:** UPSERT — safe to call multiple times with same params.

**Errors:**
- pgx errors (constraint violations, connectivity)
- `"authz AssignRole begin tx: ..."` — DB transaction start failed
- `"authz AssignRole insert: ..."` — role_assignments write failed
- `"authz AssignRole commit: ..."` — commit failed
- `"authz AssignRole casbin: ..."` — Casbin g-rule update failed

---

### RevokeRole

```go
func (s *service) RevokeRole(ctx context.Context, subject, role, domain string) error
```

**Behavior:** Marks `role_assignments.is_active = FALSE` and removes the Casbin g-rule. The `role_assignments` row is preserved for audit.

---

### GetRoles

```go
func (s *service) GetRoles(ctx context.Context, subject, domain string) ([]string, error)
```

**Returns:** Slice of role names from the Casbin in-memory model. Empty slice (not error) when subject has no roles.

**Note:** Returns current in-memory state. Reflects all `AssignRole`/`RevokeRole` calls but may differ from DB if `InvalidateCache` has not been called after direct SQL changes.

---

### HasRole

```go
func (s *service) HasRole(ctx context.Context, subject, role, domain string) (bool, error)
```

---

### GetAssignments

```go
func (s *service) GetAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error)

type RoleAssignment struct {
    ID          string
    Subject     string
    Role        string
    Domain      string
    TenantID    string
    AssignedBy  string
    DelegatedBy string
    ExpiresAt   *time.Time
    IsActive    bool
    CreatedAt   time.Time
}
```

**Returns:** All rows for subject+domain (active and inactive), ordered by `created_at DESC`. Suitable for UI display and audit screens.

---

### AddPolicy / RemovePolicy

```go
func (s *service) AddPolicy(ctx context.Context, p Policy) error
func (s *service) RemovePolicy(ctx context.Context, p Policy) error

type Policy struct {
    Subject string
    Domain  string
    Object  string
    Action  string
    Effect  string  // "allow" | "deny"
}
```

**AddPolicy errors:**
- `ErrPolicyConflict` (409) — exact rule already exists
- `ErrInvalidRequest` (400) — any required field empty
- `"authz AddPolicy: effect must be \"allow\" or \"deny\", got ..."` — invalid effect

**RemovePolicy notes:** All five fields must match exactly. Effect="" in RemovePolicy will match only rules with empty effect (which should not exist in a healthy system — always set effect on removal).

---

### GetPolicies

```go
func (s *service) GetPolicies(ctx context.Context, domain string) ([]Policy, error)
```

Returns all p-rules for the domain. Filtered by `v1 = domain` (the domain column in p-rules).

---

### Middleware

```go
func (s *service) Middleware(object, action string) fiber.Handler
```

Returns a Fiber middleware handler. Reads `Principal` from `c.Locals(LocalsKeyPrincipal)`.

**Object expansion:** If `c.Params("id")` is non-empty, object becomes `object + "/" + id`.

**HTTP responses:**
- `401 Unauthorized` — no Principal in Locals
- `403 Forbidden` — Enforce returned false
- `500 Internal Server Error` — Enforce returned error

---

### InvalidateCache

```go
func (s *service) InvalidateCache(ctx context.Context) error
```

Reloads all policies from the database. Blocking — returns when the in-memory model is fully rebuilt.

**When to call:**
- After direct SQL writes to `casbin_rule`
- After bulk policy import
- After another instance made policy changes (multi-instance sync)

---

### Error Types

```go
type Error struct {
    Code       string
    Message    string
    HTTPStatus int
}

func (e *Error) Error() string  // "[authz] CODE: message"

var (
    ErrForbidden      = &Error{"AUTHZ_FORBIDDEN",    "access denied",                   403}
    ErrUnauthorized   = &Error{"AUTHZ_UNAUTHORIZED", "authentication required",          401}
    ErrInvalidRequest = &Error{"AUTHZ_INVALID",      "subject/domain/obj/act required", 400}
    ErrPolicyConflict = &Error{"AUTHZ_DUPLICATE",    "policy already exists",           409}
)
```

---

### Helper Functions

```go
// Subject builders
authz.PlatformSubject(userID string) string   // "platform:{userID}"
authz.TenantSubject(userID string) string     // "tenant:{userID}"
authz.PortalSubject(userID string) string     // "portal:{userID}"
authz.APISubject(clientID string) string      // "api:{clientID}"

// Domain builders
authz.TenantDomain(tenantID string) string    // "{tenantID}"
authz.PortalDomain(tenantID string) string    // "{tenantID}:portal"
authz.APIDomain(tenantID string) string       // "{tenantID}:api"

// Constants
authz.DomainPlatform  = "_platform_"
authz.LocalsKeyPrincipal = "authz_principal"

// Types
authz.Principal{Subject, Domain string}

// AssignOpt constructors
authz.WithExpiry(t time.Time) AssignOpt
authz.WithAssignedBy(sub string) AssignOpt
authz.WithDelegatedBy(sub string) AssignOpt
```

---

### HTTP Admin API (Suggested Endpoints)

These endpoints are not in the authz package itself — they are implemented by the platform/tenant admin handlers that call the Service interface.

```
GET    /platform/authz/policies?domain=_platform_
POST   /platform/authz/policies
DELETE /platform/authz/policies
POST   /platform/authz/cache/reload        → svc.InvalidateCache()

GET    /api/v1/admin/users/:id/roles       → svc.GetRoles + svc.GetAssignments
POST   /api/v1/admin/users/:id/roles       → svc.AssignRole
DELETE /api/v1/admin/users/:id/roles/:role → svc.RevokeRole

GET    /api/v1/admin/policies              → svc.GetPolicies
POST   /api/v1/admin/policies              → svc.AddPolicy
DELETE /api/v1/admin/policies              → svc.RemovePolicy
```

---

Next: [Summary](./22-summary.md)
