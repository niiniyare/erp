[<-- Back to Index](README.md)

## Database Architecture

### Overview

The authz module uses exactly two tables. Both live in the same PostgreSQL database as the rest of the application, managed by the standard migration system.

```markdown
Migration Files:
  000063_authz_casbin_rule.up.sql       ← casbin_rule table
  000063_authz_casbin_rule.down.sql
  000064_authz_role_assignments.up.sql  ← role_assignments table
  000064_authz_role_assignments.down.sql
```

### Table 1: `casbin_rule`

The primary policy store. Casbin reads from and writes to this table exclusively. Every enforcement decision is ultimately backed by rows in this table.

```sql
CREATE TABLE casbin_rule (
  id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  ptype VARCHAR(10) NOT NULL,        -- "p" or "g"
  v0    VARCHAR(256) NOT NULL DEFAULT '',
  v1    VARCHAR(256) NOT NULL DEFAULT '',
  v2    VARCHAR(256) NOT NULL DEFAULT '',
  v3    VARCHAR(256) NOT NULL DEFAULT '',
  v4    VARCHAR(256) NOT NULL DEFAULT '',
  v5    VARCHAR(256) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_casbin_rule_unique
  ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);
CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
```

**Column mapping by ptype:**

| ptype | v0 | v1 | v2 | v3 | v4 | v5 |
|-------|----|----|----|----|----|----|
| `p` | subject | domain | object | action | effect | — |
| `g` | user/sub | role | domain | — | — | — |

**Example rows:**

```markdown
ptype | v0                     | v1             | v2          | v3     | v4    | v5
──────┼────────────────────────┼────────────────┼─────────────┼────────┼───────┼───
p     | role:finance-manager   | a1b2c3d4-uuid  | invoice/*   | *      | allow |
p     | role:sales-rep         | a1b2c3d4-uuid  | order/*     | read   | allow |
p     | tenant:usr_sanctioned  | a1b2c3d4-uuid  | *           | *      | deny  |
g     | tenant:usr_001         | role:finance-m | a1b2c3d4-u  |        |       |
g     | tenant:usr_002         | role:sales-rep | a1b2c3d4-u  |        |       |
p     | platform:admin         | _platform_     | tenant/*    | *      | allow |
```

**Why no FK to tenants?**

The tenant's UUID is encoded in v1 (domain field) for p-rules and v2 for g-rules. We deliberately do **not** add a FK constraint because:
1. Platform domain (`_platform_`) has no corresponding tenant row
2. Casbin's adapter interface doesn't know about tenants — it sees raw strings
3. Isolation is enforced at the application layer (matching domain in the request)

**RLS on casbin_rule:**

```sql
ALTER TABLE casbin_rule ENABLE ROW LEVEL SECURITY;
CREATE POLICY casbin_rule_app   ON casbin_rule FOR ALL TO application_role USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY casbin_rule_admin ON casbin_rule FOR ALL TO admin_role        USING (TRUE) WITH CHECK (TRUE);
```

`application_role` gets unrestricted access to casbin_rule — the domain value in v1 is what provides tenant isolation at the application layer, not at the DB row-security layer. The admin_role also gets full access for direct SQL inspection.

---

### Table 2: `role_assignments`

Metadata alongside `casbin_rule`. While `casbin_rule` is the authoritative source for enforcement, `role_assignments` provides everything the UI and audit system needs: who assigned a role, who delegated it, when it expires, and whether it is still active.

```sql
CREATE TABLE role_assignments (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  subject      VARCHAR(256) NOT NULL,   -- "tenant:user_uuid"
  role_name    VARCHAR(100) NOT NULL,
  domain       VARCHAR(256) NOT NULL,   -- matches casbin domain (v2 of g-rule)
  assigned_by  VARCHAR(256),            -- "platform:admin-uuid"
  delegated_by VARCHAR(256),            -- "tenant:manager-uuid" or NULL
  expires_at   TIMESTAMPTZ,             -- NULL = permanent
  is_active    BOOLEAN      DEFAULT TRUE,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT role_assignments_unique UNIQUE (subject, role_name, domain)
);

CREATE INDEX idx_role_assignments_subject ON role_assignments(subject, domain);
CREATE INDEX idx_role_assignments_expires ON role_assignments(expires_at)
  WHERE expires_at IS NOT NULL;
```

**Why CASCADE on tenant_id?**

When a tenant is archived and then hard-deleted, all their role assignments are automatically removed. This keeps the table clean and avoids orphaned metadata rows.

**RLS on role_assignments:**

```sql
ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY ra_tenant ON role_assignments FOR ALL TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
CREATE POLICY ra_admin ON role_assignments FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
```

Unlike `casbin_rule`, the `role_assignments` table uses true row-level security scoped by `tenant_id`. Application code can only read/write assignments for the active tenant context. This is the standard pattern from the Tenant Module.

**Platform assignments:**

Platform-level role assignments (where `domain = "_platform_"`) do not have a `tenant_id` and therefore **bypass RLS** — they must be managed exclusively by `admin_role` (platform operators), never by `application_role`.

---

### Index Strategy

```markdown
INDEX                           PURPOSE
──────────────────────────────────────────────────────────────────
idx_casbin_rule_unique          Prevents duplicate p/g rules on insert
                                Also used by adapter's ON CONFLICT DO NOTHING

idx_casbin_rule_ptype           LoadPolicy filters by ptype ("p" vs "g")
                                Also useful for admin queries

idx_role_assignments_subject    Primary lookup in GetAssignments:
                                WHERE subject=$1 AND domain=$2
                                Also used by revokeExpiredRoles

idx_role_assignments_expires    Partial index (WHERE expires_at IS NOT NULL)
                                Only indexes rows that can expire
                                Very small index — fast lazy expiry query
```

### Query Patterns

```sql
-- LoadPolicy (adapter startup / cache reload)
SELECT ptype, v0, v1, v2, v3, v4, v5
FROM casbin_rule ORDER BY ptype;

-- AddPolicy (new p or g rule)
INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT DO NOTHING;

-- RemoveFilteredPolicy (revoke all policies for a domain on tenant delete)
DELETE FROM casbin_rule WHERE ptype=$1 AND v1=$2;  -- fieldIndex=1, domain value

-- revokeExpiredRoles (hot path — called on every Enforce)
SELECT role_name FROM role_assignments
WHERE subject=$1 AND domain=$2
  AND is_active = TRUE
  AND expires_at IS NOT NULL
  AND expires_at < NOW();
-- Uses idx_role_assignments_expires → ultra-fast
-- Most tenants return 0 rows (no expired roles) → zero cost

-- AssignRole (write to role_assignments + casbin g-rule in tx)
INSERT INTO role_assignments
  (id, tenant_id, subject, role_name, domain, assigned_by, delegated_by, expires_at, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
ON CONFLICT (subject, role_name, domain)
DO UPDATE SET is_active=TRUE, assigned_by=EXCLUDED.assigned_by,
              delegated_by=EXCLUDED.delegated_by, expires_at=EXCLUDED.expires_at;
```

### Database Role Requirements

| DB Role | casbin_rule | role_assignments |
|---------|-------------|-----------------|
| `application_role` | SELECT, INSERT, UPDATE, DELETE | SELECT, INSERT, UPDATE, DELETE (tenant-scoped by RLS) |
| `admin_role` | ALL (no RLS) | ALL (no RLS) |
| `readonly_role` | SELECT | SELECT |

---

Next: [Role Management](./07-role-management.md)
