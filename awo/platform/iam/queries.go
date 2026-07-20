// Package iam — queries.go
//
// This file centralises every raw SQL statement used by AuthService.
// All SQL lives here and nowhere else in the iam package.
//
// # Why raw SQL exists in the IAM package
//
// The framework provides EntityRepository[T] for all standard CRUD operations
// on registered entities. IAM cannot use EntityRepository for every operation
// because of three specific constraints:
//
//  1. Sensitive field access — [password_hash] is declared Sensitive:true in
//     UserDefinition. EntityRepository strips sensitive fields from all query
//     results for security reasons. Login requires the raw hash for bcrypt
//     comparison. EntityRepository cannot be extended to expose sensitive fields
//     without creating a general security regression.
//
//  2. Cross-entity joins — [sqlLookupAPIToken] spans two tables (iam_api_tokens
//     JOIN iam_service_accounts). EntityRepository is scoped to a single entity
//     type; cross-entity joins are not expressible through its interface.
//
//  3. Junction tables — [sqlLoadRolePermissions] and [sqlLoadUserRoles] read
//     junction tables (iam_role_permissions, iam_user_roles) that are not
//     registered as EntityDefinitions. Registering them purely to use
//     EntityRepository would create unnecessary public API surface for tables
//     that are intentionally internal.
//
// All other SQL — session inserts, login audit inserts, revoked_at updates —
// SHOULD migrate to EntityRepository in Phase N+6C. They are listed here for
// isolation now; Phase N+6C will move them entirely.
//
// # Why set_tenant_context is called explicitly
//
// Every tenant-scoped operation must call set_tenant_context($1) at the start
// of the transaction to activate PostgreSQL Row Level Security for that
// connection. This is a framework invariant (see CLAUDE.md). The contrib/pgx
// EntityRepository implementation calls set_tenant_context automatically.
// AuthService uses raw pgx transactions and must call it explicitly here.
// When the operations below migrate to EntityRepository, these explicit calls
// will be eliminated.
//
// # SQL safety
//
// All statements use positional parameters ($1, $2 …). No SQL is constructed
// by string concatenation. All statements have been reviewed for injection risk.
package iam

// ── Tenant context ─────────────────────────────────────────────────────────────

// sqlSetTenantContext activates PostgreSQL Row Level Security for the current
// transaction by setting the transaction-local configuration variable
// app.current_tenant_id. Called at the start of every tenant-scoped transaction.
//
// Why SQL: set_tenant_context() is a PostgreSQL stored procedure — it cannot
// be called through any Go-level abstraction without still emitting this SQL.
//
// EntityRepository replacement: EntityRepository calls this internally. Once
// session/audit writes migrate to EntityRepository (Phase N+6C), this call
// will be eliminated from AuthService entirely.
//
// Limitation: PostgreSQL (stored procedure call — no Go equivalent).
const sqlSetTenantContext = `SELECT set_tenant_context($1)`

// ── Authentication ─────────────────────────────────────────────────────────────

// sqlLookupCredentials retrieves the fields required for bcrypt password
// verification for a user identified by email within the current RLS tenant
// context.
//
// Why SQL: EntityRepository strips [password_hash] because UserDefinition
// declares it Sensitive:true. Credential verification requires the raw hash.
// Exposing sensitive fields through EntityRepository would require a general
// security bypass — that is not acceptable.
//
// EntityRepository replacement: Not possible without a dedicated
// LookupCredentials method on a DatabaseQuerier interface (Phase N+6D).
//
// Limitation: Security — sensitive field policy in EntityRepository.
const sqlLookupCredentials = `
	SELECT id, password_hash, status
	FROM iam_users
	WHERE email = $1
`

// sqlUpdateLastLoginAt records the wall-clock time of the most recent
// successful login on the iam_users row. Non-critical — used for analytics
// and security dashboards.
//
// Why SQL: Direct single-field update by primary key. Could be replaced by
// EntityRepository.Update(userID, UpdateInput{Data: {"last_login_at": t}})
// once Phase N+6C is implemented.
//
// EntityRepository replacement: Yes — straightforward Update call.
// Will migrate in Phase N+6C.
//
// Limitation: Framework (EntityRepository not yet used here — transitional).
const sqlUpdateLastLoginAt = `UPDATE iam_users SET last_login_at = $1 WHERE id = $2`

// ── Session management ─────────────────────────────────────────────────────────

// sqlInsertSessionRecord creates the SQL audit record for a newly issued
// session. This mirrors the session stored in Redis (the authoritative store)
// but persists it in PostgreSQL for forensic investigation, forced-logout
// queries, and compliance reporting.
//
// gen_random_uuid() is called by PostgreSQL to produce the row ID. The token
// is never stored raw — only its SHA-256 hex digest (token_hash) is persisted.
//
// Why SQL: Could be replaced by EntityRepository.Create on the iam_session
// entity definition. Will migrate in Phase N+6C.
//
// EntityRepository replacement: Yes — direct Create call.
// Will migrate in Phase N+6C.
//
// Limitation: Framework (EntityRepository not yet used here — transitional).
const sqlInsertSessionRecord = `
	INSERT INTO iam_sessions
		(id, tenant_id, token_hash, user_id, service_account_id,
		 issued_at, expires_at, device_id, ip_address)
	VALUES
		(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
`

// sqlRevokeSessionByHash marks a single session record revoked in the SQL
// audit trail. Called during Logout after Redis revocation has succeeded.
// The SQL update is best-effort — Redis DEL is the authoritative revocation.
//
// Matches on token_hash (SHA-256 hex of raw token) and tenant_id (RLS guard).
//
// Why SQL: EntityRepository.BulkUpdate with filter.Eq("token_hash", hash)
// would express this correctly. Will migrate in Phase N+6C.
//
// EntityRepository replacement: Yes — BulkUpdate(filter.Eq("token_hash",h), patch).
// Will migrate in Phase N+6C.
//
// Limitation: Framework (EntityRepository not yet used here — transitional).
const sqlRevokeSessionByHash = `
	UPDATE iam_sessions SET revoked_at = now()
	WHERE token_hash = $1 AND tenant_id = $2
`

// sqlRevokeSessionsByHashes marks all sessions in the hash list as revoked.
// Called during RevokeUserSessions (role change, admin forced-logout) after
// Redis DEL has succeeded. The revoked_at IS NULL guard prevents double-updates.
//
// Uses PostgreSQL's = ANY($1) array operator for batch matching.
//
// Why SQL: EntityRepository.BulkUpdate with filter.In("token_hash", hashes)
// would express this correctly, though the = ANY($1) form is more efficient
// for large arrays. Will migrate in Phase N+6C; array efficiency is noted.
//
// EntityRepository replacement: Yes — BulkUpdate(filter.In("token_hash",hashes), patch).
// Phase N+6C should confirm the contrib/pgx BulkUpdate implementation emits
// = ANY($1) for In filters to preserve this efficiency.
//
// Limitation: Framework (EntityRepository not yet used here — transitional).
const sqlRevokeSessionsByHashes = `
	UPDATE iam_sessions SET revoked_at = now()
	WHERE token_hash = ANY($1) AND tenant_id = $2 AND revoked_at IS NULL
`

// ── API token validation ───────────────────────────────────────────────────────

// sqlLookupAPIToken retrieves service account identity and token metadata for
// a given token_hash. Joins iam_api_tokens and iam_service_accounts in a
// single query to avoid two round-trips on the hot-path validation flow.
//
// Why SQL: This is a cross-entity join. EntityRepository is scoped to a single
// entity type — iam_api_token or iam_service_account, not both together.
// Performing two separate EntityRepository queries and joining in Go would
// introduce a TOCTOU risk (service account status could change between reads)
// and double the latency on a hot path.
//
// EntityRepository replacement: Not directly possible. Requires a dedicated
// DatabaseQuerier.LookupAPIToken method (Phase N+6D) which encapsulates this
// join behind a typed interface.
//
// Limitation: Framework — EntityRepository does not support cross-entity joins.
// This is a deliberate framework boundary. The join should stay in an adapter
// (contrib/pgx) behind a DatabaseQuerier interface, not be broken into two
// sequential EntityRepository calls.
const sqlLookupAPIToken = `
	SELECT
		at.service_account_id,
		at.expires_at,
		at.is_revoked,
		sa.status
	FROM iam_api_tokens at
	JOIN iam_service_accounts sa ON sa.id = at.service_account_id
	WHERE at.token_hash = $1
`

// ── Roles ──────────────────────────────────────────────────────────────────────

// sqlLoadRolePermissions loads all role→permission bindings from the global
// iam_role_permissions table. Called at startup to initialise CasbinEvaluator
// Phase 2, and after any permission assignment change.
//
// iam_role_permissions is a global table — no tenant_id, no RLS. The query is
// issued directly on the pool (no transaction, no set_tenant_context).
//
// Why SQL: iam_role_permissions is not a registered EntityDefinition. It is a
// bootstrap table whose contents must be loaded before the full entity pipeline
// is available at startup. Registering it as an entity purely for EntityRepository
// access would create unnecessary public API surface for an internal table.
//
// EntityRepository replacement: Possible if iam_role_permissions is registered
// as an entity. Not recommended — bootstrap tables read before entity pipeline
// initialisation should remain outside the entity lifecycle.
//
// Limitation: IAM — intentional internal bootstrap table, not an ERP entity.
const sqlLoadRolePermissions = `
	SELECT role_name, permission_identifier
	FROM iam_role_permissions
	ORDER BY role_name, permission_identifier
`

// sqlLoadUserRoles retrieves all role names assigned to a user in the current
// tenant. Must be called within a transaction that has already established RLS
// context via set_tenant_context. The roles are embedded in the session and
// drive Casbin policy evaluation on every authenticated request.
//
// Why SQL: iam_user_roles is a junction table tracking user↔role assignments.
// It is not a registered EntityDefinition. Like iam_role_permissions, it is
// an internal IAM table that should not be exposed as an entity.
//
// EntityRepository replacement: Possible if iam_user_roles is registered as an
// entity with a composite FK to iam_users. Not recommended for the same reason
// as iam_role_permissions — it is an internal IAM implementation detail.
//
// Limitation: IAM — intentional internal junction table, not an ERP entity.
const sqlLoadUserRoles = `
	SELECT role_name FROM iam_user_roles
	WHERE tenant_id = $1 AND user_id = $2
`

// ── Login audit ────────────────────────────────────────────────────────────────

// sqlInsertLoginAudit writes an authentication event record to iam_login_audits.
// This table is append-only; updates and deletes are prohibited by the
// LoginAuditImmutableGuard hook and the absence of Write/Delete permissions.
//
// gen_random_uuid() produces the row ID in PostgreSQL. now() records the exact
// server timestamp for the event (not the Go time.Now() which could differ from
// DB time under NTP skew).
//
// Why SQL: Could be replaced by EntityRepository.Create on the iam_login_audit
// entity definition. Will migrate in Phase N+6C.
//
// Note: In Phase N+6F, iam_login_audits will be retired in favour of the
// unified audit_event table. The new table will also use EntityRepository.Create.
//
// EntityRepository replacement: Yes — direct Create call.
// Will migrate in Phase N+6C; table will be unified in Phase N+6F.
//
// Limitation: Framework (EntityRepository not yet used here — transitional).
const sqlInsertLoginAudit = `
	INSERT INTO iam_login_audits
		(id, tenant_id, event, user_id, service_account_id,
		 ip_address, device_id, failure_reason, created_at)
	VALUES
		(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now())
`
