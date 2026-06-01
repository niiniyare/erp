---
title: Authorization Guide
portal: 7 — Security
section: 07-security
audience: [backend-engineer, security, architect]
related:
  - "[Security Overview](01-security-overview.md)"
  - "[Authentication Flows](02-authentication-flows.md)"
  - "[Authorization Model](../03-platform-architecture/02-iam/03-authorization-model.md)"
  - "[Authz Middleware](../04-backend-engineering/00-module-development-guide/16-middleware-chain/03-authz-middleware.md)"
---

# Authorization Guide

## Two Authorization Layers

Every protected operation goes through two checks:

```
HTTP Request
  │
  ▼
[Authenticate middleware]     ← Is the token valid? Who is the user?
  │
  ▼
[Authorize middleware]        ← Does the user have permission for this route?
  │
  ▼
[Service method]              ← Can the user perform this action on this specific record?
  │
  ▼
[Database RLS]                ← Is this data in the user's tenant? (PostgreSQL enforces)
```

Route-level and service-level checks are complementary — route-level guards the endpoint, service-level guards business logic with full context.

## Permission String Format

```
{module}.{resource}.{action}

Examples:
  contracts.contract.create
  contracts.contract.approve
  finance.account.read
  iam.user.assign_roles
  platform.tenant.create
```

All permission strings are lowercase snake_case. No wildcards in runtime checks — always check a specific permission string.

## Route-Level Authorization

```go
// In route registration
api.Post("/contracts", authz.Authorize(authzSvc, "contracts.contract.create"), handler.Create)
api.Post("/contracts/:id/approve", authz.Authorize(authzSvc, "contracts.contract.approve"), handler.Approve)
```

The middleware calls `authzSvc.Can(ctx, principal, permission)`. Returns 403 if denied, continues if allowed.

## Service-Level Authorization

For operations requiring additional context (ownership, state checks):

```go
func (s *ContractService) Submit(ctx context.Context, sess ResolvedSession, id uuid.UUID, version int) error {
    // Check permission
    if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.submit"); err != nil {
        return ErrForbidden
    }

    // Check record state
    contract, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return err
    }
    if contract.Status != StatusDraft {
        return ErrContractNotEditable
    }
    // ...
}
```

## Casbin Policy Structure

AwoERP uses Casbin v2 with model `(sub, dom, obj, act, eft)`:

```
p = sub, dom, obj, act, eft

// Role has permission in tenant domain
p, contracts.editor, {tenant_id}, contracts.contract.create, allow
p, contracts.editor, {tenant_id}, contracts.contract.update, allow
p, contracts.editor, {tenant_id}, contracts.contract.submit, allow

// Deny policy (overrides allow when both match)
p, contracts.viewer, {tenant_id}, contracts.contract.create, deny

// User has role in domain
g, {user_id}, contracts.editor, {tenant_id}
```

`keyMatch2` is used for `obj` matching — supports `:param` style patterns for future resource-level policies.

## Fail Closed

`authzSvc.Can()` returns error (deny) on:
- Casbin enforcement error
- Missing policy
- Unknown permission

Never default to "allow" on error. If Casbin is unavailable → 403, not 200.

## Role Naming Convention

```
{module}.{verb}

Examples:
  contracts.viewer    ← read-only access to contracts
  contracts.editor    ← create, update, submit contracts
  contracts.approver  ← read + approve contracts
  finance.viewer      ← read-only finance
  finance.manager     ← full finance access
  iam.admin           ← manage users and roles
  platform.admin      ← platform-level access (superuser)
```

Custom roles created by tenants follow the same pattern but scoped under the tenant's module allocation.

## EntityScope Enforcement

EntityScope restricts which entities (org units) a user can see:

| Scope | What user sees |
|-------|----------------|
| `all` | All data in tenant |
| `subtree:{entity_id}` | Data in entity + all descendants |
| `entity:{entity_id}` | Data in that specific entity only |

The service layer applies EntityScope **after** permission check:

```go
func (s *ContractService) List(ctx context.Context, sess ResolvedSession, params ListParams) ([]Contract, error) {
    // 1. Permission check
    if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.read"); err != nil {
        return nil, ErrForbidden
    }

    // 2. Apply entity scope
    entityID := sess.EntityID()     // nil if scope is "all"
    contracts, err := s.repo.List(ctx, ListRepoParams{
        TenantID:        sess.TenantID,
        EntityID:        entityID,
        EntityScopeType: sess.EntityScope.Type,
    })
    // ...
}
```

## What NOT To Do

```go
// ❌ Checking roles directly — brittle, coupling to role names
if sess.HasRole("contracts.editor") { ... }    // HasRole doesn't exist on ResolvedSession

// ❌ Skipping the authz check in a service method
func (s *ContractService) Delete(ctx context.Context, id uuid.UUID) error {
    return s.repo.Delete(ctx, id)  // No permission check!
}

// ❌ Returning 200 on authz service failure
if err := authzSvc.Can(ctx, principal, perm); err != nil {
    // don't log and continue — return 403
}

// ✅ Always return ErrForbidden on deny
if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.approve"); err != nil {
    return ErrForbidden
}
```

## Audit Log on Permission Denial

All permission denials are logged automatically by the `Authorize` middleware:

```json
{
  "level": "warn",
  "msg": "permission denied",
  "user_id": "...",
  "tenant_id": "...",
  "permission": "contracts.contract.approve",
  "path": "/api/v1/contracts/123/approve",
  "request_id": "..."
}
```

For service-level denials, log explicitly:

```go
s.logger.WarnContext(ctx, "permission denied",
    "user_id", sess.UserID,
    "permission", "contracts.contract.submit",
    "contract_id", id,
)
```
