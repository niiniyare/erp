# Awo ERP — Finance Module
## FILE-09: Authorization, Roles & Segregation of Duties

**Document Version:** 2.0.0  
**Series:** FILE-09 of 10  
**Depends On:** FILE-01 (Domain Model — Tenant, Organisation), FILE-03 (Journal Pipeline — approval check)  
**Depended On By:** FILE-10 (Database Schema — RLS policies, API middleware)

---

## Table of Contents

1. [Why Authorization Exists as a Separate Layer](#1-why-authorization-exists-as-a-separate-layer)
2. [What You Lose Without It](#2-what-you-lose-without-it)
3. [How ERPNext, NetSuite and QuickBooks Handle Authorization](#3-how-erpnext-netsuite-and-quickbooks-handle-authorization)
4. [Permission Model (MRA Framework)](#4-permission-model-mra-framework)
5. [Role Definitions](#5-role-definitions)
6. [Segregation of Duties (SOD)](#6-segregation-of-duties-sod)
7. [Approval Authority Limits](#7-approval-authority-limits)
8. [Time-Limited Access](#8-time-limited-access)
9. [Portal and API Roles](#9-portal-and-api-roles)
10. [Row-Level Security in PostgreSQL](#10-row-level-security-in-postgresql)
11. [Audit Trail](#11-audit-trail)
12. [Business Rules & Validation](#12-business-rules--validation)
13. [Performance & Storage](#13-performance--storage)
14. [Database Schema](#14-database-schema)
15. [API Reference](#15-api-reference)
16. [Feature Flags & Configuration](#16-feature-flags--configuration)
17. [v1.0 Rollout Assessment](#17-v10-rollout-assessment)

---

## 1. Why Authorization Exists as a Separate Layer

Financial systems attract fraud. The most common financial frauds — ghost payroll employees, fictitious vendor payments, journal entry manipulation, false expense claims — all succeed when a single person controls too much of a financial process: creating the transaction, approving it, and confirming its execution.

Authorization in Awo is not just about "who can log in." It is a layered system that:

1. **Controls what actions each role can perform** on each financial resource (create a journal, approve a payment, close a period)
2. **Enforces segregation of duties** — certain role combinations are incompatible and the system prevents one person from holding both
3. **Imposes approval authority limits** — a finance manager can approve up to 500,000 KES; above that requires CFO
4. **Restricts data visibility** through Row-Level Security in the database — a user cannot even see data they are not authorised to access, regardless of which API endpoint they call
5. **Records every action permanently** in an append-only audit trail

The authorization layer is not a feature — it is a prerequisite for any financial system that will be trusted by auditors, management, or regulators.

---

## 2. What You Lose Without It

| Control | Without Authorization Layer | With Authorization Layer |
|---|---|---|
| Fraud prevention | One person can create and approve their own payments | SOD prevents self-approval programmatically |
| Audit readiness | Cannot prove who did what and when | Every action logged with user, timestamp, IP |
| Data isolation | Any user can see all data | RLS ensures users only see their tenant's data |
| Regulatory compliance (SOX, IFRS) | Cannot demonstrate control effectiveness | Controls are documented, tested, and enforced |
| Approval chain integrity | Anyone can post anything | Authority limits and approval chains are enforced |
| Period integrity | Anyone can reopen a locked period | Period operations are gated behind specific roles |

---

## 3. How ERPNext, NetSuite and QuickBooks Handle Authorization

### ERPNext

ERPNext uses a role-based permissions model where each DocType has a permissions matrix defining which roles can `read`, `write`, `create`, `delete`, `submit`, `cancel`, `amend`. Roles are composable — a user can have multiple roles.

**Strengths:** Granular per-DocType permissions. Field-level permissions (hide specific fields from certain roles). User-group-based sharing.

**Weaknesses:** SOD enforcement is manual — the system does not prevent a user from holding incompatible roles. Approval workflows are a separate system (the Workflow DocType) that operates independently from the permission system. Authority limits require custom scripting. Row-level security is application-level only (no database-level RLS).

### NetSuite

NetSuite uses a permission matrix per role per record type per access level (`view`, `create`, `edit`, `full`). Role-based restrictions can include subsidiary scope (an AP clerk for the Kenya subsidiary cannot see UAE data).

**Strengths:** Very granular. Subsidiary-level scoping. Support for SOD through role restrictions. Approval routing with authority limits is built in.

**Weaknesses:** Complex to configure — hundreds of permission combinations. No database-level enforcement (application layer only).

### QuickBooks

QuickBooks Online has four access levels: `Company Admin`, `Accountant/Bookkeeper`, `Standard User`, `Reports Only`. This is extremely coarse — there is no way to allow a user to create AP invoices but not AR invoices, for example.

**What Awo provides:** Granular action-level permissions (ERPNext-style) combined with database-level RLS (which no ERPNext or QuickBooks deployment has) and programmatic SOD enforcement.

---

## 4. Permission Model (MRA Framework)

### Module · Resource · Action

Every permission in the Finance module follows the pattern:

```
finance.{resource}.{action}
```

This is the MRA (Module-Resource-Action) framework used across all Awo ERP modules.

### Resource and Action Registry

| Resource Slug | Label | Available Actions |
|---|---|---|
| `accounts` | Chart of Accounts | `read` `create` `update` `deactivate` |
| `transactions` | Journal Entries | `read` `create` `submit` `approve` `post` `reverse` `void` |
| `periods` | Accounting Periods | `read` `close` `reopen` `lock` |
| `fiscal-years` | Fiscal Years | `read` `create` `close` `lock` |
| `currencies` | Currencies & Rates | `read` `create` `update` `load-rates` |
| `cost-centres` | Cost Centres | `read` `create` `update` `deactivate` |
| `budgets` | Budgets | `read` `create` `update` `approve` `activate` |
| `bank-accounts` | Bank Accounts | `read` `create` `update` `deactivate` |
| `reconciliation` | Bank Reconciliation | `read` `create` `match` `submit` `approve` `lock` |
| `payments` | Payment Runs | `read` `create` `approve` `post` `cancel` |
| `petty-cash` | Petty Cash | `read` `disburse` `replenish` `count` |
| `tax-config` | Tax Configuration | `read` `create` `update` |
| `intercompany` | Intercompany | `read` `create` `match` `eliminate` |
| `reports` | Financial Reports | `read` `run` `export` `schedule` `share` |
| `settings` | Finance Settings | `read` `update` |

### Permission Resolution

Permissions are resolved in the `PostingConfig` passed to the journal pipeline and in the HTTP middleware for all Finance API endpoints:

```go
type Permissions struct {
    CanReadTransactions    bool
    CanCreateTransactions  bool
    CanApproveTransactions bool
    CanPostTransactions    bool
    CanReverseTransactions bool
    ApprovalAmountLimit    *decimal.Decimal // nil = unlimited
    CanClosePeriodsRole    string           // "finance_manager" | "cfo" | ""
    CanLockPeriods         bool
    CanRunReports          bool
    CanExportReports       bool
    // ... one field per action across all resources
}

func ResolvePermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (Permissions, error) {
    roles, err := roleRepo.GetRolesForUser(ctx, userID, tenantID)
    if err != nil {
        return Permissions{}, err
    }
    return mergePermissions(roles), nil
}
```

Permissions from multiple roles are merged with OR logic — holding `ap_clerk` AND `finance_viewer` grants the union of both roles' permissions.

---

## 5. Role Definitions

### Role Hierarchy

Roles are defined as additive — each level inherits all permissions of roles below it.

```
role:report-viewer
    └── role:finance-viewer
            ├── role:ap-clerk (sibling — not inherited)
            ├── role:ar-clerk (sibling — not inherited)
            └── role:finance-manager
                    └── role:cfo
                            └── role:tenant-admin
```

`ap-clerk` and `ar-clerk` are siblings of `finance-viewer` in the hierarchy, not children — they add creation capabilities on top of read-only, but they do not inherit each other.

### role:report-viewer

**Purpose:** Non-finance staff who need to see financial figures for their own area (e.g. a department head reviewing their cost centre P&L).

| Resource | Actions |
|---|---|
| `reports` | `read` `run` `export` |
| `accounts` | `read` |
| `cost-centres` | `read` |
| `periods` | `read` |

**Scoping:** Report viewers can optionally be scoped to a single cost centre. A cost centre-scoped report viewer sees only reports filtered to their assigned cost centre.

### role:finance-viewer

**Purpose:** Read-only access to all financial data. Suitable for internal auditors during fieldwork or management who need visibility but should not create entries.

Inherits `report-viewer` plus:

| Resource | Additional Actions |
|---|---|
| `transactions` | `read` |
| `budgets` | `read` |
| `bank-accounts` | `read` |
| `reconciliation` | `read` |
| `currencies` | `read` |
| `tax-config` | `read` |
| `payments` | `read` |
| `petty-cash` | `read` |

### role:ap-clerk

**Purpose:** Accounts payable operations — entering supplier invoices, processing payments, managing petty cash.

| Resource | Actions |
|---|---|
| `transactions` | `read` `create` `submit` |
| `payments` | `read` `create` |
| `petty-cash` | `read` `disburse` `replenish` |
| `bank-accounts` | `read` |
| `accounts` | `read` |
| `currencies` | `read` |
| `reports` | `read` `run` `export` |

**Cannot:** Approve or post transactions. Cannot approve own payment run. Cannot approve own petty cash replenishment.

### role:ar-clerk

**Purpose:** Accounts receivable operations — recording customer receipts, managing customer balances, preparing bank reconciliations.

| Resource | Actions |
|---|---|
| `transactions` | `read` `create` `submit` |
| `bank-accounts` | `read` |
| `reconciliation` | `read` `create` `match` |
| `accounts` | `read` |
| `currencies` | `read` |
| `reports` | `read` `run` `export` |

**Cannot:** Approve own entries. Cannot write off bad debts without Finance Manager approval.

### role:finance-manager

**Purpose:** Full finance operations. Approves transactions, closes periods, manages reconciliations. Inherits `finance-viewer`.

| Resource | Actions | Authority Limit |
|---|---|---|
| `transactions` | `read` `create` `submit` `approve` `post` `reverse` | Up to 500,000 KES |
| `periods` | `read` `close` `reopen` | Current + prior period only |
| `budgets` | `read` `create` `update` `approve` | — |
| `bank-accounts` | `read` `create` `update` | — |
| `reconciliation` | All actions | — |
| `payments` | `read` `create` `approve` `post` | Up to 500,000 KES |
| `cost-centres` | `read` `create` `update` | — |
| `currencies` | `read` `create` `update` `load-rates` | — |
| `accounts` | `read` `create` `update` `deactivate` | — |
| `reports` | `read` `run` `export` `schedule` `share` | — |

**Cannot:** Lock fiscal years. Approve own transactions (SOD).

### role:cfo

**Purpose:** Full financial authority. Inherits all of `finance-manager` plus high-value approvals, fiscal year management, and tax configuration.

| Resource | Additional Actions | Limit |
|---|---|---|
| `transactions` | All — no amount limit | Unlimited |
| `periods` | `lock` | Any period |
| `fiscal-years` | `read` `close` `lock` | — |
| `tax-config` | `read` `create` `update` | — |
| `settings` | `read` `update` | — |
| `payments` | All — no amount limit | Unlimited |
| `intercompany` | All actions | — |
| `budgets` | `activate` | — |

### role:auditor

**Purpose:** Time-limited read-only access across all finance resources. Always assigned with an expiry date. Cannot be self-assigned.

| Resource | Actions |
|---|---|
| All finance resources | `read` only |
| `reports` | `read` `run` `export` |

The auditor role is always assigned with `expires_at`. The system automatically revokes it at expiry. Even if a CFO extends the engagement without revoking access, the system will revoke it automatically when the `expires_at` timestamp passes.

### role:budget-owner

**Purpose:** Departmental budget responsibility for a specific cost centre. Can view and propose budget amendments for their assigned area.

| Resource | Actions |
|---|---|
| `budgets` | `read` `create` (draft only, own cost centre) |
| `cost-centres` | `read` (own subtree only) |
| `reports` | `read` `run` (cost centre-scoped) |
| `transactions` | `read` (own cost centre only) |

---

## 6. Segregation of Duties (SOD)

### What SOD Is

Segregation of duties is the principle that no single person should control all steps of any significant financial process. Awo enforces SOD programmatically — it is not a policy document or a manual review process. The system prevents incompatible role combinations from being assigned to the same user.

### Incompatible Role Combinations

The following pairs are blocked at the role assignment layer. Attempting to assign the second role to a user who already holds the first returns an error:

| Role A | Role B | Reason |
|---|---|---|
| `ap-clerk` | `finance-manager` | Cannot create and approve own AP invoices |
| `ar-clerk` | `finance-manager` | Cannot create and approve own AR entries |
| `ap-clerk` | `ap-clerk` (self-approve payment run) | Enforced as self-approval prohibition |
| Budget creator | Budget approver (same budget) | Cannot self-approve own budget |
| Reconciliation preparer | Reconciliation approver | Cannot self-approve own bank rec |
| Payment creator | Payment approver | Cannot self-approve own payment run |
| User administrator | Transaction approver | Cannot create ghost approver accounts |

### Self-Approval Prohibition

Beyond incompatible roles, self-approval is prohibited at the individual transaction level — not just the role level. Even if a user holds the `finance-manager` role (which includes both `submit` and `approve`), they cannot approve their own submission. This is enforced in:

1. The `CheckApproval` pipeline stage:
```go
if entry.ApprovedBy != nil && *entry.ApprovedBy == entry.SubmittedBy {
    return ErrSelfApprovalProhibited{
        UserID: *entry.ApprovedBy,
        EntryID: entry.ID,
    }
}
```

2. The `BankReconciliation.Approve()` method:
```go
if r.PreparedBy == by {
    return ErrSelfApprovalProhibited
}
```

3. The `PaymentRun.Approve()` validation in the application service.

### SOD Enforcement Architecture

SOD enforcement happens at three levels — all three must be passed:

```
Level 1: Role assignment (prevent assigning incompatible roles to the same user)
Level 2: API middleware (check permissions before processing any request)
Level 3: Domain method (self-approval check inside aggregate methods)
```

Level 3 (domain method) is the last line of defence. It catches cases where the API middleware was bypassed (e.g. a background job, a direct service call in a test, or a future code path that forgets to check permissions).

---

## 7. Approval Authority Limits

### Amount-Based Escalation

Approval authority limits define the maximum transaction amount that each role can approve without escalation:

| Role | Manual Journal | Payment Run | Notes |
|---|---|---|---|
| `ap-clerk` | Cannot approve | Cannot approve | Clerks create, managers approve |
| `ar-clerk` | Cannot approve | N/A | Clerks create, managers approve |
| `finance-manager` | 500,000 KES | 500,000 KES | Configurable via `finance_manager_approval_limit` |
| `cfo` | Unlimited | Unlimited | No amount limit |
| `tenant-admin` | Unlimited | Unlimited | No amount limit |

When a transaction exceeds a user's approval authority limit, the system escalates automatically: the approval request is re-routed to the next authority tier. The original approver is notified that the transaction exceeded their limit and has been escalated.

```go
func (s *ApprovalService) Route(ctx context.Context, entry *JournalEntry, approver User) error {
    limit := s.getLimitForRole(approver.Role)
    if limit != nil && entry.TotalAmount.GreaterThan(*limit) {
        // Escalate to CFO
        return s.escalateToCFO(ctx, entry, approver)
    }
    return s.approve(ctx, entry, approver)
}
```

### Multi-Tier Approval for Large Amounts

Amounts above the CFO's single-approver threshold (configurable, default 5,000,000 KES) require sequential approval:

```
Finance Manager reviews → CFO approves → CEO notified
```

The CEO notification is informational — the CEO does not need to take action in the system unless the tenant's governance policy requires CEO approval (configurable via `ceo_approval_required_above`).

---

## 8. Time-Limited Access

### Auditor Access

The `auditor` role is always time-limited. The role assignment includes an `expires_at` timestamp. A background job runs hourly and revokes expired role assignments:

```go
type RoleAssignment struct {
    UserID     uuid.UUID
    TenantID   uuid.UUID
    Role       string
    AssignedBy uuid.UUID
    AssignedAt time.Time
    ExpiresAt  *time.Time  // nil = permanent (only for permanent roles)
    RevokedAt  *time.Time
    RevokedBy  *uuid.UUID
    Reason     string
}
```

The `auditor` role assignment cannot have a nil `ExpiresAt` — this is enforced at the application level. Any attempt to assign the `auditor` role without an expiry is rejected.

### Delegation

When a `finance-manager` is on leave, they can delegate their approval authority to another user for a defined period:

```go
type ApprovalDelegation struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    DelegatorID  uuid.UUID
    DelegateID   uuid.UUID
    ValidFrom    time.Time
    ValidTo      time.Time
    Scope        string        // "all" | "amount_below" | "specific_type"
    AmountLimit  *decimal.Decimal
    EntryTypes   []string
    IsActive     bool
}
```

During the delegation period, approval requests that would normally route to the delegator are also routed to the delegate. The delegator can still approve directly — the delegation does not remove their own authority.

---

## 9. Portal and API Roles

### Portal Roles (External Users)

Portal actors operate in a completely isolated domain (`{tenantID}:portal`) and cannot see internal ERP data.

| Role | Subject | Allowed Actions |
|---|---|---|
| `role:portal-customer` | Customer contact | View own invoices, download statements, see payment status |
| `role:portal-supplier` | Supplier contact | View own POs, submit invoices electronically, track payment |

Portal actors cannot access the Finance module's internal journal entries, account balances, or management reports. They can only see their own sub-ledger data (AR entries for customers, AP entries for suppliers) through purpose-built portal views.

### API Roles (Machine Clients)

API actors operate in `{tenantID}:api` domain. A compromised API key cannot escalate to tenant-user rights.

| Role | Typical Use | Permissions |
|---|---|---|
| `role:api-readonly` | BI tools, reporting integrations | `read` on all finance resources |
| `role:api-invoice-submit` | Billing systems posting invoices | `create` journal entries (integration type only) |
| `role:api-payment-notify` | Bank webhooks confirming payment | `post` payment receipts, `read` bank accounts |
| `role:api-full-access` | Trusted service accounts (internal microservices) | All actions |

API keys are tenant-scoped and cannot cross tenant boundaries. The `api-full-access` role should only be assigned to the internal Awo ERP service account, not to external integrations.

---

## 10. Row-Level Security in PostgreSQL

### Why Database-Level RLS

Application-level permission checks are necessary but not sufficient. A bug in the application, a direct database query from a support tool, a misconfigured API endpoint, or a SQL injection vulnerability could all bypass application-level checks. Database-level RLS is the final boundary — it is enforced by the database engine regardless of how a query reaches it.

Every table in the Finance module has RLS enabled and forced:

```sql
ALTER TABLE journal_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_entries FORCE  ROW LEVEL SECURITY;  -- applies to table owners too
```

Two policies cover all cases:

```sql
-- application_role: full access to the current tenant's rows
CREATE POLICY {table}_app ON {table} FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- readonly_role: SELECT only, for reporting and read replicas
CREATE POLICY {table}_ro ON {table} FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Tenant Context Setting

The tenant ID is set in the PostgreSQL session at connection acquisition time:

```go
func (p *Pool) AcquireForTenant(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Conn, error) {
    conn, err := p.pool.Acquire(ctx)
    if err != nil {
        return nil, err
    }
    _, err = conn.Exec(ctx,
        "SELECT set_config('app.current_tenant', $1, true)",
        tenantID.String(),
    )
    if err != nil {
        conn.Release()
        return nil, err
    }
    return conn, nil
}
```

The `true` parameter to `set_config` means the setting is transaction-local — it is reset at the end of each transaction, preventing a leaked tenant context from affecting subsequent queries on the same connection.

### admin_role Bypass

The `admin_role` database role is assigned `BYPASSRLS`. This role is used exclusively for:
- Schema migrations
- Support debugging under explicit authorisation
- Backup and restore operations

It is never used by the application server. Any query executed as `admin_role` is logged in the PostgreSQL audit log.

---

## 11. Audit Trail

### What Is Logged

Every action that changes the state of a financial record is logged in the `audit_log` table:

```go
type AuditEntry struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    TableName   string
    RecordID    uuid.UUID
    Action      string    // INSERT | UPDATE | DELETE
    ChangedBy   uuid.UUID
    ChangedAt   time.Time
    OldValues   json.RawMessage  // nil for INSERT
    NewValues   json.RawMessage  // nil for DELETE
    IPAddress   string
    UserAgent   string
    SessionID   string
}
```

### Audit Trigger

A generic PostgreSQL trigger captures all changes:

```sql
CREATE OR REPLACE FUNCTION fn_audit_trigger()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
    INSERT INTO audit_log (
        tenant_id, table_name, record_id, action,
        changed_by, changed_at, old_values, new_values,
        ip_address, session_id
    ) VALUES (
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        current_setting('app.current_user_id', true)::UUID,
        NOW(),
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE row_to_json(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE row_to_json(NEW) END,
        current_setting('app.client_ip', true),
        current_setting('app.session_id', true)
    );
    RETURN COALESCE(NEW, OLD);
END;$$;
```

The audit trigger is attached to:
- `journal_entries` (all operations)
- `journal_lines` (INSERT only — updates/deletes are blocked on posted entries)
- `accounting_periods` (UPDATE — status changes)
- `accounts` (INSERT, UPDATE)
- `account_mappings` (all operations)
- `budgets` and `budget_lines` (all operations)
- `bank_reconciliations` (UPDATE — status changes)
- `payment_runs` (UPDATE — status changes)

### Audit Log Immutability

The `audit_log` table has no `UPDATE` or `DELETE` policies — only `INSERT`:

```sql
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE  ROW LEVEL SECURITY;

CREATE POLICY audit_app_insert ON audit_log FOR INSERT TO application_role
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY audit_app_select ON audit_log FOR SELECT TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- No UPDATE or DELETE policy — these operations are impossible for application_role
```

Even the application server cannot delete or modify audit log entries. Only `admin_role` (BYPASSRLS) can delete them, and doing so would require a deliberate action that itself would be visible in the PostgreSQL server log.

### Audit Log Shipping

Audit entries are shipped to an external SIEM (Security Information and Event Management) system within 60 seconds via a Temporal workflow that tails the `audit_log` table using a cursor. This provides an offsite copy that remains accessible even if the primary database is compromised.

---

## 12. Business Rules & Validation

### Authorization Rules

| Rule ID | Rule | Enforcement |
|---|---|---|
| `FIN-AUTH-001` | A user cannot approve their own submissions | Domain method + pipeline stage |
| `FIN-AUTH-002` | Approval authority limits are enforced — a finance manager cannot approve above their limit | ApprovalService.Route() |
| `FIN-AUTH-003` | SOD: incompatible role combinations cannot be assigned to the same user | RoleAssignmentService.Assign() |
| `FIN-AUTH-004` | The `auditor` role must always have an `expires_at` — permanent assignment is blocked | Application hard block |
| `FIN-AUTH-005` | Period lock requires CFO role — no exception | PeriodService.Lock() permission check |
| `FIN-AUTH-006` | Fiscal year lock requires CFO role | FiscalYearService.Lock() permission check |
| `FIN-AUTH-007` | Accessing another tenant's data is impossible via the application_role | PostgreSQL RLS |
| `FIN-AUTH-008` | A portal user cannot access internal journal entries or account balances | Portal domain isolation |
| `FIN-AUTH-009` | Admin_role access is logged in the PostgreSQL audit log and requires explicit authorisation | Operational policy |

### Alternatives Considered

**Alternative: Attribute-based access control (ABAC) instead of RBAC.**
Deferred to v2. ABAC would allow fine-grained rules like "user X can approve transactions for cost centre Y but not cost centre Z." This level of granularity is not required at v1.0. RBAC with cost centre scoping covers the initial use cases.

**Alternative: Row-level security in the application only (no DB RLS).**
Rejected. Application-only enforcement creates a gap exploitable by support tooling, background jobs, and API bugs. DB-level RLS closes this gap definitively.

**Alternative: Permissive audit log (application can delete entries).**
Rejected. An audit log that can be tampered with by the application is not an audit log — it is just another table. The immutability is the entire point.

---

## 13. Performance & Storage

### Permission Check Latency

Permission checks happen on every API request. They must be fast.

| Operation | Mechanism | Target |
|---|---|---|
| Load user roles | Redis cache (keyed by user_id + tenant_id), TTL 5 minutes | < 1ms (cache hit) |
| Resolve merged permissions | In-memory bitfield merge | < 0.1ms |
| Self-approval check (domain method) | In-memory UUID comparison | Negligible |
| Authority limit check | In-memory decimal comparison | Negligible |

Cache invalidation: when a user's roles change, their cache entry is invalidated immediately. Role changes are rare (a few per month per user) — the cost of a cache miss is one DB query (~5ms), which is acceptable.

### Audit Log Volume

| Scenario | Log Entries/Day | Storage/Day (avg 800 bytes/row) |
|---|---|---|
| Single station (light activity) | ~500 | ~400 KB |
| 5-site operator | ~5,000 | ~4 MB |
| 20-site operator | ~20,000 | ~16 MB |

Audit log retention is 7 years minimum (regulatory requirement). For a 5-site operator: **~14 GB** of audit log over 7 years. For a 20-site operator: **~56 GB**.

Audit log rows are append-only and compressed well (high repetition in `table_name`, `action`, `tenant_id`). With TOAST compression, actual storage is approximately 40% of the uncompressed estimate.

**Partitioning:** The `audit_log` table is range-partitioned by `changed_at` (monthly partitions). Old partitions are moved to cheaper storage after 2 years but remain queryable. This keeps the active partition small and fast.

```sql
CREATE TABLE audit_log (
    tenant_id   UUID          NOT NULL,
    id          UUID          NOT NULL DEFAULT gen_random_uuid(),
    changed_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    -- ... other fields
) PARTITION BY RANGE (changed_at);

CREATE TABLE audit_log_2025_01 PARTITION OF audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

---

## 14. Database Schema

```sql
-- ── role_assignments ──────────────────────────────────────────────────────────

CREATE TABLE role_assignments (
    tenant_id   UUID        NOT NULL REFERENCES tenants(id),
    id          UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id     UUID        NOT NULL,
    role        TEXT        NOT NULL,
    assigned_by UUID        NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,          -- NULL only for permanent roles; mandatory for 'auditor'
    revoked_at  TIMESTAMPTZ,
    revoked_by  UUID,
    reason      TEXT,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, user_id, role),
    CONSTRAINT chk_auditor_must_expire
        CHECK (role <> 'auditor' OR expires_at IS NOT NULL)
);

CREATE INDEX idx_ra_user    ON role_assignments (tenant_id, user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_ra_expires ON role_assignments (expires_at)
    WHERE expires_at IS NOT NULL AND revoked_at IS NULL;

ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_assignments FORCE  ROW LEVEL SECURITY;
CREATE POLICY ra_app ON role_assignments FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

-- ── approval_delegations ──────────────────────────────────────────────────────

CREATE TABLE approval_delegations (
    tenant_id    UUID        NOT NULL REFERENCES tenants(id),
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    delegator_id UUID        NOT NULL,
    delegate_id  UUID        NOT NULL,
    valid_from   TIMESTAMPTZ NOT NULL,
    valid_to     TIMESTAMPTZ NOT NULL,
    scope        TEXT        NOT NULL DEFAULT 'all'
                     CHECK (scope IN ('all','amount_below','specific_type')),
    amount_limit NUMERIC(18,4),
    entry_types  TEXT[],
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    created_by   UUID        NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT chk_different_users CHECK (delegator_id <> delegate_id),
    CONSTRAINT chk_valid_range     CHECK (valid_to > valid_from)
);

CREATE INDEX idx_deleg_active ON approval_delegations (tenant_id, delegate_id, valid_from, valid_to)
    WHERE is_active = TRUE;

-- ── audit_log ─────────────────────────────────────────────────────────────────

CREATE TABLE audit_log (
    tenant_id   UUID          NOT NULL,
    id          UUID          NOT NULL DEFAULT gen_random_uuid(),
    changed_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    table_name  TEXT          NOT NULL,
    record_id   UUID          NOT NULL,
    action      TEXT          NOT NULL CHECK (action IN ('INSERT','UPDATE','DELETE')),
    changed_by  UUID,
    old_values  JSONB,
    new_values  JSONB,
    ip_address  TEXT,
    user_agent  TEXT,
    session_id  TEXT,

    PRIMARY KEY (id, changed_at)  -- partition key must be in PK
) PARTITION BY RANGE (changed_at);

-- Create initial partitions (new partitions added by a monthly cron job)
CREATE TABLE audit_log_2025_01 PARTITION OF audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE audit_log_2025_02 PARTITION OF audit_log
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- ... etc.

CREATE INDEX idx_audit_tenant_table ON audit_log (tenant_id, table_name, changed_at DESC);
CREATE INDEX idx_audit_record       ON audit_log (tenant_id, record_id, changed_at DESC);
CREATE INDEX idx_audit_user         ON audit_log (tenant_id, changed_by, changed_at DESC);

-- Audit log: INSERT only for application_role — no UPDATE or DELETE
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE  ROW LEVEL SECURITY;

CREATE POLICY audit_insert ON audit_log FOR INSERT TO application_role
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY audit_select ON audit_log FOR SELECT TO application_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY audit_ro ON audit_log FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

---

## 15. API Reference

### Role Assignments

```
GET    /auth/users/{id}/roles              List roles for a user
POST   /auth/users/{id}/roles              Assign role (with optional expires_at)
DELETE /auth/users/{id}/roles/{role}       Revoke role assignment (soft revoke — records revoker)
GET    /auth/sod-matrix                    View incompatible role combinations
POST   /auth/sod-check                     Check if a proposed role assignment creates a SOD violation
```

**Assign role:**
```json
{
  "role": "auditor",
  "expires_at": "2025-03-31T23:59:59Z",
  "reason": "FY2024 annual audit engagement"
}
```

### Approval Delegations

```
GET    /finance/delegations                List active delegations
POST   /finance/delegations               Create delegation
DELETE /finance/delegations/{id}          Deactivate delegation
```

### Audit Trail

```
GET    /finance/audit-log                  Query audit entries (?table=, ?record_id=, ?user_id=, ?from=, ?to=)
GET    /finance/audit-log/{id}             Get single audit entry with full before/after values
POST   /finance/audit-log/export           Export audit log for a date range (async, returns execution_id)
```

---

## 16. Feature Flags & Configuration

| Flag | Type | Default | Description |
|---|---|---|---|
| `finance_manager_approval_limit` | decimal | `500000` | Maximum KES amount a finance_manager can approve. Amounts above this auto-escalate to CFO. |
| `ceo_approval_required_above` | decimal | `null` | If set, amounts above this require CEO notification (not approval unless `ceo_approval_blocks = true`). |
| `ceo_approval_blocks` | bool | `false` | If true, CEO must approve (not just be notified) for amounts above `ceo_approval_required_above`. |
| `multi_approver_threshold` | decimal | `5000000` | Above this, sequential Finance Manager + CFO approval is required. |
| `sod_enforcement_mode` | enum | `block` | `block` — prevent SOD violations at role assignment; `warn` — allow with logged warning (not recommended for production). |
| `auditor_role_max_days` | int | `90` | Maximum duration for an auditor role assignment. |
| `audit_log_shipping_enabled` | bool | `false` | Enable shipping audit logs to external SIEM. |
| `audit_log_siem_endpoint` | string | `null` | SIEM webhook URL for audit log shipping. |
| `audit_log_shipping_delay_seconds` | int | `60` | Maximum delay between audit entry creation and SIEM shipping. |
| `permission_cache_ttl_seconds` | int | `300` | How long to cache user permissions in Redis. |
| `delegation_max_days` | int | `30` | Maximum duration for an approval delegation. |
| `require_reason_for_period_reopen` | bool | `true` | Period reopen requires a non-empty reason. |
| `require_dual_auth_for_fiscal_year_lock` | bool | `false` | Fiscal year lock requires both CFO and CEO actions (v1.2 feature). |

---

## 17. v1.0 Rollout Assessment

### Must Have at v1.0

- All role definitions and permission keys
- SOD enforcement at role assignment level (block mode)
- Self-approval prohibition in domain methods and pipeline stage
- Authority limits for finance_manager (configurable amount)
- PostgreSQL RLS on all Finance module tables (non-negotiable)
- Audit trigger on all critical tables (`journal_entries`, `journal_lines`, `accounting_periods`)
- Tenant context setting pattern in connection pool
- Permission caching with Redis (in-process cache acceptable at v1.0 if Redis is not deployed)
- Role assignment API (assign, revoke)

### Can Be Deferred

| Feature | Suggested Release |
|---|---|
| Auditor role with auto-expiry | v1.1 — manual revocation acceptable at v1.0 |
| Approval delegation | v1.1 |
| Multi-tier approval (Finance Manager + CFO sequential) | v1.1 |
| Audit log partitioning | v1.1 — start with a single table; partition once volume warrants it |
| Audit log shipping to SIEM | v1.1 |
| Portal roles | v1.2 (when customer/supplier portal is built) |
| Cost centre scoped report viewer | v1.1 |

### Never Defer

- RLS enabled and forced on every table — `FORCE ROW LEVEL SECURITY` ensures even table owners are subject to RLS
- Self-approval prohibition at domain method level (not just API layer)
- Audit trigger on `journal_entries` and `accounting_periods`
- Immutable audit log (no UPDATE/DELETE policy for application_role)
- `chk_auditor_must_expire` DB constraint — prevents accidental permanent auditor assignments

---

*End of FILE-09. Proceeding to FILE-10: Database Schema, Workflow Orchestration, API Reference & Configuration Catalogue.*
