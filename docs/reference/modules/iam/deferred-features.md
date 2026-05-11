[<-- Back to Index](README.md)

## Deferred Features — Not Active in v1.0

> This document is the authoritative list of IAM/access features that are **designed but not
> implemented** in v1.0. No code in this list executes at runtime.
>
> These are not abandoned — they are explicitly reserved for v2.0 and later. The schema
> migrations are deployed (to avoid future downtime migrations) but no service code reads
> or writes these tables during normal operation.

---

### ABAC — Attribute-Based Access Control

**Code location**: `internal/core/access/`

**Gate**: Every `.go` file in `internal/core/access/` carries `//go:build ignore`. The package does not compile into the binary.

**What was designed**:
- `Service` interface (in `internal/core/access/service.go`): access request management, approval workflows, conditional access evaluation, direct permission grant/revoke.
- `conditional` sub-package: `ConditionalAccessPolicy` evaluation by time, location, device trust, risk score.
- Types for `DeviceInfo`, `LocationInfo`, `SecurityContext`, `TimeContext`, `BusinessHours`.

**Why deferred**: ABAC requires a runtime attribute collection pipeline (device fingerprinting, geo-lookup, risk scoring) that does not exist in v1.0. RBAC via Casbin covers all v1.0 authorization needs without this complexity.

**Target version**: v2.0

---

### Access Requests and Approval Workflows

**What was designed**:
- `CreateAccessRequest` — users request temporary or permanent access to a resource.
- `ProcessAccessRequest` — approvers approve or reject with comments and conditions.
- `ApprovalWorkflow` — multi-step approval chains with timeouts, required approver counts, and step conditions.
- `ListAccessRequests` — requestor and approver views.

**Why deferred**: Requires a notification system (email, in-app), a task queue for timeout handling, and UI pages for approval inbox. None of these exist in v1.0.

**Target version**: v2.0

---

### Conditional Access Policies

**What was designed**:
- `EvaluateConditionalAccess(req)` — evaluates a policy set against an `AccessContext` (IP address, device info, location, security context, time context).
- `CreateConditionalAccessPolicy` / `UpdateConditionalAccessPolicy` — admin management of policy rules.
- Policy conditions: time of day, day of week, business hours, country, IP range, device trust level, risk score threshold.

**Why deferred**: Requires runtime attribute collection at request time — GeoIP lookup, device fingerprinting, risk scoring engine. Not in scope for v1.0.

**Target version**: v2.0

---

### Direct Permission Grant/Revoke (non-RBAC)

**What was designed**:
- `GrantPermission(req)` — assign a specific resource/action permission directly to a user, bypassing roles.
- `RevokePermission(req)` — remove a direct permission grant.
- `ListUserPermissions(userID, entityID)` — enumerate direct grants for a user.

**Why deferred**: In v1.0, all permissions flow through Casbin roles (RBAC). Direct user-level grants would complicate the enforcement model without clear benefit at this stage.

**Target version**: v2.0

---

### SAML 2.0 SSO

**Current state**: OAuth 2.0 / OIDC (Google, Microsoft) is implemented. SAML 2.0 is not.

**Why deferred**: SAML requires XML signature validation, IdP metadata fetching, and assertion parsing libraries that add significant dependency weight. Enterprise customers requiring SAML can use an IdP proxy (e.g. Keycloak) in front of the OIDC endpoint.

**Target version**: v2.0

---

### SMS / Email OTP MFA

**Current state**: TOTP (RFC 6238) via authenticator app is implemented. SMS and email OTP are not.

**Why deferred**: SMS OTP requires a telecom gateway integration (Twilio, etc.) and is vulnerable to SIM-swapping. Email OTP requires an email delivery system. TOTP is sufficient for v1.0 and more secure than SMS.

**Target version**: v2.0 (email OTP), v3.0 (SMS, if ever)

---

### Roles Table (ABAC Role Model)

**Migration**: `000405_auth_create_roles.up.sql` — the `roles` table is deployed.

**What is deployed but not used**:
- `roles` table with `is_system_role`, `role_type` (SYSTEM/TENANT/ENTITY/CUSTOM/FUNCTIONAL), `parent_role_id`, `conditions JSONB` (time/location/device rules), `permissions JSONB` (cached permissions).
- Migrations 000406 (`role_permissions` mapping) and 000407 (`user_role_assignments` linking table).

**Current v1.0 state**: Role names are plain strings in Casbin g-rules. The `roles` table is populated by seed migrations but is not read or written by any service code in normal operation.

**Target version**: v2.0 — the `roles` table becomes the canonical role registry when ABAC is activated.

---

### Policy Count Limit (DoS Prevention)

**Design**: `AddPolicy()` should check the current policy count for a domain and reject if over a configurable limit (default: 10,000).

**Current state**: Not implemented. The `AuthzConfig` struct has no `MaxPoliciesPerDomain` field. No count check exists in `AddPolicy()`.

**Reference**: `tasks.md` AUTHZ-5.

**Target version**: v1.0 blocker — should be implemented before production exposure of the policy management API.

---

### Service-Layer Platform Domain Write Guard

**Design**: `AddPolicy()` and `AssignRole()` should reject calls that would write to `_platform_` domain from a non-platform actor.

**Current state**: Not implemented at the service layer. The guard is expected to be enforced by middleware (Casbin enforcement on the policy management endpoint), but there is no defence-in-depth at the service layer itself.

**Reference**: `tasks.md` AUTHZ-4.

**Target version**: v1.0 — should be implemented before production.

---

### Security Event Audit Log

**Design**: Every `AssignRole`, `RevokeRole`, `AddPolicy`, `RemovePolicy` should emit a structured audit event persisted to an audit log table (not just OTel spans).

**Current state**: OTel span attributes capture the operation context, but spans are not persisted as audit records. The `role_assignments` table provides partial auditability for role changes.

**Reference**: `tasks.md` AUTHZ-7.

**Target version**: v1.0 — should be in place before exposing role/policy management to tenant admins.

---

### Redis Pub/Sub for Sub-Second Policy Propagation

**Current state**: `StartAutoLoadPolicy(30s)` is active — all instances converge within 30 seconds.

**Design**: A Redis pub/sub channel (`authz:policy-changed:{tenantID}`) would allow the instance executing a policy mutation to notify all other instances to reload immediately, reducing the propagation window from 30 seconds to < 1 second.

**Target version**: v2.0 — acceptable for v1.0 where 30-second drift is tolerable.

---

Next: [RBAC Enforcement](./rbac-enforcement.md)
