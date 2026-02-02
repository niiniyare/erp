[<-- Back to Index](README.md)

## Business Rules & Validation

### Status Transition Rules

```markdown
VALID STATUS TRANSITIONS:

From          To              Condition
─────────────────────────────────────────────────
PENDING    → ACTIVE          Admin activation or auto-activate
ACTIVE     → SUSPENDED       Payment/policy/security reason required
ACTIVE     → ARCHIVED        Archive reason + retention required
SUSPENDED  → ACTIVE          Reactivation reason required
SUSPENDED  → ARCHIVED        Archive reason + retention required
ARCHIVED   → (none)          Terminal state - no transitions allowed

INVALID TRANSITIONS (blocked):
PENDING    → SUSPENDED       Must activate first
PENDING    → ARCHIVED        Must activate first
ARCHIVED   → ACTIVE          Cannot reactivate archived tenant
ARCHIVED   → SUSPENDED       Already permanently deactivated
SUSPENDED  → PENDING         Cannot revert to pending
```

### Data Validation Rules

```markdown
TENANT CREATION RULES:

name:
├── Required
├── Max 255 characters
├── Must not be empty string
└── Trimmed of leading/trailing whitespace

email:
├── Required
├── Must be valid email format
├── Max 255 characters
└── Unique across non-deleted tenants

subdomain:
├── Optional
├── Max 63 characters (DNS limit)
├── Lowercase alphanumeric + hyphens only
├── Cannot start or end with hyphen
├── Unique across non-deleted tenants
└── Cannot be reserved word (admin, api, www, app, etc.)

slug:
├── Auto-generated from name
├── Max 50 characters
├── Lowercase alphanumeric + hyphens
├── Unique across non-deleted tenants
└── UUID suffix appended if collision

currency_code:
├── 3 characters exactly
├── ISO 4217 format
└── Default: USD

timezone:
├── Valid IANA timezone string
├── e.g., Africa/Nairobi, UTC, America/New_York
└── Default: UTC

company_size:
├── Must be: Small, Medium, Large, Enterprise
└── Default: Small

status:
├── Must be: PENDING, ACTIVE, SUSPENDED, ARCHIVED
└── Default: PENDING (on creation)
```

### Resource Limit Rules

```markdown
LIMIT ENFORCEMENT RULES:

1. User Limit:
   Rule: active_users < max_users
   Checked: Before creating new user
   Error: "Tenant has reached maximum user limit"
   Override: Platform admin can temporarily increase

2. Entity Limit:
   Rule: total_entities < max_entities
   Checked: Before creating new company/branch
   Error: "Tenant has reached maximum entity limit"

3. Transaction Limit:
   Rule: total_transactions < max_transactions_per_month
   Checked: Before posting any transaction
   Error: "Monthly transaction limit reached"
   Reset: First day of each month

4. Storage Limit:
   Rule: storage_used_mb < storage_quota_mb
   Checked: Before file upload
   Error: "Storage quota exceeded"
   Threshold: Warning at 80% utilization

5. API Rate Limit:
   Rule: requests_in_window < rate_limit
   Checked: Per request via middleware
   Error: 429 Too Many Requests
   Window: Per minute (configurable)
```

### Soft Delete Rules

```markdown
SOFT DELETE BEHAVIOR:

On Delete:
├── deleted_at = NOW() (not physical delete)
├── All queries filter: WHERE deleted_at IS NULL
├── Slug and subdomain become reusable (unique constraint scoped)
└── Audit log entry created

Constraints:
├── Unique slug:      WHERE deleted_at IS NULL
├── Unique subdomain: WHERE deleted_at IS NULL
├── Unique email:     WHERE deleted_at IS NULL
└── Cascade: Soft delete does NOT cascade to child tables

Restoration:
├── Set deleted_at = NULL
├── Verify slug/subdomain still available
└── Requires platform admin action
```

### Configuration Rules

```markdown
CONFIGURATION VALIDATION:

fiscal_year_start_month:
├── Must be 1-12
└── Cannot be changed mid-fiscal-year (warning)

accounting_method:
├── Must be: FIFO, LIFO, or WEIGHTED_AVERAGE
└── Changing method requires audit trail entry

password_policy:
├── min_length: Must be >= 8
├── max_age_days: Must be > 0 or null (no expiry)
└── At least one complexity rule recommended

api_rate_limits:
├── requests_per_minute: Must be > 0
├── Must not exceed plan maximum
└── Burst limit <= 2x per-minute rate

allowed_modules:
├── Must be valid module names
├── "financial" and "selling" always included
└── Cannot enable modules above plan tier
```

### Isolation Rules

```markdown
TENANT ISOLATION INVARIANTS:

1. Every tenant-scoped table has RLS enabled
2. Every query through application_role is filtered
3. Foreign keys cannot reference cross-tenant records
4. Tenant context must be set before any data access
5. Tenant context is transaction-scoped (auto-cleared)
6. Admin role bypasses RLS (platform admin only)
7. Readonly role can see all tenants (monitoring only)
8. Superuser access is prohibited in production
```

---

Next: [Summary](./22-summary.md)
