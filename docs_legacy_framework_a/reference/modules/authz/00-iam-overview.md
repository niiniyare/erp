> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## IAM Module — Philosophy & Scope

### The Three Questions Every Request Answers

Every HTTP request that enters Awo ERP answers three questions before touching business data:

1. **Authentication (AuthN): Who are you?** Verify the claimed identity. Produce a trusted session.
2. **Authorization (AuthZ): What can you do?** Given your identity, decide allow or deny for this operation.
3. **Configuration: How should this behave?** Given your tenant's flags and settings, which features are on and how do they operate?

These are distinct concerns with distinct mechanisms, but they all resolve at the same moment — **login** — and travel together in the session object. After login, every answer to all three questions is an in-memory lookup with no database hit.

---

### The Unified Key Namespace

The most important architectural decision in this module is that **permissions, feature flags, and settings all share the same dot-notation key hierarchy**:

```
{module}.{resource}.{action}     →  permission key (leaf level)
{module}.{resource}              →  resource-level flag or setting key
{module}                         →  module-level flag key

finance                                      module flag: is Finance on?
finance.transactions                         resource flag: are Transactions enabled?
finance.transactions.approval_workflow       setting: is approval workflow active?
finance.transactions.approval_threshold      setting: approval amount threshold
finance.transactions.approve                 permission: can this user approve?
```

The Module/Resource/Action (MRA) tables are the single source of truth that anchors all three systems. You do not write new key strings in Go code for flags or settings — you derive them from the MRA slugs.

- Adding a new module seeds its flag definition automatically (DB trigger).
- Adding a new action seeds its permission key automatically.

The key namespace is shared; the tables are not. A flag is binary (feature exists or not). A setting is a value (how the feature behaves). Different resolution logic, different UI controls, different permission requirements to change.

---

### The Trust Chain

```
Browser / API Client
      │
      ▼
Fiber HTTP Server
      ├── Recovery, Logger, RateLimit, CORS, SecurityHeaders
      │
      ├── Authenticate()              AuthN: session cookie or Bearer token
      │     → loads ResolvedSession into context
      │     → ResolvedSession carries: identity + permissions + flags + settings + entity scope
      │
      ├── SetDBPool()                 pool selection from session.UserType
      │     platform user  → admin_role pool (BYPASSRLS)
      │     all others     → application_role pool (RLS active, per-connection settings)
      │
      ├── RequireTenantContext()      validates X-Tenant-ID header
      │
      ├── RequirePermission()         O(1) map lookup from session — zero DB hit
      ├── RequireFlag()               O(1) map lookup from session — zero DB hit
      │
      └── Handler                    business logic; reads only from session and request
```

---

### What This Module Owns

**Owns:**
- User lifecycle (create, invite, suspend, deactivate)
- Authentication: login, logout, MFA (TOTP), OAuth/OIDC/SAML, passwords, API keys
- Session management with pre-computed permissions/flags/settings
- RBAC with Casbin (roles, policies, temporal assignments, deny-override)
- The MRA registry (modules, resources, actions tables)
- Feature flag catalogue and per-tenant flag values
- Tenant settings catalogue and per-tenant values
- User preferences
- UI navigation generation (BootService)
- Entity hierarchy resolution (EntityScope)

**Does not own:** HTTP routing, DB connection pools, notification delivery, audit log persistence, tenant provisioning (responds to it via events), business module schemas.

---

### The Platform Facade

Every business module receives one `*platform.Platform`. It never imports individual platform files:

```go
// internal/platform/service.go

type Platform struct {
    IAM      *IAMService
    Tenant   *TenantService
    Flags    *FlagService
    Settings *SettingService
    Boot     *BootService
    Audit    *AuditService
    Notify   *NotifyService
}

// Usage in any business module:
type FinanceService struct {
    platform *platform.Platform
    repo     FinanceRepository
}

func (s *FinanceService) PostTransaction(ctx context.Context,
    params PostTransactionParams) (*Transaction, error) {

    session := domain.SessionFromContext(ctx)
    if !session.Can("finance.transactions", "post") {
        return nil, domain.ErrForbidden
    }
    // ...
}
```

---

Next: [Code Architecture & Conventions](./00b-code-architecture.md)
