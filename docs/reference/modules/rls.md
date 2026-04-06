# PostgreSQL Schema & RLS Architecture — Awo ERP

**Version:** 1.0.0  
**Status:** Implementation Reference  
**Audience:** Engineering  
**Database:** PostgreSQL 15+

---

## Table of Contents

1. [Role Architecture](#1-role-architecture)
2. [Schema Conventions](#2-schema-conventions)
3. [Core Infrastructure Tables](#3-core-infrastructure-tables)
4. [Tables — Design & RLS](#4-tables--design--rls)
5. [Views — Security Invoker Pattern](#5-views--security-invoker-pattern)
6. [Functions & Stored Procedures](#6-functions--stored-procedures)
7. [Triggers](#7-triggers)
8. [Query Patterns](#8-query-patterns)
9. [Session Context Management](#9-session-context-management)
10. [Adding a New Module](#10-adding-a-new-module)
11. [Verification Queries](#11-verification-queries)

---

## 1. Role Architecture

Awo ERP uses exactly three database roles. No application code ever connects as a superuser. Each role has a distinct contract with the database.

### Role Definitions

```sql
-- ── application_role ─────────────────────────────────────────────────────────
-- Used by the Go backend for all tenant-facing operations.
-- Subject to RLS on every table. Never bypasses policies.
-- Connects via connection pool (PgBouncer, pgxpool).
CREATE ROLE application_role
    NOLOGIN          -- actual login credentials are on a user that INHERITS this role
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    NOBYPASSRLS;     -- ← CRITICAL: must never bypass RLS

-- ── admin_role ───────────────────────────────────────────────────────────────
-- Used by platform administrators, migration tools, and internal ops scripts.
-- Bypasses RLS intentionally — used only for cross-tenant operations,
-- backfills, and schema migrations. Never used in tenant-facing request paths.
CREATE ROLE admin_role
    NOLOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    BYPASSRLS;       -- ← intentional: admin sees all tenants

-- ── readonly_role ────────────────────────────────────────────────────────────
-- Used by analytics dashboards, BI tools (Metabase, Grafana, Redash),
-- and reporting queries. Read-only on all tables and views.
-- Subject to RLS — dashboard queries are always tenant-scoped.
CREATE ROLE readonly_role
    NOLOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    NOBYPASSRLS;     -- ← subject to RLS, same as application_role

-- ── Actual login users (grant the role above) ────────────────────────────────
CREATE USER awo_app      PASSWORD '<strong-secret>' INHERIT IN ROLE application_role;
CREATE USER awo_admin    PASSWORD '<strong-secret>' INHERIT IN ROLE admin_role;
CREATE USER awo_readonly PASSWORD '<strong-secret>' INHERIT IN ROLE readonly_role;
```

### Role Capability Matrix

| Capability | `application_role` | `admin_role` | `readonly_role` |
|---|---|---|---|
| SELECT tenant data | ✅ own tenant only | ✅ all tenants | ✅ own tenant only |
| INSERT / UPDATE / DELETE | ✅ own tenant only | ✅ all tenants | ❌ |
| RLS enforced | ✅ yes | ❌ bypassed | ✅ yes |
| Schema changes (DDL) | ❌ | ✅ | ❌ |
| Cross-tenant reads | ❌ | ✅ | ❌ |
| Used in request path | ✅ | ❌ | ✅ (analytics) |

### Schema Privileges (applied once, covers all current and future objects)

```sql
-- Revoke default public access first
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE ledger FROM PUBLIC;

-- application_role
GRANT CONNECT ON DATABASE ledger TO application_role;
GRANT USAGE   ON SCHEMA   public TO application_role;

-- admin_role
GRANT CONNECT ON DATABASE ledger TO admin_role;
GRANT USAGE   ON SCHEMA   public TO admin_role;
GRANT CREATE  ON SCHEMA   public TO admin_role;

-- readonly_role
GRANT CONNECT ON DATABASE ledger TO readonly_role;
GRANT USAGE   ON SCHEMA   public TO readonly_role;

-- Default privileges — automatically apply to future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO application_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO readonly_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON TABLES TO admin_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO application_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO admin_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO application_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO readonly_role;
```

---

## 2. Schema Conventions

These conventions apply to every table in every module. Deviating from them requires explicit justification in the PR.

### Naming

| Object | Convention | Example |
|---|---|---|
| Tables | `snake_case`, plural | `employees`, `payroll_runs` |
| Columns | `snake_case` | `tenant_id`, `created_at` |
| Indexes | `idx_{table}_{columns}` | `idx_employees_tenant_id` |
| Foreign keys | `fk_{table}_{target}` | `fk_employees_tenants` |
| Policies | `{table}_{role}_{command}` | `employees_app_all` |
| Functions | `{verb}_{noun}` | `set_tenant_context`, `audit_stamp` |
| Triggers | `trg_{table}_{event}` | `trg_employees_before_update` |

### Mandatory Columns — Every Tenant-Scoped Table

```sql
-- Every tenant-scoped table must have exactly these columns in this order
-- before any domain-specific columns.

tenant_id   UUID        NOT NULL REFERENCES tenants(id),
id          UUID        NOT NULL DEFAULT gen_random_uuid(),
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
created_by  UUID,       -- references users(id); nullable for system-generated rows
updated_by  UUID,       -- references users(id)

PRIMARY KEY (id),
-- Composite index: tenant_id first, then id
-- Satisfies RLS policy check AND primary key lookup in one index scan
UNIQUE (tenant_id, id)
```

### Non-Tenant (Global/Shared) Tables

Tables without `tenant_id` are global: `tenants`, `jurisdiction_packages`, `system_config`. These tables either have no RLS or have their own access policies that do not reference `app.current_tenant`.

---

## 3. Core Infrastructure Tables

### Tenants

```sql
CREATE TABLE tenants (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT        NOT NULL UNIQUE,
    name            TEXT        NOT NULL,
    jurisdiction    CHAR(2)     NOT NULL DEFAULT 'KE',
    base_currency   CHAR(3)     NOT NULL DEFAULT 'KES',
    status          TEXT        NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active','suspended','terminated')),
    config          JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- No RLS on tenants table — it is a global lookup table.
-- application_role may SELECT; only admin_role may INSERT/UPDATE.
REVOKE INSERT, UPDATE, DELETE ON tenants FROM application_role;
GRANT  SELECT                  ON tenants TO application_role;
GRANT  SELECT                  ON tenants TO readonly_role;
```

### Outbox Events (per-module, example for HR module)

```sql
-- Each module has its own outbox table to avoid cross-module contention.
-- Named {module}_outbox_events.

CREATE TABLE hr_outbox_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    event_type      TEXT        NOT NULL,
    schema_version  INT         NOT NULL DEFAULT 1,
    aggregate_type  TEXT        NOT NULL,
    aggregate_id    UUID        NOT NULL,
    payload         JSONB       NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ,
    publish_attempts INT        NOT NULL DEFAULT 0,
    last_error      TEXT,

    CONSTRAINT chk_publish_attempts CHECK (publish_attempts >= 0)
);

CREATE INDEX idx_hr_outbox_unpublished
    ON hr_outbox_events (occurred_at)
    WHERE published_at IS NULL;

-- RLS
ALTER TABLE hr_outbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE hr_outbox_events FORCE  ROW LEVEL SECURITY;

CREATE POLICY hr_outbox_app_all ON hr_outbox_events
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- admin_role bypasses RLS (BYPASSRLS flag) — no policy needed
-- readonly_role gets no policy on outbox — dashboards don't need raw events
```

### Inbox Events

```sql
CREATE TABLE hr_inbox_events (
    event_id        UUID        PRIMARY KEY,   -- the IntegrationEvent.EventID
    event_type      TEXT        NOT NULL,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    consumer        TEXT        NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hr_inbox_tenant ON hr_inbox_events (tenant_id);

ALTER TABLE hr_inbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE hr_inbox_events FORCE  ROW LEVEL SECURITY;

CREATE POLICY hr_inbox_app_all ON hr_inbox_events
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 4. Tables — Design & RLS

This section shows the complete pattern for a representative set of HR module tables. Every other module follows the identical structure.

### Template — Applying RLS to Any Table

```sql
-- Step 1: Create the table (mandatory columns first)
CREATE TABLE {module}_{resource} (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    -- domain columns here

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

-- Step 2: Indexes
CREATE INDEX idx_{module}_{resource}_tenant
    ON {module}_{resource} (tenant_id);

-- Step 3: Enable and force RLS
ALTER TABLE {module}_{resource} ENABLE ROW LEVEL SECURITY;
ALTER TABLE {module}_{resource} FORCE  ROW LEVEL SECURITY;

-- Step 4: application_role policy — tenant-scoped read/write
CREATE POLICY {module}_{resource}_app_all
    ON {module}_{resource}
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- Step 5: readonly_role policy — tenant-scoped read only
CREATE POLICY {module}_{resource}_ro_select
    ON {module}_{resource}
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- admin_role has BYPASSRLS — no policy needed.
-- It sees everything without a policy entry.
```

### Employees

```sql
CREATE TABLE employees (
    -- mandatory
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    -- domain
    employee_number TEXT        NOT NULL,
    first_name      TEXT        NOT NULL,
    last_name       TEXT        NOT NULL,
    date_of_birth   DATE        NOT NULL,
    gender          TEXT        NOT NULL CHECK (gender IN ('male','female','other')),
    status          TEXT        NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','pending','active','suspended','terminated')),
    department_id   UUID,
    position_id     UUID,
    grade_id        UUID,
    manager_id      UUID        REFERENCES employees(id),
    biometric_pin   TEXT,
    join_date       DATE        NOT NULL,
    identifiers     JSONB       NOT NULL DEFAULT '{}',

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, employee_number),
    UNIQUE (tenant_id, biometric_pin)
);

CREATE INDEX idx_employees_tenant        ON employees (tenant_id);
CREATE INDEX idx_employees_tenant_status ON employees (tenant_id, status);
CREATE INDEX idx_employees_manager       ON employees (manager_id) WHERE manager_id IS NOT NULL;
CREATE INDEX idx_employees_department    ON employees (tenant_id, department_id);

-- Full-text search on name (used by AMIS-UI autocomplete)
CREATE INDEX idx_employees_name_trgm
    ON employees USING gin (
        (first_name || ' ' || last_name) gin_trgm_ops
    );

ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees FORCE  ROW LEVEL SECURITY;

-- application_role: full CRUD, own tenant only
CREATE POLICY employees_app_all
    ON employees
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- readonly_role: SELECT only, own tenant only
CREATE POLICY employees_ro_select
    ON employees
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Payroll Runs

```sql
CREATE TABLE payroll_runs (
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    id                  UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID,
    updated_by          UUID,

    period_start        DATE        NOT NULL,
    period_end          DATE        NOT NULL,
    pay_date            DATE        NOT NULL,
    jurisdiction        CHAR(2)     NOT NULL,
    currency            CHAR(3)     NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'draft'
                            CHECK (status IN (
                                'draft','processing','review',
                                'approved','posted','reversed','failed'
                            )),
    rate_snapshot       JSONB       NOT NULL DEFAULT '{}',
    exchange_rates      JSONB       NOT NULL DEFAULT '{}',
    employee_count      INT         NOT NULL DEFAULT 0,
    total_gross         NUMERIC(18,4),
    total_net           NUMERIC(18,4),
    total_tax           NUMERIC(18,4),
    total_employer_cost NUMERIC(18,4),
    approved_by         UUID,
    posted_at           TIMESTAMPTZ,
    finalized_at        TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),

    CONSTRAINT chk_period CHECK (period_start < period_end),
    CONSTRAINT chk_pay_date CHECK (pay_date >= period_end),
    CONSTRAINT chk_totals CHECK (
        total_gross IS NULL OR total_gross >= 0
    )
);

CREATE INDEX idx_payroll_runs_tenant        ON payroll_runs (tenant_id);
CREATE INDEX idx_payroll_runs_tenant_status ON payroll_runs (tenant_id, status);
CREATE INDEX idx_payroll_runs_period        ON payroll_runs (tenant_id, period_start, period_end);

ALTER TABLE payroll_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE payroll_runs FORCE  ROW LEVEL SECURITY;

-- application_role: full CRUD
CREATE POLICY payroll_runs_app_all
    ON payroll_runs
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- readonly_role: SELECT — finance dashboards need payroll data
CREATE POLICY payroll_runs_ro_select
    ON payroll_runs
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Payslips

```sql
CREATE TABLE payslips (
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    id                  UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID,
    updated_by          UUID,

    run_id              UUID        NOT NULL REFERENCES payroll_runs(id),
    employee_id         UUID        NOT NULL REFERENCES employees(id),
    period_start        DATE        NOT NULL,
    period_end          DATE        NOT NULL,
    currency            CHAR(3)     NOT NULL,
    gross_earnings      NUMERIC(18,4) NOT NULL,
    total_deductions    NUMERIC(18,4) NOT NULL,
    net_pay             NUMERIC(18,4) NOT NULL,
    ytd_gross           NUMERIC(18,4) NOT NULL DEFAULT 0,
    ytd_tax             NUMERIC(18,4) NOT NULL DEFAULT 0,
    lines               JSONB       NOT NULL DEFAULT '[]',

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, run_id, employee_id),   -- one payslip per employee per run

    CONSTRAINT chk_net_pay CHECK (net_pay >= 0)
);

CREATE INDEX idx_payslips_tenant      ON payslips (tenant_id);
CREATE INDEX idx_payslips_employee    ON payslips (tenant_id, employee_id);
CREATE INDEX idx_payslips_run         ON payslips (tenant_id, run_id);

ALTER TABLE payslips ENABLE ROW LEVEL SECURITY;
ALTER TABLE payslips FORCE  ROW LEVEL SECURITY;

-- application_role: full access (HR admin, Finance manager)
CREATE POLICY payslips_app_all
    ON payslips
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- application_role self-read: employee reads own payslip
-- This is a SECOND policy — PostgreSQL combines policies with OR for the same role.
-- An employee satisfies EITHER the above OR the one below.
-- The Go layer enforces which users get which app.employee_id setting.
CREATE POLICY payslips_app_self_read
    ON payslips
    FOR SELECT
    TO application_role
    USING (
        tenant_id   = current_setting('app.current_tenant')::UUID
        AND employee_id = current_setting('app.current_employee', true)::UUID
    );

-- readonly_role: tenant-scoped read for analytics
CREATE POLICY payslips_ro_select
    ON payslips
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

> **Note on multiple policies:** PostgreSQL evaluates multiple PERMISSIVE policies with OR logic. An HR admin satisfies `payslips_app_all` (full tenant access). An employee satisfies `payslips_app_self_read` (own payslip only). Both use the same `application_role` — the difference is whether `app.current_employee` is set in the session. The Go middleware sets it for employee sessions and does not set it for admin sessions.

### Leave Requests

```sql
CREATE TABLE leave_requests (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    employee_id     UUID        NOT NULL REFERENCES employees(id),
    leave_type_id   UUID        NOT NULL,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    days_requested  NUMERIC(6,2) NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'draft'
                        CHECK (status IN (
                            'draft','submitted','approved',
                            'rejected','cancelled','completed'
                        )),
    comment         TEXT,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    CONSTRAINT chk_dates CHECK (start_date <= end_date),
    CONSTRAINT chk_days  CHECK (days_requested > 0)
);

CREATE INDEX idx_leave_requests_tenant   ON leave_requests (tenant_id);
CREATE INDEX idx_leave_requests_employee ON leave_requests (tenant_id, employee_id);
CREATE INDEX idx_leave_requests_status   ON leave_requests (tenant_id, status)
    WHERE status IN ('submitted','approved');

ALTER TABLE leave_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_requests FORCE  ROW LEVEL SECURITY;

CREATE POLICY leave_requests_app_all
    ON leave_requests
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY leave_requests_ro_select
    ON leave_requests
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Attendance Records

```sql
CREATE TABLE attendance_records (
    tenant_id           UUID        NOT NULL REFERENCES tenants(id),
    id                  UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID,
    updated_by          UUID,

    employee_id         UUID        NOT NULL REFERENCES employees(id),
    date                DATE        NOT NULL,
    source              TEXT        NOT NULL
                            CHECK (source IN ('biometric','mobile','api','manual')),
    source_ref          TEXT,
    scheduled_start     TIMESTAMPTZ,
    scheduled_end       TIMESTAMPTZ,
    actual_clock_in     TIMESTAMPTZ,
    actual_clock_out    TIMESTAMPTZ,
    break_minutes       INT         NOT NULL DEFAULT 0,
    overtime_minutes    INT         NOT NULL DEFAULT 0,
    absence_type        TEXT        NOT NULL DEFAULT 'present'
                            CHECK (absence_type IN (
                                'present','absent','half_day','public_holiday','leave'
                            )),
    leave_request_id    UUID        REFERENCES leave_requests(id),
    approved_ot         BOOLEAN     NOT NULL DEFAULT FALSE,
    notes               TEXT,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, employee_id, date)
);

CREATE INDEX idx_attendance_tenant   ON attendance_records (tenant_id);
CREATE INDEX idx_attendance_employee ON attendance_records (tenant_id, employee_id);
CREATE INDEX idx_attendance_date     ON attendance_records (tenant_id, date);

ALTER TABLE attendance_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE attendance_records FORCE  ROW LEVEL SECURITY;

CREATE POLICY attendance_app_all
    ON attendance_records
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY attendance_ro_select
    ON attendance_records
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Discrepancy Records (Employee Accountability)

```sql
CREATE TABLE discrepancy_records (
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    employee_id     UUID        NOT NULL REFERENCES employees(id),
    type            TEXT        NOT NULL
                        CHECK (type IN (
                            'cash_short','cash_over','stock_shortage',
                            'stock_surplus','damage','equipment_loss'
                        )),
    incident_date   DATE        NOT NULL,
    shift_id        UUID,
    description     TEXT        NOT NULL,
    expected_value  NUMERIC(18,4) NOT NULL,
    actual_value    NUMERIC(18,4) NOT NULL,
    variance_amount NUMERIC(18,4) NOT NULL
                        CHECK (variance_amount >= 0),
    currency        CHAR(3)     NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'identified'
                        CHECK (status IN (
                            'identified','under_review','disputed',
                            'confirmed','waived','recovered','written_off'
                        )),
    recovery_method TEXT
                        CHECK (recovery_method IN (
                            'payroll_deduction','direct_payment','waived','write_off'
                        )),
    evidence_refs   JSONB       NOT NULL DEFAULT '[]',
    raised_by       UUID        NOT NULL,
    reviewed_by     UUID,
    resolved_at     TIMESTAMPTZ,
    notes           TEXT,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_discrepancy_tenant   ON discrepancy_records (tenant_id);
CREATE INDEX idx_discrepancy_employee ON discrepancy_records (tenant_id, employee_id);
CREATE INDEX idx_discrepancy_status   ON discrepancy_records (tenant_id, status)
    WHERE status NOT IN ('recovered','written_off');

ALTER TABLE discrepancy_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE discrepancy_records FORCE  ROW LEVEL SECURITY;

CREATE POLICY discrepancy_app_all
    ON discrepancy_records
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY discrepancy_ro_select
    ON discrepancy_records
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 5. Views — Security Invoker Pattern

Every view must carry `security_invoker = true`. Without it, the view runs as its owner (which may be a superuser), silently bypassing RLS and returning all tenants' data.

### The Rule

```sql
-- WRONG — view runs as owner; bypasses RLS if owner has BYPASSRLS
CREATE VIEW active_employees AS
SELECT id, tenant_id, first_name, last_name, status
FROM employees
WHERE status = 'active';

-- CORRECT — view runs as the caller; RLS applies as normal
CREATE VIEW active_employees AS
SELECT id, tenant_id, first_name, last_name, status
FROM employees
WHERE status = 'active';

ALTER VIEW active_employees SET (security_invoker = true);
```

### HR Module Views

```sql
-- ── active_employees ─────────────────────────────────────────────────────────
CREATE OR REPLACE VIEW active_employees AS
SELECT
    e.tenant_id,
    e.id,
    e.employee_number,
    e.first_name,
    e.last_name,
    e.first_name || ' ' || e.last_name  AS full_name,
    e.gender,
    e.status,
    e.join_date,
    e.department_id,
    e.position_id,
    e.grade_id,
    e.manager_id
FROM employees e
WHERE e.status = 'active';

ALTER VIEW active_employees SET (security_invoker = true);

GRANT SELECT ON active_employees TO application_role;
GRANT SELECT ON active_employees TO readonly_role;
-- admin_role inherits SELECT via DEFAULT PRIVILEGES


-- ── pending_approvals ────────────────────────────────────────────────────────
-- Surfaces all records waiting for manager or HR action.
-- Used by the "What Needs My Attention" dashboard widget.
CREATE OR REPLACE VIEW pending_approvals AS
SELECT
    lr.tenant_id,
    'leave_request'             AS approval_type,
    lr.id                       AS record_id,
    lr.employee_id,
    lr.status,
    lr.created_at               AS raised_at,
    NULL::UUID                  AS approver_id
FROM leave_requests lr
WHERE lr.status = 'submitted'

UNION ALL

SELECT
    pr.tenant_id,
    'payroll_run'               AS approval_type,
    pr.id                       AS record_id,
    NULL::UUID                  AS employee_id,
    pr.status,
    pr.created_at               AS raised_at,
    NULL::UUID                  AS approver_id
FROM payroll_runs pr
WHERE pr.status = 'review'

UNION ALL

SELECT
    dr.tenant_id,
    'discrepancy'               AS approval_type,
    dr.id                       AS record_id,
    dr.employee_id,
    dr.status,
    dr.created_at               AS raised_at,
    NULL::UUID                  AS approver_id
FROM discrepancy_records dr
WHERE dr.status = 'under_review';

ALTER VIEW pending_approvals SET (security_invoker = true);

GRANT SELECT ON pending_approvals TO application_role;
GRANT SELECT ON pending_approvals TO readonly_role;


-- ── payroll_run_summary ──────────────────────────────────────────────────────
-- Pre-joined summary for Finance dashboard — avoids ad-hoc joins in queries.
CREATE OR REPLACE VIEW payroll_run_summary AS
SELECT
    pr.tenant_id,
    pr.id               AS run_id,
    pr.period_start,
    pr.period_end,
    pr.pay_date,
    pr.jurisdiction,
    pr.currency,
    pr.status,
    pr.employee_count,
    pr.total_gross,
    pr.total_net,
    pr.total_tax,
    pr.total_employer_cost,
    pr.total_gross - pr.total_net   AS total_deductions,
    pr.created_at,
    pr.finalized_at
FROM payroll_runs pr;

ALTER VIEW payroll_run_summary SET (security_invoker = true);

GRANT SELECT ON payroll_run_summary TO application_role;
GRANT SELECT ON payroll_run_summary TO readonly_role;


-- ── employee_leave_balances ──────────────────────────────────────────────────
-- Denormalised leave balance view for the self-service dashboard.
CREATE OR REPLACE VIEW employee_leave_balances AS
SELECT
    lb.tenant_id,
    lb.employee_id,
    lt.code             AS leave_type_code,
    lt.name             AS leave_type_name,
    lt.paid,
    lb.entitled_days,
    lb.taken_days,
    lb.pending_days,
    lb.carry_forward_days,
    (lb.entitled_days + lb.carry_forward_days)
        - lb.taken_days
        - lb.pending_days               AS available_days
FROM leave_balances lb
JOIN leave_types    lt ON lt.id = lb.leave_type_id
WHERE lb.leave_year = EXTRACT(YEAR FROM CURRENT_DATE);

ALTER VIEW employee_leave_balances SET (security_invoker = true);

GRANT SELECT ON employee_leave_balances TO application_role;
GRANT SELECT ON employee_leave_balances TO readonly_role;
```

### Analytics Views (readonly_role primary consumer)

```sql
-- ── monthly_payroll_cost ─────────────────────────────────────────────────────
-- Aggregate payroll cost by month for the Finance P&L dashboard.
-- readonly_role queries this; RLS ensures tenant scoping.
CREATE OR REPLACE VIEW monthly_payroll_cost AS
SELECT
    pr.tenant_id,
    DATE_TRUNC('month', pr.period_start)    AS month,
    pr.currency,
    COUNT(*)                                AS run_count,
    SUM(pr.total_gross)                     AS total_gross,
    SUM(pr.total_employer_cost)             AS total_employer_cost,
    SUM(pr.total_gross + pr.total_employer_cost) AS total_labour_cost,
    SUM(pr.total_tax)                       AS total_paye,
    SUM(pr.employee_count)                  AS total_headcount
FROM payroll_runs pr
WHERE pr.status IN ('posted','reversed')
GROUP BY pr.tenant_id, DATE_TRUNC('month', pr.period_start), pr.currency;

ALTER VIEW monthly_payroll_cost SET (security_invoker = true);

GRANT SELECT ON monthly_payroll_cost TO readonly_role;
GRANT SELECT ON monthly_payroll_cost TO application_role;


-- ── attendance_summary ───────────────────────────────────────────────────────
CREATE OR REPLACE VIEW attendance_summary AS
SELECT
    ar.tenant_id,
    ar.employee_id,
    DATE_TRUNC('month', ar.date)            AS month,
    COUNT(*) FILTER (WHERE ar.absence_type = 'present')         AS days_present,
    COUNT(*) FILTER (WHERE ar.absence_type = 'absent')          AS days_absent,
    COUNT(*) FILTER (WHERE ar.absence_type = 'leave')           AS days_on_leave,
    SUM(ar.overtime_minutes)                                    AS total_ot_minutes
FROM attendance_records ar
GROUP BY ar.tenant_id, ar.employee_id, DATE_TRUNC('month', ar.date);

ALTER VIEW attendance_summary SET (security_invoker = true);

GRANT SELECT ON attendance_summary TO readonly_role;
GRANT SELECT ON attendance_summary TO application_role;
```

---

## 6. Functions & Stored Procedures

### Session Context Functions

```sql
-- ── set_tenant_context ───────────────────────────────────────────────────────
-- Called by the Go RunInTx wrapper at the start of every transaction.
-- is_local=true means the setting expires at transaction end automatically.
CREATE OR REPLACE FUNCTION set_tenant_context(
    p_tenant_id     UUID,
    p_employee_id   UUID    DEFAULT NULL,
    p_user_type     TEXT    DEFAULT NULL
)
RETURNS VOID
LANGUAGE plpgsql
SECURITY INVOKER   -- runs as the calling role; respects its privileges
AS $$
BEGIN
    PERFORM set_config('app.current_tenant',   p_tenant_id::TEXT,   true);

    IF p_employee_id IS NOT NULL THEN
        PERFORM set_config('app.current_employee', p_employee_id::TEXT, true);
    END IF;

    IF p_user_type IS NOT NULL THEN
        PERFORM set_config('app.user_type',        p_user_type,         true);
    END IF;
END;
$$;

GRANT EXECUTE ON FUNCTION set_tenant_context TO application_role;
GRANT EXECUTE ON FUNCTION set_tenant_context TO readonly_role;


-- ── clear_tenant_context ─────────────────────────────────────────────────────
-- Defensive reset. Called when releasing a connection back to the pool.
-- Redundant with is_local=true but provides a safety net.
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS VOID
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
BEGIN
    -- set_config with empty string; current_setting will return '' not NULL
    -- Policies use ::UUID cast which will raise an error on empty string,
    -- ensuring no accidental cross-tenant access if is_local somehow failed.
    PERFORM set_config('app.current_tenant',   '', false);
    PERFORM set_config('app.current_employee', '', false);
    PERFORM set_config('app.user_type',        '', false);
END;
$$;

GRANT EXECUTE ON FUNCTION clear_tenant_context TO application_role;
GRANT EXECUTE ON FUNCTION clear_tenant_context TO readonly_role;


-- ── current_tenant_id ────────────────────────────────────────────────────────
-- Helper: returns the current tenant UUID or raises a clear error.
-- Use inside other functions to avoid repeating the cast and NULL check.
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID
LANGUAGE plpgsql
SECURITY INVOKER
STABLE   -- does not modify data; result stable within a transaction
AS $$
DECLARE
    v_tenant TEXT;
BEGIN
    v_tenant := current_setting('app.current_tenant', true);
    IF v_tenant IS NULL OR v_tenant = '' THEN
        RAISE EXCEPTION 'tenant context not set — call set_tenant_context() first'
            USING ERRCODE = 'insufficient_privilege';
    END IF;
    RETURN v_tenant::UUID;
END;
$$;

GRANT EXECUTE ON FUNCTION current_tenant_id TO application_role;
GRANT EXECUTE ON FUNCTION current_tenant_id TO readonly_role;
```

### Audit Stamp Function

```sql
-- ── audit_stamp ──────────────────────────────────────────────────────────────
-- Returns a row containing the current tenant and user for audit columns.
-- Used by triggers to populate created_by / updated_by.
CREATE OR REPLACE FUNCTION audit_stamp()
RETURNS TABLE (tenant UUID, actor UUID)
LANGUAGE plpgsql
SECURITY INVOKER
STABLE
AS $$
DECLARE
    v_employee TEXT;
BEGIN
    v_employee := current_setting('app.current_employee', true);
    RETURN QUERY SELECT
        current_tenant_id(),
        CASE WHEN v_employee IS NOT NULL AND v_employee <> ''
             THEN v_employee::UUID
             ELSE NULL
        END;
END;
$$;

GRANT EXECUTE ON FUNCTION audit_stamp TO application_role;
```

### Business Logic Functions

```sql
-- ── activate_employee ────────────────────────────────────────────────────────
-- Transitions an employee from pending → active.
-- Validates preconditions then performs the update atomically.
-- Declared SECURITY INVOKER so RLS on employees still applies —
-- only the calling role's policies govern what it can update.
CREATE OR REPLACE FUNCTION activate_employee(
    p_employee_id   UUID,
    p_activated_by  UUID
)
RETURNS employees
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
DECLARE
    v_employee employees;
BEGIN
    -- Lock the row for update within this transaction
    SELECT * INTO v_employee
    FROM employees
    WHERE id = p_employee_id
      AND tenant_id = current_tenant_id()
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'employee not found: %', p_employee_id
            USING ERRCODE = 'no_data_found';
    END IF;

    IF v_employee.status <> 'pending' THEN
        RAISE EXCEPTION 'employee % cannot be activated from status %',
            p_employee_id, v_employee.status
            USING ERRCODE = 'invalid_parameter_value';
    END IF;

    UPDATE employees
    SET
        status     = 'active',
        updated_at = NOW(),
        updated_by = p_activated_by
    WHERE id        = p_employee_id
      AND tenant_id = current_tenant_id()
    RETURNING * INTO v_employee;

    RETURN v_employee;
END;
$$;

GRANT EXECUTE ON FUNCTION activate_employee TO application_role;
-- NOT granted to readonly_role — mutations only via application_role


-- ── finalise_payroll_run ──────────────────────────────────────────────────────
-- Locks a payroll run into 'posted' status and records finalization timestamp.
-- Only called from the Temporal activity after GL confirmation.
CREATE OR REPLACE FUNCTION finalise_payroll_run(
    p_run_id        UUID,
    p_approved_by   UUID,
    p_journal_ref   TEXT
)
RETURNS payroll_runs
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
DECLARE
    v_run payroll_runs;
BEGIN
    SELECT * INTO v_run
    FROM payroll_runs
    WHERE id = p_run_id
      AND tenant_id = current_tenant_id()
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'payroll run not found: %', p_run_id;
    END IF;

    IF v_run.status <> 'approved' THEN
        RAISE EXCEPTION 'cannot post run % from status %',
            p_run_id, v_run.status;
    END IF;

    UPDATE payroll_runs
    SET
        status       = 'posted',
        approved_by  = p_approved_by,
        posted_at    = NOW(),
        finalized_at = NOW(),
        updated_at   = NOW(),
        updated_by   = p_approved_by
    WHERE id        = p_run_id
      AND tenant_id = current_tenant_id()
    RETURNING * INTO v_run;

    RETURN v_run;
END;
$$;

GRANT EXECUTE ON FUNCTION finalise_payroll_run TO application_role;


-- ── get_pending_approvals_count ───────────────────────────────────────────────
-- Fast count of items pending action for the notification badge.
-- Declared STABLE — safe to call repeatedly in a transaction.
CREATE OR REPLACE FUNCTION get_pending_approvals_count()
RETURNS INT
LANGUAGE sql
SECURITY INVOKER
STABLE
AS $$
    SELECT COUNT(*)::INT FROM pending_approvals
    WHERE tenant_id = current_tenant_id();
$$;

GRANT EXECUTE ON FUNCTION get_pending_approvals_count TO application_role;
GRANT EXECUTE ON FUNCTION get_pending_approvals_count TO readonly_role;
```

### Admin-Only Functions

```sql
-- ── backfill_tenant_context ──────────────────────────────────────────────────
-- Used by admin_role only during data migrations.
-- Sets tenant context without the is_local restriction (session-scoped).
-- NEVER callable by application_role or readonly_role.
CREATE OR REPLACE FUNCTION backfill_tenant_context(p_tenant_id UUID)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER   -- runs as the function owner (must be admin user)
AS $$
BEGIN
    -- Verify the caller is admin_role
    IF current_user NOT IN ('awo_admin') THEN
        RAISE EXCEPTION 'backfill_tenant_context may only be called by awo_admin'
            USING ERRCODE = 'insufficient_privilege';
    END IF;
    PERFORM set_config('app.current_tenant', p_tenant_id::TEXT, false);
END;
$$;

-- Deliberately not granted to application_role or readonly_role
GRANT EXECUTE ON FUNCTION backfill_tenant_context TO admin_role;
```

---

## 7. Triggers

### Updated At — Automatic Timestamp

```sql
-- ── fn_set_updated_at ────────────────────────────────────────────────────────
-- Generic trigger function. Attach to every table.
CREATE OR REPLACE FUNCTION fn_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- Attach to each table (shown for employees; repeat for every table)
CREATE TRIGGER trg_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_payroll_runs_updated_at
    BEFORE UPDATE ON payroll_runs
    FOR EACH ROW
    EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_payslips_updated_at
    BEFORE UPDATE ON payslips
    FOR EACH ROW
    EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_leave_requests_updated_at
    BEFORE UPDATE ON leave_requests
    FOR EACH ROW
    EXECUTE FUNCTION fn_set_updated_at();
```

### Tenant Guard — Prevent Cross-Tenant Writes

```sql
-- ── fn_enforce_tenant_id ─────────────────────────────────────────────────────
-- Ensures tenant_id on INSERT always matches the session context.
-- Also prevents UPDATE from changing a row's tenant_id.
-- This is a second layer of defence behind the WITH CHECK policy.
CREATE OR REPLACE FUNCTION fn_enforce_tenant_id()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
DECLARE
    v_ctx_tenant UUID;
BEGIN
    -- Skip enforcement for admin_role (BYPASSRLS already handles it)
    IF current_user = 'awo_admin' THEN
        RETURN NEW;
    END IF;

    v_ctx_tenant := current_tenant_id();  -- raises if not set

    IF TG_OP = 'INSERT' THEN
        IF NEW.tenant_id <> v_ctx_tenant THEN
            RAISE EXCEPTION
                'tenant_id mismatch on INSERT into %: got %, expected %',
                TG_TABLE_NAME, NEW.tenant_id, v_ctx_tenant
                USING ERRCODE = 'insufficient_privilege';
        END IF;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF NEW.tenant_id <> OLD.tenant_id THEN
            RAISE EXCEPTION
                'tenant_id is immutable — cannot change on %',
                TG_TABLE_NAME
                USING ERRCODE = 'insufficient_privilege';
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

-- Attach to every tenant-scoped table
CREATE TRIGGER trg_employees_enforce_tenant
    BEFORE INSERT OR UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();

CREATE TRIGGER trg_payroll_runs_enforce_tenant
    BEFORE INSERT OR UPDATE ON payroll_runs
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();

CREATE TRIGGER trg_payslips_enforce_tenant
    BEFORE INSERT OR UPDATE ON payslips
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();

CREATE TRIGGER trg_leave_requests_enforce_tenant
    BEFORE INSERT OR UPDATE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();

CREATE TRIGGER trg_discrepancy_enforce_tenant
    BEFORE INSERT OR UPDATE ON discrepancy_records
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();
```

### Immutability Guard — Protect Finalised Records

```sql
-- ── fn_protect_posted_run ────────────────────────────────────────────────────
-- Prevents any modification to a payroll run once it is posted or reversed.
-- Corrections must go through the reversal workflow.
CREATE OR REPLACE FUNCTION fn_protect_posted_run()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
BEGIN
    IF OLD.status IN ('posted', 'reversed') THEN
        RAISE EXCEPTION
            'payroll run % is % and cannot be modified — create a reversal run',
            OLD.id, OLD.status
            USING ERRCODE = 'invalid_parameter_value';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_payroll_runs_immutability
    BEFORE UPDATE ON payroll_runs
    FOR EACH ROW EXECUTE FUNCTION fn_protect_posted_run();


-- ── fn_protect_payslip ───────────────────────────────────────────────────────
-- Payslips are immutable once created. No UPDATE allowed via trigger.
CREATE OR REPLACE FUNCTION fn_protect_payslip()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
BEGIN
    RAISE EXCEPTION
        'payslips are immutable — reverse the run to correct payslip %', OLD.id
        USING ERRCODE = 'invalid_parameter_value';
END;
$$;

CREATE TRIGGER trg_payslips_immutability
    BEFORE UPDATE ON payslips
    FOR EACH ROW EXECUTE FUNCTION fn_protect_payslip();
```

### Outbox — Auto-Insert on Domain Events

```sql
-- ── fn_employee_outbox ───────────────────────────────────────────────────────
-- Automatically writes an outbox event whenever an employee changes status.
-- This keeps domain event emission inside the DB transaction,
-- complementing the Go-layer outbox writer for cases where the trigger
-- catches status changes that happen via direct SQL (e.g. admin backfill).
CREATE OR REPLACE FUNCTION fn_employee_status_outbox()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO hr_outbox_events (
            tenant_id,
            event_type,
            aggregate_type,
            aggregate_id,
            payload,
            occurred_at
        ) VALUES (
            NEW.tenant_id,
            'hr.employee.' || NEW.status,     -- e.g. hr.employee.active
            'Employee',
            NEW.id,
            jsonb_build_object(
                'employee_id',  NEW.id,
                'tenant_id',    NEW.tenant_id,
                'old_status',   OLD.status,
                'new_status',   NEW.status,
                'changed_by',   NEW.updated_by,
                'changed_at',   NOW()
            ),
            NOW()
        );
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_employees_status_outbox
    AFTER UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION fn_employee_status_outbox();
```

---

## 8. Query Patterns

All queries in this section are written as they appear in the Go application. The session context (`set_tenant_context`) is always called at the start of the transaction before any of these run.

### Standard CRUD Queries

```sql
-- ── LIST employees (paginated) ───────────────────────────────────────────────
-- RLS automatically filters to current tenant.
-- No explicit WHERE tenant_id needed — the policy adds it.
SELECT
    id,
    employee_number,
    first_name || ' ' || last_name  AS full_name,
    status,
    department_id,
    grade_id,
    join_date
FROM employees
WHERE status = $1              -- filter parameter from application
ORDER BY last_name, first_name
LIMIT  $2
OFFSET $3;


-- ── GET single employee ──────────────────────────────────────────────────────
SELECT *
FROM employees
WHERE id = $1;
-- RLS policy: if $1 belongs to a different tenant, returns 0 rows (not an error)
-- Application layer interprets 0 rows as 404, not 403, per security best practice


-- ── INSERT employee ──────────────────────────────────────────────────────────
-- tenant_id is set explicitly to the session value.
-- The fn_enforce_tenant_id trigger will reject a mismatch.
INSERT INTO employees (
    tenant_id,
    employee_number,
    first_name,
    last_name,
    date_of_birth,
    gender,
    join_date,
    department_id,
    position_id,
    grade_id,
    created_by,
    updated_by
) VALUES (
    current_tenant_id(),   -- function call ensures it matches session context
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10
)
RETURNING *;


-- ── UPDATE employee status ───────────────────────────────────────────────────
-- The WHERE clause includes tenant_id even though RLS covers it —
-- belt-and-suspenders, and makes EXPLAIN plans clearer.
UPDATE employees
SET
    status     = $1,
    updated_by = $2,
    updated_at = NOW()
WHERE id        = $3
  AND tenant_id = current_tenant_id()   -- explicit + RLS = double safety
RETURNING *;


-- ── DELETE (soft) — mark as terminated, never hard delete ────────────────────
UPDATE employees
SET
    status     = 'terminated',
    updated_by = $1,
    updated_at = NOW()
WHERE id        = $2
  AND tenant_id = current_tenant_id();
```

### Payroll Queries

```sql
-- ── LIST payroll runs for a tenant ───────────────────────────────────────────
SELECT
    id,
    period_start,
    period_end,
    pay_date,
    status,
    employee_count,
    total_gross,
    total_net,
    currency
FROM payroll_runs
ORDER BY period_start DESC
LIMIT $1 OFFSET $2;


-- ── GET payroll run with payslip summary ─────────────────────────────────────
SELECT
    pr.*,
    COUNT(ps.id)            AS payslip_count,
    SUM(ps.gross_earnings)  AS computed_gross,
    SUM(ps.net_pay)         AS computed_net
FROM payroll_runs pr
LEFT JOIN payslips ps ON ps.run_id = pr.id
                      AND ps.tenant_id = pr.tenant_id
WHERE pr.id = $1
GROUP BY pr.id;


-- ── Employee's own payslips (self-service) ───────────────────────────────────
-- Requires app.current_employee to be set in the session.
-- The payslips_app_self_read policy limits results to own payslips.
SELECT
    id,
    period_start,
    period_end,
    currency,
    gross_earnings,
    total_deductions,
    net_pay,
    ytd_gross,
    ytd_tax
FROM payslips
WHERE employee_id = current_setting('app.current_employee')::UUID
ORDER BY period_start DESC;
```

### Outbox Relay Query

```sql
-- ── Fetch unpublished outbox events ──────────────────────────────────────────
-- FOR UPDATE SKIP LOCKED: allows multiple relay workers to run concurrently
-- without competing for the same rows.
-- This query is run by admin_role (BYPASSRLS) — fetches across all tenants.
SELECT
    id,
    tenant_id,
    event_type,
    schema_version,
    aggregate_type,
    aggregate_id,
    payload,
    occurred_at
FROM hr_outbox_events
WHERE published_at     IS NULL
  AND publish_attempts <  5
ORDER BY occurred_at
LIMIT $1
FOR UPDATE SKIP LOCKED;


-- ── Mark event published ──────────────────────────────────────────────────────
UPDATE hr_outbox_events
SET
    published_at      = NOW(),
    publish_attempts  = publish_attempts + 1
WHERE id = $1;


-- ── Record publish failure ────────────────────────────────────────────────────
UPDATE hr_outbox_events
SET
    publish_attempts = publish_attempts + 1,
    last_error       = $1
WHERE id = $2;
```

### Analytics / Dashboard Queries (readonly_role)

```sql
-- ── Headcount by department ──────────────────────────────────────────────────
-- readonly_role query; RLS filters to the tenant set in session context.
SELECT
    department_id,
    status,
    COUNT(*) AS headcount
FROM employees
GROUP BY department_id, status
ORDER BY department_id, status;


-- ── Monthly labour cost trend ────────────────────────────────────────────────
-- Queries the view — RLS applied at view level via security_invoker.
SELECT
    month,
    currency,
    total_gross,
    total_employer_cost,
    total_labour_cost,
    total_headcount
FROM monthly_payroll_cost
WHERE month >= DATE_TRUNC('month', NOW() - INTERVAL '12 months')
ORDER BY month;


-- ── Pending approvals count for notification badge ───────────────────────────
SELECT get_pending_approvals_count();


-- ── Attendance rate this month ────────────────────────────────────────────────
SELECT
    SUM(days_present)                                               AS total_present,
    SUM(days_present + days_absent + days_on_leave)                AS total_working_days,
    ROUND(
        100.0 * SUM(days_present)
              / NULLIF(SUM(days_present + days_absent + days_on_leave), 0),
        1
    )                                                               AS attendance_rate_pct
FROM attendance_summary
WHERE month = DATE_TRUNC('month', CURRENT_DATE);
```

---

## 9. Session Context Management

### Go Wrapper (canonical implementation)

```go
// pkg/db/tenant.go

// RunInTx executes fn within a transaction that has tenant context set.
// The context is scoped to the transaction (is_local = true) and expires
// automatically on commit or rollback — even if the connection is reused
// from a pool.
func (db *DB) RunInTx(
    ctx context.Context,
    tenantID uuid.UUID,
    opts TxOptions,
    fn func(ctx context.Context, tx pgx.Tx) error,
) error {
    if tenantID == uuid.Nil {
        return ErrNoTenantContext
    }

    tx, err := db.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // Set tenant context — is_local=true means it expires with the transaction.
    // Using the function rather than a raw SET gives us the clear error message
    // from current_tenant_id() if something goes wrong downstream.
    if _, err := tx.Exec(ctx,
        "SELECT set_tenant_context($1, $2, $3)",
        tenantID,
        opts.EmployeeID,    // nil for non-employee sessions
        opts.UserType,      // "hr_admin", "employee", "line_manager", etc.
    ); err != nil {
        return fmt.Errorf("set tenant context: %w", err)
    }

    if err := fn(ctx, tx); err != nil {
        return err
    }

    return tx.Commit(ctx)
}

// RunReadOnly is used by analytics and dashboard queries (readonly_role conn pool).
// Same pattern but uses the readonly connection pool.
func (db *DB) RunReadOnly(
    ctx context.Context,
    tenantID uuid.UUID,
    fn func(ctx context.Context, tx pgx.Tx) error,
) error {
    return db.readonlyPool.RunInTx(ctx, tenantID, TxOptions{}, fn)
}
```

### Two Connection Pools

Awo maintains two connection pools targeting the same database, using different roles:

```go
// cmd/server/main.go

appPool, _ := pgxpool.New(ctx, os.Getenv("DB_APP_URL"))
// DB_APP_URL connects as awo_app (inherits application_role)

roPool, _ := pgxpool.New(ctx, os.Getenv("DB_READONLY_URL"))
// DB_READONLY_URL connects as awo_readonly (inherits readonly_role)

// admin_role is not pooled — used only by migration scripts and Temporal activities
// that run outside the request path.
```

### PgBouncer Configuration

```ini
; pgbouncer.ini

[databases]
ledger_app      = host=db port=5432 dbname=ledger user=awo_app
ledger_readonly = host=db port=5432 dbname=ledger user=awo_readonly

[pgbouncer]
; CRITICAL: Transaction pooling is the only safe mode with RLS.
; Session pooling would allow tenant context to bleed between requests.
; Statement pooling would lose the context mid-transaction.
pool_mode = transaction

server_reset_query = SELECT clear_tenant_context();
```

---

## 10. Adding a New Module

This section is the step-by-step procedure every developer follows when adding a new module to Awo ERP. Following it exactly ensures the module is RLS-compliant, properly integrated with all three roles, and consistent with every existing module.

### Step 1 — Create the Migration File

Every change to the schema is a numbered migration file. Never apply schema changes directly to a production database.

```
migrations/
  0001_init_tenants.sql
  0002_hr_employees.sql
  0003_hr_payroll.sql
  ...
  NNNN_{module}_{description}.sql   ← new file
```

The migration file has three sections: `-- +migrate Up`, `-- +migrate Down`, and `-- +migrate Verify`.

### Step 2 — Define Tables

For each table in the new module, use the mandatory column template from §2, then add domain columns.

```sql
-- migrations/NNNN_inventory_stock_items.sql
-- +migrate Up

-- ── stock_items ───────────────────────────────────────────────────────────────
CREATE TABLE stock_items (
    -- mandatory columns (always first, always in this order)
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    -- domain columns
    sku             TEXT        NOT NULL,
    name            TEXT        NOT NULL,
    category        TEXT        NOT NULL,
    unit_of_measure TEXT        NOT NULL DEFAULT 'unit',
    reorder_level   NUMERIC(12,4) NOT NULL DEFAULT 0,
    status          TEXT        NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active','inactive','discontinued')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, sku)
);
```

### Step 3 — Add Indexes

```sql
-- Mandatory: tenant_id index (satisfies RLS policy scan)
CREATE INDEX idx_stock_items_tenant
    ON stock_items (tenant_id);

-- Domain-specific indexes
CREATE INDEX idx_stock_items_sku
    ON stock_items (tenant_id, sku);

CREATE INDEX idx_stock_items_category
    ON stock_items (tenant_id, category);
```

### Step 4 — Enable and Force RLS

```sql
-- Both commands are required. One without the other is a security hole.
ALTER TABLE stock_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_items FORCE  ROW LEVEL SECURITY;
```

### Step 5 — Create Policies for All Three Roles

```sql
-- application_role: full CRUD, own tenant only
CREATE POLICY stock_items_app_all
    ON stock_items
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- readonly_role: SELECT only, own tenant only
CREATE POLICY stock_items_ro_select
    ON stock_items
    FOR SELECT
    TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- admin_role: no policy needed — BYPASSRLS is set on the role.
-- Document this explicitly in a comment so reviewers don't add a spurious policy.
-- admin_role intentionally bypasses RLS on all tables via BYPASSRLS attribute.
```

### Step 6 — Attach Standard Triggers

```sql
-- updated_at trigger (mandatory on every table)
CREATE TRIGGER trg_stock_items_updated_at
    BEFORE UPDATE ON stock_items
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- tenant guard trigger (mandatory on every tenant-scoped table)
CREATE TRIGGER trg_stock_items_enforce_tenant
    BEFORE INSERT OR UPDATE ON stock_items
    FOR EACH ROW EXECUTE FUNCTION fn_enforce_tenant_id();

-- Module-specific triggers (if needed)
-- Example: auto-write outbox event when stock level changes
CREATE TRIGGER trg_stock_items_level_outbox
    AFTER UPDATE ON stock_items
    FOR EACH ROW EXECUTE FUNCTION fn_stock_level_changed_outbox();
```

### Step 7 — Create Outbox and Inbox Tables

Every module that emits integration events needs its own outbox. Every module that consumes events needs its own inbox.

```sql
-- Outbox (if this module emits events)
CREATE TABLE inventory_outbox_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    event_type      TEXT        NOT NULL,
    schema_version  INT         NOT NULL DEFAULT 1,
    aggregate_type  TEXT        NOT NULL,
    aggregate_id    UUID        NOT NULL,
    payload         JSONB       NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ,
    publish_attempts INT        NOT NULL DEFAULT 0,
    last_error      TEXT
);

CREATE INDEX idx_inventory_outbox_unpublished
    ON inventory_outbox_events (occurred_at)
    WHERE published_at IS NULL;

ALTER TABLE inventory_outbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_outbox_events FORCE  ROW LEVEL SECURITY;

CREATE POLICY inventory_outbox_app_all
    ON inventory_outbox_events
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- Inbox (if this module consumes events)
CREATE TABLE inventory_inbox_events (
    event_id        UUID        PRIMARY KEY,
    event_type      TEXT        NOT NULL,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    consumer        TEXT        NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE inventory_inbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_inbox_events FORCE  ROW LEVEL SECURITY;

CREATE POLICY inventory_inbox_app_all
    ON inventory_inbox_events
    FOR ALL
    TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Step 8 — Create Views

```sql
-- Every view must have security_invoker = true.
-- Create first, then alter. This order works even in a transaction.

CREATE OR REPLACE VIEW active_stock_items AS
SELECT
    tenant_id,
    id,
    sku,
    name,
    category,
    unit_of_measure,
    reorder_level
FROM stock_items
WHERE status = 'active';

ALTER VIEW active_stock_items SET (security_invoker = true);

GRANT SELECT ON active_stock_items TO application_role;
GRANT SELECT ON active_stock_items TO readonly_role;
```

### Step 9 — Write the Down Migration

Every migration must be reversible. The down migration drops objects in reverse dependency order.

```sql
-- +migrate Down

DROP VIEW  IF EXISTS active_stock_items;
DROP TABLE IF EXISTS inventory_inbox_events;
DROP TABLE IF EXISTS inventory_outbox_events;
DROP TABLE IF EXISTS stock_items;
```

### Step 10 — Write the Verify Section

The verify section is run by the inspection script after the migration. It asserts that every required property is present.

```sql
-- +migrate Verify

-- RLS enabled
SELECT relrowsecurity, relforcerowsecurity
FROM pg_class WHERE relname = 'stock_items';
-- Expected: t, t

-- Policies exist for both roles
SELECT policyname, roles, cmd
FROM pg_policies WHERE tablename = 'stock_items'
ORDER BY policyname;
-- Expected: stock_items_app_all (application_role, ALL), stock_items_ro_select (readonly_role, SELECT)

-- tenant_id index exists
SELECT indexname FROM pg_indexes
WHERE tablename = 'stock_items' AND indexdef LIKE '%tenant_id%';
-- Expected: idx_stock_items_tenant

-- Views have security_invoker
SELECT relname, reloptions FROM pg_class
WHERE relname = 'active_stock_items';
-- Expected: reloptions contains security_invoker=true

-- Triggers are attached
SELECT tgname FROM pg_trigger
WHERE tgrelid = 'stock_items'::regclass;
-- Expected: trg_stock_items_updated_at, trg_stock_items_enforce_tenant
```

### Step 11 — Run the Schema Inspection Script

Before opening a PR, run the inspection script against a local database with the migration applied:

```bash
DB_URL="postgresql://awo_app:secret@localhost:5432/ledger" ./scripts/db/schema_inspect.sh \
  | jq '.summary'
```

The PR must not be merged if `overall_health` is anything other than `"✅  clean"`.

### Checklist Before Merging a New Module

```
Schema
  [ ] All tables have mandatory columns (tenant_id, id, created_at, updated_at, created_by, updated_by)
  [ ] All tables have ENABLE ROW LEVEL SECURITY
  [ ] All tables have FORCE ROW LEVEL SECURITY
  [ ] All tables have both USING and WITH CHECK in application_role policy
  [ ] All tables have a readonly_role SELECT policy
  [ ] All tables have idx_{table}_tenant index on tenant_id
  [ ] No policy uses USING (true) without a tenant_id filter

Views
  [ ] Every view has ALTER VIEW x SET (security_invoker = true)
  [ ] Every view has GRANT SELECT to application_role and readonly_role

Functions
  [ ] All functions are SECURITY INVOKER (unless explicitly justified as SECURITY DEFINER)
  [ ] SECURITY DEFINER functions validate the caller's role explicitly
  [ ] GRANT EXECUTE is applied to the correct roles

Triggers
  [ ] trg_{table}_updated_at attached to every table
  [ ] trg_{table}_enforce_tenant attached to every table

Outbox / Inbox
  [ ] Module outbox table has RLS + application_role policy
  [ ] Module inbox table has RLS + application_role policy
  [ ] Outbox index on occurred_at WHERE published_at IS NULL

Migration
  [ ] Down migration drops all objects in reverse order
  [ ] Verify section asserts RLS, policies, indexes, and views

Inspection
  [ ] schema_inspect.sh passes with overall_health = "✅  clean"
```

---

## 11. Verification Queries

Run these directly in psql to audit the current state of the schema.

```sql
-- ── Tables missing tenant_id ──────────────────────────────────────────────────
SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_schema NOT IN ('pg_catalog','information_schema')
  AND table_type = 'BASE TABLE'
  AND table_name NOT IN ('tenants','jurisdiction_packages','system_config')
  AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns c
      WHERE c.table_name   = tables.table_name
        AND c.column_name  = 'tenant_id'
  )
ORDER BY table_name;


-- ── Tables with RLS disabled or not forced ────────────────────────────────────
SELECT
    n.nspname           AS schema,
    c.relname           AS table,
    c.relrowsecurity    AS rls_enabled,
    c.relforcerowsecurity AS rls_forced
FROM pg_class     c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog','information_schema')
  AND (c.relrowsecurity = false OR c.relforcerowsecurity = false)
ORDER BY c.relname;


-- ── Policies missing WITH CHECK (inserts unprotected) ────────────────────────
SELECT tablename, policyname, cmd, qual, with_check
FROM pg_policies
WHERE with_check IS NULL
  AND cmd IN ('ALL','INSERT','UPDATE')
ORDER BY tablename;


-- ── Views missing security_invoker ───────────────────────────────────────────
SELECT n.nspname AS schema, c.relname AS view
FROM pg_class     c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'v'
  AND n.nspname NOT IN ('pg_catalog','information_schema')
  AND NOT (
      c.reloptions IS NOT NULL AND
      EXISTS (
          SELECT 1 FROM unnest(c.reloptions) opt
          WHERE opt ILIKE 'security_invoker=true'
      )
  )
ORDER BY c.relname;


-- ── Tables missing tenant_id index ───────────────────────────────────────────
SELECT t.relname AS table
FROM pg_class     t
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE t.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog','information_schema')
  AND EXISTS (
      SELECT 1 FROM information_schema.columns c
      WHERE c.table_name  = t.relname
        AND c.column_name = 'tenant_id'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM pg_index  ix
      JOIN pg_class  i ON i.oid = ix.indexrelid
      JOIN pg_attribute a ON a.attrelid = t.oid
          AND a.attnum = ANY(ix.indkey)
          AND a.attname = 'tenant_id'
      WHERE ix.indrelid = t.oid
  )
ORDER BY t.relname;


-- ── Roles with BYPASSRLS (should only be admin_role) ─────────────────────────
SELECT rolname, rolbypassrls
FROM pg_roles
WHERE rolbypassrls = true
ORDER BY rolname;


-- ── Full policy audit ─────────────────────────────────────────────────────────
SELECT
    tablename,
    policyname,
    roles,
    permissive,
    cmd,
    qual        AS using_expr,
    with_check  AS check_expr
FROM pg_policies
ORDER BY tablename, policyname;


-- ── Current session context (run as application_role to debug) ────────────────
SELECT
    current_user                                                AS role,
    current_setting('app.current_tenant',   true)              AS tenant_id,
    current_setting('app.current_employee', true)              AS employee_id,
    current_setting('app.user_type',        true)              AS user_type;
```

---

*This document is the authoritative schema implementation reference for Awo ERP. Every table, view, function, trigger, and migration must conform to the patterns and checklists defined here.*
