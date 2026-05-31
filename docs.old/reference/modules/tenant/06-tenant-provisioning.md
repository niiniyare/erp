[<-- Back to Index](README.md)

## Tenant Provisioning

### Overview

Provisioning is the process of creating a fully operational tenant from a signup request. The AWO ERP uses a multi-step provisioning workflow handled by the `ProvisioningService`.

### Provisioning Flow

```markdown
PROVISIONING WORKFLOW:

Step 1: Create Tenant Record
│  Input: name, email, subdomain, industry, company_size,
│         currency_code, timezone, settings
│  Action: INSERT into tenants table
│  Result: UUID generated, slug auto-created
│  Status: PENDING
│
Step 2: Create Configuration
│  Action: Triggered automatically (create_default_tenant_config)
│  Result: tenant_configurations row with plan defaults
│  Sets: max_users, max_entities, storage quota, etc.
│
Step 3: Create Admin User
│  Input: admin name, email, password
│  Action: Create user record linked to tenant
│  Result: First user with admin role
│
Step 4: Enable Modules
│  Input: List of module names (from plan)
│  Action: Update allowed_modules in configuration
│  Result: Tenant can access enabled modules
│
Step 5: Send Welcome Notification
│  Action: Dispatch welcome email to admin
│  Result: Admin receives onboarding instructions
│
Step 6: Activate Tenant
│  Action: Update status PENDING → ACTIVE
│  Result: Tenant fully operational
│  Log: Audit entry for provisioning completion
```

### Real-World Example

```markdown
EXAMPLE: Provisioning Coastal Coffee Co.

REQUEST:
  Name:         Coastal Coffee Co.
  Email:        admin@coastalcoffee.co.ke
  Subdomain:    coastalcoffee
  Industry:     Agriculture
  Company Size: Medium
  Currency:     KES
  Timezone:     Africa/Nairobi
  Plan:         Professional
  Admin User:   James Mwangi (james@coastalcoffee.co.ke)

STEP 1 - Tenant Created:
  ID:        f7a8b9c0-1234-5678-9abc-def012345678
  Slug:      coastal-coffee-co
  Subdomain: coastalcoffee.awo-erp.com
  Status:    PENDING

STEP 2 - Configuration Applied:
  max_users:            50
  max_entities:         20
  max_transactions:     10,000/month
  storage_quota:        10,240 MB
  api_rate_limit:       500 req/min
  fiscal_year_start:    January
  accounting_method:    FIFO
  allowed_modules:      ["financial", "selling", "buying", "inventory", "hr"]

STEP 3 - Admin User Created:
  User:     James Mwangi
  Email:    james@coastalcoffee.co.ke
  Role:     Tenant Administrator
  Tenant:   f7a8b9c0-...

STEP 4 - Modules Enabled:
  ✓ Financial Module
  ✓ Selling Module
  ✓ Buying Module
  ✓ Inventory Module
  ✓ HR Module

STEP 5 - Welcome Email Sent:
  To: james@coastalcoffee.co.ke
  Subject: Welcome to AWO ERP - Coastal Coffee Co.

STEP 6 - Tenant Activated:
  Status: PENDING → ACTIVE
  Audit: "Tenant provisioned and activated successfully"
```

### Provisioning Request Structure

```markdown
COMPREHENSIVE PROVISION REQUEST:

Tenant Info:
├── name           : Company name (required)
├── email          : Contact email (required)
├── subdomain      : Custom subdomain (optional)
├── industry       : Business sector (optional)
├── company_size   : Small/Medium/Large/Enterprise
├── plan_type      : basic/professional/enterprise
├── currency_code  : ISO 4217 code (default: USD)
├── timezone       : IANA timezone (default: UTC)
└── settings       : Custom JSON settings

Admin User:
├── name           : Admin full name (required)
├── email          : Admin email (required)
└── password       : Initial password (required)

Initial Settings:
├── fiscal_year_start_month : 1-12
├── accounting_method       : FIFO/LIFO/WEIGHTED_AVERAGE
├── enabled_modules         : List of module names
└── password_policy         : Complexity requirements
```

### Database-Level Provisioning

The `provision_tenant_complete()` SQL function handles the core tenant creation:

```markdown
FUNCTION: provision_tenant_complete()

Parameters:
├── p_name          VARCHAR(255)  - Tenant name
├── p_email         VARCHAR(255)  - Contact email
├── p_subdomain     VARCHAR(63)   - Subdomain (optional)
├── p_industry      VARCHAR(50)   - Industry (optional)
├── p_company_size  VARCHAR(20)   - Default: 'Small'
├── p_currency_code CHAR(3)       - Default: 'USD'
├── p_timezone      VARCHAR(50)   - Default: 'UTC'
└── p_settings      JSONB         - Default: '{}'

Process:
1. Generate UUID: gen_random_uuid()
2. Generate slug: lowercase + replace non-alnum with hyphens
3. Ensure uniqueness: Append UUID prefix if slug exists
4. INSERT tenant with status = 'PENDING'
5. RETURN tenant UUID

Trigger Side Effects:
- Slug auto-generated if NULL (generate_tenant_slug trigger)
- Default configuration created (create_default_tenant_config trigger)
```

### Error Handling

```markdown
PROVISIONING FAILURE SCENARIOS:

1. Duplicate Subdomain:
   Error: UNIQUE constraint violation on subdomain
   Action: Return error, suggest alternative subdomain

2. Invalid Email Format:
   Error: CHECK constraint on email field
   Action: Return validation error

3. Database Connection Failure:
   Error: Connection timeout
   Action: Retry with exponential backoff

4. Admin User Creation Failure:
   Error: User service unavailable
   Action: Rollback tenant creation, return error
   Note: Entire provisioning is transactional

5. Slug Collision (rare):
   Error: None - handled automatically
   Action: UUID prefix appended to slug
```

---

Next: [Tenant Configuration Management](./07-tenant-configuration-management.md)
