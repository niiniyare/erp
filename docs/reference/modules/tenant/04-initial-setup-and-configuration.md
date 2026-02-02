[<-- Back to Index](README.md)

## Initial Setup & Configuration

### Prerequisites

Before tenants can be provisioned, the platform requires these database objects:

```markdown
REQUIRED DATABASE SETUP:

1. Core Functions:
   ├── set_tenant_context(UUID)      - Sets session tenant ID
   ├── current_tenant_id()           - Reads current tenant from session
   ├── clear_tenant_context()        - Clears tenant session variable
   └── update_updated_at_column()    - Timestamp trigger function

2. Database Roles:
   ├── admin_role       - Full access to all tenant data
   ├── application_role - Standard API access with RLS
   └── readonly_role    - Read-only monitoring access

3. Migrations Applied (in order):
   ├── 000101 - Core tenant tables and functions
   ├── 000102 - Configuration tables
   ├── 000103 - Usage tracking tables
   ├── 000104 - Provisioning function
   ├── 000105 - Bulk operations tables
   ├── 000106 - Tenant context validation
   └── 000107 - Tenant isolation enforcement
```

### Platform Configuration

The platform administrator configures global defaults before onboarding tenants:

```markdown
DEFAULT TENANT SETTINGS:

Timezone:       UTC (overridden per tenant)
Currency:       USD (overridden per tenant)
Company Size:   Small (default for new signups)
Status:         PENDING (always starts here)

DEFAULT RESOURCE LIMITS (tenant_configurations):

Plan: Basic
├── max_users:                  10
├── max_entities:               5
├── max_transactions_per_month: 1,000
├── storage_quota_mb:           1,024 (1 GB)
└── api_rate_limit:             100 req/min

Plan: Professional
├── max_users:                  50
├── max_entities:               20
├── max_transactions_per_month: 10,000
├── storage_quota_mb:           10,240 (10 GB)
└── api_rate_limit:             500 req/min

Plan: Enterprise
├── max_users:                  Unlimited
├── max_entities:               Unlimited
├── max_transactions_per_month: Unlimited
├── storage_quota_mb:           102,400 (100 GB)
└── api_rate_limit:             2,000 req/min
```

### Database Role Setup

```markdown
ROLE PERMISSIONS MATRIX:

                        admin_role    application_role    readonly_role
tenants                 CRUD          CRUD (via RLS)      SELECT
tenant_configurations   CRUD          CRUD (via RLS)      SELECT
tenant_usage_stats      CRUD          CRUD (via RLS)      SELECT
tenant_bulk_operations  CRUD          CRUD (via RLS)      SELECT
tenant_bulk_op_results  CRUD          CRUD (via RLS)      SELECT

Functions:
set_tenant_context()         ✓              ✓                ✓
current_tenant_id()          ✓              ✓                ✓
clear_tenant_context()       ✓              ✓                ✓
validate_and_set_context()   ✓              ✓                ✓
provision_tenant_complete()  ✓              ✓                ✗
check_tenant_limits()        ✓              ✓                ✓
get_bulk_operation_summary() ✓              ✓                ✓
update_bulk_operation_counts()✓             ✓                ✗
```

### Slug Generation

Tenant slugs are auto-generated from the company name:

```markdown
SLUG GENERATION RULES:

Input:  "Savannah Electronics Ltd."
Step 1: Lowercase          → "savannah electronics ltd."
Step 2: Replace non-alnum  → "savannah-electronics-ltd-"
Step 3: Trim trailing dash → "savannah-electronics-ltd"

UNIQUENESS HANDLING:
If "savannah-electronics-ltd" already exists:
  → Append UUID prefix: "savannah-electronics-ltd-a1b2c3d4"

TRIGGER: generate_tenant_slug
  Fires: BEFORE INSERT on tenants
  Action: Auto-generates slug if not provided
```

### Validation Constraints

```markdown
TENANT TABLE CONSTRAINTS:

name:          NOT NULL, max 255 chars
email:         NOT NULL, max 255 chars, valid format
slug:          UNIQUE (where deleted_at IS NULL), max 50 chars
subdomain:     UNIQUE (where deleted_at IS NULL), max 63 chars
status:        Must be: PENDING, ACTIVE, SUSPENDED, ARCHIVED
currency_code: 3-character ISO code (e.g., KES, USD)
timezone:      Valid timezone string (e.g., Africa/Nairobi)
company_size:  Must be: Small, Medium, Large, Enterprise

CONFIGURATION TABLE CONSTRAINTS:

max_users:                  > 0
max_entities:               > 0
max_transactions_per_month: > 0
storage_quota_mb:           > 0
fiscal_year_start_month:    1-12
accounting_method:          FIFO, LIFO, or WEIGHTED_AVERAGE
```

---

Next: [Core Database Architecture](./05-core-database-architecture.md)
