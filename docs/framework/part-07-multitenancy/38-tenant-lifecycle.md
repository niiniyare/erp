---
title: "Chapter 38: Tenant Lifecycle"
part: "Part VII — Multi-Tenancy Operations"
chapter: 38
section: "38-tenant-lifecycle"
related:
  - "[Chapter 14: Multi-Tenancy Middleware](../part-03-api/14-multitenancy-middleware.md)"
  - "[Chapter 27: Defining Workflows](../part-05-workflow/27-defining-workflows.md)"
---

# Chapter 38: Tenant Lifecycle

A tenant in Awo is the top-level isolation boundary — a business customer running their ERP on the platform. All tenants share a single PostgreSQL database and schema; isolation is enforced by Row-Level Security (RLS) keyed on `tenant_id`. This chapter covers provisioning, status management, offboarding, and cloning.

---

## 38.1. Tenant Entity and Status Machine

### 38.1.1. Tenant Entity

```go
type Tenant struct {
    ID            uuid.UUID
    Slug          string     // URL-safe identifier: "acme-petroleum"
    Name          string     // Display name: "Acme Petroleum Ltd"
    Domain        *string    // Custom domain: "erp.acmepetroleum.co.ke"
    Plan          string     // "starter" | "growth" | "enterprise"
    Status        string     // PENDING | ACTIVE | SUSPENDED | ARCHIVED
    // Kenya business identity
    KraPIN        *string
    CompanyRegNo  *string
    // Configuration
    Timezone      string     // default "Africa/Nairobi"
    Currency      string     // default "KES"
    FiscalYearEnd string     // default "06-30" (June 30 — Kenya fiscal year)
    // Lifecycle timestamps
    TrialEndsAt   *time.Time
    SuspendedAt   *time.Time
    SuspendReason *string
    ArchivedAt    *time.Time
    ProvisionedAt *time.Time
}
```

No `DBSchema` field — there is no per-tenant schema. The tenant's data lives in shared tables filtered by `tenant_id = id` via RLS.

### 38.1.2. Status Machine

```
PENDING ──────────────────────────────► ARCHIVED
   │                                        ▲
   │ (provisioning complete)                │
   ▼                                        │
ACTIVE ──────── (payment failure) ──► SUSPENDED
   ▲                                        │
   │                                        │
   └────────── (payment received) ──────────┘
```

Valid transitions:
- `PENDING → ACTIVE`: provisioning workflow completes
- `ACTIVE → SUSPENDED`: payment failed or manual admin action
- `SUSPENDED → ACTIVE`: payment resolved
- `ACTIVE → ARCHIVED`: tenant requests account deletion (after data retention period)
- `SUSPENDED → ARCHIVED`: non-payment beyond grace period
- `PENDING → ARCHIVED`: provisioning abandoned

`ARCHIVED` is terminal. The tenant's rows remain in the database during the retention period but the RLS context will reject any API access (set_tenant_context() checks `status = 'ACTIVE'`).

### 38.1.3. HTTP Responses by Status

| Status | HTTP | Enforced by |
|---|---|---|
| PENDING | 503 + `Retry-After: 60` | Middleware status check |
| SUSPENDED | 402 Payment Required | Middleware + stored procedure |
| ARCHIVED | 410 Gone | Middleware status check |

---

## 38.2. Provisioning Workflow

### 38.2.1. Why a Workflow?

Provisioning involves multiple steps that can fail independently: inserting seed data, creating users, configuring feature flags, sending the welcome email. A Temporal workflow provides automatic retry, durable state across restarts, and a complete audit trail in the event history.

Unlike a schema-per-tenant model, there is **no DDL** during tenant provisioning (no `CREATE SCHEMA`, no `CREATE TABLE`). All tables already exist. Provisioning is purely data-level: inserting reference rows tagged with the new `tenant_id`.

### 38.2.2. Provisioning Workflow

```go
func TenantProvisioningWorkflow(ctx workflow.Context, params TenantProvisionParams) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts:        5,
            InitialInterval:        time.Second,
            BackoffCoefficient:     2.0,
            NonRetryableErrorTypes: []string{"TENANT_ALREADY_PROVISIONED"},
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Seed reference data (chart of accounts, leave types, tax bands)
    //    All rows inserted with tenant_id = params.TenantID
    workflow.ExecuteActivity(ctx, activities.SeedTenantReferenceData,
        SeedInput{TenantID: params.TenantID, Plan: params.Plan})

    // 2. Create platform admin user for tenant
    workflow.ExecuteActivity(ctx, activities.CreateTenantAdminUser,
        AdminUserInput{
            TenantID: params.TenantID,
            Email:    params.AdminEmail,
            Name:     params.AdminName,
        })

    // 3. Assign default roles
    workflow.ExecuteActivity(ctx, activities.AssignDefaultRoles, params.TenantID)

    // 4. Set up default feature flags
    workflow.ExecuteActivity(ctx, activities.ConfigureDefaultFeatureFlags,
        FeatureFlagInput{TenantID: params.TenantID, Plan: params.Plan})

    // 5. Mark tenant ACTIVE
    workflow.ExecuteActivity(ctx, activities.ActivateTenant, params.TenantID)

    // 6. Send welcome email
    workflow.ExecuteActivity(ctx, activities.SendWelcomeEmail,
        WelcomeEmailInput{
            AdminEmail: params.AdminEmail,
            TenantSlug: params.Slug,
            LoginURL:   fmt.Sprintf("https://%s.awo.so/login", params.Slug),
        })

    return nil
}
```

### 38.2.3. Seed Data Inserted on Provisioning

All seed rows are tagged `tenant_id = params.TenantID`. RLS ensures they are only visible to that tenant:

| Category | Rows seeded |
|---|---|
| Chart of accounts | 40+ accounts (Kenya KRA-aligned) |
| Leave types | Annual, Sick, Maternity, Paternity, Compassionate |
| PAYE bands | Current Kenya PAYE bands |
| NHIF/NSSF brackets | Current contribution tables |
| Roles | admin, accountant, cashier, salesperson, hr_manager, store_keeper, viewer |
| Feature flag overrides | Plan-appropriate defaults |
| Fiscal year | Current Kenya fiscal year (July–June) |
| Tenant config | Default timezone, currency, VAT settings |

### 38.2.4. Idempotency

The workflow ID is `{tenant_slug}.provision`. Re-triggering uses `REJECT_DUPLICATE` — the second call returns the first run's result without re-seeding. Each seed activity uses `INSERT ... ON CONFLICT DO NOTHING` to be safe even if somehow called twice.

---

## 38.3. Suspension and Reactivation

### 38.3.1. Suspension Workflow

```go
func TenantSuspensionWorkflow(ctx workflow.Context, params SuspendParams) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Invalidate all active sessions for this tenant's users
    workflow.ExecuteActivity(ctx, activities.InvalidateAllTenantSessions, params.TenantID)

    // 2. Mark tenant SUSPENDED in DB
    //    set_tenant_context() will now reject all API requests for this tenant
    workflow.ExecuteActivity(ctx, activities.SetTenantStatus,
        StatusUpdate{TenantID: params.TenantID, Status: "SUSPENDED", Reason: params.Reason})

    // 3. Pause Temporal Schedules for this tenant
    workflow.ExecuteActivity(ctx, activities.PauseTenantSchedules, params.TenantID)

    // 4. Notify admin users
    workflow.ExecuteActivity(ctx, activities.NotifyTenantSuspension,
        SuspendNotifyInput{TenantID: params.TenantID, Reason: params.Reason})

    return nil
}
```

After step 2, the `set_tenant_context()` stored procedure will raise an exception for this tenant ID because its status is no longer `ACTIVE`. Any in-flight request that reaches the DB after the status update will be rejected at the database level, not just the middleware level.

### 38.3.2. Reactivation

```go
func TenantReactivationWorkflow(ctx workflow.Context, params ReactivateParams) error {
    // 1. Mark tenant ACTIVE
    workflow.ExecuteActivity(ctx, activities.SetTenantStatus,
        StatusUpdate{TenantID: params.TenantID, Status: "ACTIVE"})

    // 2. Resume paused schedules
    workflow.ExecuteActivity(ctx, activities.ResumeTenantSchedules, params.TenantID)

    // 3. Notify admin users
    workflow.ExecuteActivity(ctx, activities.NotifyTenantReactivated, params.TenantID)

    return nil
}
```

---

## 38.4. Offboarding and Archival

### 38.4.1. Data Retention Before Archival

Kenya data retention requirements:
- Financial records: 7 years (Companies Act, KRA)
- Employee records: 5 years after termination
- Customer records: 3 years after last transaction

The archival workflow does not immediately delete data. It:
1. Exports a full tenant data extract (SQL/JSON, keyed by `WHERE tenant_id = ?`)
2. Uploads to S3 with AES-256 encryption
3. Marks tenant `ARCHIVED`
4. Schedules physical row deletion after the retention expiry date

### 38.4.2. Offboarding Workflow

```go
func TenantOffboardingWorkflow(ctx workflow.Context, params OffboardParams) error {
    longAO := workflow.ActivityOptions{StartToCloseTimeout: 2 * time.Hour}
    ctx = workflow.WithActivityOptions(ctx, longAO)

    // 1. Export all rows WHERE tenant_id = params.TenantID across all tables
    var exportPath string
    workflow.ExecuteActivity(ctx, activities.ExportTenantData,
        ExportInput{TenantID: params.TenantID}).
        Get(ctx, &exportPath)

    // 2. Upload to S3
    var s3URI string
    workflow.ExecuteActivity(ctx, activities.UploadTenantArchive,
        UploadInput{LocalPath: exportPath, TenantID: params.TenantID}).
        Get(ctx, &s3URI)

    // 3. Mark ARCHIVED (point of no return for API access)
    workflow.ExecuteActivity(ctx, activities.SetTenantStatus,
        StatusUpdate{TenantID: params.TenantID, Status: "ARCHIVED"})

    // 4. Purge Redis keys for tenant
    workflow.ExecuteActivity(ctx, activities.PurgeTenantRedisKeys, params.TenantID)

    // 5. Schedule physical row deletion at retention expiry
    workflow.ExecuteActivity(ctx, activities.SchedulePhysicalDeletion,
        DeletionSchedule{TenantID: params.TenantID, DeleteAfter: params.RetentionExpiry})

    // 6. Send confirmation email with archive download link
    workflow.ExecuteActivity(ctx, activities.SendOffboardingConfirmation,
        OffboardConfirmInput{
            AdminEmail: params.AdminEmail,
            ArchiveURI: s3URI,
            ExpiresAt:  params.RetentionExpiry,
        })

    return nil
}
```

### 38.4.3. Physical Row Deletion

At retention expiry, a scheduled workflow runs:

```go
func TenantPhysicalDeletionWorkflow(ctx workflow.Context, params DeletionParams) error {
    // DELETE WHERE tenant_id = ? across all tenant-scoped tables
    // Done in batches to avoid long-running transactions
    workflow.ExecuteActivity(ctx, activities.DeleteTenantRows,
        DeleteInput{TenantID: params.TenantID, BatchSize: 10000})
    return nil
}
```

The `DeleteTenantRows` activity deletes 10,000 rows at a time across each table, sleeping briefly between batches to avoid table locks and WAL pressure. This is safe because the tenant is ARCHIVED — no queries are running against these rows.

### 38.4.4. GDPR/Data Subject Deletion

Individual employee or customer data deletion requests (Kenya Data Protection Act 2019):

```go
func DataSubjectDeletionWorkflow(ctx workflow.Context, params DeletionParams) error {
    // Anonymise PII — replace with pseudonymous identifiers
    // Financial records (JournalEntry, Payslip) are retained but de-identified
    workflow.ExecuteActivity(ctx, activities.AnonymisePII,
        AnonymiseInput{
            TenantID:    params.TenantID,
            SubjectType: params.SubjectType, // "employee" | "customer"
            SubjectID:   params.SubjectID,
        })

    // Delete non-financial personal records (notes, activity log, preferences)
    workflow.ExecuteActivity(ctx, activities.DeletePersonalRecords, params)

    // Revoke portal access
    workflow.ExecuteActivity(ctx, activities.RevokePortalAccess, params.SubjectID)

    // Record in audit log (the fact of deletion must itself be retained)
    workflow.ExecuteActivity(ctx, activities.RecordDataDeletion, params)

    return nil
}
```

Anonymisation replaces PII fields: name → `[DELETED]`, email → `anon_{sha256(id)}@deleted.local`, phone → `0000000000`, NationalID → `XXXXXXXXX`. The `tenant_id` and record ID are preserved so financial totals remain correct.

---

## 38.5. Tenant Cloning

### 38.5.1. Use Cases

- **Staging copy**: replicate a tenant's configuration from production to a staging tenant for testing
- **Franchise setup**: clone a template tenant's chart of accounts and roles to a new franchise member
- **Disaster recovery test**: test restore procedures on a cloned tenant without affecting production

### 38.5.2. Clone Workflow

```go
func TenantCloneWorkflow(ctx workflow.Context, params CloneParams) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Seed empty target tenant
    workflow.ExecuteActivity(ctx, activities.SeedTenantReferenceData,
        SeedInput{TenantID: params.TargetTenantID, Plan: params.Plan})

    // 2. Copy configuration entities from source tenant
    //    INSERT INTO ... SELECT ... FROM ... WHERE tenant_id = source_id
    //    with tenant_id replaced by target_id
    workflow.ExecuteActivity(ctx, activities.CloneConfiguration,
        CloneConfigInput{
            SourceTenantID: params.SourceTenantID,
            TargetTenantID: params.TargetTenantID,
            Tables: []string{
                "accounts", "cost_centres", "roles", "role_permissions",
                "leave_types", "customer_tiers", "item_groups", "warehouses",
                "feature_flag_overrides", "tenant_config",
            },
        })

    // 3. Optionally copy transactional data
    if params.IncludeData {
        workflow.ExecuteActivity(ctx, activities.CloneTransactionalData,
            CloneDataInput{
                SourceTenantID: params.SourceTenantID,
                TargetTenantID: params.TargetTenantID,
                AsOfDate:       params.DataAsOfDate,
            })
    }

    // 4. Reset user passwords in the clone
    workflow.ExecuteActivity(ctx, activities.ResetClonedUserPasswords, params.TargetTenantID)

    // 5. Activate cloned tenant
    workflow.ExecuteActivity(ctx, activities.ActivateTenant, params.TargetTenantID)

    return nil
}
```

**Configuration clone** uses `INSERT ... SELECT` rewriting `tenant_id`:

```sql
INSERT INTO accounts (id, tenant_id, code, name, type, ...)
SELECT gen_random_uuid(), current_tenant_id(), code, name, type, ...
FROM accounts
WHERE tenant_id = $source_tenant_id;
```

No file exports, no schema manipulation — pure SQL row-level copying within the same shared database.
