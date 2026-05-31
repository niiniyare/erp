[<-- Back to Index](README.md)

## API Keys and Service Accounts

> **Implementation status**: API key create/validate/revoke/list are [IMPLEMENTED].
> Scopes field is [PARTIAL] — stored in DB but not enforced via Casbin in v1.0; stored in session configuration only.
> Full Casbin-backed API key authorization is [PLANNED - NOT IN v1.0].

---

### 1. API Key Lifecycle

#### Creation

API keys are created via `APIKeyService.CreateAPIKey()` (see `internal/core/iam/service/apikey.go`).

Who can create:
- Any authenticated tenant user with permission to create API keys in their tenant
- The key is scoped to the **creator's tenant** — cross-tenant keys are not possible

Required fields (`domain.CreateAPIKeyRequest`):
- `Name` — human-readable identifier for the key (e.g., "CI Pipeline - Invoice Import")
- `CreatedBy` — UUID of the authenticated user creating the key (for audit trail)
- `Scopes` — list of permission strings (acts as a ceiling; see Section 1.3)

Optional fields:
- `ExpiresAt` — expiry time; `nil` means the key never expires

The `TenantID` is derived from the request context (`current_tenant_id()`) — it is never passed in the request body.

#### Key Format and Prefix

The generated raw token format:
```
eak_<64 hex characters>
```

- Prefix `eak_` stands for "ERP API Key" — makes the token recognisable in logs and tooling without exposing the secret
- 64 hex chars = 32 bytes of cryptographically random entropy (via `crypto/rand`)
- Total raw token length: 68 characters

The raw token is returned **once** at creation and never stored anywhere. After that, only the SHA-256 hex hash (`key_hash`) is persisted in the `api_keys` table.

#### Scopes Field [PARTIAL]

The `scopes` column stores a `TEXT[]` array of permission strings. Design intent:

- Scopes act as a **ceiling**: a key can never exceed the creating user's own permissions at creation time
- Current enforcement: Scopes are loaded into `ResolvedSession.Configuration` when the key is validated — they are accessible to application code but are **not** automatically enforced by Casbin
- Full Casbin integration (mapping scopes to p-rules) is planned for v1.0+ — see `deferred-features.md`

For now, when using API keys for machine integrations, design handlers to explicitly check scopes from the session configuration if needed, or use role-based permissions by assigning appropriate Casbin roles to the service account's user record.

#### Expiry

Set `ExpiresAt` when creating keys that should automatically expire:
- Short-lived keys for CI/CD pipelines (e.g., 90 days)
- Integration keys with known contract durations
- Keys for temporary contractors

`APIKey.IsActive()` checks both `RevokedAt` and `ExpiresAt` — an expired key returns `nil` from `ValidateAPIKey`.

#### Revocation

`APIKeyService.RevokeAPIKey(keyID)` sets `revoked_at` on the DB row immediately. Subsequent `ValidateAPIKey` calls that hit the DB will return `nil`.

**Cache eviction limitation**: The API key validation path caches `ResolvedSession` in Redis with a 5-minute TTL (`apikey:{sha256hex(rawToken)}`). When a key is revoked, the DB is updated immediately, but the in-flight cache entry can serve requests for up to 5 more minutes.

For emergency revocation:
1. Call `RevokeAPIKey(keyID)` via the API
2. Manually evict the Redis key: `DEL apikey:{sha256hex(rawToken)}` — requires knowing the raw token's hash
3. Or wait ≤ 5 minutes for cache expiry

This is a known limitation documented in [Security Considerations](./17-security-considerations.md) under T10.

#### Last-Used Tracking

The `last_used_at` column on `api_keys` is updated asynchronously on each validated request. This enables audit queries:
```sql
SELECT name, last_used_at, expires_at FROM api_keys
WHERE tenant_id = $1 AND revoked_at IS NULL
ORDER BY last_used_at DESC NULLS LAST;
```

---

### 2. Service Account Model

#### Service Accounts as Users

In Awo ERP, machine principals are modelled as users with `user_type = 'API'` or `'SERVICE'`. This means:
- Each integration/service has a user record in the `users` table
- The user has `UserType = "API"` or `"SERVICE"`, which maps to `ActorAPI` in the domain model
- The service account can be assigned Casbin roles in the `<tenantID>:api` domain
- The service account has one or more API keys associated with it

This design gives service accounts the same role/permission management tools as human users.

#### Authentication Flow for API Keys

The API key session does **not** use the `user_sessions` table. Instead (from `service/apikey.go`):

1. Client sends `Authorization: Bearer eak_<token>` header
2. `ValidateAPIKey` is called with the raw token
3. Cache-aside: check Redis `apikey:{sha256hex(token)}`
4. On cache miss: `repo.GetByHash()` executes a cross-tenant hash lookup (requires `admin_role` at DB level — no tenant context yet)
5. `buildAPIKeySession` creates a minimal `ResolvedSession` from the key's fields
6. Session is cached in Redis for 5 minutes
7. The `ResolvedSession.UserID` is set to the key's `CreatedBy` (creator's UUID) for audit trail

This design avoids `user_sessions` table bloat on high-frequency API calls.

**Note on ActorAPI sessions**: The `ToPrincipal()` method on `ResolvedSession` currently falls back to `APISubject(userID)` using the creator's user UUID, not a dedicated `ClientID`. This is noted as a TODO in the code and is functionally correct for policy lookups but semantically imprecise. Full service-account-specific subject handling is planned.

#### Entity Scope for Service Accounts

API key sessions currently default to `EntityScopeAll` (see `buildAPIKeySession` in `service/apikey.go`). This means the API key is not restricted by entity scope — it can access all entities in the tenant.

For tightly scoped integrations (e.g., a POS integration that should only access one terminal), restrict access via Casbin role policies rather than entity scope in v1.0.

#### Recommended Role Structure for M2M

For machine-to-machine integrations:

1. Create a dedicated Casbin role for the integration (e.g., `role:invoice_import_service`)
2. Grant only the minimum necessary permissions:
   ```
   finance.receivables.invoices create
   finance.receivables.invoices read
   ```
3. Assign that role to the service account user in the `<tenantID>:api` domain
4. Create one API key per integration — never share keys across integrations
5. Set a reasonable expiry (e.g., 1 year for long-running integrations)

---

### 3. Security Guarantees

#### Key Hash at Rest

Only `SHA-256(raw_token)` is stored in the `key_hash` column. The raw token is never persisted — not in DB, not in logs, not in backups. A DB compromise does not expose raw API key values.

#### Prefix for Identification

The `eak_` prefix enables identification of Awo API keys in logs, code repositories, or leak scanning tools without exposing the secret portion. Secret scanning tools can flag patterns matching `eak_[0-9a-f]{64}` to detect accidental exposure.

#### Revocation Is Immediate at DB Level

The DB row's `revoked_at` is set immediately. New requests that miss the Redis cache will see the revocation on the first DB lookup after revocation. The only delay is the 5-minute Redis cache TTL for already-cached sessions.

#### Keys Are Tenant-Scoped

The RLS policy on `api_keys` enforces `tenant_id = current_tenant_id()` for `application_role` queries. Cross-tenant hash lookups (for authentication, before tenant context is known) use `admin_role` — this bypass is necessary and controlled. See migration `000310_iam_api_keys.up.sql`.

#### DB-Level RLS for api_keys

Three RLS policies are defined on `api_keys`:
- `api_keys_tenant_isolation` (application_role): Tenant-scoped read/write
- `api_keys_admin_access` (admin_role): Full bypass — required for cross-tenant hash lookups during auth
- `api_keys_ro_select` (readonly_role): Tenant-scoped read

---

### 4. Operational Best Practices

#### Short-Lived Keys for Automation Pipelines

CI/CD pipelines and automated tasks should use time-limited keys:
- Set `ExpiresAt` to 90 days or less
- Rotate before expiry — create new key, update pipeline secret, then revoke old key
- Never store raw tokens in plaintext config files — use a secrets manager

#### Dedicated Role per Integration

Each integration should have:
- Its own user record (service account)
- Its own API key(s)
- Its own Casbin role with minimum necessary permissions

This ensures that:
- Revoking one integration's access does not affect other integrations
- Audit logs identify which integration performed each action
- Permissions can be independently adjusted per integration

#### Regular Rotation

- Review all API keys quarterly: `ListAPIKeys()` per tenant
- Revoke keys that are unused (`last_used_at` older than 90 days)
- Rotate all keys before they expire (create new, update secrets, revoke old)
- Revoke keys when the integration is decommissioned

#### Audit Logging

Every API request via key authentication creates an OTel span with:
- `apikey.id` — the key UUID
- `tenant.id` — the tenant UUID
- `authz.subject`, `authz.domain`, `authz.object`, `authz.action` — the enforcement decision context

Query patterns for API key audit:
```sql
-- All active keys and their last usage
SELECT name, id, expires_at, last_used_at
FROM api_keys
WHERE tenant_id = $1
  AND revoked_at IS NULL
ORDER BY last_used_at DESC NULLS LAST;

-- Keys expiring within 30 days
SELECT name, expires_at
FROM api_keys
WHERE tenant_id = $1
  AND revoked_at IS NULL
  AND expires_at BETWEEN NOW() AND NOW() + INTERVAL '30 days';
```

---

See also:
- [Tenant Administration](./23-tenant-administration.md)
- [Entity Scope](./25-user-entity-scope.md)
- [Security Considerations](./17-security-considerations.md)
- [Authentication](./06b-authentication.md)
