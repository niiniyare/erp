[<-- Back to Index](README.md)

## RLS: End-to-End Flow, Testing, and Troubleshooting

### End-to-End Request Flow

Every API request that touches tenant data goes through the same chain before a single row is read or written.

```
HTTP Request (e.g. GET /api/v1/invoices)
│
├── 1. Router resolves tenant from subdomain / header / JWT claim
│       → tenant slug: "acme-corp"
│
├── 2. Middleware: TenantContextMiddleware
│       → SELECT id FROM tenants WHERE slug = 'acme-corp' AND deleted_at IS NULL
│       → tenant_id = 'xxxxxxxx-...'
│
├── 3. DB connection acquired from pool
│       → conn.Exec("SELECT set_tenant_context($1)", tenant_id)
│
│       set_tenant_context() internals:
│         PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);
│                                                                          ^^^^
│                                                               is_local = true
│                                                        (resets at transaction end)
│
├── 4. Application query executes (no WHERE clause needed):
│       SELECT * FROM invoices;
│
│       PostgreSQL evaluates RLS policy BEFORE returning rows:
│         USING (tenant_id = current_tenant_id())
│
│       current_tenant_id() reads:
│         current_setting('app.current_tenant_id', true)::uuid
│
│       Effective SQL (internal rewrite):
│         SELECT * FROM invoices
│         WHERE tenant_id = 'xxxxxxxx-...'   ← injected by RLS
│
├── 5. Rows returned: only rows where tenant_id matches
│
└── 6. Connection returned to pool
        → transaction ends → is_local=true resets app.current_tenant_id automatically
```

### Policy Evaluation Rules

```
Role              Policy                    Effect
─────────────────────────────────────────────────────────────────
application_role  tenant_isolation_policy   tenant_id = current_tenant_id()
                                            (both USING and WITH CHECK)
admin_role        admin_full_access_policy  USING (true) — sees all rows
readonly_role     readonly_access_policy    SELECT only, USING (true)
superuser         (bypasses RLS entirely)   sees all rows regardless
─────────────────────────────────────────────────────────────────
```

> **Important**: If `SET ROLE` is not called and the connection runs as a superuser
> (e.g. `postgres`), RLS is bypassed. Always connect as `application_role` in tests.

---

### Testing RLS

#### Prerequisites

```sql
-- Confirm the DB roles exist
SELECT rolname FROM pg_roles
WHERE rolname IN ('application_role', 'admin_role', 'readonly_role');

-- Confirm RLS is enabled on a representative table
SELECT relname, relrowsecurity, relforcerowsecurity
FROM pg_class
WHERE relname IN ('invoices', 'persons', 'user_sessions', 'entities');
-- relrowsecurity should be TRUE for all of these
```

#### Test 1 — No context returns 0 rows

```sql
BEGIN;
  SET ROLE application_role;

  -- No set_tenant_context call → current_tenant_id() returns NULL
  SELECT COUNT(*) FROM persons;
  -- Expected: 0  (NULL ≠ any tenant_id, so no rows pass the policy)

  SELECT current_setting('app.current_tenant_id', true);
  -- Expected: '' (empty string)
ROLLBACK;
```

#### Test 2 — Per-tenant isolation

```sql
-- Get two tenant IDs from the seed data
SELECT id, slug FROM tenants WHERE slug IN ('acme-corp', 'globex') ORDER BY slug;
-- Copy the UUIDs: acme_id and globex_id

BEGIN;
  SET ROLE application_role;

  -- Set context to ACME
  SELECT set_tenant_context('<acme_id>');
  SELECT COUNT(*) FROM persons;             -- only ACME persons
  SELECT DISTINCT tenant_id FROM persons;   -- should show only acme_id

  -- Switch context to Globex
  SELECT set_tenant_context('<globex_id>');
  SELECT COUNT(*) FROM persons;             -- only Globex persons
  SELECT DISTINCT tenant_id FROM persons;   -- should show only globex_id

ROLLBACK;
```

#### Test 3 — Admin role bypasses RLS

```sql
BEGIN;
  SET ROLE admin_role;
  -- No set_tenant_context needed
  SELECT COUNT(*) FROM persons;             -- ALL persons across all tenants
  SELECT COUNT(DISTINCT tenant_id) FROM persons; -- should be > 1
ROLLBACK;
```

#### Test 4 — Context is transaction-local

```sql
-- Transaction 1: set context and confirm it's visible within the transaction
BEGIN;
  SET ROLE application_role;
  SELECT set_tenant_context('<acme_id>');
  SELECT current_setting('app.current_tenant_id', true); -- shows acme_id
ROLLBACK;  -- transaction ends → context reset

-- Transaction 2: context is gone
BEGIN;
  SET ROLE application_role;
  SELECT current_setting('app.current_tenant_id', true); -- shows '' (empty)
  SELECT COUNT(*) FROM persons;                           -- 0
ROLLBACK;
```

#### Test 5 — WITH CHECK blocks cross-tenant writes

```sql
BEGIN;
  SET ROLE application_role;
  SELECT set_tenant_context('<acme_id>');

  -- Try to insert a row with a DIFFERENT tenant_id
  INSERT INTO persons (id, tenant_id, entity_id, first_name, last_name, email, type)
  VALUES (
    gen_random_uuid(),
    '<globex_id>',      -- different tenant!
    (SELECT uuid FROM entities WHERE tenant_id = '<acme_id>' LIMIT 1),
    'Evil', 'Hacker', 'evil@example.com', 'EMPLOYEE'
  );
  -- Expected: ERROR: new row violates row-level security policy for table "persons"

ROLLBACK;
```

#### Test 6 — enforce_tenant_isolation trigger

```sql
BEGIN;
  SET ROLE application_role;
  SELECT set_tenant_context('<acme_id>');

  -- Try to reference an entity that belongs to a different tenant
  INSERT INTO persons (id, tenant_id, entity_id, first_name, last_name, email, type)
  VALUES (
    gen_random_uuid(),
    '<acme_id>',
    (SELECT uuid FROM entities WHERE tenant_id = '<globex_id>' LIMIT 1),  -- cross-tenant entity!
    'Jane', 'Doe', 'jane@acme.com', 'EMPLOYEE'
  );
  -- Expected: ERROR: Entity <uuid> does not belong to tenant <acme_id>

ROLLBACK;
```

#### Quick sanity check (copy-paste DO block)

```sql
DO $$
DECLARE
  v_acme  UUID;
  v_globex UUID;
  n_acme  INT;
  n_globex INT;
  n_none  INT;
BEGIN
  SELECT id INTO v_acme  FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex FROM tenants WHERE slug = 'globex';

  -- Role must be application_role for RLS to apply
  -- (cannot SET ROLE inside DO block — run this as application_role user)

  PERFORM set_tenant_context(v_acme);
  SELECT COUNT(*) INTO n_acme FROM persons;

  PERFORM set_tenant_context(v_globex);
  SELECT COUNT(*) INTO n_globex FROM persons;

  PERFORM clear_tenant_context();
  SELECT COUNT(*) INTO n_none FROM persons;

  RAISE NOTICE 'ACME persons: %, Globex persons: %, No-context persons: %',
    n_acme, n_globex, n_none;

  -- Assertions
  ASSERT n_acme  > 0,   'ACME should have persons';
  ASSERT n_globex > 0,  'Globex should have persons';
  ASSERT n_none  = 0,   'No context must return 0 rows';
  ASSERT n_acme != n_globex OR n_acme = n_globex,
    'Pass (counts may coincidentally match)';

  RAISE NOTICE 'RLS isolation: PASS';
END $$;
-- Run this connected as application_role, not superuser
```

---

### Troubleshooting RLS

#### Issue 1 — RLS is not filtering rows (everyone sees everything)

```
SYMPTOM:
  application_role query returns rows from all tenants
  even after set_tenant_context() is called

DIAGNOSE:
  -- Check RLS is enabled on the table
  SELECT relname, relrowsecurity
  FROM pg_class
  WHERE relname = '<table_name>';
  -- relrowsecurity must be TRUE

  -- Check policies exist for application_role
  SELECT policyname, cmd, roles, qual
  FROM pg_policies
  WHERE tablename = '<table_name>';
  -- Should show tenant_isolation_policy FOR application_role

  -- Confirm current role
  SELECT current_user, session_user;
  -- If current_user = 'postgres' or any superuser → RLS is bypassed!

RESOLUTION:
  a) If relrowsecurity = FALSE:
     ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;
     (check migration for this table; may be missing)

  b) If no policy for application_role:
     The migration that creates the policy may not have run.
     Run: SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;

  c) If running as superuser:
     Connect as application_role:  SET ROLE application_role;
     Or use: ALTER TABLE <table_name> FORCE ROW LEVEL SECURITY;
     (forces RLS even for table owner — use only in dev)
```

#### Issue 2 — Empty context returns rows (should return 0)

```
SYMPTOM:
  Without calling set_tenant_context(), queries still return rows

DIAGNOSE:
  SELECT current_setting('app.current_tenant_id', true);
  -- If this returns a non-empty UUID, a previous transaction
  -- leaked context (is_local=false was used somewhere)

  -- Check if set_config was called with is_local = false
  -- This makes the setting persist beyond the transaction
  SHOW app.current_tenant_id;

RESOLUTION:
  a) Never call set_config('app.current_tenant_id', val, FALSE)
     Always use is_local = true (the set_tenant_context() function does this)

  b) If leaked context is found, reset it:
     SELECT clear_tenant_context();
     -- or:
     SELECT set_config('app.current_tenant_id', '', false);

  c) If the problem is in Go code, ensure every DB call uses
     set_tenant_context() inside the same transaction, not as
     a separate connection-level call.
```

#### Issue 3 — INSERT fails with "row-level security policy" error

```
SYMPTOM:
  ERROR: new row violates row-level security policy for table "X"

CAUSES:
  a) Inserting with tenant_id that doesn't match current context
  b) Forgetting to set the tenant_id column on insert
  c) The WITH CHECK clause rejects the row

DIAGNOSE:
  -- What context is set?
  SELECT current_setting('app.current_tenant_id', true);

  -- What tenant_id is in the insert?
  -- Compare the two — they must match

RESOLUTION:
  Ensure the row being inserted has tenant_id = current_tenant_id().
  The application layer should always populate tenant_id from the
  authenticated tenant context, never from user-supplied input.
```

#### Issue 4 — Cross-tenant data leak (rows from wrong tenant visible)

```
SYMPTOM:
  User of Tenant A can see data belonging to Tenant B

DIAGNOSE:
  Step 1: Confirm RLS is enabled
    SELECT relname, relrowsecurity FROM pg_class WHERE relname = '<table>';

  Step 2: Confirm the policy uses current_tenant_id()
    SELECT qual FROM pg_policies
    WHERE tablename = '<table>' AND roles @> ARRAY['application_role'];
    -- qual should contain: (tenant_id = current_tenant_id())

  Step 3: Check if any query bypasses the ORM / uses raw SQL
    -- Raw queries that join across tenants can pull in foreign rows
    -- Example leak: SELECT i.* FROM invoices i JOIN orders o ON i.order_id = o.id
    --               if orders has a cross-tenant entry, invoices may leak

  Step 4: Check enforce_tenant_isolation trigger is present
    SELECT trigger_name, event_manipulation, event_object_table
    FROM information_schema.triggers
    WHERE trigger_name = 'check_tenant_isolation';

RESOLUTION:
  a) If policy is missing or wrong → re-run the migration for that table
  b) If a raw query bypasses RLS → refactor to go through the ORM or
     use CTEs that stay within tenant scope
  c) If trigger is missing → re-run migration 000201 (or wherever it's defined)
```

#### Issue 5 — `current_tenant_id()` returns NULL unexpectedly

```
SYMPTOM:
  SELECT current_tenant_id() returns NULL mid-request

CAUSES:
  a) set_tenant_context() was never called
  b) The connection was returned to pool and reused without re-setting context
  c) An explicit COMMIT was called in application code mid-transaction,
     causing is_local=true settings to reset

DIAGNOSE:
  -- In the problematic transaction:
  SELECT current_setting('app.current_tenant_id', true);
  -- If empty → context was never set or was reset

RESOLUTION:
  a) Ensure middleware always calls set_tenant_context() at the start
     of every request, within the same transaction that runs queries.
  b) Never commit mid-request. The tenant context is transaction-local;
     a mid-request COMMIT clears it.
  c) If using pgbouncer or a connection pool in transaction mode,
     ensure set_tenant_context() is the FIRST statement in each transaction.
```

#### Issue 6 — `validate_and_set_tenant_context()` raises "Tenant is not active"

```
SYMPTOM:
  EXCEPTION: Tenant is not active: <uuid> (status: SUSPENDED)

DIAGNOSE:
  SELECT id, "Status", deleted_at FROM tenants WHERE id = '<uuid>';

RESOLUTION:
  PENDING  → Tenant not yet activated. Use ActivateTenant() API or admin action.
  SUSPENDED → Reactivate via admin panel or ReactivateTenant() service method.
  ARCHIVED  → Cannot be reactivated. Restore from backup if needed.

  For testing, reset status directly:
    UPDATE tenants SET "Status" = 'ACTIVE' WHERE id = '<uuid>';
    -- Only in dev/test environments
```

#### Issue 7 — RLS slowing down queries

```
SYMPTOM:
  Queries on tenant-scoped tables are slow; EXPLAIN shows Seq Scan

DIAGNOSE:
  EXPLAIN (ANALYZE, BUFFERS)
  SELECT * FROM persons WHERE first_name = 'John';
  -- Check if tenant_id index is being used

  -- Confirm index exists
  SELECT indexname, indexdef
  FROM pg_indexes
  WHERE tablename = 'persons' AND indexdef LIKE '%tenant_id%';

RESOLUTION:
  Every tenant-scoped table should have a composite index with tenant_id
  as the leading column:

    CREATE INDEX IF NOT EXISTS idx_persons_tenant
    ON persons(tenant_id, id);

  For queries filtered by other columns:
    CREATE INDEX IF NOT EXISTS idx_persons_tenant_email
    ON persons(tenant_id, email);

  The RLS WHERE clause adds tenant_id = current_tenant_id() to every query.
  If there's no index on tenant_id, PostgreSQL will do a full table scan
  before applying the filter.
```

---

Next: [Tenant Context Management](./09-tenant-context-management.md)
