-- ------------------------------------------------------------------------------------------------
-- TENANT TABLE — TRIGGER FUNCTIONS AND TRIGGERS
-- ------------------------------------------------------------------------------------------------
-- All trigger functions in this file follow the same structure:
--   1. Function definition with detailed inline comments.
--   2. DROP TRIGGER IF EXISTS (idempotent re-runs).
--   3. CREATE TRIGGER pinned to the correct event and timing.
--
-- Trigger inventory:
--   A. update_updated_at_column          — BEFORE UPDATE, generic timestamp maintenance
--   B. update_last_activity_at_column    — BEFORE UPDATE, conditional activity tracking
--   C. generate_unique_slug_from_name    — BEFORE INSERT, slug auto-generation
--   D. enforce_slug_immutability         — BEFORE UPDATE, slug change guard
--   E. check_subdomain_not_reserved      — BEFORE INSERT OR UPDATE, reservation guard
--   F. check_tenant_hierarchy_depth      — BEFORE INSERT OR UPDATE, loop/depth guard
--
-- Security: all SECURITY DEFINER functions pin SET search_path = pg_catalog, public
-- to prevent search path injection (a malicious user cannot shadow the tenants table).
--
-- NOTE: update_updated_at_column() is generic and reused by other tables.
--       Its trigger on tenants is registered here; other tables add their own
--       triggers in their respective migrations.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- A. update_updated_at_column
-- ------------------------------------------------------------------------------------------------
-- Generic trigger function: sets updated_at to NOW() before any UPDATE.
-- Shared across multiple tables — reused by any table with an updated_at column.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION update_updated_at_column() IS
  'Generic BEFORE UPDATE trigger: sets updated_at = NOW(). '
  'Safe to reuse across multiple tables.';

DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
CREATE TRIGGER update_tenants_updated_at
  BEFORE UPDATE ON tenants
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- B. update_last_activity_at_column
-- ------------------------------------------------------------------------------------------------
-- Conditional trigger: only updates last_activity_at when the application has NOT
-- already set it to a new value in this statement.
--
-- Why conditional: firing on EVERY row update conflates "the row was touched" with
-- "the tenant was active", making last_activity_at useless as a churn-detection signal.
-- The trigger acts as a safe default, not an override. Application code should set
-- last_activity_at explicitly on meaningful events (logins, API calls) that do not
-- update other columns:
--   UPDATE tenants SET last_activity_at = NOW() WHERE id = $1
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_last_activity_at_column()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  -- Only auto-update when the application has not explicitly set a new value.
  IF OLD.last_activity_at IS NOT DISTINCT FROM NEW.last_activity_at THEN
    NEW.last_activity_at = NOW();
  END IF;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION update_last_activity_at_column() IS
  'BEFORE UPDATE trigger: sets last_activity_at = NOW() only when the '
  'application has not already set a new value in this statement. '
  'Application code should set last_activity_at explicitly on meaningful '
  'events (logins, API calls) that do not update other columns.';

DROP TRIGGER IF EXISTS update_tenants_last_activity ON tenants;
CREATE TRIGGER update_tenants_last_activity
  BEFORE UPDATE ON tenants
  FOR EACH ROW
  EXECUTE FUNCTION update_last_activity_at_column();

-- ------------------------------------------------------------------------------------------------
-- C. generate_unique_slug_from_name
-- ------------------------------------------------------------------------------------------------
-- Auto-generates a URL-safe slug from the tenant name on INSERT when no slug is
-- explicitly provided.
--
-- TOCTOU note: the EXISTS check and the INSERT are not atomic. A concurrent
-- transaction can win the race between them. The UNIQUE constraint on tenants.slug
-- is the authoritative collision guard — the loop handles only the sequential common
-- case (two tenants named "Acme"). The UNIQUE constraint makes races surface as a
-- retryable unique_violation at the application layer.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION generate_unique_slug_from_name()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
DECLARE
  v_base    TEXT;
  v_final   TEXT;
  v_counter INTEGER := 1;
BEGIN
  -- If the caller supplied a slug, trust it and skip generation.
  -- The UNIQUE constraint and valid_slug CHECK will validate it.
  IF NEW.slug IS NOT NULL THEN
    RETURN NEW;
  END IF;

  -- Sanitise: lowercase → collapse whitespace to hyphens →
  --   strip non-word/non-hyphen characters → trim leading/trailing hyphens.
  v_base := lower(
    trim(BOTH '-' FROM
      regexp_replace(
        regexp_replace(NEW.name, '[^\w\s-]', '', 'g'),
        '\s+', '-', 'g'
      )
    )
  );

  -- Guard: if the name produces an empty string after sanitisation
  -- (e.g. the name was all emoji or special characters), use 'tenant'.
  IF v_base = '' THEN
    v_base := 'tenant';
  END IF;

  v_final := v_base;

  -- Walk candidates until we find one not already taken.
  WHILE v_counter <= 1000 LOOP
    IF NOT EXISTS (
      SELECT 1 FROM tenants WHERE slug = v_final
    ) THEN
      NEW.slug := v_final;
      RETURN NEW;
    END IF;

    v_final   := v_base || '-' || v_counter;
    v_counter := v_counter + 1;
  END LOOP;

  -- Exhausted 1000 candidates — caller should supply an explicit slug.
  RAISE EXCEPTION
    'generate_unique_slug_from_name: could not find a unique slug for "%" '
    'after 1000 attempts — supply an explicit slug instead',
    NEW.name
    USING ERRCODE = 'unique_violation';
END;
$$;

COMMENT ON FUNCTION generate_unique_slug_from_name() IS
  'BEFORE INSERT trigger: auto-generates a URL-safe slug from tenant name '
  'when slug is not supplied. The UNIQUE constraint on tenants.slug is the '
  'authoritative collision guard — this loop handles the sequential common case. '
  'Slugs are not regenerated on UPDATE (see enforce_slug_immutability trigger).';

DROP TRIGGER IF EXISTS tenant_slug_trigger ON tenants;
CREATE TRIGGER tenant_slug_trigger
  BEFORE INSERT ON tenants
  FOR EACH ROW
  EXECUTE FUNCTION generate_unique_slug_from_name();

-- ------------------------------------------------------------------------------------------------
-- D. enforce_slug_immutability
-- ------------------------------------------------------------------------------------------------
-- Prevents slug changes after creation. Slugs appear in public URLs, API paths,
-- webhook endpoints, and stored references in third-party integrations. Allowing
-- a slug change would silently break all of them.
--
-- To override for a legitimate slug change, an admin must:
--   ALTER TABLE tenants DISABLE TRIGGER enforce_tenants_slug_immutability;
--   UPDATE tenants SET slug = 'new-slug' WHERE id = '...';
--   ALTER TABLE tenants ENABLE TRIGGER enforce_tenants_slug_immutability;
-- This makes the change deliberate and leaves a DDL audit trail.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enforce_slug_immutability()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  IF OLD.slug IS DISTINCT FROM NEW.slug THEN
    RAISE EXCEPTION
      'enforce_slug_immutability: slug is immutable after creation — '
      'old: %, attempted: %. '
      'To change a slug deliberately, disable this trigger (requires admin DDL access).',
      OLD.slug, NEW.slug
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION enforce_slug_immutability() IS
  'BEFORE UPDATE trigger: prevents slug changes after INSERT. '
  'Slugs appear in public URLs, API paths, and third-party integrations — '
  'changing them silently would break all stored references. '
  'To change a slug deliberately, an admin must disable this trigger, '
  'make the change, then re-enable it (produces a DDL audit event).';

DROP TRIGGER IF EXISTS enforce_tenants_slug_immutability ON tenants;
CREATE TRIGGER enforce_tenants_slug_immutability
  BEFORE UPDATE ON tenants
  FOR EACH ROW
  WHEN (OLD.slug IS DISTINCT FROM NEW.slug)
  EXECUTE FUNCTION enforce_slug_immutability();

-- ------------------------------------------------------------------------------------------------
-- E. check_subdomain_not_reserved
-- ------------------------------------------------------------------------------------------------
-- Blocks tenants from registering subdomains that clash with platform infrastructure
-- routes. Implemented as a trigger (not a CHECK constraint) because PostgreSQL CHECK
-- constraints cannot reliably query other tables — they only re-validate the modified
-- row, not concurrent inserts. A BEFORE INSERT OR UPDATE trigger fires correctly
-- under REPEATABLE READ and SERIALIZABLE isolation.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_subdomain_not_reserved()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public   -- search path pin: prevents shadow table injection
AS $$
BEGIN
  -- Only check when subdomain is actually being set.
  IF NEW.subdomain IS NULL THEN
    RETURN NEW;
  END IF;

  -- Only run the check when subdomain is being inserted or changed.
  IF TG_OP = 'UPDATE' AND OLD.subdomain IS NOT DISTINCT FROM NEW.subdomain THEN
    RETURN NEW;
  END IF;

  IF EXISTS (
    SELECT 1 FROM reserved_subdomains WHERE subdomain = NEW.subdomain
  ) THEN
    RAISE EXCEPTION
      'check_subdomain_not_reserved: subdomain "%" is reserved for platform '
      'infrastructure and cannot be registered by a tenant. '
      'Choose a different subdomain.',
      NEW.subdomain
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION check_subdomain_not_reserved() IS
  'BEFORE INSERT OR UPDATE trigger: blocks tenants from registering subdomains '
  'that appear in the reserved_subdomains table. '
  'Implemented as a trigger (not a CHECK constraint) because CHECK constraints '
  'cannot reliably query other tables in PostgreSQL. '
  'SECURITY DEFINER with search_path pin to prevent shadow table injection.';

DROP TRIGGER IF EXISTS tenant_subdomain_reserved_check ON tenants;
CREATE TRIGGER tenant_subdomain_reserved_check
  BEFORE INSERT OR UPDATE ON tenants
  FOR EACH ROW
  EXECUTE FUNCTION check_subdomain_not_reserved();

-- ------------------------------------------------------------------------------------------------
-- F. check_tenant_hierarchy_depth
-- ------------------------------------------------------------------------------------------------
-- Prevents circular references (A→B→A) and enforces a maximum hierarchy depth
-- in the parent_tenant_id self-reference.
--
-- The trigger caps depth at 5 levels and detects circular references by walking
-- the parent chain. This is a safety net; the application UI should enforce a
-- tighter limit (e.g. 3 levels) for UX reasons.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_tenant_hierarchy_depth()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public   -- search path pin
AS $$
DECLARE
  v_max_depth  CONSTANT INTEGER := 5;
  v_current_id UUID;
  v_depth      INTEGER := 0;
BEGIN
  -- No parent set — root tenant, nothing to validate.
  IF NEW.parent_tenant_id IS NULL THEN
    RETURN NEW;
  END IF;

  -- A tenant cannot be its own parent.
  IF NEW.parent_tenant_id = NEW.id THEN
    RAISE EXCEPTION
      'check_tenant_hierarchy_depth: tenant cannot reference itself as parent — id: %',
      NEW.id
      USING ERRCODE = 'check_violation';
  END IF;

  -- Walk up the parent chain looking for cycles and enforcing depth limit.
  v_current_id := NEW.parent_tenant_id;

  WHILE v_current_id IS NOT NULL LOOP
    v_depth := v_depth + 1;

    -- Depth exceeded before we found a NULL parent — hierarchy too deep.
    IF v_depth > v_max_depth THEN
      RAISE EXCEPTION
        'check_tenant_hierarchy_depth: hierarchy depth exceeds maximum of % levels — '
        'tenant id: %, parent chain leads beyond limit',
        v_max_depth, NEW.id
        USING ERRCODE = 'check_violation';
    END IF;

    -- Circular reference detected — an ancestor points back to the current row.
    IF v_current_id = NEW.id THEN
      RAISE EXCEPTION
        'check_tenant_hierarchy_depth: circular parent_tenant_id reference detected — '
        'tenant id: % appears in its own ancestor chain',
        NEW.id
        USING ERRCODE = 'check_violation';
    END IF;

    -- Move one level up the tree.
    SELECT parent_tenant_id
      INTO v_current_id
      FROM tenants
     WHERE id = v_current_id;
  END LOOP;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION check_tenant_hierarchy_depth() IS
  'BEFORE INSERT OR UPDATE trigger: validates the parent_tenant_id self-reference. '
  'Prevents a tenant from being its own parent, detects circular references '
  '(A→B→C→A), and enforces a maximum hierarchy depth of 5 levels. '
  'Application UI should enforce a tighter limit (e.g. 3) for UX reasons. '
  'SECURITY DEFINER with search_path pin.';

DROP TRIGGER IF EXISTS tenant_hierarchy_depth_check ON tenants;
CREATE TRIGGER tenant_hierarchy_depth_check
  BEFORE INSERT OR UPDATE OF parent_tenant_id ON tenants
  FOR EACH ROW
  EXECUTE FUNCTION check_tenant_hierarchy_depth();
