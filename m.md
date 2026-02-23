please fix followin 

— what's broken, what's architecturally weak, and concretely how each fix translates to a tenant benefit.

---

## Settings / Configuration Schema

### Critical: `config_definitions` has `tenant_id` and `entity_id` columns

This is the most structurally wrong thing in the file. `config_definitions` is a **system-wide schema registry** — it answers "what configuration keys are valid across the platform?" Putting `tenant_id` on it means you'd need to duplicate every definition once per tenant. That turns a 15-row table into a 15 × N-tenants row table, and when you add a new config key you have to insert it for every existing tenant rather than once.

The UNIQUE constraint `(module_name, config_key)` already conflicts with this — if `tenant_id` were meaningful, the constraint would need to be `(tenant_id, module_name, config_key)`. The RLS policy then undermines it further by using `USING (true)`, meaning all tenants can see all definitions regardless of tenant_id. The column is functionally dead but architecturally misleading.

**Fix:** Drop both columns from `config_definitions`. It is a global platform table. Tenant-specific overrides live in `tenant_configurations.settings` JSONB and `entities.settings` JSONB — that IS the 3-level inheritance model. The config_definitions table is just the schema contract, not the values.

**Tenant benefit:** Adding a new configuration option (e.g. a new HR payroll setting) is a single INSERT to config_definitions. It becomes available to all tenants instantly. Without this fix, a new config requires a bulk INSERT for every existing tenant — an operational risk and a migration bottleneck.

### `is_overridable` is a single boolean — too coarse

The current schema has one flag. But from the PRD, `inventory.default_valuation_method` should be overridable at tenant level but NOT at entity level (accounting consistency requires it be consistent within a tenant). A single `is_overridable = true/false` can't express this distinction.

**Fix:** Split into `is_tenant_overridable BOOLEAN` and `is_entity_overridable BOOLEAN`. This is what migration 008 implemented. The combination matrix then becomes:

- `true / true` → full inheritance cascade (most settings)
- `true / false` → tenant sets policy, entities must follow (valuation method, accounting basis)
- `false / false` → system default is locked for everyone (compliance-critical settings)
- `false / true` → system locks tenant but allows entity variance (unusual but valid for some regional configs)

**Tenant benefit:** A tenant administrator can configure valuation method once and know that no branch manager can override it and create an inventory valuation inconsistency across the organisation. Without this, enforcing accounting consistency requires application-layer guards that are easy to miss.

### `configuration_templates.tenant_id` — ambiguous ownership

The templates table has `tenant_id` which suggests templates are per-tenant. But industry templates (Manufacturing, Services, Retail) are system-wide and should be available to all tenants. Custom templates a tenant creates should be private to that tenant.

Right now you can't distinguish between these two cases in a query. There's also no `is_system BOOLEAN` or `scope VARCHAR` to mark the distinction.

**Fix:** Add a `scope VARCHAR(10) CHECK (scope IN ('SYSTEM', 'TENANT'))` column. System-scoped templates have `tenant_id = NULL` and are readable by all tenants but writable only by admin_role. Tenant-scoped templates are readable and writable only by their own tenant (enforced by RLS). The RLS policy becomes:

```sql
CREATE POLICY configuration_templates_read ON configuration_templates
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = current_tenant_id())
  );
```

**Tenant benefit:** Tenants can create and share their own internal templates (e.g. "Regional Branch Standard" or "Acquisition Onboarding") without those templates being visible to other tenants. System-provided industry templates remain available without duplicating them per tenant.

### `template_applications.template_id ON DELETE CASCADE` — destroys audit history

If an admin deletes a template, all records of where it was applied disappear. This is a compliance problem. You can no longer answer "which tenants were affected by the Manufacturing template v1.2?" after deletion.

**Fix:** Change to `ON DELETE RESTRICT` (prevents deletion of templates that have been applied) or `ON DELETE SET NULL` with a `template_name_snapshot VARCHAR` column that captures the name at application time for audit purposes.

### `configuration_audit.config_key` stores a combined string

The column is `VARCHAR(150)` storing `"finance.invoice_prefix"` as a combined string. This means filtering by module requires `WHERE config_key LIKE 'finance.%'` — a leading-wildcard LIKE that can't use a B-tree index. Split into `module_name VARCHAR(50)` and `config_key VARCHAR(100)` to match the structure in `config_definitions`, and index them separately.

### Missing: `display_name` on `config_definitions`

The description column is there but no `display_name`. The Settings UI needs a short label for input fields ("Invoice Prefix") separate from the full `description` paragraph. Without it, the UI either shows the raw `invoice_prefix` key (bad UX) or the entire description paragraph as the label (worse UX).

---

## Entities / Organization Hierarchy Schema

### `hierarchy_paths.entity_id` is always equal to `descendant_id`

The `maintain_entity_id` trigger exists solely to enforce `entity_id = descendant_id`. This is an invariant enforced by application logic on a column that shouldn't exist. In a closure table, `(ancestor_id, descendant_id, depth)` is the complete information — `entity_id` is pure redundancy. It adds 16 bytes per row to a table that can have O(depth²) rows per entity tree, and the trigger adds write overhead.

**Fix:** Drop `entity_id` from `hierarchy_paths`. Every query that filters by `entity_id` should use `descendant_id` directly. The `maintain_entity_id` function and trigger are eliminated entirely.

**Tenant benefit:** For a tenant with 200 entities in a 5-level hierarchy, hierarchy_paths holds ~20,000 rows. Eliminating the redundant column and its trigger reduces write latency on every entity insert (which must insert O(depth) hierarchy_paths rows) and reduces table size.

### `hierarchy_paths` has validation columns that don't belong there

`version`, `last_validation_run`, `validation_status`, `validation_errors` on a **relationship table** make no sense. A closure table row is a mathematical fact (A is an ancestor of B at depth N). It isn't validated; it's either correct or it isn't. Validation state belongs on `entities` rows, not on the edges connecting them.

**Fix:** Drop these columns from `hierarchy_paths`. Keep them on `entities` where they're meaningful.

### `v_active_entities` — the CASE WHEN EXISTS join pattern will be extremely slow

This view uses:
```sql
LEFT JOIN entities c ON (
  CASE
    WHEN e.type = 'COMPANY' THEN e.uuid = c.uuid
    ELSE EXISTS (SELECT 1 FROM hierarchy_paths hp WHERE ...)
  END
)
```

A `CASE WHEN EXISTS (subquery)` inside a JOIN condition forces a correlated subquery evaluation per row. The query planner cannot flatten this into an efficient hash or merge join. For a tenant with 500 entities this view will be unusably slow.

**Fix:** Use a straightforward `hierarchy_paths` join to find the nearest ancestor of each type:

```sql
CREATE VIEW v_active_entities AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  company.name AS company_name,
  region.name AS regional_name,
  dept.name AS department_name
FROM entities e
JOIN tenants t ON e.tenant_id = t.id
-- Find company ancestor via closure table
LEFT JOIN hierarchy_paths hp_c ON hp_c.descendant_id = e.uuid
LEFT JOIN entities company ON company.uuid = hp_c.ancestor_id AND company.type = 'COMPANY' AND company.deleted_at IS NULL
-- Find region ancestor
LEFT JOIN hierarchy_paths hp_r ON hp_r.descendant_id = e.uuid
LEFT JOIN entities region ON region.uuid = hp_r.ancestor_id AND region.type IN ('REGION','REGIONAL') AND region.deleted_at IS NULL
-- Find department ancestor
LEFT JOIN hierarchy_paths hp_d ON hp_d.descendant_id = e.uuid
LEFT JOIN entities dept ON dept.uuid = hp_d.ancestor_id AND dept.type = 'DEPARTMENT' AND dept.deleted_at IS NULL
WHERE e.deleted_at IS NULL AND e.is_active = TRUE;
```

**Tenant benefit:** Reports on org structure and active entities are a daily workflow for HR, Finance, and management dashboards. A slow view blocks those workflows. With proper hierarchy_paths-based joins the query becomes a set operation the planner can optimise with indexes.

### Views hardcode a 4-level hierarchy (COST_CENTER → DEPARTMENT → REGION → COMPANY)

`v_entity_structure`, `v_cost_center_info`, `v_cost_center_summary` all assume this exact structure with these exact type names. A tenant that uses BRANCH → DIVISION → COMPANY (3 levels) or TEAM → COST_CENTER → DEPARTMENT → REGION → SUBSIDIARY → COMPANY (6 levels) will see empty results or wrong results from these views — silently, with no error.

**Fix:** Either make these views use the `hierarchy_paths` depth column to find the nearest typed ancestor (flexible) or document them clearly as "example views for the canonical 4-level structure" and build a configurable reporting layer in the application. Don't ship rigid structural assumptions in production views.

### `v_entity_changes` has ORDER BY without LIMIT

`ORDER BY e.updated_at DESC` in a view definition is meaningless (ORDER BY in a view is not guaranteed to be preserved when queried) and dangerous when the view is used in a subquery or CTE where the planner might materialise the full sorted result. Remove it and let callers add ORDER BY.

### `entitystate` is missing the `config` JSONB column from the PRD

The PRD explicitly calls for adding `config JSONB` to `entitystate` for document sequence formatting (prefix, suffix, pad_length, reset_frequency, format_template). This is what gives entities their "BRANCH-INV-000001" vs "INV-2025-000001" format control. The column is absent.

```sql
ALTER TABLE entitystate ADD COLUMN config JSONB NOT NULL DEFAULT '{}';
COMMENT ON COLUMN entitystate.config IS
  'Document sequence formatting config. '
  'Schema: { prefix, suffix, pad_length, reset_frequency, format_template }. '
  'Values here override tenant_configurations.settings for this entity+doctype.';
```

**Tenant benefit:** Each branch gets its own invoice format without code changes or a new table. A manufacturing company with 40 branches can give each a distinct prefix (NORTH-INV-, SOUTH-INV-) that routes documents to the right P&L centre in their accounting system.

### Circular reference prevention is only a CHECK constraint, not a trigger

`no_self_parent CHECK (uuid != parent_id)` prevents immediate self-reference but not A→B→A cycles. Unlike the tenants table which has a `check_tenant_hierarchy_depth` trigger walking the parent chain, entities relies solely on the application layer to prevent cycles. A bug in the app can create an infinite loop that crashes any recursive CTE touching that subtree.

**Fix:** Port the hierarchy depth check trigger from tenants to entities with a `max_depth = 8` (deeper trees are valid for organisations).

---

## Modules / Resources / Actions Schema

### The schema is half built — there's no permissions table

You have `modules`, `resources`, and `actions` as three separate tables, but no table connecting them. A permission is `(resource, action)` — "can READ Invoice API", "can APPROVE Purchase Order". Without a `permissions` table defining these pairs, nothing can be granted. The IAM system is missing its central artifact.

**Fix:** Add the missing tables to complete the IAM model:

```sh 
  db/migration/000404_auth_define_permissions.up.sql
  db/migration/000405_auth_create_roles.up.sql                               
  db/migration/000406_auth_map_role_permissions.up.sql             
  db/migration/000407_auth_assign_user_roles.up.sql
```

**Tenant benefit:** Without this, every permission check in the application is either hard-coded or uses an ad-hoc JSONB structure. With it, tenant administrators can create custom roles ("Regional Finance Viewer", "Branch HR Manager"), assign them to users, and scope them to specific entities in their org tree — all through a UI without code changes.

### `modules` and `actions` have no `tenant_id` and no RLS

These are commented out. For system modules and standard CRUD actions this is fine — they're read-only platform data. But the comment block that would add per-tenant modules/actions is entirely commented out with no migration path. If a tenant needs a custom module or a custom action type, they have no place to put it.

**Fix:** Implement the scope pattern from templates: add `scope VARCHAR(10) CHECK (scope IN ('SYSTEM', 'TENANT'))` and `tenant_id UUID` (nullable for SYSTEM scope). RLS then selectively exposes system-scoped rows to all tenants and tenant-scoped rows only to their owner.

### No `updated_at` on modules, resources, or actions

These tables have `created_at` but no `updated_at`. When a module's `display_name` or `description` is corrected, there's no way to know the row was changed or when. This matters for cache invalidation in services that preload module metadata.

### `actions.requires_approval` is a flag with nowhere to go

The column exists but there's no `approval_workflows` table, no link to an approver role, no escalation path. A flag that records "this action needs approval" without defining who approves it or how is documentation masquerading as functionality.

**Fix:** Either remove the column until the approval workflow module is built, or add a minimal `approver_role_id UUID REFERENCES roles(id)` so the flag has actionable meaning. A boolean alone can't drive a workflow.

### `resources.path` has no format validation

The column is VARCHAR(500) storing anything from `/api/v1/invoices` to `finance.invoice.create` to just `invoices`. Without a format convention, every consuming service parses the path field differently and lookups become string matching nightmares.

---

## How All Three Schemas Benefit the Tenant Together

The Settings, Entities, and IAM schemas are not independent — they form the tenant's customisation surface. The connection points matter most:

**Configuration scoped to org hierarchy:** `entities.settings` stores entity-level overrides, but the ConfigurationService must know which entity to look up. The `hierarchy_paths` closure table is what makes "entity X inherits from its DEPARTMENT, which inherits from REGION, which inherits from tenant" fast. A broken or slow closure table means configuration resolution either fails or scans the entire entities table on every request.

**Permission scoped to entity:** The `user_roles.entity_id` FK I proposed above is the critical link. A Finance Manager at the Manchester branch should be able to approve invoices for Manchester entities but not London entities. Without `entity_id` on role assignments, you can only do tenant-wide permission grants, which is too coarse for any organisation with more than one location.

**Feature flags scoped by plan_tier:** The `required_feature_flag` column on `config_definitions` connects to the `tenants.plan_tier` column from the tenants migration. The FeatureFlag service evaluates whether `inventory.lot_tracking` is enabled for this tenant before the Settings module resolves that config key. If the IAM schema has `modules.required_plan_tier`, the same `plan_tier` column drives both which modules are visible and which configs are available — one source of truth for what a tenant has paid for.

The cleanest summary: fix `config_definitions` ownership (drop tenant_id), complete the IAM schema (add permissions + roles tables), fix the closure table (drop the redundant entity_id column and slow views), and add `entitystate.config` for document formatting. Those four changes unlock the full three-layer configuration inheritance the PRD describes, with org-scoped permissions and flexible document numbering, all from the existing infrastructure.
