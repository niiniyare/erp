---
title: Data Security
portal: 7 — Security
section: 07-security
audience: [backend-engineer, security, architect]
related:
  - "[Security Overview](01-security-overview.md)"
  - "[RLS Enforcement](../03-platform-architecture/01-multi-tenancy/02-rls-enforcement.md)"
  - "[Schema Conventions](../03-platform-architecture/03-data-architecture/02-schema-conventions.md)"
---

# Data Security

## Tenant Isolation: PostgreSQL RLS

Row-Level Security is the primary data isolation mechanism. Every table has a policy:

```sql
-- Enable RLS
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

-- Policy: users only see their tenant's rows
CREATE POLICY tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

The GUC `app.tenant_id` is set at the start of every transaction via `SET LOCAL`:

```go
func (s *Store) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(*pgx.Tx) error) error {
    return pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
        _, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
        if err != nil {
            return err
        }
        return fn(&tx)
    })
}
```

`SET LOCAL` scopes the GUC to the transaction. It cannot "leak" to subsequent transactions in the same connection.

### RLS Bypass Risk: Superuser

PostgreSQL superusers bypass RLS. Application DB user must **never** be a superuser:

```sql
-- Application user (used by pgxpool)
CREATE ROLE awoerp_app WITH LOGIN PASSWORD '...' NOSUPERUSER;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO awoerp_app;

-- Migration user (used only during migrations)
CREATE ROLE awoerp_migrate WITH LOGIN PASSWORD '...' NOSUPERUSER;
GRANT ALL ON ALL TABLES IN SCHEMA public TO awoerp_migrate;
```

Migrations use a separate role. The `awoerp_app` role has no DDL permissions.

## Sensitive Data Handling

| Data type | Storage | Transit | At rest |
|-----------|---------|---------|---------|
| Passwords | bcrypt hash only — plaintext never stored | HTTPS | DB encryption |
| Session tokens | Redis (ephemeral, not DB) | HTTPS header | Redis auth + TLS |
| Personal data (email, name) | Postgres | HTTPS | DB encryption |
| Financial amounts | `numeric(20,6)` — no float | JSON string | DB encryption |
| Audit logs | Append-only table | Internal | DB encryption |

### Never Log Sensitive Data

```go
// ❌ Never log passwords, tokens, or PII in structured log fields
logger.Info("login attempt", "password", req.Password)
logger.Info("session created", "token", token)

// ✅ Log only non-sensitive identifiers
logger.Info("login attempt", "email_domain", emailDomain(req.Email))
logger.Info("session created", "user_id", userID, "tenant_id", tenantID)
```

## SQL Injection Prevention

All queries go through SQLC-generated code. SQLC compiles parameterized queries:

```go
// SQLC generates this — $1, $2 are always parameters, never concatenated
const getByID = `SELECT id, title, status FROM contracts WHERE id = $1 AND tenant_id = $2`
```

Rules:
- Never build SQL strings with `fmt.Sprintf`
- Never use `string` in a query where a UUID is expected — always use `uuid.UUID` type
- If a raw query is unavoidable, use `$N` placeholders, never string interpolation

## Input Validation

Validation happens in the handler layer before passing to the service:

```go
type CreateContractRequest struct {
    ContractNumber string          `json:"contract_number" validate:"required,max=50"`
    Title          string          `json:"title"           validate:"required,max=255"`
    TotalValue     decimal.Decimal `json:"total_value"     validate:"required,min=0"`
    Currency       string          `json:"currency"        validate:"required,len=3,uppercase"`
    StartDate      civil.Date      `json:"start_date"      validate:"required"`
}
```

UUID path params are parsed by the framework — invalid UUIDs return 400 before reaching the handler.

## Cross-Tenant Data Leakage Prevention

Defense in depth:

1. **PostgreSQL RLS** — row-level filter at DB
2. **`TenantID` in all queries** — explicit filter even with RLS (belt and suspenders)
3. **`ResolvedSession.TenantID`** — always sourced from the verified session, never from URL params or request body
4. **No `SECURITY DEFINER` functions** — all DB functions run as calling user (RLS applies)
5. **Zero-UUID guard** — `app.tenant_id = '00000000-...'` must never match real data

## Encryption at Rest

Database disk encryption handled at infrastructure level:
- Cloud: AWS RDS encryption (AES-256)
- Self-hosted: PostgreSQL data directory on encrypted volume

Redis encryption at rest:
- Use Redis Enterprise with encryption, or mount Redis data on encrypted volume
- Session data has TTL — loss is a user re-login, not a data loss event

## Audit Trail

All state-changing operations write an audit log entry:

```go
// In service, async after successful write
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    defer func() { recover() }()
    s.audit.Log(ctx, audit.Entry{
        TenantID:   sess.TenantID,
        UserID:     sess.UserID,
        Action:     "contracts.contract.create",
        ResourceID: contract.ID,
        Before:     nil,
        After:      contract,
    })
}()
```

Audit log is append-only — no UPDATE or DELETE on audit records. Separate DB user for audit writes has INSERT-only permission.

## Vulnerability Prevention Checklist

| OWASP | Control |
|-------|---------|
| A01 Broken Access Control | Casbin RBAC + RLS + EntityScope |
| A02 Cryptographic Failures | bcrypt passwords, TLS in transit, AES at rest, no MD5/SHA1 |
| A03 Injection | SQLC parameterized queries; no string concat in SQL |
| A04 Insecure Design | Fail-closed authz; no superuser app DB role |
| A05 Security Misconfiguration | No default credentials; secrets via k8s Secrets |
| A06 Vulnerable Components | Dependabot + `govulncheck` in CI |
| A07 Auth Failures | Rate limiting on login; bcrypt cost 12; instant revocation |
| A08 Software Integrity | Image signing; locked `go.sum`; pinned k8s manifests |
| A09 Logging Failures | Structured logs; no sensitive data in logs; audit trail |
| A10 Server-Side Request Forgery | No user-controlled URL fetching in server code |
