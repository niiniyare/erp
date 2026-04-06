-- ------------------------------------------------------------------------------------------------
-- PROVISION_TENANT_COMPLETE FUNCTION
-- ------------------------------------------------------------------------------------------------
-- Atomically provisions a new tenant record in a single function call. Generates a UUID and
-- a URL-safe slug from p_name, appending 8 chars of the UUID if a slug collision is detected.
-- The tenant is created with STATUS='PENDING'; callers must transition it to ACTIVE separately.
--
-- NOTE: Depends on the tenants table (previous migrations) and the
--       create_tenant_configuration_trigger (000102) which auto-creates the config row.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION provision_tenant_complete(
  p_name          VARCHAR(255),
  p_email         VARCHAR(255),
  p_subdomain     VARCHAR(63)  DEFAULT NULL,
  p_industry      VARCHAR(50)  DEFAULT NULL,
  p_company_size  VARCHAR(20)  DEFAULT 'Small',
  p_currency_code CHAR(3)      DEFAULT 'USD',
  p_timezone      VARCHAR(50)  DEFAULT 'UTC',
  p_settings      JSONB        DEFAULT '{}'
) RETURNS TABLE(id UUID) AS $body$
DECLARE
  v_tenant_id UUID;
  v_slug      VARCHAR(50);
BEGIN
  -- Generate UUID and slug
  v_tenant_id := gen_random_uuid();
  v_slug := lower(regexp_replace(p_name, '[^a-zA-Z0-9]+', '-', 'g'));

  -- Ensure slug uniqueness
  WHILE EXISTS (
    SELECT 1 FROM tenants WHERE slug = v_slug AND deleted_at IS NULL
  ) LOOP
    v_slug := v_slug || '-' || substring(v_tenant_id::text, 1, 8);
  END LOOP;

  -- Create tenant record
  INSERT INTO tenants (
    id,
    slug,
    name,
    email,
    subdomain,
    STATUS,
    industry,
    company_size,
    currency_code,
    timezone,
    settings
  ) VALUES (
    v_tenant_id,
    v_slug,
    p_name,
    p_email,
    p_subdomain,
    'PENDING',
    p_industry,
    p_company_size,
    p_currency_code,
    p_timezone,
    p_settings
  );

  -- Return tenant information
  RETURN QUERY SELECT v_tenant_id AS tenant_id;
END;
$body$ LANGUAGE plpgsql;
