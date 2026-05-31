# IAM Module Audit Review

**Document Classification:** Internal Engineering Audit
**Audit Date:** 2026-05-10
**Review Scope:** `internal/core/iam`, `internal/core/access`, `db/migration` (IAM-related), `db/queries` (IAM-related), `db/sqlc` (IAM-related), `docs/reference/modules/iam`
**Reviewer Role:** Principal Security Architect + Staff Engineer + Product Risk
**Verdict Preview:** NOT PRODUCTION READY — multiple critical and high-severity findings block safe deployment

---

## Executive Summary

The IAM module contains a significant architectural split that represents the most dangerous risk in the codebase: **two completely separate authorization systems exist in parallel, only one is operational, and the codebase pretends they are integrated.** The Casbin-based RBAC system (`casbin_rule` + `role_assignments`) is the only live authorization engine. The entire relational ABAC layer (`permissions`, `roles`, `role_permissions`, `user_roles`, `policies`, `access_requests` tables — migrations 000404–000413) is dead schema with zero service implementation. The `internal/core/access` package defines a large service interface but every subdirectory (`approval`, `conditional`, `permission`, `request`, `execution`) contains only model files. No service is implemented.

Beyond the phantom ABAC layer, the audit found: a revoked API key remains valid for up to 5 minutes by design, forcibly logged-out users retain valid cached sessions for the entire TTL window, a JIT bootstrap silently grants tenant_admin to any roleless tenant user (privilege escalation), the Casbin enforcer in-memory state is NOT synchronized across horizontal replicas (authorization decisions diverge post-deploy), and every single role-management DB integration test is commented out.

This system will pass local testing and fail in production under adversarial conditions, under horizontal scaling, and under routine operations like key revocation and emergency user termination.

**Production Readiness Score: 3.5/10**

---

## Overall Architecture Assessment

The IAM module attempts a layered clean architecture (domain → service → repository → DB) and largely succeeds at the structural level. Package boundaries, facade pattern at `iam.go`, and interface-first design are positive signals.

However, two architectural choices threaten the entire system:

**Split Authorization Brain.** Two independent authorization models were built at different phases and never reconciled. Casbin RBAC governs actual request enforcement. The relational ABAC model (roles, permissions tables) is aspirational schema that has never been wired to enforcement. The documentation presents both as equally valid. Enterprise customers running compliance audits will discover this immediately.

**Single-Node Casbin.** The Casbin enforcer is a struct field on `authzService`. Each process instance holds its own in-memory policy graph. Policy mutations (AddPolicy, AssignRole, RevokeRole) update only the local instance. Horizontal deployment means instances diverge. This is not a future concern — it is a present architectural defect that makes horizontal scaling impossible without silent authorization inconsistency.

---

## Documentation Assessment

**Finding: Documentation describes features that do not exist.**

The documentation references conditional access, ABAC policy evaluation, approval workflows, and fine-grained permissions. None of these are implemented. A CTO or enterprise compliance officer reading the docs would conclude capabilities exist that do not. This is a liability for enterprise sales and a deception risk.

**Finding: Casbin model definition location is undocumented.**

`domain.CasbinModel` is referenced throughout but `internal/core/iam/domain/domain.go` does not exist according to the file scan. The domain package contains only `errors.go`. Where `CasbinModel` is defined is ambiguous. If it lives in an unlisted file, it is undiscoverable from documentation.

**Finding: Dual role-naming conventions are undocumented.**

`SeedDefaultRoles` uses `"role:tenant.admin"` format. `BootstrapTenantAdmin` uses `"tenant_admin"` format. Two different role name conventions exist for what appears to be the same administrative role. No documentation explains this discrepancy or whether they are the same role.

---

## Domain Model Assessment

**Finding: Domain model exposes HTTP status codes.**

`domain.Error` embeds `HTTPStatus int`. This couples the domain to the HTTP transport layer. Domain errors should be transport-agnostic. Mapping HTTP statuses in the domain violates clean architecture.

**Finding: Two disconnected role identity models.**

The domain exposes `RoleAssignment` which references roles by `string` name (e.g., `"tenant_admin"`). The database has a `roles` table with UUID primary keys, hierarchy levels, JSONB permissions cache, and module associations. These two models are completely disconnected — no service layer bridges them. Role names in Casbin rules bear no referential integrity relationship to the `roles` table.

**Finding: EntityScope computed at login is never invalidated.**

`EntityScope` is computed at login and embedded in the session. If a user's entity assignment changes after login, the session retains the old scope until expiry. There is no invalidation path for entity scope changes short of forcing user logout.

**Finding: JIT admin bootstrap violates least-privilege principle.**

```go
// service/session.go:404
if len(roles) == 0 && domain.ActorTypeFromUserType(user.UserType) == domain.ActorTenant && user.TenantID != uuid.Nil {
    if bErr := s.authz.BootstrapTenantAdmin(ctx, user.TenantID, user.ID); bErr != nil {
```

Any tenant user with zero role assignments is automatically granted `tenant_admin` at first login. This is designed as a migration aid but constitutes a privilege escalation path. A user intentionally created with no roles — for example, a deprovisioned account being kept for audit retention, or a restricted service account — will be silently elevated to tenant admin on their next login attempt.

---

## Access Control Assessment

### Finding: AssignRole is non-atomic — Casbin and DB can diverge

**Severity: Critical**
**Location:** `internal/core/iam/service/authz.go:236–271`

```go
if _, err := s.enforcer.AddGroupingPolicy(subject, role, domainName); err != nil {
    // ... returns error
}
// Casbin updated. Now update the DB:
if err := s.repo.UpsertRoleAssignment(ctx, ...); err != nil {
    // DB failed — but Casbin already has the assignment in memory
    return err
}
```

**Problem:** Casbin `AddGroupingPolicy` and `UpsertRoleAssignment` are not wrapped in a distributed transaction or a compensating rollback. If `UpsertRoleAssignment` fails (DB error, connection loss), Casbin holds the role in memory. On the next `LoadPolicy` (restart, `InvalidateCache`), the role disappears — but until then, the user has unauthorized access that was never durably recorded.

**Impact:** User holds permissions not reflected in the audit log. On restart the access vanishes silently. Authorization state is non-deterministic across restarts.

**Recommendation:** Persist to DB first in a transaction. On success, update Casbin. On DB failure, do not update Casbin. Alternatively: make Casbin the write-through adapter that calls the DB synchronously on every mutation (which the pgxAdapter `AddPolicy` method does — but `AssignRole` calls `AddGroupingPolicy` which also calls the adapter, so verify the adapter's `AddPolicy` is called for g-type rows, not just p-type).

---

### Finding: RevokeRole error from Casbin is silently ignored

**Severity: High**
**Location:** `internal/core/iam/service/authz.go:297`

```go
s.enforcer.DeleteRoleForUserInDomain(subject, role, domainName)
// return value discarded — no error check
```

Casbin's `DeleteRoleForUserInDomain` can fail. The return value is discarded. If Casbin fails to remove the role from in-memory state, the DB row is marked inactive but Casbin still enforces the grant. The user retains access despite being "revoked."

**Recommendation:** Check and propagate the return value. Treat Casbin removal failure as a critical error.

---

### Finding: Casbin in-memory state does not synchronize across instances

**Severity: Critical**
**Location:** `internal/core/iam/service/authz.go` — entire service

The `authzService` holds a `*casbin.Enforcer` in a struct field. In any horizontally-scaled deployment (2+ replicas), each instance maintains its own independent policy graph. When `AssignRole`, `RevokeRole`, `AddPolicy`, or `RemovePolicy` is called on instance A:

- Instance A's enforcer is updated immediately
- Instance B's enforcer retains stale state until restart or explicit `InvalidateCache` call
- `InvalidateCache` has no broadcast mechanism — calling it on one instance does not trigger reloads on others

**Impact:** After a role revocation on any instance, the user may still receive `allowed=true` from other instances for an unbounded period. In a load-balanced deployment, ~50% of requests (across 2 instances) continue to authorize revoked users. This is a complete security failure for emergency access revocation scenarios (e.g., employee termination, credential compromise).

**Recommendation:** Register a Casbin Watcher (e.g., Redis pub/sub watcher from `casbin-go-redis-watcher`) that broadcasts policy changes to all instances. All instances must reload their enforcer on watcher notification. This is a standard Casbin deployment requirement for distributed systems.

---

### Finding: revokeExpiredRoles executes a DB query on every Enforce call

**Severity: High**
**Location:** `internal/core/iam/service/authz.go:153`, `575–588`

```go
if err := s.revokeExpiredRoles(ctx, r.Subject, r.Domain); err != nil {
    // non-fatal, continues
}
```

Every call to `Enforce` — which happens on every authenticated HTTP request — triggers `ListExpiredActiveRoleNames` as a DB query. This query hits the database even when no roles are expired, on every single request. For a system processing 1,000 req/s, this is 1,000 unnecessary DB round-trips per second per instance.

There is no caching of "this subject has no expired roles," no debounce, no circuit breaker. Under load, this query becomes a hot-path bottleneck and a DoS vector (crafted high-frequency requests from a valid session).

**Recommendation:** Cache the "no expired roles" result for the subject+domain for a short TTL (e.g., 30 seconds). Only execute the DB query on cache miss or on expiry window approach. Alternatively, move expiry enforcement to a background job that pre-cleans expired rows.

---

### Finding: buildPermissions failure silently creates empty-permission sessions

**Severity: High**
**Location:** `internal/core/iam/service/session.go:196–202`

```go
if err := s.buildPermissions(ctx, user); err != nil {
    s.log.WarnContext(ctx, "buildPermissions failed, proceeding with empty map", ...)
    perms = make(map[string]bool)
}
```

If permission computation fails (Casbin error, DB timeout), the session is created with an empty permission map. The user logs in successfully but has zero permissions. Depending on how `ResolvedSession.Can()` is implemented, this could manifest as the user seeing "access denied" on everything — or, if the empty map is interpreted as "no restrictions," as full access.

Additionally, the JIT bootstrap path (lines 404–412) catches `BootstrapTenantAdmin` failure but still proceeds with whatever roles were loaded (possibly none). The next call to `GetImplicitRoles` may return empty if the bootstrap failed, leading to the empty-permission session scenario described above.

**Recommendation:** `buildPermissions` failure should fail the login, not silently proceed. A session with empty permissions is not a safe fallback — it is an inconsistent state. Return an error to the caller and reject the login.

---

## Tenant Isolation Assessment

### Finding: casbin_rule RLS provides zero tenant isolation

**Severity: Critical**
**Location:** `db/migration/000063_authz_casbin_rule.up.sql:34`

```sql
CREATE POLICY casbin_rule_app ON casbin_rule FOR ALL TO application_role USING (TRUE) WITH CHECK (TRUE);
```

The `casbin_rule` table RLS policy for `application_role` is `USING (TRUE)` — full table access for any application-role connection. Every tenant using `application_role` can read and write ALL casbin rules for ALL tenants.

The comment in the migration acknowledges this: "Tenant-level isolation is NOT enforced via RLS here; the domain value carried in v1 enforces multi-tenant isolation entirely at the application layer."

This is an acceptable architectural decision only if the application layer is provably correct. But the pgxAdapter's `LoadPolicy` uses a raw pgxpool connection that bypasses the `WithTenantFromCtx` path entirely. The adapter bypasses RLS context-setting. Any bug in domain filtering at the Casbin layer exposes all tenant policies to cross-tenant reads or writes.

**Recommendation:** Either accept this documented risk explicitly and add compensating controls (audit logging every casbin_rule write, separate DB roles per tenant tier), or scope `LoadPolicy` to use tenant-aware queries.

---

### Finding: casbin_rule SavePolicy uses TRUNCATE — cross-tenant destructive

**Severity: Critical**
**Location:** `internal/core/iam/repository/authz.go:248–283`

```go
if _, err := tx.Exec(ctx, `TRUNCATE casbin_rule`); err != nil {
```

`SavePolicy` truncates the entire `casbin_rule` table before re-inserting. This is a full data wipe affecting ALL tenants. Even though `AutoSave(true)` means `SavePolicy` is called only on explicit `SavePolicy()` invocation (not on incremental changes), this method exists and is callable. Any code path that calls `SavePolicy` — including disaster recovery procedures, migration scripts, or test cleanup — will wipe all authorization rules for all tenants simultaneously.

**Recommendation:** Remove or replace `SavePolicy` with a tenant-scoped version. Document explicitly that `SavePolicy` must never be called in production. Add a production guard (environment check) that panics or returns error if called.

---

### Finding: set_tenant_context uses is_local=true but pool connections are reused

**Severity: High**
**Location:** `db/migration/000306_security_active_tenant_guard.up.sql:32`

```sql
PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, true);
```

`is_local=true` makes the GUC setting transaction-local. This is correct behavior. However, connection pool connections are reused across requests. If a transaction is not properly committed or rolled back (e.g., application crash mid-transaction), the connection is returned to the pool with potentially stale session state. Whether pgx's pool correctly resets session state on connection return must be verified. If not, a subsequent request on the same connection could inherit a previous tenant's context.

**Recommendation:** Verify that `pgxpool` resets session-local GUCs on connection return. Add an integration test that verifies tenant context is clean on a recycled connection.

---

## Permission System Assessment

### Finding: The entire fine-grained ABAC permission system is unimplemented

**Severity: Critical**
**Location:** `internal/core/access/` (entire package), `db/migration/000404–000413`

The database schema defines:
- `permissions` table (resource + action + ABAC conditions)
- `roles` table (hierarchical, module-scoped, entity-scoped)
- `role_permissions` table (role → permission mapping)
- `user_roles` table (user → role assignment)
- `user_permissions` table (direct permission grants with DENY override)
- `policies` table (ABAC, RBAC, HYBRID policy evaluation)
- `access_requests` table (approval workflow)

None of these tables are queried by the authorization enforcement path. The `access.Service` interface in `internal/core/access/service.go` declares 13 methods. Zero are implemented — only model structs exist in subdirectories. The `access` package has a `service_mock.go` but no `service.go` implementation.

The only operational authorization system is Casbin using `casbin_rule` and `role_assignments`. Every other table is dead schema.

**Impact:**
- Data filters (`permissions.data_filters`) are never applied — row-level data filtering is silent no-op
- Field restrictions (`permissions.field_restrictions`) are never applied — column-level restrictions are silent no-op
- ABAC conditions (`permissions.conditions`) are never evaluated
- Access requests are never processed — `auto_revoke` never fires
- Role hierarchy (`roles.parent_role_id`, `roles.level`) is never used by the enforcer
- Direct permission grants (`user_permissions`) are never evaluated

An enterprise customer purchasing this ERP believing ABAC and approval workflows are functional will find they are not.

**Recommendation:** Either (a) remove all dead schema and document that ABAC is roadmapped but not implemented, (b) wire the relational ABAC system into the enforcement path, or (c) implement the `access.Service` interface. Option (a) is the only honest choice for a production system today.

---

### Finding: API key revocation does not invalidate Redis cache

**Severity: High**
**Location:** `internal/core/iam/service/apikey.go:142–156`

```go
func (s *apiKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
    // ...
    if err := s.repo.Revoke(ctx, keyID); err != nil { ... }
    // Best-effort cache eviction — we don't know the raw token so we can't
    // derive the cache key here.  The cached entry expires naturally after TTL.
    s.metrics.IncrementCounter("iam.apikey.revoked", nil)
    return nil
}
```

The code explicitly acknowledges it cannot evict the API key from Redis on revocation. A revoked key remains valid for up to 5 minutes (the cache TTL). For M2M integrations, this is an unacceptable security window. If a key is compromised and revoked, the attacker retains access for 5 minutes — enough to exfiltrate data or escalate.

**Recommendation:** Store the `key_hash` alongside the `key_id` in the API key repository (the DB already has it). At revocation time, query by `key_id` to get `key_hash`, then evict `apikey:{key_hash}` from Redis. This is a single extra DB read on an infrequent revocation path.

---

### Finding: Force logout does not evict session cache

**Severity: High**
**Location:** `internal/core/iam/service/session.go:333–376`, `internal/core/iam/repository/session.go:235–251`

```go
func (r *sessionRepo) InvalidateByUser(ctx context.Context, userID uuid.UUID) error {
    if err := r.store.InvalidateSessionsByUser(ctx, userID); err != nil { ... }
    // No cache eviction
    return nil
}
```

`LogoutAllForUser` and `LogoutAllForTenant` mark DB rows inactive but do NOT evict Redis cache entries. The service-layer comment says "Existing cache entries expire naturally within the session TTL window."

For a user terminated at 9:00 AM with a 24-hour session TTL, active session caches remain valid until 9:00 AM the next day. The user retains full system access after forced logout. This is unacceptable for employment termination, credential compromise, and incident response scenarios.

**Impact:** This is a SOX/SOC2 compliance failure. Terminated employee access cannot be immediately revoked. Security incident response cannot immediately invalidate compromised sessions.

**Recommendation:** On `InvalidateByUser`, enumerate the active session tokens from the DB (before marking inactive), then delete each from Redis. Add a session enumeration query if one doesn't exist. This is a critical path for access revocation SLAs.

---

## SQL Migration Assessment

### Finding: roles table entity_id is NOT NULL but Casbin never references roles table

**Severity: High**
**Location:** `db/migration/000405_auth_create_roles.up.sql:16`

```sql
entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
```

The `roles` table requires a non-null `entity_id`. This means every role must be scoped to an entity. But the Casbin role assignment system (`casbin_rule`, `role_assignments`) operates entirely by string names with no FK relationship to the `roles` table. The roles used in Casbin (e.g., `"tenant_admin"`, `"role:tenant.admin"`) have no corresponding rows in the `roles` table.

**Impact:** The `roles` table with its hierarchical structure, conditions JSONB, and permissions cache is completely orphaned. The ON DELETE RESTRICT on entity_id means deleting an entity would fail if roles exist — but since no roles in the operational system are in this table, this constraint provides no protection.

**Recommendation:** Remove the NOT NULL constraint on `roles.entity_id` to make it optional, document the decoupling, or (preferably) reconcile the two role systems.

---

### Finding: assign_user_role and revoke_user_role DB functions have no idempotency

**Severity: Medium**
**Location:** `db/migration/000412_user_roles_functions.up.sql:21–49`

`assign_user_role` is a plain INSERT with no ON CONFLICT clause. Calling it twice for the same (user_id, role_id, entity_id) will fail with a unique constraint violation (assuming the table has one) or silently create duplicates (if not). `revoke_user_role` is a DELETE with no check for existence. Neither function is used by the operational IAM service.

---

### Finding: Migration 000064 UNIQUE constraint excludes tenant_id

**Severity: Medium**
**Location:** `db/migration/000064_authz_role_assignments.up.sql:27`

```sql
CONSTRAINT role_assignments_unique UNIQUE (subject, role_name, domain)
```

The unique constraint is on `(subject, role_name, domain)` without `tenant_id`. The `domain` field carries the tenant identifier in Casbin convention, so this may be functionally correct. However, if a subject string format ever collides across tenants, or if the domain convention changes, this constraint provides no tenant-level protection. The constraint should include `tenant_id` for defense in depth.

---

### Finding: Down migrations are destructive with no data preservation

**Severity: Medium**
**Location:** `db/migration/000063_authz_casbin_rule.down.sql`

```sql
DROP TABLE IF EXISTS casbin_rule CASCADE;
```

`CASCADE` drops all dependent objects. Rolling back migration 000063 drops `casbin_rule` and cascades to anything dependent. In production, this wipes all authorization rules. There is no backup-first, soft-delete, or rename-then-drop pattern. Rolling back this migration in production is equivalent to wiping all IAM data.

**Recommendation:** Production migration rollbacks should rename tables (`ALTER TABLE casbin_rule RENAME TO casbin_rule_deprecated_YYYYMMDD`) rather than drop them. The data is recoverable if needed.

---

## SQL Query Assessment

### Finding: ListRoleAssignments returns all history including inactive — cache pollution risk

**Severity: Medium**
**Location:** `db/queries/authz.sql:34–42`

```sql
-- name: ListRoleAssignments :many
SELECT ... FROM role_assignments
WHERE subject = $1 AND domain = $2
ORDER BY created_at DESC;
```

This query has no `is_active` filter. It returns ALL role assignments including revoked and expired ones. The result is cached for 5 minutes (`roleAssignmentCacheTTL`). When `GetAssignments` is called (used for audit display), it returns revoked assignments — which is correct for audit. But the same cached data might be consumed by other callers expecting active-only assignments. The distinction between audit queries and active queries should be separate cached entities.

---

### Finding: ListAssignmentHistory uses optional text parameter via IS NULL trick — unsafe at scale

**Severity: Low**
**Location:** `db/queries/authz.sql:79–89`

```sql
WHERE tenant_id = $1
  AND ($2::text IS NULL OR subject = $2)
```

The `$2::text IS NULL OR subject = $2` pattern prevents PostgreSQL from using an index on `subject` when `$2` is NULL (full scan). For tenants with large assignment histories, this query performs a full table scan filtered by `tenant_id` alone. Add a separate query for the "all subjects" case.

---

### Finding: UpsertRoleAssignment query mismatch — role_slug not populated by service

**Severity: Medium**
**Location:** `db/queries/authz.sql:1–17`, `internal/core/iam/repository/authz.go:74–84`

The `UpsertRoleAssignment` SQL expects `role_slug` (parameter `$5`). The SQLC-generated `UpsertRoleAssignmentParams` struct includes `RoleSlug`. But the Go call site in `authzRepo.UpsertRoleAssignment` only passes `role` (the role name string). The `RoleSlug` field is not set — it receives its zero value. The `role_slug` column is documented as "denormalized slug for display/audit without JOIN" but is never populated. Audit logs dependent on `role_slug` are empty.

---

## SQLC Layer Assessment

### Finding: authz.sql.go UpsertRoleAssignment params structure inconsistency

**Severity: Medium**
**Location:** `db/sqlc/authz.sql.go`

The generated `UpsertRoleAssignmentParams` includes a `GrantedBy` field (UUID). The Go repository code passes `AssignedBy` and `DelegatedBy` (both `*string`) but the `GrantedBy` UUID field (added in migration 000412) is not wired up. The repo call passes only 8 parameters but the SQL has 10 positional parameters (`$1` through `$10` for `id, tenant_id, subject, role_name, role_slug, domain, assigned_by, granted_by, delegated_by, expires_at`). If SQLC regeneration is not current, there may be a parameter count mismatch at runtime.

---

## Transaction & Consistency Assessment

### Finding: AssignRole writes to two systems without a saga or compensation

Already documented above (Critical). The Casbin write and DB write are separate non-atomic operations.

### Finding: buildAndPersistSession permission computation happens before session row insertion

**Severity: Medium**
**Location:** `internal/core/iam/service/session.go:190–282`

Permission computation (including `GetImplicitRoles`, `GetPolicies`) happens before the session row is inserted. If the session insertion fails, the computed permissions are discarded — that is correct. But if `buildPermissions` itself mutates state (e.g., the JIT bootstrap path calls `BootstrapTenantAdmin` which writes to both Casbin and the DB), those writes persist even if the subsequent session creation fails. A failed login can silently bootstrap tenant_admin for a user.

---

### Finding: UpdateLastSeen goroutine uses cancelled context

**Severity: Medium**
**Location:** `internal/core/iam/repository/session.go:253–257`

```go
func (r *sessionRepo) UpdateLastSeen(ctx context.Context, hash string) {
    go func() {
        _ = r.store.UpdateSessionLastSeen(ctx, hash)
    }()
}
```

The goroutine captures the HTTP request context. When the HTTP handler returns (before the goroutine completes), the context is cancelled. `UpdateSessionLastSeen` will then fail with a context cancellation error on almost every call because the DB round-trip takes longer than the remaining handler lifetime. The `last_seen_at` column will rarely be updated correctly.

**Recommendation:** Use `context.WithoutCancel(ctx)` to detach from the request lifecycle, or use `context.Background()` with a short timeout.

---

## Concurrency & Race Condition Assessment

### Finding: Concurrent revokeExpiredRoles calls can double-deactivate roles

**Severity: Medium**
**Location:** `internal/core/iam/service/authz.go:575–588`

`revokeExpiredRoles` runs on every `Enforce` call. For a subject with expired roles, concurrent requests will each call `ListExpiredActiveRoleNames` (returning the same expired roles), then each call `DeactivateRoleAssignment` for the same roles. The DB UPDATE is idempotent (`SET is_active=FALSE` on an already-false row). But `DeleteRoleForUserInDomain` is called multiple times concurrently on the Casbin enforcer. Casbin's in-memory operations must be verified as thread-safe for concurrent deletions of the same role. If Casbin's role manager is not fully goroutine-safe for concurrent duplicate removes, this is a race condition.

---

### Finding: No rate limiting on CompleteMFALogin

**Severity: High**
**Location:** `internal/core/iam/service/session.go:145–175`

`CompleteMFALogin` validates a TOTP code. TOTP codes are 6 decimal digits (1,000,000 combinations). There is no rate limiting, no attempt counter, and no lockout on the pending MFA token. An attacker who obtains the pending token (e.g., via network interception or log leakage) can brute-force TOTP codes exhaustively. The 5-minute pending token TTL (300 seconds) with 1M combinations means ~3,333 attempts per second would be needed — feasible with no rate limiting.

**Recommendation:** Limit `CompleteMFALogin` to 5 attempts per pending token. On fifth failure, delete the pending token and require re-authentication.

---

## Caching & Invalidation Assessment

### Finding: Session cache is invalidated by token hash, but force-logout cannot enumerate hashes

As documented above. `InvalidateByUser` and `InvalidateByTenant` cannot evict Redis caches because they don't enumerate token hashes.

### Finding: Role assignment cache key includes subject string — injection risk

**Severity: Low**
**Location:** `internal/core/iam/repository/authz.go:212–214`

```go
func roleAssignmentCacheKey(subject, domainName string) string {
    return "authz:ra:" + subject + ":" + domainName
}
```

If `subject` or `domainName` contain the separator character `:`, the cache key is ambiguous. For example, `subject="authz:ra:tenant:user1"` + `domain="tenant1"` produces the key `authz:ra:authz:ra:tenant:user1:tenant1` which conflicts with `subject="authz:ra:tenant"` + `domain="user1:tenant1"`. While the current subject format `"tenant:uuid"` is not adversarially constructed, this pattern is fragile and should use a delimiter-safe encoding (e.g., hash the inputs or use a non-colon separator).

---

## Security Assessment

### Finding: SSO bypasses MFA unconditionally

**Severity: High**
**Location:** `internal/core/iam/service/session.go:177–186`

```go
// MFA is intentionally skipped — the IdP is the second factor.
func (s *sessionService) LoginWithSSO(ctx context.Context, user *domain.User) (*domain.ResolvedSession, string, error) {
```

SSO login bypasses MFA entirely. If an SSO provider supports password-only authentication (no MFA enforced at the IdP), users can register an SSO provider and use it as an MFA bypass. There is no configuration to require IdP MFA before granting session. An insider threat can configure a weak SSO provider (e.g., one they control with no MFA) and bypass the platform's MFA requirement.

**Recommendation:** Add a `require_idp_mfa` flag to the SSO provider config. When true, verify the IdP's auth response includes MFA claim (AMR claim for OIDC). Block sessions from IdPs that don't assert MFA when the flag is set.

---

### Finding: Session TTL is tenant-configurable with no maximum bound

**Severity: High**
**Location:** `internal/core/iam/service/session.go:524–531`

```go
func (s *sessionService) resolveTTL(cfg domain.Configuration) time.Duration {
    if v, ok := cfg.Settings["iam.session_ttl_hours"]; ok && v != "" {
        if hours, err := strconv.Atoi(v); err == nil && hours > 0 {
            return time.Duration(hours) * time.Hour
        }
    }
    return s.cfg.SessionTTL
}
```

A tenant can set `iam.session_ttl_hours` to any positive integer. Setting it to `876000` (100 years) creates sessions that never effectively expire. Combined with the fact that force-logout doesn't evict cache, a compromised session from a misconfigured tenant has an effectively infinite window.

**Recommendation:** Add a hard cap (e.g., 72 hours maximum) that cannot be overridden by tenant settings.

---

### Finding: tenantIDFromCtx silently returns uuid.Nil

**Severity: High**
**Location:** `internal/core/iam/service/session.go:482–489`

```go
func tenantIDFromCtx(ctx context.Context) uuid.UUID {
    if v, ok := ctx.Value(cache.TenantIDKey).(string); ok && v != "" {
        if id, err := uuid.Parse(v); err == nil {
            return id
        }
    }
    return uuid.Nil
}
```

If the context lacks a tenant ID (misconfigured middleware, race condition, or code path that doesn't set tenant context), `uuid.Nil` is returned silently and embedded in the session. A session with `tenant_id = uuid.Nil` would either fail DB insertion (FK to tenants(id)) or, if somehow persisted, represent a session not scoped to any tenant. The caller receives no error indication.

**Recommendation:** Return `(uuid.UUID, error)`. The session creation path must reject sessions with no tenant ID.

---

### Finding: API key scopes are stored as text array with no validation

**Severity: Medium**
**Location:** `db/migration/000310_iam_api_keys.up.sql:17`

```sql
scopes TEXT[] NOT NULL DEFAULT '{}',
```

Scopes are free-form strings. The "cannot exceed creator's permissions" ceiling is documented but there is no enforcement in the repository layer — the caller could pass any scopes. The `buildAPIKeySession` function converts scopes directly to permissions:

```go
for _, scope := range key.Scopes {
    perms[scope] = true
}
```

If scopes include `"*"` (the wildcard sentinel), the API key session bypasses all permission checks. There is no validation that scopes are a subset of the creator's actual permissions at creation time.

**Recommendation:** At `CreateAPIKey` time, validate that each requested scope is present in the creator's resolved permissions. Reject any scope not present in the creator's permission set.

---

## Auditability & Compliance Assessment

### Finding: role_slug is never populated

Documented above. Audit log entries for role assignments show empty `role_slug`, making audit trails non-human-readable without joining to Casbin string names.

### Finding: Access request approval is never executed

The `access_requests` table is designed for SOX/SOC2 approval workflows. The `auto_revoke` column triggers automatic access revocation on expiry. None of this executes. Approved access requests are never fulfilled. Expired access is never revoked.

### Finding: No audit event for JIT bootstrap

When `BootstrapTenantAdmin` grants tenant_admin at login, there is no audit log entry. A user receiving escalated privileges at login leaves no trace.

### Finding: Policy evaluation cache (`policy_evaluations` table) has no TTL enforcement

The `policy_evaluations` table caches authorization decisions. There is no query or job that removes stale/expired entries. Over time this table grows unboundedly.

---

## Performance & Scalability Assessment

### Finding: LoadPolicy loads all tenants' policies into memory — O(n tenants) memory

**Location:** `internal/core/iam/repository/authz.go:228–246`

```sql
SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule ORDER BY ptype
```

At startup and on every `InvalidateCache`, all Casbin rules for all tenants are loaded into the enforcer's memory. With 10,000 tenants × 25 default policies each = 250,000 rows minimum. With complex permission sets: millions of rows. The enforcer holds this entirely in RAM. This is a fundamental architectural constraint of the single-enforcer-per-process Casbin model.

**Recommendation:** Evaluate Casbin's per-domain filtering capability or tenant-sharded enforcer pools for multi-tenant scale.

---

### Finding: buildPermissions calls GetPolicies which loads all domain policies on every login

**Location:** `internal/core/iam/service/session.go:415`

```go
policies, err := s.authz.GetPolicies(ctx, domainName)
```

`GetPolicies` calls `GetFilteredPolicy(1, domainName)` on the Casbin enforcer. While this filters in-memory (no DB call at this point), for a tenant with hundreds of roles and complex ACLs, this iterates the full policy list and copies it. This runs on every login. No caching of the policy result for a domain.

---

## Testing & QA Assessment

### Finding: roles_test.go is entirely commented out — zero role integration test coverage

**Severity: Critical**
**Location:** `internal/core/iam/roles_test.go`

The entire file is a block comment. All DB-backed tests for `AssignRole`, `RevokeRole`, `GetRoles`, `HasRole`, `GetAssignments`, and temporal expiry are commented out. There are 30+ test cases that exist as comments but have never been run. This means:

- Role assignment correctness has never been verified against a real database
- Expiry/revocation behavior has never been integration-tested
- The "idempotent upsert" guarantee has never been verified

**Recommendation:** Uncomment and fix. These are not optional tests — role management correctness is foundational.

---

### Finding: No test coverage for cross-instance cache invalidation

There are no tests verifying behavior when `InvalidateByUser` is called while a valid cached session exists. No test verifies that a user can still access resources after being force-logged-out via the DB-only path. This behavior is demonstrably broken (as documented above) and untested.

---

### Finding: AuthzEnforceSuite uses in-memory enforcer — does not test pgxAdapter

The `authz_enforce_test.go` tests use `NewInMemoryAuthzService` which bypasses the pgxAdapter entirely. The Casbin pgx adapter (`AddPolicy`, `AddPolicies`, `RemovePolicy`, `RemoveFilteredPolicy`, `LoadPolicy`, `SavePolicy`) has zero test coverage.

---

### Finding: No fuzzing on permission string inputs

Subject, domain, object, and action are user-controlled strings passed directly to Casbin. No fuzzing tests verify behavior with: empty strings (caught by validation), extremely long strings, Unicode, SQL injection fragments, Casbin wildcard characters (`*`, `?`), path traversal patterns. An input of `../../../` in the object field could match wildcard policies unexpectedly.

---

### Finding: No test for JIT bootstrap idempotency

The JIT bootstrap in `buildPermissions` calls `BootstrapTenantAdmin` which calls `AddPolicy` and `AssignRole`. If these are called concurrently (two simultaneous logins for a roleless user), both calls may try to bootstrap simultaneously. The `ErrPolicyConflict` handling is silently ignored, but the `AssignRole` call is not idempotent by itself — the upsert in the repo handles it, but the Casbin `AddGroupingPolicy` may create duplicate in-memory entries. No concurrent test exists.

---

## Operational Readiness Assessment

### Finding: No health check or readiness probe for Casbin enforcer state

There is no endpoint or signal that indicates whether the Casbin enforcer has successfully loaded policies. On startup, `NewAuthzService` calls `casbin.NewEnforcer` which triggers `LoadPolicy`. If the DB is unavailable at startup, the enforcer starts with an empty policy set. All authorization checks return `denied` (default deny — safe for security, catastrophic for availability). There is no alerting or readiness gate for this condition.

---

### Finding: No metrics for Casbin policy count or stale state age

Metrics track enforcement counts and durations but not:
- Current policy count in the enforcer
- Time since last `LoadPolicy`
- Number of expired role cleanup operations
- Number of JIT bootstraps triggered

These are critical operational signals.

---

## Maintainability Assessment

### Finding: Dead code duplication — ruleToValues/filterEmpty/joinRule defined twice

**Location:** `internal/core/iam/adapter.go:4–29`, `internal/core/iam/repository/authz.go:361–384`

The three helper functions `ruleToValues`, `filterEmpty`, and `joinRule` are defined in both the `iam` package (`adapter.go`) and the `repository` package (`authz.go`). The `iam`-package versions appear to be unused dead code since the repository has its own copies. This creates maintenance confusion and drift risk.

---

### Finding: Service alias creates naming confusion

**Location:** `internal/core/iam/iam.go:154`

```go
// Service is a backward-compatible alias for AuthzService.
// Prefer AuthzService in new code.
Service = iamservice.AuthzService
```

`SeedDefaultRoles` uses `Service` while `NewUserServiceWithConfig` takes `AuthzService`. Two names for the same interface creates cognitive overhead and increases the chance of a future interface divergence going unnoticed.

---

### Finding: access package service_mock.go exists but no implementation

`internal/core/access/service_mock.go` (generated by mockgen) mocks the `Service` interface. But there is no concrete implementation. The mock is used for nothing since no test currently exercises the access package methods. This is misleading infrastructure.

---

## Critical Findings

| # | Finding | Location |
|---|---------|----------|
| C1 | AssignRole is non-atomic between Casbin and DB | `service/authz.go:236` |
| C2 | Casbin enforcer state not synchronized across horizontal replicas | `service/authz.go` (entire) |
| C3 | casbin_rule RLS is USING(TRUE) — no DB-level tenant isolation | migration 000063 |
| C4 | SavePolicy TRUNCATEs all tenants' rules | `repository/authz.go:256` |
| C5 | Entire ABAC permission system is unimplemented dead schema | `internal/core/access/` |

---

## High Severity Findings

| # | Finding | Location |
|---|---------|----------|
| H1 | RevokeRole ignores Casbin return value | `service/authz.go:297` |
| H2 | API key revocation cannot evict Redis cache (5-min window) | `service/apikey.go:142` |
| H3 | Force logout (InvalidateByUser/Tenant) does not evict session cache | `repository/session.go:235` |
| H4 | buildPermissions failure silently creates empty-permission sessions | `service/session.go:196` |
| H5 | JIT bootstrap grants tenant_admin to any roleless tenant user | `service/session.go:404` |
| H6 | No rate limiting on CompleteMFALogin (TOTP brute-force possible) | `service/session.go:145` |
| H7 | SSO bypasses MFA unconditionally with no configurable override | `service/session.go:179` |
| H8 | Session TTL tenant-configurable with no maximum cap | `service/session.go:524` |
| H9 | tenantIDFromCtx silently returns uuid.Nil | `service/session.go:482` |
| H10 | revokeExpiredRoles DB query on every Enforce call (hot path) | `service/authz.go:153` |
| H11 | roles_test.go entirely commented out | `roles_test.go` |

---

## Medium Severity Findings

| # | Finding | Location |
|---|---------|----------|
| M1 | UpdateLastSeen goroutine uses cancelled HTTP context | `repository/session.go:253` |
| M2 | roles table entity_id NOT NULL orphaned from Casbin system | migration 000405 |
| M3 | role_slug never populated in role assignments | `repository/authz.go:74` |
| M4 | ListAssignmentHistory optional-subject via IS NULL prevents index use | `queries/authz.sql:87` |
| M5 | UpsertRoleAssignment SQLC params not fully wired (GrantedBy) | `repository/authz.go:74` |
| M6 | Down migrations use CASCADE DROP — production data loss risk | migration down files |
| M7 | Concurrent revokeExpiredRoles may race on Casbin delete | `service/authz.go:575` |
| M8 | API key scopes have no validation against creator's permissions | `service/apikey.go:74` |
| M9 | policy_evaluations table grows unbounded (no TTL cleanup) | migration 000307 |
| M10 | Domain error type embeds HTTP status (transport coupling) | `domain/errors.go` |
| M11 | buildPermissions concurrent JIT bootstrap race (two simultaneous logins) | `service/session.go:404` |

---

## Low Severity Findings

| # | Finding | Location |
|---|---------|----------|
| L1 | roleAssignmentCacheKey uses colon separator — ambiguous with colon in values | `repository/authz.go:212` |
| L2 | ruleToValues/filterEmpty/joinRule defined twice (dead code) | `adapter.go`, `repository/authz.go` |
| L3 | Service alias creates naming confusion | `iam.go:154` |
| L4 | role_assignments UNIQUE constraint excludes tenant_id | migration 000064 |
| L5 | assign_user_role DB function has no ON CONFLICT clause | migration 000412 |
| L6 | pgxAdapter LoadPolicy uses context.Background() ignoring cancellation | `repository/authz.go:229` |

---

## Architectural Smells

1. **Phantom ABAC layer** — Schema, documentation, and interface definitions for a permission system that does not execute. The codebase pretends completeness it does not have.

2. **Single-enforcer anti-pattern for SaaS** — Global Casbin enforcer holding all tenant policies in one RAM blob. This is the standard Casbin getting-started pattern, not a production SaaS pattern.

3. **Two role models with zero reconciliation** — Casbin string-named roles and relational UUID-keyed roles in `roles` table. No bridge, no FK, no documentation of intentional decoupling.

4. **Authorization logic split between three packages** — `internal/core/iam/service` (Casbin enforcement), `internal/core/iam/repository` (role persistence), `internal/core/access` (stub). Authorization concerns are fragmented.

5. **JIT mutation in read path** — `buildPermissions` (called during login, a mostly-read operation) mutates authorization state (calls BootstrapTenantAdmin). Read paths should not have write side effects.

---

## Dangerous Assumptions

1. **"The domain field enforces multi-tenant isolation at the application layer"** — This is only true if the application never passes the wrong domain. Every Casbin query must pass the correct tenant domain. There is no DB-level backstop.

2. **"MFA is skipped for SSO — the IdP is the second factor"** — Only true if the IdP actually enforces MFA. This assumption is not verified.

3. **"Sessions expire naturally after TTL"** — True for Redis cache. Not true for sessions that are forcibly invalidated via DB-only path (bulk logout). The cache survives invalidation.

4. **"AutoSave is true so SavePolicy is never called"** — AutoSave=true means individual mutation methods (AddPolicy, etc.) persist immediately via the adapter. SavePolicy is a bulk replace and is only called explicitly. But the method exists, is callable, and is catastrophically dangerous.

5. **"API key scope ceiling is enforced at creation time"** — Not enforced in code. Documented as design intent only.

---

## Missing Safeguards

1. No distributed Casbin Watcher for multi-instance policy synchronization
2. No maximum session TTL cap
3. No rate limiting on MFA completion
4. No Redis cache eviction on bulk user/tenant logout
5. No Redis cache eviction on API key revocation
6. No audit log for JIT privilege bootstrap
7. No readiness gate for Casbin policy load on startup
8. No validation of API key scopes against creator permissions
9. No enforcement that tenantIDFromCtx returns a valid UUID before session creation
10. No background job processing `access_requests.auto_revoke`

---

## Production Failure Scenarios

1. **Horizontal scaling deployment:** Two instances start. Admin revokes a user's role on instance A. Instance B continues granting access to revoked user. Duration: until next restart or `InvalidateCache` call. Probability: 100% in any load-balanced deployment.

2. **Employee termination:** HR terminates user, admin calls `LogoutAllForUser`. DB sessions marked inactive. Redis cache survives. User continues accessing system for up to session TTL (default 8 hours, configurable up to infinite). Probability: 100% currently.

3. **API key compromise:** Compromised key is reported and revoked via admin console. Key DB row marked revoked. Redis cache survives for 5 minutes. Attacker continues using key for 5 minutes. Probability: 100% currently.

4. **New tenant user first login:** Any tenant user with no explicit role assignments is granted `tenant_admin` at first login via JIT bootstrap. If this is not the intended behavior (e.g., new employee accounts, service accounts), they receive unrestricted admin access. Probability: 100% for any intentionally role-less account.

5. **Casbin startup with DB unavailable:** Service starts, `LoadPolicy` fails, enforcer has empty policy set. All authorization checks return denied. Service is fully unavailable for legitimate users. No readiness gate prevents traffic from reaching this state. Probability: non-trivial in cloud environments.

6. **TOTP brute-force:** Attacker intercepts MFA pending token (network log, Redis breach). 5-minute window, no rate limiting, 6-digit TOTP = 1M combinations. With 10,000 req/s possible against unprotected endpoint: full brute-force in under 2 minutes. Probability: low but possible.

---

## Top 10 Highest Risk Issues

1. **Casbin not distributed** — authorization state diverges across instances immediately on scaling
2. **Force logout cache bypass** — terminated users retain access for full TTL
3. **JIT tenant_admin privilege escalation** — any roleless tenant user gets admin at login
4. **Dead ABAC system** — advertised permission model does not execute
5. **AssignRole non-atomic** — ghost Casbin assignments after DB failure
6. **API key revocation cache bypass** — 5-minute window post-revocation
7. **casbin_rule USING(TRUE) RLS** — no DB-level cross-tenant policy isolation
8. **SavePolicy TRUNCATE** — callable method that wipes all authorization data
9. **MFA no rate limiting** — brute-force viable via pending token
10. **revokeExpiredRoles on every Enforce** — hot-path DB query DoS vector

---

## Refactoring Priorities

1. **Immediate:** Register a Casbin Redis Watcher for distributed policy sync before any horizontal deployment
2. **Immediate:** Implement Redis cache eviction in `InvalidateByUser` and `InvalidateByTenant`
3. **Immediate:** Fix `RevokeAPIKey` to look up and evict the Redis cache entry by key_hash
4. **Immediate:** Add maximum session TTL cap in `resolveTTL`
5. **Immediate:** Add MFA rate limiting in `CompleteMFALogin`
6. **Short-term:** Remove or guard the JIT bootstrap — require explicit admin action to grant admin role
7. **Short-term:** Make `tenantIDFromCtx` return an error instead of uuid.Nil
8. **Short-term:** Fix `UpdateLastSeen` goroutine to use a detached context
9. **Short-term:** Check and propagate `DeleteRoleForUserInDomain` return value in `RevokeRole`
10. **Short-term:** Uncomment all tests in `roles_test.go` and verify they pass
11. **Medium-term:** Decide fate of ABAC layer — implement it or remove the dead schema
12. **Medium-term:** Implement API key scope validation against creator permissions
13. **Long-term:** Evaluate per-tenant Casbin enforcer sharding for multi-tenant scale

---

## Immediate Action Items

| Priority | Action | Risk Blocked |
|----------|--------|-------------|
| P0 | Do NOT deploy behind a load balancer until Casbin Watcher is implemented | C2 |
| P0 | Fix `InvalidateByUser`/`InvalidateByTenant` to evict Redis cache | H3 |
| P0 | Audit all existing "roleless" users — they will receive tenant_admin on next login | H5 |
| P1 | Fix `RevokeAPIKey` to evict cache | H2 |
| P1 | Add MFA attempt rate limiting | H6 |
| P1 | Add session TTL hard cap | H8 |
| P1 | Fix `RevokeRole` error check | H1 |
| P1 | Fix `UpdateLastSeen` context | M1 |
| P2 | Uncomment and run `roles_test.go` | H11 |
| P2 | Remove `SavePolicy` TRUNCATE or add production guard | C4 |

---

## Positive Findings

1. **Token storage security is correct.** SHA-256 hashes for sessions, API keys, and password reset tokens. Raw tokens never persisted. Well-implemented.

2. **SSO client secret encryption is correct.** AES-256-GCM encryption for SSO provider secrets documented and implemented.

3. **Casbin domain isolation logic is correct.** The Casbin model correctly scopes role inheritance and policy matching to the domain field. Cross-domain isolation in the enforcer itself is sound (verified by `TestV5_CrossTenantPolicyLeak_Denied`).

4. **Deny-override is correctly implemented.** The `buildPermissions` and Casbin model both handle deny-overrides. Explicit deny beats any allow for the same object.action.

5. **Password history enforcement schema exists.** The password history JSONB approach prevents password reuse. Implementation is pragmatic.

6. **Tenant context guard at DB level.** `set_tenant_context` rejects non-ACTIVE tenants at the DB function level. Defense in depth.

7. **RLS is consistently applied.** Every IAM-related table has RLS enabled with FORCE ROW LEVEL SECURITY. The pattern is consistent.

8. **Observability is thorough.** Tracing, metrics, and structured logging are present throughout. The system is observable.

9. **Lazy role expiry design is conceptually sound.** Expiry checked at enforcement time prevents clock skew between instances causing inconsistent revocation. The mechanism is correct; the hot-path cost is the concern.

10. **In-memory enforcer for tests is a good pattern.** `NewInMemoryAuthzService` allows fast unit tests without a database. The test infrastructure is well-designed even if coverage is incomplete.

---

## Final Verdict

The IAM module is **not production-ready for a multi-tenant SaaS deployment.**

The core enforcement logic (Casbin RBAC with domain isolation) is functionally correct in a single-instance deployment. The token storage, encryption, and credential handling patterns are sound.

However, four architectural defects independently make this system unsafe for production:

1. Horizontal scaling breaks authorization correctness immediately (no distributed policy sync)
2. Emergency access revocation does not work (cache survives DB invalidation)
3. Privilege escalation exists for any roleless tenant user (JIT bootstrap)
4. The documented ABAC permission system does not execute (phantom compliance)

Each of these alone would be a production blocker. Together they represent a system that will fail in the specific scenarios where IAM matters most: under load, during incidents, during compliance audits, and during adversarial use.

The codebase shows serious engineering effort and good instincts in several areas. The path to production-readiness is not a rewrite — it requires targeted fixes to the 10 P0/P1 items listed above, genuine integration test coverage, and an honest documentation revision that reflects what is actually implemented.

---

## Production Readiness Score: 3.5 / 10

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Security posture | 3/10 | Cache bypass on revocation, MFA brute-force, JIT escalation, no scope validation |
| Tenant isolation | 4/10 | DB-level isolation via Casbin domain works single-instance; RLS on casbin_rule is USING(TRUE) |
| Scalability | 2/10 | Single-instance Casbin, hot-path DB query per request, no distributed invalidation |
| Correctness | 4/10 | Core Casbin enforcement correct; non-atomic writes; empty-perm session risk |
| Test coverage | 3/10 | Unit tests for enforcement exist; roles integration tests fully commented out; DB adapter untested |
| Auditability | 4/10 | Audit fields exist; role_slug unpopulated; ABAC audit trail dead |
| Operational readiness | 3/10 | No startup readiness gate; no max TTL; no distributed cache invalidation |
| Code quality | 6/10 | Clean structure, good observability; dead code duplication; stub package misleads |
| Documentation | 4/10 | Documentation references unimplemented features; model is honest in places |
| Compliance | 2/10 | Force-logout doesn't work; approval workflow not implemented; JIT escalation unlogged |
