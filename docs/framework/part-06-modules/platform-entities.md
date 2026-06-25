---
title: "Part VI: Platform Entities"
part: "Part VI — Platform Entities"
chapter: 34
section: "platform-entities"
related:
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
---

# Part VI: Platform Entities

Platform entities are EntityDefinitions owned by the `awo.so/framework` package itself. Every Awo deployment, regardless of which domain modules it enables, has these entities. They form the infrastructure on which tenancy, identity, access control, observability, and runtime configurability are built.

This document is the authoritative reference for all platform entities. It explains why each entity belongs in the framework rather than in a module, what its fields mean, and how it integrates with the surrounding systems.

---

## Why These Entities Are in the Framework

A distinction runs through the codebase between what belongs in the framework and what belongs in modules.

**Framework entities** satisfy at least one of these criteria:
- Every Awo application needs them unconditionally.
- Other framework systems (middleware, permission resolution, RLS, feature flags) depend on them directly in Go code.
- Their schema is referenced by migrations that ship with the framework itself.

**Module entities** (Finance, CRM, Inventory, HR, Forecourt) are domain-specific. They live in separate Go module repositories and import `awo.so/framework`, but the framework never imports them. A deployment that only runs HR has no reason to have a `FuelTank` table.

The platform entities documented here are: Tenant, OrgNode, User, Role, Permission, UserRole, RolePermission, Session, AuditLog, FeatureFlag, FeatureFlagOverride, Notification, TenantConfig, CustomFieldDef, ReportDefinition.

---

## Chapter 34 — Tenant and Organisation

### 34.1 The Tenant Entity

A tenant is the top-level isolation boundary. It represents one business running on the platform.

All tenants share a single PostgreSQL database and schema. Isolation is enforced by Row-Level Security (RLS) keyed on `tenant_id`. There is no `CREATE SCHEMA` per tenant, no per-tenant migration, and no `db_schema` field on the Tenant entity.

```go
type Tenant struct {
    ID            uuid.UUID
    Slug          string      // URL-safe identifier: "acme-petroleum"
    Name          string      // Display name: "Acme Petroleum Ltd"
    Domain        *string     // Custom domain: "erp.acmepetroleum.co.ke"
    Plan          string      // "starter" | "growth" | "enterprise"
    Status        string      // "PENDING" | "ACTIVE" | "SUSPENDED" | "ARCHIVED"

    // Kenya business identity
    KraPIN        *string
    CompanyRegNo  *string

    // Locale and fiscal defaults
    Timezone      string      // default "Africa/Nairobi"
    Currency      string      // default "KES"
    FiscalYearEnd string      // default "06-30" (Kenya: July–June)

    // Lifecycle timestamps
    TrialEndsAt   *time.Time
    SuspendedAt   *time.Time
    SuspendReason *string
    ArchivedAt    *time.Time
    ProvisionedAt *time.Time

    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### Status Machine

```
PENDING ─────────────────────────────► ARCHIVED
   │                                       ▲
   │ (provisioning complete)               │
   ▼                                       │
ACTIVE ──────── (payment failure) ──► SUSPENDED
   ▲                                       │
   └────────── (payment received) ─────────┘
```

Valid transitions:
- `PENDING → ACTIVE`: provisioning workflow completes successfully
- `ACTIVE → SUSPENDED`: payment failed or manual admin action
- `SUSPENDED → ACTIVE`: payment resolved
- `ACTIVE → ARCHIVED`: tenant requests account deletion (after data retention period)
- `SUSPENDED → ARCHIVED`: non-payment beyond grace period
- `PENDING → ARCHIVED`: provisioning abandoned

`ARCHIVED` is terminal. The tenant's rows remain in the database during the data retention period, but the `set_tenant_context()` stored procedure will reject all API access for an archived tenant (it checks `status = 'ACTIVE'`).

#### HTTP Responses by Status

| Status | HTTP | Enforced by |
|--------|------|-------------|
| PENDING | 503 + `Retry-After: 60` | Middleware status check |
| SUSPENDED | 402 Payment Required | Middleware + `set_tenant_context()` |
| ARCHIVED | 410 Gone | Middleware status check |

### 34.2 OrgNode — Company and Division Hierarchy

Within a tenant, organisational structure is represented by the OrgNode entity. This models the business units that users, documents, and cost allocations are linked to.

```go
type OrgNode struct {
    ID        uuid.UUID
    TenantID  uuid.UUID
    ParentID  *uuid.UUID  // nil for the root company node
    Type      string      // "COMPANY" | "DIVISION" | "DEPARTMENT" | "BRANCH" | "COST_CENTRE"
    Code      string      // Short identifier, unique within tenant
    Name      string
    Path      string      // Materialised path: "acme/nairobi/retail" for fast ancestor queries
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

The materialised path enables efficient queries such as "all nodes under the Nairobi division":

```sql
SELECT * FROM org_nodes
WHERE tenant_id = current_setting('app.current_tenant_id')::uuid
  AND path LIKE 'acme/nairobi/%';
```

OrgNode is used for:
- **Cost allocation**: GL entries linked to a cost centre OrgNode
- **Permission scoping**: a branch manager sees only records where `org_node_id` is within their branch subtree
- **HR structure**: employees assigned to department OrgNodes
- **Reporting**: aggregating financial and operational data by division

Every tenant has exactly one root OrgNode of type `COMPANY` created during provisioning.

### 34.3 Tenant Provisioning

Provisioning is entirely data-level. There is no DDL. The provisioning Temporal workflow:

1. Inserts seed rows tagged with `tenant_id`: chart of accounts, leave types, PAYE bands, roles, feature flag overrides, tenant config defaults, root OrgNode.
2. Creates the admin user and assigns the `admin` role.
3. Sets tenant status to `ACTIVE`.
4. Sends the welcome email.

All seed inserts use `INSERT ... ON CONFLICT DO NOTHING` for idempotency. The workflow ID is `{slug}.provision` with `REJECT_DUPLICATE` policy — re-triggering is safe and returns the first run's result.

For full provisioning workflow code, see Chapter 41.

---

## Chapter 35 — IAM: Users, Roles, and Permissions

### 35.1 The User Entity

```go
type User struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    Email        string      // Unique per tenant, not globally
    Name         string
    PasswordHash string      // bcrypt, excluded from API responses (Sensitive field)
    Status       string      // "INVITED" | "ACTIVE" | "SUSPENDED" | "DELETED"
    MFASecret    *string     // TOTP secret, encrypted at rest
    MFAEnabled   bool
    LastLoginAt  *time.Time
    OrgNodeID    *uuid.UUID  // Primary organisational assignment
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

`Email` is unique within a tenant — the same email address can exist as a user in two different tenants. This is by design: an employee at two franchisees of the same chain has separate accounts per tenant with potentially different roles.

System users (service accounts used by background workflows) have `Status = "ACTIVE"` but no password hash and no MFA secret. They authenticate via service tokens rather than session cookies.

### 35.2 The Role Entity

```go
type Role struct {
    ID          uuid.UUID
    TenantID    *uuid.UUID  // nil = system role (applies to all tenants)
    Name        string      // "admin" | "accountant" | "cashier" | ...
    Description string
    IsSystem    bool        // System roles cannot be deleted
    CreatedAt   time.Time
}
```

Built-in system roles seeded at provisioning:

| Role | Primary capabilities |
|------|---------------------|
| `admin` | Full access, role management, tenant configuration |
| `accountant` | GL, invoicing, reports, period close |
| `cashier` | POS, shift open/close, receipts |
| `salesperson` | Sales orders, customers, quotations |
| `hr_manager` | Employee records, payroll, leave approval |
| `store_keeper` | Inventory entries, item master |
| `viewer` | Read-only access to all modules |

Tenant administrators can create additional roles or further restrict system roles by creating tenant-scoped Role rows and assigning narrower permissions.

### 35.3 UserRole — Role Assignment

```go
type UserRole struct {
    UserID    uuid.UUID
    RoleID    uuid.UUID
    TenantID  uuid.UUID
    AssignedBy uuid.UUID    // User ID who made the assignment
    AssignedAt time.Time
}
```

A user can hold multiple roles. Permissions from all roles are merged using union semantics — if any role grants a permission, the user has it.

### 35.4 The Permission Entity and RolePermission

Permissions are defined per entity name and operation:

```go
type Permission struct {
    ID         uuid.UUID
    EntityName string      // e.g. "journal_entry", "customer", "shift"
    Operation  string      // "create" | "read" | "write" | "delete" | "submit" | "cancel" | "amend"
    CreatedAt  time.Time
}

type RolePermission struct {
    RoleID          uuid.UUID
    PermissionID    uuid.UUID
    TenantID        *uuid.UUID   // nil = system-level grant
    // Field-level restrictions stored as JSONB
    FieldRestrictions map[string]FieldPermission `json:"field_restrictions,omitempty"`
}

type FieldPermission struct {
    Readable bool
    Writable bool
}
```

The permission resolution pipeline (§53.3) evaluates the full matrix: load user's roles → load RolePermission rows for each role × entity × operation → merge via union → apply tenant overrides → output a `Permissions` struct.

### 35.5 The Session Entity

Sessions are the primary authentication mechanism for browser clients. They are stored in Redis for fast lookup and in PostgreSQL for audit trail.

```go
type Session struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    UserID      uuid.UUID
    Token       string      // Opaque random token, stored hashed
    IP          string
    UserAgent   string
    CreatedAt   time.Time
    LastSeenAt  time.Time
    ExpiresAt   time.Time
    RevokedAt   *time.Time
    RevokedBy   *uuid.UUID  // User ID or service that revoked
}
```

Redis key: `session:{tenant_id}:{session_id}`. TTL matches `ExpiresAt`. The full Session row in PostgreSQL is kept for audit; the Redis entry is the hot path for every request.

On tenant suspension, `InvalidateAllTenantSessions` deletes all Redis session keys for the tenant's users. Subsequent requests find no session and receive 401, which the middleware converts to 402 after checking tenant status.

---

## Chapter 36 — Audit Log

### 36.1 What the Audit Log Tracks

The AuditLog entity records every significant data change across all auditable entities. By default, all platform entities are auditable. Module entities declare `Auditable: true` on their EntityDefinition to opt in.

```go
type AuditLog struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    EntityName    string      // "journal_entry", "user", "customer", ...
    RecordID      uuid.UUID   // Primary key of the changed record
    Operation     string      // "CREATE" | "UPDATE" | "DELETE"
    UserID        *uuid.UUID  // nil for system/workflow operations
    WorkflowID    *string     // Temporal workflow ID if triggered from a workflow
    RequestID     *string     // HTTP request ID if triggered from an API call
    IP            *string
    Timestamp     time.Time
    PreviousValue map[string]any  // Field values before the change (UPDATE/DELETE)
    NewValue      map[string]any  // Field values after the change (CREATE/UPDATE)
}
```

`Sensitive` fields are excluded from `PreviousValue` and `NewValue` — password hashes, MFA secrets, and any field marked `Sensitive: true` on the EntityDefinition are never written to the audit log.

### 36.2 Write Path

Awo uses a dual write approach:

**DB trigger** (`audit_log_trigger`): A PostgreSQL trigger on every auditable table captures INSERT/UPDATE/DELETE operations at the database level. This fires even for direct SQL access (migrations, emergency fixes) and is the completeness layer.

```sql
CREATE OR REPLACE FUNCTION audit_log_trigger_fn()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (
        id, tenant_id, entity_name, record_id, operation,
        timestamp, previous_value, new_value
    ) VALUES (
        gen_random_uuid(),
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        now(),
        CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) ELSE NULL END,
        CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) ELSE NULL END
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

**Application hook** (`after_save`): The framework's `after_save` hook on auditable entities enriches the audit log entry with application context — `user_id`, `request_id`, `workflow_id`, `ip` — which the DB trigger cannot see. The hook looks up the most recently inserted trigger row for this operation (by record ID and timestamp within a 1-second window) and updates it with the enriched metadata.

This combination means: the trigger catches everything; the hook provides the who-and-why.

### 36.3 Retention and Search

Kenya's Companies Act 2015 mandates financial record retention for 7 years. The AuditLog partition strategy:

```sql
CREATE TABLE audit_log (
    -- columns
) PARTITION BY RANGE (timestamp);

CREATE TABLE audit_log_2024_01 PARTITION OF audit_log
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

Monthly partitions are created automatically by a scheduled job. Partitions older than the retention period are dropped, which is far more efficient than `DELETE WHERE timestamp < cutoff` on a large table.

Search endpoint: `GET /api/v1/audit-log?entity=&record_id=&user_id=&from=&to=&limit=&cursor=`

The endpoint respects the tenant's RLS context — users can only search audit records for their tenant's data.

---

## Chapter 37 — Feature Flags

### 37.1 The FeatureFlag Entity

```go
type FeatureFlag struct {
    Key          string      // Stable identifier: "forecourt.wetstock_alerts"
    Name         string      // Human-readable: "Wetstock Variance Alerts"
    Description  string
    Type         string      // "boolean" | "string" | "percentage"
    DefaultValue string      // JSON-encoded default
    Status       string      // "draft" | "active" | "deprecated" | "removed"
    CreatedAt    time.Time
}

type FeatureFlagOverride struct {
    ID        uuid.UUID
    FlagKey   string
    TenantID  *uuid.UUID   // nil = system-wide override
    UserID    *uuid.UUID   // non-nil = user-level override
    Value     string       // JSON-encoded value
    EnabledAt time.Time
    CreatedBy uuid.UUID
}
```

System-level flags (no `TenantID`) set the global default. Tenant-level overrides (non-nil `TenantID`, nil `UserID`) customise per tenant. User-level overrides (both non-nil) are for beta testing with specific users.

### 37.2 Evaluation Engine

Evaluation is deterministic and cached:

```go
func EvaluateFlag(ctx context.Context, key string) (string, error) {
    tenantID := contextutil.TenantID(ctx)
    userID   := contextutil.UserID(ctx)

    // 1. System default
    flag, err := registry.GetFlag(key)
    value := flag.DefaultValue

    // 2. Tenant override
    if override, ok := tenantOverrides[tenantID][key]; ok {
        value = override.Value
    }

    // 3. User override
    if userID != uuid.Nil {
        if override, ok := userOverrides[tenantID][userID][key]; ok {
            value = override.Value
        }
    }

    // 4. Percentage rollout (only if flag type is "percentage")
    if flag.Type == "percentage" {
        threshold, _ := strconv.Atoi(value)
        hash := fnv32(tenantID.String() + ":" + key) % 100
        if int(hash) >= threshold {
            return "false", nil
        }
        return "true", nil
    }

    return value, nil
}
```

Percentage rollout is stable: the same tenant always gets the same result for a given threshold, because the hash is deterministic. Rolling a flag from 10% to 20% adds new tenants to the enabled set without removing previously enabled ones.

### 37.3 Redis Caching

The full flag set for a tenant is loaded into Redis at tenant boot and on any override change:

```
flags:{tenant_id}   → Redis Hash
  "forecourt.wetstock_alerts"  → "true"
  "reporting.advanced_charts"  → "false"
  "ui.new_dashboard"           → "75"
```

TTL is 60 seconds. Most flag changes propagate within one minute. Critical flag changes (emergency kill switches) can call `redis.Del("flags:{tenant_id}")` directly to force immediate cache invalidation.

### 37.4 Telemetry

Every flag evaluation emits an OpenTelemetry span attribute: `feature_flag.key`, `feature_flag.value`, `feature_flag.tenant_id`. This enables dashboards showing which flags are active, for which tenants, and what percentage of traffic they affect.

Dead flag detection runs weekly: if a flag's evaluation always returns the same value (because the default and all overrides agree), it is reported as a candidate for removal.

---

## Chapter 38 — Notifications

### 38.1 The Notification Entity

```go
type Notification struct {
    ID        uuid.UUID
    TenantID  uuid.UUID
    UserID    uuid.UUID       // Recipient
    Title     string
    Body      string
    Type      string          // "info" | "warning" | "error" | "success" | "action_required"
    Channel   string          // "in_app" | "email" | "sms"
    Status    string          // "PENDING" | "SENT" | "DELIVERED" | "FAILED" | "READ"
    ReadAt    *time.Time
    Metadata  map[string]any  // JSONB — entity name, record ID, action URL
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

`Metadata` typically carries a link back to the originating record so the amis UI can render a "View Invoice" or "Approve Leave" button directly in the notification:

```json
{
  "entity_name": "leave_request",
  "record_id": "018f3a2b-...",
  "action_url": "/leave/requests/018f3a2b-...",
  "action_label": "Review Request"
}
```

### 38.2 In-App Notifications via SSE

The in-app channel uses Server-Sent Events. Each authenticated user maintains a long-lived SSE connection:

```
GET /api/v1/notifications/stream
Accept: text/event-stream
Cookie: session=...

data: {"id":"018f...","title":"Leave request pending approval","type":"action_required",...}

data: {"id":"018g...","title":"Month-end GL posted","type":"success",...}
```

The Fiber handler subscribes to a Redis pub/sub channel `notifications:{tenant_id}:{user_id}`. Notification delivery (from workflows or `after_save` hooks) publishes to this channel. The SSE handler receives the message and writes it to the client's response stream.

This architecture means:
- Horizontal scaling works — any server instance can serve any user's SSE stream.
- The Redis channel is the fan-out; SSE is the final delivery.
- If the SSE connection drops, the client reconnects and requests unread notifications via `GET /api/v1/notifications?status=PENDING&limit=50`.

### 38.3 Email Notifications

Email delivery uses Go `html/template`. One template file per notification type:

```
templates/notifications/
  leave_request_approval.html
  invoice_submitted.html
  shift_close_pending.html
  tenant_suspended.html
  welcome.html
```

Each template receives a `NotificationEmailData` struct:

```go
type NotificationEmailData struct {
    Tenant       *Tenant
    Recipient    *User
    Notification *Notification
    // Type-specific data
    Data any
}
```

SMTP configuration is stored in TenantConfig (`notifications.smtp_host`, `notifications.smtp_port`, etc.). Tenants on the enterprise plan can configure their own SMTP relay; others use the platform default (SendGrid or similar).

### 38.4 SMS Notifications

SMS delivery uses Africa's Talking for the Kenya market. Only `action_required` and `error` notifications are sent via SMS by default (SMS has cost and character constraints). Users can opt in or out per notification type.

### 38.5 Delivery Workflow

All delivery goes through a `SendNotificationWorkflow` Temporal workflow, not synchronous HTTP calls. This ensures retry on transient email/SMS gateway failures:

```go
func SendNotificationWorkflow(ctx workflow.Context, n Notification) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 2 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts:    5,
            InitialInterval:    5 * time.Second,
            BackoffCoefficient: 2.0,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    switch n.Channel {
    case "in_app":
        return workflow.ExecuteActivity(ctx, activities.PublishSSENotification, n).Get(ctx, nil)
    case "email":
        return workflow.ExecuteActivity(ctx, activities.SendEmailNotification, n).Get(ctx, nil)
    case "sms":
        return workflow.ExecuteActivity(ctx, activities.SendSMSNotification, n).Get(ctx, nil)
    }
    return nil
}
```

---

## Chapter 39 — Settings: TenantConfig

### 39.1 The TenantConfig Entity

```go
type TenantConfig struct {
    ID        uuid.UUID
    TenantID  *uuid.UUID  // nil = system default row
    Category  string      // "locale" | "modules" | "integrations" | "limits" | "branding" | "notifications"
    Key       string      // Dotted path: "locale.timezone"
    Value     string      // String-encoded value; type depends on key
    UpdatedAt time.Time
    UpdatedBy *uuid.UUID
}
```

System default rows have `TenantID = NULL`. When a tenant config value is read:

```go
func GetString(ctx context.Context, key string) string {
    tenantID := contextutil.TenantID(ctx)

    // Tenant-specific row takes precedence
    if val, ok := tenantCache[tenantID][key]; ok {
        return val
    }
    // Fall back to system default
    if val, ok := systemDefaults[key]; ok {
        return val
    }
    return ""
}
```

### 39.2 Configuration Keys Reference (partial)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `locale.timezone` | string | `Africa/Nairobi` | IANA timezone |
| `locale.date_format` | string | `DD/MM/YYYY` | Display date format |
| `locale.currency` | string | `KES` | Base currency |
| `locale.language` | string | `en` | `en` or `sw` (Swahili) |
| `modules.forecourt.enabled` | bool | `false` | Enable forecourt module |
| `modules.finance.enabled` | bool | `true` | Enable finance module |
| `integrations.etims.kra_pin` | string | `` | KRA taxpayer PIN |
| `integrations.etims.env` | string | `sandbox` | `sandbox` or `production` |
| `limits.api_rate_per_minute` | int | `600` | Per-tenant API rate limit |
| `branding.primary_colour` | string | `#1a56db` | Primary UI colour |
| `notifications.email_from` | string | `noreply@awo.so` | Sender address |

### 39.3 Module Enablement

Modules are enabled per tenant by setting the corresponding config key to `true`. The `ModuleEnabled(ctx, "forecourt")` helper is called by module middleware to gate route registration and page builder execution.

Module dependencies are enforced in Go: if `modules.finance.enabled = false`, the Inventory GL posting activity returns `NonRetryableError` with code `MODULE_NOT_ENABLED`.

---

## Chapter 40 — Metadata and Reporting

### 40.1 CustomFieldDef

CustomFieldDef allows tenants to extend system entities with additional JSONB fields without a code deployment or migration.

```go
type CustomFieldDef struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    EntityName      string      // "customer", "employee", "journal_entry"
    Key             string      // JSONB path key: "kra_exemption_code"
    Label           string      // "KRA Exemption Code"
    FieldType       string      // Subset of EntityDefinition field types
    Options         []string    // For Select fields
    Required        bool
    MaxLength       *int
    RegexPattern    *string
    UISection       string      // Which form section to inject into
    UIOrder         int         // Position within section
    CreatedAt       time.Time
}
```

Supported field types for custom fields: `Data`, `SmallText`, `LongText`, `Int`, `Float`, `Currency`, `Bool`, `Date`, `DateTime`, `Select`, `MultiSelect`.

At tenant boot, all `CustomFieldDef` rows for the tenant are loaded and merged into the in-memory EntityDefinition for the relevant entity. The page builder then automatically injects the custom field into the form at the declared `UISection` and `UIOrder` position.

The system entity's table has a `custom_fields JSONB` column. Custom field values are stored and retrieved via the JSONB path key:

```sql
-- Find all customers with a specific exemption code
SELECT * FROM customers
WHERE tenant_id = current_setting('app.current_tenant_id')::uuid
  AND custom_fields->>'kra_exemption_code' = 'EXEMPT_2024';
```

For performance, tenants can request a GIN index on specific custom field keys via the admin UI. The index is created via a migration:

```sql
CREATE INDEX CONCURRENTLY idx_customers_cf_exemption
ON customers USING GIN ((custom_fields->'kra_exemption_code'));
```

### 40.2 ReportDefinition

ReportDefinition allows tenants to define named reports that can be executed via the API and scheduled for delivery.

```go
type ReportDefinition struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    Name         string      // URL slug: "monthly-sales-by-branch"
    Label        string      // "Monthly Sales by Branch"
    QueryType    string      // "sql" | "filter_dsl"
    Query        string      // SQL template or filter DSL JSON
    Parameters   []ReportParam
    Columns      []ReportColumn
    ChartConfig  *ChartConfig  // Optional chart over the report data
    IsPublic     bool          // Visible to all users (vs only creator)
    CreatedBy    uuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type ReportParam struct {
    Name     string   // "from_date"
    Label    string   // "From Date"
    Type     string   // "date" | "entity" | "select" | "string"
    Required bool
}
```

SQL reports use Go template syntax for parameter substitution. RLS is automatic — the database connection has `set_tenant_context()` called before any query, so even raw SQL reports can only see the current tenant's data.

```sql
-- Example SQL report template
SELECT
    o.name AS branch,
    SUM(si.grand_total) AS total_sales,
    COUNT(*) AS invoice_count
FROM sales_invoices si
JOIN org_nodes o ON si.org_node_id = o.id
WHERE si.posting_date BETWEEN '{{.from_date}}' AND '{{.to_date}}'
  AND si.status = 'SUBMITTED'
GROUP BY o.name
ORDER BY total_sales DESC;
```

The `{{.param_name}}` values are sanitised against the declared parameter types before substitution — SQL injection via report parameters is not possible.

Reports are accessible at:
```
GET /api/v1/reports/{name}?from_date=2024-01-01&to_date=2024-01-31
```

Response follows the standard envelope with the report results as `data` and column metadata in `meta`.

Export endpoints:
```
GET /api/v1/reports/{name}/export/csv?{params}
GET /api/v1/reports/{name}/export/pdf?{params}
```

PDF export uses the print template system (see Chapter 25). The report definition may specify a custom `PrintTemplate` for its PDF layout; otherwise the framework uses a generic table-format template.

Scheduled reports are configured via TenantConfig or the admin UI: pick a report, set parameters and a delivery schedule, and the framework creates a Temporal schedule that runs the report and emails the results.
