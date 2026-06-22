---
title: "Chapter 11: Database Migrations"
part: "Part II — The EntityDefinition System"
chapter: 11
section: "11-database-migrations"
related:
  - "[Chapter 8: The Persistence Interface](08-persistence-interface.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
---

# Chapter 11: Database Migrations

Awo uses Atlas CLI as its migration engine. All schema changes flow through versioned, reviewed, human-readable SQL files. Auto-migrate (applying schema changes without review) is available for development but is explicitly disabled in production builds.

---

## 11.1. Migration Strategy Overview

### 11.1.1. Why Manual Reviewed Migrations in Production

Auto-migrate tools (ent's `client.Schema.Create`, GORM's `AutoMigrate`) are convenient in development but dangerous in production for several reasons:

- **Silent destructive operations**: dropping a column or index happens without warning
- **No rollback**: most auto-migrate tools have no down-migration concept
- **No drift detection**: if someone manually altered a table, auto-migrate may silently diverge
- **No review gate**: a developer pushing code on Friday night auto-migrates a 50M-row table in production

Awo's production migration path requires:
1. A human-reviewed SQL file committed to version control
2. A matching down migration for every up migration
3. CI validation that migrations are safe (Atlas lint)
4. Explicit application via `awo entity migrate --apply`

### 11.1.2. Atlas as the Migration Tool

Atlas CLI handles the diff between the desired schema (ent-generated) and the current database state (inspected from a running DB or a dev DB). It generates the SQL, you review it, it applies it.

Atlas does NOT understand Awo business logic — it only knows SQL. Awo's `awo entity migrate` command wraps Atlas with multi-tenant awareness.

### 11.1.3. The Migration Lifecycle

```
1. Developer changes ent schema
2. awo entity migrate --diff        → generates versioned SQL file
3. Developer reviews SQL file
4. PR review + CI Atlas lint
5. Merge to main
6. Deploy: awo entity migrate --apply  → applies to all tenant schemas
7. awo entity migrate --verify         → confirms applied state matches files
```

---

## 11.2. The `awo entity migrate` Command

### 11.2.1. What It Does Under the Hood

`awo entity migrate` connects to your development database, inspects the current schema, compares it against the desired schema generated from ent schema files, and delegates to Atlas for diff/apply/lint operations.

### 11.2.2. `--dry-run` — Preview SQL Without Writing

```bash
awo entity migrate --dry-run
# Prints the SQL that would be applied. No files written, no DB changes.
```

### 11.2.3. `--diff` — Generate a Versioned Migration File

```bash
awo entity migrate --diff --name="add_kra_customs_code_to_invoices"
# Creates: db/migration/20250704120000_add_kra_customs_code_to_invoices.sql
```

The timestamp prefix ensures migrations apply in creation order.

### 11.2.4. `--apply` — Execute Pending Migrations

```bash
awo entity migrate --apply
# Applies all pending migrations to all tenant schemas in sequence
```

### 11.2.5. `--verify` — Drift Detection

```bash
awo entity migrate --verify
# Compares applied migration checksums against migration files.
# Exits non-zero if drift detected.
```

---

## 11.3. The Versioned Migration File Format

### 11.3.1. File Naming

```
db/migration/
  20250101000000_initial_schema.sql
  20250215083000_add_invoice_status.sql
  20250704120000_add_kra_customs_code_to_invoices.sql
```

Format: `{YYYYMMDDHHMMSS}_{description_slug}.sql`

### 11.3.2. File Structure

```sql
-- 20250704120000_add_kra_customs_code_to_invoices.sql
-- atlas:sum h1:abc123...  (checksum — do not edit)

-- +goose Up
ALTER TABLE invoices ADD COLUMN kra_customs_code varchar(20);
CREATE INDEX CONCURRENTLY ON invoices(kra_customs_code)
    WHERE kra_customs_code IS NOT NULL;

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS invoices_kra_customs_code_idx;
ALTER TABLE invoices DROP COLUMN IF EXISTS kra_customs_code;
```

Both Up and Down sections are required. A migration without a Down section fails CI validation.

### 11.3.3. Down Migration — Required

Down migrations enable partial rollback (rolling back one bad release) and schema reversibility testing in CI. They are also the escape hatch when `--apply` fails mid-fleet.

### 11.3.4. Checksum — Tamper Detection

Atlas embeds a checksum (`atlas:sum`) in each migration file. If the file is modified after generation, `--verify` and `--apply` reject it with `ErrChecksumMismatch`. This prevents accidental or malicious editing of applied migrations.

To legitimately edit a not-yet-applied migration: delete the file and regenerate with `--diff`.

### 11.3.5. Safe vs Unsafe Manual Edits

**Safe to edit after generation (before application):**
- Adding comments
- Adding `IF NOT EXISTS` guards
- Changing `CREATE INDEX` to `CREATE INDEX CONCURRENTLY`
- Changing column defaults

**Never edit:**
- The `atlas:sum` line
- Already-applied migrations
- Operations that change the schema differently than ent expects

---

## 11.4. Multi-Tenant Migration Strategy

### 11.4.1. Schema-per-Tenant Layout

Each tenant has a dedicated PostgreSQL schema: `tenant_{uuid}`. The platform data lives in the `public` schema. Migrations must be applied to every tenant schema.

```
public schema:
  tenants, users, feature_flags, ...

tenant_a1b2c3d4 schema:
  invoices, customers, products, ...

tenant_e5f6g7h8 schema:
  invoices, customers, products, ...
```

### 11.4.2. Migrating All Tenant Schemas

`awo entity migrate --apply` iterates over all active tenants from the `tenants` table and applies pending migrations to each tenant's schema in sequence:

```go
// Pseudocode for multi-tenant migration
tenants := listActiveTenants(ctx)
for _, tenant := range tenants {
    setSearchPath(conn, tenant.SchemaName)
    if err := applyPendingMigrations(conn, migrationDir); err != nil {
        log.Errorf("migration failed for tenant %s: %v", tenant.ID, err)
        return err  // stop on first failure
    }
    log.Infof("migrated tenant %s successfully", tenant.ID)
}
```

### 11.4.3. Sequential by Default

Migrations apply to tenants **sequentially** (not concurrently) by default. Reasons:

- A failed migration for tenant 5 is detectable before tenant 6 is affected
- Sequential execution allows clean rollback of a bad migration to a known state
- Parallel migration multiplies connection pressure — with 1000 tenants, parallel = 1000 concurrent DDL operations

For very large tenant fleets (>500 tenants), use `--concurrency=N` to apply to N tenants in parallel, with automatic pause-on-failure.

### 11.4.4. Handling a Failed Migration Mid-Fleet

If migration fails at tenant #47 out of 200:

```bash
# Check which tenants are at which migration version
awo entity migrate --status

# Apply only to specific tenant for diagnosis
awo entity migrate --apply --tenant=a1b2c3d4

# After fixing, resume from tenant #47
awo entity migrate --apply --skip-migrated
```

Never re-run `--apply` against already-migrated tenants — Atlas checksums prevent double-application, but the attempt wastes time and logs noise.

---

## 11.5. Rolling Migrations — Zero-Downtime Patterns

### 11.5.1. Expand-Contract Pattern

The safest zero-downtime column change:

**Phase 1 — Expand** (deploy with old code reading old column):
```sql
ALTER TABLE invoices ADD COLUMN new_status varchar(50);
```

**Phase 2 — Dual-write** (deploy new code writing both columns):
```go
// Write both old and new column
record.Set("status", newStatus)
record.Set("new_status", newStatus)
```

**Phase 3 — Contract** (deploy new code reading new column only):
```sql
ALTER TABLE invoices DROP COLUMN status;
ALTER TABLE invoices RENAME COLUMN new_status TO status;
```

This approach ensures zero downtime across three deployments, at the cost of temporarily storing data twice.

### 11.5.2. Multi-Phase Column Renames

Never rename a column in a single migration in a live system. Use expand-contract (§11.5.1). The single-step rename pattern (`ALTER TABLE ... RENAME COLUMN`) is only safe with a full maintenance window.

### 11.5.3. `CREATE INDEX CONCURRENTLY`

Always use `CONCURRENTLY` for index creation in production:

```sql
-- BLOCKS TABLE ACCESS — never in production
CREATE INDEX ON invoices(customer_id);

-- NON-BLOCKING — always use this
CREATE INDEX CONCURRENTLY ON invoices(customer_id);
```

Atlas lint (`atlas migrate lint`) flags non-concurrent index creation as a warning. CI should fail on this warning for production migration files.

### 11.5.4. Operations That Require Downtime

| Operation | Downtime required | Alternative |
|---|---|---|
| Adding NOT NULL to existing column | Yes (unless default set first) | Add nullable → backfill → add constraint |
| Dropping a column | No (data loss is immediate though) | Soft-drop: mark unused, remove later |
| Changing column type | Usually yes | Add new column, dual-write, swap |
| Adding a FK constraint | No (use `NOT VALID` then `VALIDATE`) | |
| `VACUUM FULL` | Yes | Schedule maintenance window |

---

## 11.6. Drift Detection

### 11.6.1. What Drift Is

Drift occurs when the live database schema differs from what the migration files would produce. Common causes:
- A developer ran manual `ALTER TABLE` in production to "quick-fix" an issue
- An ORM ran auto-migrate (should be disabled in production)
- A migration file was edited after application

### 11.6.2. Running Drift Detection in CI

```yaml
# .github/workflows/ci.yml
- name: Check migration drift
  run: awo entity migrate --verify
  env:
    DATABASE_URL: ${{ secrets.CI_DATABASE_URL }}
```

The CI database is seeded by applying all migration files from scratch. `--verify` then confirms that the resulting schema matches what Atlas would generate from the current ent schemas.

### 11.6.3. Resolving Drift

If drift is detected:
1. **Identify the change**: `awo entity migrate --status --verbose` shows which objects differ
2. **If the manual change is correct**: generate a new migration file that formalises it
3. **If the manual change is incorrect**: revert it manually, then run `--verify` again

Never edit an applied migration file to paper over drift. The checksum will mismatch and `--apply` will refuse to run.

---

## 11.7. Migration Rollback

### 11.7.1. Running a Down Migration

```bash
# Roll back the last applied migration across all tenants
awo entity migrate --rollback --steps=1

# Roll back to a specific version
awo entity migrate --rollback --to=20250215083000
```

### 11.7.2. Partial Rollback — One Tenant

```bash
awo entity migrate --rollback --steps=1 --tenant=a1b2c3d4
```

Useful for isolating a problematic tenant after a failed migration, while leaving other tenants at the new version.

### 11.7.3. When Rollback Is Not Possible

Destructive operations cannot be truly rolled back:
- `DROP TABLE` — data is gone
- `DROP COLUMN` — column data is gone
- `TRUNCATE` — data is gone
- `DELETE WHERE` — deleted rows are gone

For these cases, the down migration can restore the structure (recreate the table/column) but not the data. This is why:
1. **Never use `DROP TABLE` in production migrations** without verified backup
2. **Always take a snapshot before destructive migrations**
3. **The correct zero-downtime alternative is always "soft removal first"**

---

## 11.8. Atlas CI Integration

### 11.8.1. Atlas Migrate Lint in CI

```bash
atlas migrate lint --dev-url="postgres://..." --dir="db/migration"
```

Atlas lint detects:
- Missing down migrations
- Destructive operations (DROP TABLE, DROP COLUMN)
- Non-concurrent index creation
- Missing `IF NOT EXISTS` guards
- Lock-heavy operations (`ALTER TABLE ... ALTER COLUMN TYPE`)

Configure lint to fail CI on destructive operations and warn on lock-heavy operations.

### 11.8.2. Gate Merges on Unapproved Migrations

Add a required CI check: `migration-review`. This check:
1. Detects if any new migration files were added in the PR
2. Requires a `migration-approved` label from a designated DB reviewer before allowing merge
3. Runs Atlas lint and fails if any errors are found

```yaml
# .github/workflows/migration-gate.yml
on: pull_request

jobs:
  migration-review:
    runs-on: ubuntu-latest
    steps:
      - name: Check for new migrations
        run: |
          NEW_MIGRATIONS=$(git diff --name-only origin/main...HEAD -- 'db/migration/*.sql')
          if [ -n "$NEW_MIGRATIONS" ]; then
            echo "New migrations detected: $NEW_MIGRATIONS"
            # Check for required label
            gh pr view ${{ github.event.pull_request.number }} \
              --json labels --jq '.labels[].name' | grep -q 'migration-approved' \
              || (echo "Missing migration-approved label" && exit 1)
          fi
```
