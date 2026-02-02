[<-- Back to Index](README.md)

## Row-Level Security & Isolation

### Overview

Row-Level Security (RLS) is the cornerstone of AWO ERP's multi-tenant isolation. Every table with tenant data has RLS policies that ensure queries only return data belonging to the current tenant.

### How RLS Works

```markdown
RLS MECHANISM:

1. API request arrives for Tenant A
2. Middleware calls: set_tenant_context('tenant-a-uuid')
3. PostgreSQL stores tenant ID in session variable
4. Every query automatically filtered:

   SELECT * FROM invoices;

   PostgreSQL internally rewrites to:
   SELECT * FROM invoices
   WHERE tenant_id = current_tenant_id();

5. Tenant A CANNOT see Tenant B's invoices
   Even if they craft a manual WHERE clause - RLS overrides it
```

### RLS Policy Pattern

Every tenant-scoped table follows the same RLS pattern:

```markdown
STANDARD RLS SETUP (applied to every table):

-- Enable RLS
ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;

-- Application role: Isolated by tenant context
CREATE POLICY tenant_isolation_policy ON <table_name>
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());

-- Admin role: Full access (platform administration)
CREATE POLICY admin_full_access_policy ON <table_name>
  FOR ALL TO admin_role
  USING (true);

-- Readonly role: Read access to all (monitoring)
CREATE POLICY readonly_access_policy ON <table_name>
  FOR SELECT TO readonly_role
  USING (true);
```

### Context Functions

```markdown
TENANT CONTEXT FUNCTIONS:

1. set_tenant_context(p_tenant_id UUID)
   ├── Sets: app.current_tenant_id = p_tenant_id
   ├── Scope: Current transaction (local = true)
   └── Used by: API middleware on every request

2. current_tenant_id() → UUID
   ├── Reads: app.current_tenant_id from session
   ├── Returns: UUID of current tenant
   └── Used by: All RLS USING clauses

3. clear_tenant_context()
   ├── Resets: app.current_tenant_id to empty
   └── Used by: Request cleanup / connection pool return

4. validate_and_set_tenant_context(p_tenant_id UUID)
   ├── Validates: Tenant exists
   ├── Validates: Tenant not soft-deleted
   ├── Validates: Status is 'active' or 'pending'
   ├── Updates: last_activity_at = NOW()
   ├── Sets: Tenant context if all checks pass
   ├── Returns: (tenant_name, tenant_status)
   └── Raises: Exception if validation fails
```

### Cross-Tenant Isolation Enforcement

Beyond RLS, the `enforce_tenant_isolation()` trigger prevents cross-tenant foreign key references:

```markdown
CROSS-TENANT FK VALIDATION:

Scenario: Person record references an Entity

  INSERT INTO persons (entity_id, tenant_id, name, ...)
  VALUES ('entity-uuid', 'tenant-a-uuid', 'John', ...);

  Trigger fires: enforce_tenant_isolation()

  Check: Does 'entity-uuid' belong to 'tenant-a-uuid'?

  ├── YES → INSERT proceeds normally
  │
  └── NO  → RAISE EXCEPTION
            'Entity entity-uuid does not belong to tenant tenant-a-uuid'

This prevents data corruption even if application code has bugs.
```

### Security Example

```markdown
EXAMPLE: Two Tenants on Same Database

Tenant: Savannah Electronics (ID: aaa-111)
  Invoices: INV-001, INV-002, INV-003
  Users: Alice, Bob

Tenant: Coastal Coffee (ID: bbb-222)
  Invoices: INV-001, INV-002
  Users: James, Mary

QUERY BY ALICE (Savannah Electronics):
  Session: app.current_tenant_id = 'aaa-111'

  SELECT * FROM invoices;

  Result:
  ┌──────────┬────────────┬──────────┐
  │ Invoice  │ Amount     │ Tenant   │
  ├──────────┼────────────┼──────────┤
  │ INV-001  │ KES 50,000 │ aaa-111  │
  │ INV-002  │ KES 30,000 │ aaa-111  │
  │ INV-003  │ KES 75,000 │ aaa-111  │
  └──────────┴────────────┴──────────┘

  Coastal Coffee's invoices: INVISIBLE

QUERY BY JAMES (Coastal Coffee):
  Session: app.current_tenant_id = 'bbb-222'

  SELECT * FROM invoices;

  Result:
  ┌──────────┬────────────┬──────────┐
  │ Invoice  │ Amount     │ Tenant   │
  ├──────────┼────────────┼──────────┤
  │ INV-001  │ KES 12,000 │ bbb-222  │
  │ INV-002  │ KES 8,500  │ bbb-222  │
  └──────────┴────────────┴──────────┘

  Savannah's invoices: INVISIBLE
```

### Tables with RLS Enabled

```markdown
ALL TENANT-SCOPED TABLES HAVE RLS:

Tenant Module:
├── tenants (admin/app policies - no tenant_id filter)
├── tenant_configurations
├── tenant_usage_stats
├── tenant_bulk_operations
└── tenant_bulk_operation_results

Other Modules (all use tenant_id = current_tenant_id()):
├── entities, persons, employees
├── accounts, journals, journal_entries
├── invoices, payments, credit_notes
├── customers, vendors
├── products, inventory_items
└── ... (every tenant-scoped table)
```

---

Next: [Tenant Context Management](./09-tenant-context-management.md)
