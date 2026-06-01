---
title: Authorization Model
portal: 3 — Platform Architecture
section: 02-iam
audience: [architect, backend-engineer, tech-lead]
related:
  - "[IAM Overview](01-iam-overview.md)"
  - "[Session Architecture](02-session-architecture.md)"
  - "[Service-Level Authorization](../../04-backend-engineering/00-module-development-guide/06-service-layer/04-authorization.md)"
---

# Authorization Model

AwoERP uses Casbin v2 with a domain-scoped RBAC model. Policies live in the database and are loaded into Casbin at startup.

## Casbin Model

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom \
    && keyMatch2(r.obj, p.obj) \
    && keyMatch(r.act, p.act)
```

- `sub`: principal subject — `role:{name}` or `user:{uuid}`
- `dom`: tenant domain — `tenant:{uuid}`
- `obj`: resource — permission string or resource path
- `act`: action — `allow` (route-level) or specific action
- `eft`: effect — `allow` or `deny`

## Two Authorization Styles

### Route-Level (middleware)

Checks a permission string against `allow`:

```
sub=role:contracts_editor, dom=tenant:acme, obj=contracts.contract.create, act=allow
```

Policy:
```
p, role:contracts_editor, tenant:acme, contracts.contract.create, allow, allow
```

### Service-Level (resource instance)

Checks a specific resource path:

```
sub=role:contracts_viewer, dom=tenant:acme, obj=tenants/acme/contracts/*, act=read
```

Policy:
```
p, role:contracts_viewer, tenant:acme, tenants/*/contracts/*, read, allow
```

`keyMatch2` enables wildcard matching on `obj`.

## Role Naming Convention

```
{module}.{role_name}
```

Examples:
```
contracts.viewer
contracts.editor
contracts.reviewer
contracts.admin
finance.accountant
finance.admin
hr.manager
```

Platform-wide roles:
```
platform.superadmin    — all tenants, all actions (ops only)
tenant.admin           — all actions within a tenant
```

## Policy Management

Policies are stored in the `policies` table and loaded into Casbin's in-memory enforcer at startup:

```go
// Load all policies for a tenant on first enforce call
enforcer.LoadFilteredPolicy(filter.NewFilter(
    []string{"", tenantID.String()},
))
```

Policy changes (add role, grant permission) update the database **and** call `enforcer.LoadPolicy()` to refresh the in-memory model without restart.

## Deny Policy (Explicit Deny)

Casbin supports `deny` effect. An explicit deny overrides any allow:

```
p, user:blocked-user-uuid, tenant:acme, contracts.*, allow, deny
```

This allows temporary blocking of specific users without revoking their roles.

## Permission Check in Practice

```go
// Route middleware — coarse check
allowed, err := authzSvc.Enforce(ctx, iam.Request{
    Subject: session.ToPrincipal().Subject,
    Domain:  session.ToPrincipal().Domain,
    Object:  "contracts.contract.create",
    Action:  "allow",
})

// Service layer — fine-grained instance check
allowed, err := authzSvc.Enforce(ctx, iam.Request{
    Subject: principal.Subject,
    Domain:  principal.Domain,
    Object:  fmt.Sprintf("tenants/%s/contracts/%s", tenantID, contractID),
    Action:  "update",
})
```
