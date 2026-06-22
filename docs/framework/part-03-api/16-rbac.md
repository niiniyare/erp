---
title: "Chapter 16: Role-Based Access Control"
part: "Part III — The API Layer"
chapter: 16
section: "16-rbac"
related:
  - "[Chapter 15: Authentication and Session Management](15-auth-sessions.md)"
  - "[Chapter 9: Privacy Policies](../part-02-entity-system/09-privacy-policies.md)"
---

# Chapter 16: Role-Based Access Control

Awo's RBAC is implemented via Casbin with a PostgreSQL adapter and a Redis-backed SyncedEnforcer. It controls what operations an actor can perform on entity types — complemented by Privacy Policies (Chapter 9) which control which rows within those entity types are accessible.

---

## 16.1. Role Model

### 16.1.1. System Roles — Defined in Framework Code

System roles are seeded by the framework during tenant bootstrapping and cannot be deleted:

| Role | Purpose |
|---|---|
| `role:platform-admin` | Full access to all tenants, platform configuration |
| `role:tenant.admin` | Full access within one tenant, cannot change platform settings |
| `role:tenant.user` | Standard user access, customised per-tenant |
| `role:api-client` | Machine-to-machine access, limited to configured scopes |

Subject format in Casbin: `user:{uuid}` for users, `role:{name}` for roles.
Domain format: `_platform_` for platform-level policies, tenant UUID for tenant-scoped policies.

### 16.1.2. Tenant Roles — Defined Per Tenant

Tenants create custom roles via the admin UI:

```
role:finance_manager   — full access to Finance module
role:sales_rep         — create/read own invoices, read customer records
role:collection_agent  — read/update assigned invoices only
role:viewer            — read-only across all entities
```

These roles exist only in the tenant's Casbin domain and do not affect other tenants.

### 16.1.3. Role Hierarchy — How Permissions Accumulate

Casbin supports role inheritance via `g` (grouping) assertions:

```
g, role:tenant.admin, role:finance_manager, tenant_uuid
g, role:finance_manager, role:sales_rep, tenant_uuid
```

`tenant.admin` inherits all permissions of `finance_manager`, which inherits all permissions of `sales_rep`. A user assigned `tenant.admin` has all permissions from all inherited roles.

### 16.1.4. The Superuser Role — What It Bypasses

`role:platform-admin` bypasses Casbin enforcement entirely in the authz service:

```go
func (s *authzService) Enforce(ctx context.Context, sub, dom, obj, act string) (bool, error) {
    a := actor.FromContext(ctx)
    // Platform admin bypasses all checks
    if a.Type == actor.Platform {
        return true, nil
    }
    return s.enforcer.Enforce(sub, dom, obj, act)
}
```

`platform-admin` cannot bypass:
- Privacy policies (tenant data isolation)
- The `Sensitive` field redaction (unless they also hold `invoice:view_sensitive`)
- Audit logging (all actions are logged regardless of role)

---

## 16.2. Permission Model

### 16.2.1. Permission Levels

| Level | HTTP methods | ERP meaning |
|---|---|---|
| `read` | GET | View records |
| `create` | POST | Create new records |
| `write` | PUT, PATCH | Edit existing records |
| `delete` | DELETE | Delete or soft-delete records |
| `submit` | POST /{id}/submit | Finalise a document |
| `cancel` | POST /{id}/cancel | Cancel a submitted document |
| `amend` | POST /{id}/amend | Create an amendment |

### 16.2.2. Per-EntityDefinition Permission Matrix

Casbin policy format: `p, subject, domain, object, action`:

```
p, role:finance_manager, tenant_uuid, Invoice, read
p, role:finance_manager, tenant_uuid, Invoice, create
p, role:finance_manager, tenant_uuid, Invoice, write
p, role:finance_manager, tenant_uuid, Invoice, submit
p, role:sales_rep,       tenant_uuid, Invoice, read
p, role:sales_rep,       tenant_uuid, Invoice, create
```

### 16.2.3. Field-Level Permissions — Per-Role Field Restrictions

Some roles can see a form but should not see certain fields (e.g. a collection agent sees invoice total but not the customer's bank account number):

```go
type FieldPermission struct {
    EntityName  string
    FieldName   string
    Roles       []string
    Mode        string  // "hidden" | "readonly" | "write"
}
```

Field permissions are evaluated in the SDUI page builder and in the `Privacy Policy` field visibility rules (Chapter 9).

### 16.2.4. Action Permissions — Beyond CRUD

Custom actions (e.g. `approve`, `dispatch`, `refund`) have their own permission entries:

```
p, role:finance_manager, tenant_uuid, Invoice:approve, execute
p, role:logistics_manager, tenant_uuid, SalesOrder:dispatch, execute
```

These are checked via `authzService.Enforce(ctx, subject, domain, "Invoice:approve", "execute")`.

---

## 16.3. Permission Evaluation

### 16.3.1. `RequirePermission` Middleware

Applied to individual routes:

```go
api.Post("/invoices/:id/submit",
    middleware.RequirePermission("Invoice", "submit"),
    handlers.SubmitInvoice,
)
```

The middleware resolves the actor's subject and the tenant domain from context, then calls Casbin:

```go
func RequirePermission(entity, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        actorID := c.Locals("actor_id").(uuid.UUID)
        tenantID := middleware.GetTenantID(c)

        subject := "user:" + actorID.String()
        domain := tenantID.String()

        allowed, err := authzService.Enforce(c.UserContext(), subject, domain, entity, action)
        if err != nil {
            return err
        }
        if !allowed {
            return c.Status(403).JSON(ForbiddenError(entity, action))
        }
        return c.Next()
    }
}
```

### 16.3.2. Permission Resolution Order

For a given `(user, tenant, entity, action)` tuple:
1. Check user-specific overrides (rare, for temporary access grants)
2. Check role-based permissions (via role inheritance chain)
3. Check tenant defaults
4. Deny if none match

### 16.3.3. Merging Permissions From Multiple Roles — Union

A user with multiple roles receives the **union** of all their roles' permissions:

```
User has: role:sales_rep, role:collection_agent
sales_rep can: read Invoice, create Invoice
collection_agent can: read Invoice, update Invoice (assigned only, enforced by privacy policy)
→ Union: read, create, update Invoice
```

There is no "intersection" mode in Awo RBAC. If a user is given a role, they get all of that role's permissions.

### 16.3.4. Permission Caching — Per-Request Cache

The SyncedEnforcer caches policies in memory and reloads every 30 seconds from PostgreSQL. Additionally, Redis pub/sub notifies all instances when a policy changes:

```go
// When a role assignment changes:
redis.Publish(ctx, "casbin:policy_changed", tenantID.String())

// Each instance listens and invalidates its in-memory cache:
go func() {
    sub := redis.Subscribe(ctx, "casbin:policy_changed")
    for msg := range sub.Channel() {
        enforcer.LoadPolicy()
    }
}()
```

This ensures permission changes propagate across all application instances within milliseconds, not 30-second reload cycles.

---

## 16.4. Surfacing Permissions in the SDUI Layer

### 16.4.1. The `Permissions` Struct Passed to Every Page Builder

```go
type Permissions struct {
    CanView   func(entity string) bool
    CanCreate func(entity string) bool
    CanEdit   func(entity string) bool
    CanDelete func(entity string) bool
    CanSubmit func(entity string) bool
    CanCancel func(entity string) bool
    CanExecute func(entity, action string) bool
}

func buildPermissions(ctx context.Context, tenantID uuid.UUID, authz AuthzService) *Permissions {
    check := func(entity, action string) bool {
        allowed, _ := authz.Enforce(ctx, subjectFromCtx(ctx), tenantID.String(), entity, action)
        return allowed
    }
    return &Permissions{
        CanView:    func(e string) bool { return check(e, "read") },
        CanCreate:  func(e string) bool { return check(e, "create") },
        CanEdit:    func(e string) bool { return check(e, "write") },
        CanDelete:  func(e string) bool { return check(e, "delete") },
        CanSubmit:  func(e string) bool { return check(e, "submit") },
        CanCancel:  func(e string) bool { return check(e, "cancel") },
        CanExecute: func(e, a string) bool { return check(e+":"+a, "execute") },
    }
}
```

### 16.4.2. Conditional Visibility in amis JSON

```go
// In a page builder:
toolbar := amis.NewToolbar()
if perms.CanCreate("Invoice") {
    toolbar.Add(amis.NewButton("New Invoice").
        ActionType("link").
        Href("/invoices/new"))
}
if perms.CanSubmit("Invoice") {
    toolbar.Add(amis.NewButton("Submit").
        ActionType("ajax").
        API("POST /api/v1/invoices/${id}/submit"))
}
```

Buttons and sections that the user cannot use are simply not included in the generated amis JSON — the client never knows they exist.

### 16.4.3. Hiding Sections vs Disabling Fields

**Hide sections** when the user cannot access the data at all (no `read` permission).
**Disable fields** when the user can see the data but not change it (`read` but no `write`).

Never rely on disabled UI fields for security. The API must enforce permissions independently — a determined user can always bypass the UI and call the API directly.

---

## 16.5. Auditing Permission Checks

### 16.5.1. What Is Logged

Every Casbin enforcement decision is logged asynchronously:

```go
type PermissionAuditEntry struct {
    Timestamp   time.Time `json:"ts"`
    RequestID   string    `json:"request_id"`
    TenantID    uuid.UUID `json:"tenant_id"`
    ActorID     uuid.UUID `json:"actor_id"`
    Entity      string    `json:"entity"`
    Action      string    `json:"action"`
    Allowed     bool      `json:"allowed"`
    Roles       []string  `json:"roles"`
}
```

### 16.5.2. Denied Access Log — Security Monitoring

Denied access attempts are logged at `WARN` level and fed into the security event stream. Alert on:
- >10 denied access attempts per user per minute (possible privilege escalation attempt)
- Access denied for `platform-admin` role (should never happen — indicates misconfiguration)
- Access attempts to tenant B from a user authenticated to tenant A

### 16.5.3. Permission Audit Report

The `GET /api/v1/admin/permissions/audit` endpoint (platform-admin only) returns a report of all permission checks in a time window, used for compliance audits and access reviews.
