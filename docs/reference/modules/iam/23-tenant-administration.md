[<-- Back to Index](README.md)

## Tenant Administration — Operational Guide

> **Implementation status**: This document describes the v1.0 operational model as implemented.
> Features labeled **[PLANNED - NOT IN v1.0]** are defined in `deferred-features.md`.

---

### Who Is This For?

This guide is for **tenant administrators** — the people responsible for managing users, roles, and access within a single organisation (tenant) in Awo ERP. It also serves platform engineers and product managers who need to understand what a tenant admin can and cannot do.

---

### 1. Tenant Bootstrap

#### What is `BootstrapTenantAdmin`?

When the very first user registers for a new tenant, the system calls `BootstrapTenantAdmin` from the `AuthzService`. This function (see `internal/core/iam/service/authz.go`) performs two idempotent steps:

1. **Seeds a wildcard allow policy** for the `tenant_admin` role in the new tenant's Casbin domain:
   ```
   Policy: subject=tenant_admin, domain=<tenantID>, object=*, action=*, effect=allow
   ```
   This means the `tenant_admin` role has unrestricted access to all resources within the tenant.

2. **Assigns the `tenant_admin` role** to the registering user in the tenant's Casbin domain.

The function is safe to call repeatedly — duplicate policies and role assignments are silently skipped (`ErrPolicyConflict` is ignored).

#### Additional Role Seeding: `SeedDefaultRoles`

Separately, `SeedDefaultRoles` (see `internal/core/iam/seed.go`) seeds explicit permission policies for the `role:tenant.admin` named role in a new tenant. This covers:

- Finance module: invoices, bills, accounts, transactions (read/create/update/delete/approve/export)
- People module: employees, persons (read/create/update/delete)
- Settings module: iam and general settings (read/update)

Both functions work together: `BootstrapTenantAdmin` creates the user–role link; `SeedDefaultRoles` populates the role's permission policies.

#### When Does Bootstrap Run?

Bootstrap runs **only** during `RegisterNewUser` — the first user registration flow for a new tenant. It does **not** run on subsequent logins, session refreshes, or user updates.

This is a deliberate design decision: bootstrap is a one-time provisioning event, not a session-time concern. See `tasks.md` (BLOCK-6) for historical context.

#### What If No Admin Exists?

If bootstrap is skipped or fails and no `tenant_admin` assignment exists:
- The tenant has no user with administrative access.
- No users can manage IAM settings (roles, policies, user assignments) because those endpoints require the `settings.iam update` permission.
- Recovery requires platform-level intervention: a platform operator must manually seed the bootstrap policies via the platform admin API or direct DB access (admin_role).

---

### 2. What Tenant Admins Can Do

A user holding the `tenant_admin` role (or `role:tenant.admin`) can:

**User Management**
- Create new users within their tenant
- Update user profiles, display names, email addresses
- Deactivate or delete users in their tenant
- Reset user passwords
- Assign and revoke roles for users in their tenant

**Role Management**
- Create custom tenant roles (see Section 3)
- Assign roles to users with optional entity scope and expiry
- Revoke role assignments

**Settings and Configuration**
- Read and update `settings.iam` — IAM-related settings (session TTL, MFA requirements)
- Read and update `settings.general` — general tenant settings (locale, timezone, etc.)

**Tenant Feature Flags**
- View feature flag state for the tenant
- Override feature flags within platform-permitted bounds (flags with `is_system=false`)
- Cannot toggle `is_system=true` flags — these are platform-controlled

**What Tenant Admins Cannot Do**
- Create or manage platform users (user_type=SYSADMIN/PLATFORM)
- Modify system roles (`is_system_role=true`) — see Section 3
- Access data belonging to other tenants
- Create policies in the `_platform_` Casbin domain
- Grant themselves or others platform-level authority

Cross-tenant operations are impossible by architectural design: Casbin's `r.dom == p.dom` matcher, combined with PostgreSQL RLS (`tenant_id = current_tenant_id()`), means tenant admin actions are always scoped to their own tenant UUID. See [Security Considerations](./17-security-considerations.md) — T2.

---

### 3. Custom Tenant Roles

#### Creating a Custom Role

Tenant admins can create roles with `role_type = 'CUSTOM'` or `'TENANT'` within their tenant. The `roles` table (migration `000405_auth_create_roles.up.sql`) stores role definitions.

Required fields:
- `name` — unique within the tenant (enforced by `UNIQUE (tenant_id, name)` constraint)
- `tenant_id` — automatically set from session context
- `entity_id` — the entity scope context for the role
- `role_type` — `CUSTOM` or `TENANT` for tenant-defined roles

Optional fields:
- `display_name` — human-readable label
- `description` — free-text explanation
- `parent_role_id` — the parent role in the hierarchy (for inheritance)
- `module_id` — which module this role belongs to

**Note**: The `conditions` JSONB column (time/location/device rules) and `entity_scope` JSONB on the roles table are schema fields reserved for v2.0 ABAC features. They exist in the DB but are not read at runtime in v1.0. Casbin policies are the sole authorization mechanism.

#### Role Hierarchy and Inheritance

The `parent_role_id` field on the `roles` table defines a self-referencing hierarchy. Casbin represents inheritance as g-rules (grouping policies):

```
g(role:accountant, role:finance_viewer, tenantID)
```

This means `role:accountant` inherits all permissions granted to `role:finance_viewer`. When Casbin evaluates a request, it traverses the role graph to check all effective permissions.

Use `GetImplicitRoles()` to see all inherited roles for a subject, including those acquired through the hierarchy.

#### Tenant Isolation of Role Hierarchy

Role inheritance is strictly domain-scoped. A g-rule for `role:A` inheriting from `role:B` in `tenantDomain_X` does not apply in `tenantDomain_Y`. Cross-tenant inheritance is structurally impossible: the Casbin matcher `g(r.sub, p.sub, r.dom)` uses `r.dom` as the scoping context.

#### System Roles Are Immutable

Roles with `is_system_role = true` are seeded by the platform and protected:
- The schema CHECK constraint allows `role_type` values of `SYSTEM`, `TENANT`, `ENTITY`, `CUSTOM`, or `FUNCTIONAL`.
- System roles (`role_type = 'SYSTEM'`) should never be modified via the tenant admin API.
- Casbin policies for system roles are seeded during tenant provisioning and should only be modified by platform operators.

#### Naming Guidance and Least-Privilege

Follow these conventions:
- Use `role:` prefix in Casbin role names (e.g., `role:accounts_payable_clerk`)
- Name roles after job functions, not after people
- Assign the minimum permissions required for the job function
- Prefer explicit permission grants over wildcards
- Reserve wildcard `*` actions only for the `tenant_admin` role
- Create separate roles for read-only vs read-write access to the same resource

---

### 4. User Lifecycle Management

#### Creating a User and Assigning a Role

The standard flow for provisioning a new employee:

1. **Create the user**: `UserService.RegisterNewUser()` — sets `user_type`, `email`, `username`, `tenant_id`
2. **Assign a role**: `AuthzService.AssignRole(tenantID, subject, role, domain)` — creates the Casbin g-rule and a `role_assignments` row
3. **Set entity scope**: Pass an `entity_id` and scope type as part of role assignment context

The `user_roles` table (migration `000407_auth_assign_user_roles.up.sql`) records role assignments at the application layer. The Casbin `role_assignments` table (managed by the authz repository) is the enforcement source of truth.

#### Activating and Deactivating a User

User deactivation is handled at the `users` table level (an `is_active` or similar flag). Deactivated users cannot create new sessions. For immediate session termination, call `SessionService.LogoutAllForUser()` which invalidates all active DB sessions for the user.

For a complete access suspension, combine:
1. Deactivate the user record
2. `AuthzService.RevokeRole()` for all role assignments — this also calls `InvalidateByUser()`, evicting Redis session cache immediately

#### Role Assignment Flow

When `AssignRole` is called:
1. Casbin `AddGroupingPolicy(subject, role, domain)` is called — the in-memory enforcer is updated immediately
2. A `role_assignments` row is upserted — stores audit metadata (assigned_by, delegated_by, expires_at)
3. On the next request from that user, the new role is effective

When `RevokeRole` is called:
1. Casbin `DeleteRoleForUserInDomain(subject, role, domain)` — in-memory update is immediate
2. The `role_assignments` row is deactivated in DB
3. `SessionInvalidator.InvalidateByUser()` is called — Redis cache entries for the user are evicted
4. The user's next request will see the revocation

The window between role revocation and enforcement is effectively zero for the current server instance. Other instances converge within 30 seconds via `StartAutoLoadPolicy`.

#### Temporal Role Assignments

Role assignments can be time-limited using `WithExpiry(t time.Time)`:
```go
authzSvc.AssignRole(ctx, tenantID, subject, role, domain, iam.WithExpiry(expiresAt))
```

Expired roles are lazily cleaned up on the next `Enforce()` call for that subject via `revokeExpiredRoles()`. This is a non-blocking, best-effort cleanup — enforcement still denies expired roles because the Casbin model evaluates `IsEffective()`.

#### Entity Scope Options

See the dedicated guide: [Entity Scope](./25-user-entity-scope.md).

Summary of scope types:
- `EntityScopeAll` (`"all"`) — access to all entities in the tenant; typically for CFO, HR Director, platform admins
- `EntityScopeSubtree` (`"subtree"`) — access to the user's entity and all its descendants; for branch managers, regional supervisors
- `EntityScopeEntity` (`"entity"`) — access only to the user's directly assigned entity; for cashiers, station attendants, front-line workers

---

### 5. Separation of Authorities

#### Platform Authority vs Tenant Authority

| Concern | Platform Authority | Tenant Authority |
|---|---|---|
| Create platform users | Platform super admin only | Cannot |
| Provision new tenants | Platform operators | Cannot |
| Modify system roles | Platform operators | Cannot |
| Modify tenant custom roles | Cannot (per-tenant) | Tenant admin |
| Create tenant users | Cannot (per-tenant) | Tenant admin |
| Toggle `is_system` feature flags | Platform operators | Cannot |
| Toggle non-system feature flags | Platform operators can set bounds | Tenant admin within bounds |
| Access other tenants' data | Platform operators (support roles only) | Cannot |
| Manage Casbin policies in own tenant | Cannot (per-tenant) | Tenant admin |
| Manage Casbin policies in `_platform_` | Platform super admin only | Cannot |

#### Why Cross-Tenant Operations Are Impossible

Three independent layers enforce tenant boundaries:

1. **Casbin domain match** (`r.dom == p.dom`): A user in tenant domain `A` cannot trigger policies scoped to domain `B`, even if the Casbin rule names are identical. Domain comes from the authenticated session, never from user input.

2. **PostgreSQL RLS**: Every table with `tenant_id` enforces `tenant_id = current_tenant_id()`. A Go bug cannot return another tenant's data; the DB silently filters it out.

3. **Session construction**: `ToPrincipal()` derives the Casbin domain from the session's `TenantID` field, which is set at authentication time from the authenticated user's DB record, not from any request parameter.

---

See also:
- [Platform Administration](./24-platform-administration.md)
- [Entity Scope](./25-user-entity-scope.md)
- [Role Management](./07-role-management.md)
- [Security Considerations](./17-security-considerations.md)
