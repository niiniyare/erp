[<-- Back to Index](README.md)

## Subscription Plans & Limits

### Plan Types

AWO ERP offers three subscription tiers defined in the Go domain model:

```markdown
PLAN DEFINITIONS:

1. BASIC
   Target: Small businesses, freelancers
   Users:  Up to 10
   Price:  Entry-level

2. PROFESSIONAL
   Target: Growing SMEs
   Users:  Up to 50
   Price:  Mid-tier

3. ENTERPRISE
   Target: Large organizations
   Users:  Unlimited
   Price:  Custom pricing
```

### Detailed Plan Comparison

```markdown
FEATURE MATRIX:

                            Basic       Professional    Enterprise
────────────────────────────────────────────────────────────────────
RESOURCE LIMITS:
  Max Users                 10          50              Unlimited
  Max Entities              5           20              Unlimited
  Transactions/Month        1,000       10,000          Unlimited
  Storage                   1 GB        10 GB           100 GB
  API Rate (req/min)        100         500             2,000

MODULES:
  Financial                 ✓           ✓               ✓
  Selling                   ✓           ✓               ✓
  Buying                    ✗           ✓               ✓
  Inventory                 ✗           ✓               ✓
  HR                        ✗           ✗               ✓
  Manufacturing             ✗           ✗               ✓
  Projects                  ✗           ✗               ✓

SECURITY:
  Password Min Length       8           10              12
  MFA                       Optional    Optional        Required
  Password Expiry           None        90 days         60 days
  Audit Log Retention       30 days     90 days         365 days

SUPPORT:
  Email Support             ✓           ✓               ✓
  Phone Support             ✗           ✓               ✓
  Dedicated Account Mgr     ✗           ✗               ✓
  SLA                       99%         99.5%           99.9%
```

### Plan Upgrade Flow

```markdown
EXAMPLE: Highland Textiles Upgrading Basic → Professional

CURRENT STATE:
  Tenant:     Highland Textiles
  Plan:       Basic
  Users:      8 / 10 (80% utilized)
  Entities:   4 / 5 (80% utilized)
  Problem:    Need to add Buying and Inventory modules

UPGRADE REQUEST:
  New Plan:   Professional
  Requested:  2024-06-01 by admin@highlandtextiles.co.ke

CHANGES APPLIED:
  max_users:            10 → 50
  max_entities:         5 → 20
  max_transactions:     1,000 → 10,000
  storage_quota_mb:     1,024 → 10,240
  api_rate_limit:       100 → 500 req/min
  allowed_modules:      ["financial", "selling"]
                        → ["financial", "selling", "buying", "inventory"]

RESULT:
  Status: Upgrade completed
  Billing: Pro-rated from June 1
  Users:  8 / 50 (16% utilized - room to grow)
  Modules: Buying and Inventory now accessible
```

### Plan Downgrade Considerations

```markdown
DOWNGRADE VALIDATION:

SCENARIO: Coastal Coffee wants to downgrade Enterprise → Professional

PRE-CHECK:
  Current active_users:     82
  Professional max_users:   50
  → BLOCKED: Must reduce users to 50 or fewer first

  Current entities:         15
  Professional max_entities: 20
  → OK: Within limits

  Current storage:          8,500 MB
  Professional storage:     10,240 MB
  → OK: Within limits

  Currently using HR module: Yes
  Professional includes HR:  No
  → BLOCKED: Must migrate off HR module first

DOWNGRADE BLOCKERS:
  1. Reduce active users from 82 to ≤ 50
  2. Disable and migrate data from HR module
  Only after resolving blockers can downgrade proceed.
```

### Tenant Limits in Go Domain Model

```markdown
GO STRUCT: TenantLimits

type TenantLimits struct {
    MaxUsers         int
    MaxEntities      int
    MaxTransactions  int
    StorageQuotaMB   int
}

Used by:
├── Service.ValidateTenantAccess()  - Check tenant is within limits
├── Service.UpdateTenant()          - Enforce limits on changes
└── Provisioning.ProvisionTenant()  - Set initial limits from plan
```

---

Next: [Bulk Operations](./12-bulk-operations.md)
