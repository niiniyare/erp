-- =============================================================================
-- FIX: Split is_overridable into two axes on config_definitions
-- =============================================================================
-- WHY: The 3-level inheritance model (System → Tenant → Entity) requires two
--      independent override gates:
--        is_tenant_overridable — can a tenant admin change the system default?
--        is_entity_overridable — can an entity manager change the tenant value?
--      A single is_overridable boolean collapses both gates to the same bit,
--      making it impossible to express "tenant can change this, but entities
--      cannot" (e.g. inventory.default_valuation_method).
--
-- HOW: Add the two new columns with defaults that preserve current behaviour.
--      Migrate is_overridable = true  → both new columns = true
--      Migrate is_overridable = false → both new columns = false
--      The old column is left in place and marked deprecated via comment.
--      Drop it in a subsequent migration once application code is updated.
-- =============================================================================

ALTER TABLE config_definitions
  ADD COLUMN IF NOT EXISTS is_tenant_overridable BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS is_entity_overridable BOOLEAN NOT NULL DEFAULT TRUE;

-- Migrate existing values: preserve the single-bit intent
UPDATE config_definitions
   SET is_tenant_overridable = is_overridable,
       is_entity_overridable = is_overridable
 WHERE is_overridable IS NOT NULL;

COMMENT ON COLUMN config_definitions.is_tenant_overridable IS
  'Whether tenant administrators may override the system default value. '
  'FALSE = system default is locked for all tenants.';

COMMENT ON COLUMN config_definitions.is_entity_overridable IS
  'Whether entity managers may override the tenant-level value. '
  'FALSE = tenant value is locked for all entities within that tenant. '
  'Requires is_tenant_overridable = TRUE to have any effect.';

COMMENT ON COLUMN config_definitions.is_overridable IS
  'DEPRECATED — use is_tenant_overridable and is_entity_overridable instead. '
  'Kept for backward compatibility. Will be dropped in a future migration '
  'once application code references the new columns.';
