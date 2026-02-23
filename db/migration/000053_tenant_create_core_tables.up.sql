-- =============================================================================
-- MIGRATION 003 UP: Tenants Table
-- =============================================================================
-- Architecture Decision (ADR-003):
--   The tenants table is the single authoritative record for each customer
--   organisation in this multi-tenant ERP. It is intentionally kept flat —
--   new capabilities are added as columns, not as new tables, to avoid
--   premature normalization while the domain is still evolving.
--
--   The only exception is config_definitions (migration 008), which belongs
--   to the Settings module and has its own ownership boundary.
--
-- Column groupings (document order):
--   1. Identity        — id, slug, name
--   2. Contact         — email, billing_email, billing_contact_name, subdomain
--   3. Lifecycle       — Status, plan_tier, last_activity_at
--   4. Locale          — timezone, currency_code
--   5. Flexible data   — metadata, settings
--   6. Classification  — industry, company_size
--   7. Compliance      — tax_id, registration_number, legal_entity_type
--   8. Hierarchy       — parent_tenant_id
--   9. Audit provenance— created_by, deleted_by, created_at, updated_at, deleted_at
--
-- Architecture Decision (ADR-004): metadata vs. settings JSONB columns
--   TWO JSONB columns are intentional and serve different owners:
--   • metadata  = integration data owned by third-party systems and ops
--                 (webhook URLs, Stripe customer IDs, Salesforce org IDs)
--   • settings  = product-owned configuration managed by the Settings
--                 module's ConfigurationService (feature flags, UX prefs,
--                 module defaults). Other services READ this via the
--                 GetEffectiveConfiguration() interface; they do NOT write
--                 directly to this column.
--   Merging them into one column would couple unrelated ownership domains.
--
-- Architecture Decision (ADR-005): plan_tier on tenants (not a new table)
--   plan_tier is a denormalized column. A full subscriptions table is the
--   long-term home, but that belongs to the Billing module and should be
--   built when that module is ready. In the interim plan_tier gives the
--   FeatureFlag and Settings services a typed, indexable signal without
--   requiring a cross-module JOIN. When the Billing module is built it
--   can own plan details and write back to this column via an event.
--
-- Architecture Decision (ADR-006): parent_tenant_id self-reference
--   Supports enterprise hierarchies (parent org → subsidiary tenants) and
--   white-label reseller chains without a schema change. NULL means a
--   root-level tenant. Depth is application-enforced (recommend max 3).
--   Circular reference prevention is handled by a trigger (migration 007).
--
-- Architecture Decision (ADR-007): created_by / deleted_by provenance
--   Soft delete without actor identity is operationally useless. These
--   columns store the UUID of the actor (human user or service account)
--   that performed the action. The application layer is responsible for
--   populating them. The Audit Service captures the full event payload;
--   these columns are a fast-path convenience for the ops dashboard.
-- =============================================================================

SET row_security = ON;

CREATE TABLE IF NOT EXISTS tenants (

  -- =========================================================================
  -- 1. IDENTITY
  -- =========================================================================

  id    UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,

  -- slug: URL-safe tenant identifier used in subdirectory routing and API paths.
  -- Auto-generated from name on INSERT by trigger if not supplied.
  -- IMMUTABLE after creation — changing it breaks bookmarked URLs and
  -- stored API paths. Immutability enforced by trigger (migration 007).
  -- UNIQUE constraint is the authoritative collision guard for the slug
  -- generation trigger (see ADR-008 in migration 007).
  slug  VARCHAR(50)  NOT NULL
        CONSTRAINT tenants_slug_key UNIQUE
        CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),

  name  VARCHAR(255) NOT NULL CONSTRAINT tenants_name_key UNIQUE,

  -- =========================================================================
  -- 2. CONTACT
  -- =========================================================================

  -- Primary contact email: used for account notifications and authentication.
  -- Distinct from billing_email so finance comms can go to a different inbox
  -- without changing the account login email.
  email                 VARCHAR(255) NOT NULL,

  -- Billing contact: receives invoices, payment failure notices, receipts.
  -- NULL means fall back to the primary email (application-enforced).
  billing_email         VARCHAR(255),

  -- Human-readable name for the billing contact person or team.
  -- Printed on invoices; not used for authentication.
  billing_contact_name  VARCHAR(255),

  -- Subdomain tenants use for vanity routing (e.g. acme.yourplatform.com).
  -- NULL means the tenant uses default path-based routing.
  -- Format enforced by CHECK; reserved names blocked by trigger (007).
  subdomain  VARCHAR(63) UNIQUE
             CHECK (
               subdomain IS NULL
               OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'
             ),

  -- =========================================================================
  -- 3. LIFECYCLE
  -- =========================================================================

  -- Status column retains the quoted "Status" name from the original schema
  -- to preserve ORM mappings that quote the identifier.
  -- PostgreSQL folds unquoted identifiers to lowercase; quoting here is
  -- cosmetic but must be consistent with application-layer queries.
  "Status"   VARCHAR(20) NOT NULL DEFAULT 'PENDING'
             CHECK ("Status" IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),

  -- plan_tier: denormalized signal for the FeatureFlag and Settings services.
  -- See ADR-005 above. The Billing module owns the authoritative subscription
  -- record and writes back to this column via event.
  plan_tier  VARCHAR(20) NOT NULL DEFAULT 'STARTER'
             CHECK (plan_tier IN ('STARTER', 'GROWTH', 'PROFESSIONAL', 'ENTERPRISE')),

  -- last_activity_at: tracks the most recent meaningful interaction.
  -- Updated by trigger on any row change AND by the application on significant
  -- events (logins, API calls) that do not touch other columns.
  -- Trigger only sets this when the application has NOT already changed it
  -- (see ADR-009 in migration 007 for the conditional trigger logic).
  last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- =========================================================================
  -- 4. LOCALE
  -- =========================================================================

  timezone      VARCHAR(50) NOT NULL DEFAULT 'UTC',
  currency_code CHAR(3)     NOT NULL DEFAULT 'USD'
                CHECK (currency_code ~* '^[A-Z]{3}$'),

  -- =========================================================================
  -- 5. FLEXIBLE DATA  (see ADR-004)
  -- =========================================================================

  -- NOT NULL DEFAULT '{}' — v1 allowed NULL here which caused silent
  -- JSON merge failures in integration code. Empty object is correct sentinel.
  metadata  JSONB NOT NULL DEFAULT '{}',
  settings  JSONB NOT NULL DEFAULT '{}',

  -- =========================================================================
  -- 6. CLASSIFICATION
  -- =========================================================================

  industry      VARCHAR(50),
  company_size  VARCHAR(20)
                CHECK (company_size IN
                  ('STARTUP','SMALL','MEDIUM','LARGE','ENTERPRISE')),

  -- =========================================================================
  -- 7. COMPLIANCE
  -- =========================================================================
  -- These three columns live here for now because the Compliance module does
  -- not yet exist. When it does, extract these to a tenant_legal_profiles
  -- table owned by that module. Until then, keeping them here avoids
  -- premature normalization for data that 80%+ of tenants will leave NULL.

  tax_id               VARCHAR(50),
  registration_number  VARCHAR(50),
  legal_entity_type    VARCHAR(50),

  -- =========================================================================
  -- 8. HIERARCHY  (see ADR-006)
  -- =========================================================================

  -- Self-referential FK for enterprise org trees and reseller chains.
  -- Deferrable INITIALLY DEFERRED allows bulk-inserting a parent and its
  -- children in the same transaction without FK ordering constraints.
  -- Circular reference prevention is handled by the check_tenant_hierarchy
  -- trigger in migration 007.
  parent_tenant_id  UUID REFERENCES tenants(id)
                    ON DELETE RESTRICT          -- never orphan children silently
                    DEFERRABLE INITIALLY DEFERRED,

  -- =========================================================================
  -- 9. AUDIT PROVENANCE  (see ADR-007)
  -- =========================================================================

  -- UUID of the actor (user or service account) that created this tenant.
  -- Application layer is responsible for populating on INSERT.
  -- NULL means created by a migration or seed script.
  created_by  UUID,

  -- UUID of the actor that issued the soft-delete (set deleted_at != NULL).
  -- Application layer sets both deleted_at and deleted_by together.
  -- NULL when the tenant is not deleted.
  deleted_by  UUID,

  -- Standard audit timestamps — both maintained by triggers in migration 007.
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- Soft-delete: NULL means the tenant is active.
  -- Hard deletes are prohibited at the database level (no cascade DELETE).
  -- Archive workflow: set Status = 'ARCHIVED' first, then deleted_at after
  -- a configurable retention window.
  deleted_at  TIMESTAMPTZ
);

-- -------------------------------------------------------------------------
-- TABLE AND COLUMN COMMENTS
-- -------------------------------------------------------------------------

COMMENT ON TABLE tenants IS
  'Core tenant registry for the multi-tenant ERP. '
  'One row per customer organisation. '
  'The Settings module reads/writes the settings JSONB column via its '
  'ConfigurationService — other services must not write directly to it. '
  'Soft-delete only: set deleted_at + deleted_by; never hard-DELETE.';

COMMENT ON COLUMN tenants.id                  IS 'Immutable UUID — used in all external API references and foreign keys.';
COMMENT ON COLUMN tenants.slug                IS 'URL-safe identifier. Auto-generated from name on INSERT. IMMUTABLE after creation (trigger-enforced).';
COMMENT ON COLUMN tenants.name                IS 'Human-readable display name. Must be globally unique.';
COMMENT ON COLUMN tenants.email               IS 'Primary account email — used for auth and account notifications.';
COMMENT ON COLUMN tenants.billing_email       IS 'Billing contact email — receives invoices and payment alerts. Falls back to email when NULL.';
COMMENT ON COLUMN tenants.billing_contact_name IS 'Billing contact name printed on invoices. Optional.';
COMMENT ON COLUMN tenants.subdomain           IS 'Vanity subdomain for tenant routing. NULL = use path-based routing. Reserved names blocked by trigger.';
COMMENT ON COLUMN tenants."Status"            IS 'Lifecycle state. New tenants start PENDING until email verification + payment setup complete.';
COMMENT ON COLUMN tenants.plan_tier           IS 'Denormalized subscription tier for FeatureFlag and Settings services. Billing module is source of truth; writes back here via event.';
COMMENT ON COLUMN tenants.last_activity_at    IS 'Last meaningful interaction (login, API call, row change). Updated by trigger unless application already set it in this transaction.';
COMMENT ON COLUMN tenants.timezone            IS 'IANA timezone string (e.g. America/New_York). Used for date-relative operations in the ERP.';
COMMENT ON COLUMN tenants.currency_code       IS 'ISO 4217 three-letter currency code. Default is USD.';
COMMENT ON COLUMN tenants.metadata            IS 'Integration metadata owned by ops/third-party systems: Stripe IDs, webhook URLs, external system refs. NOT managed by Settings module.';
COMMENT ON COLUMN tenants.settings            IS 'Application-owned config managed exclusively by the Settings module (ConfigurationService). Other services must use GetEffectiveConfiguration() — do not write here directly.';
COMMENT ON COLUMN tenants.industry            IS 'Industry vertical for analytics and template selection (e.g. Manufacturing, Retail, Services).';
COMMENT ON COLUMN tenants.company_size        IS 'Approximate headcount band. Used for template selection and capacity planning.';
COMMENT ON COLUMN tenants.tax_id              IS 'Government tax identifier (VAT, EIN, GST etc.). Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.registration_number IS 'Company registry number. Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.legal_entity_type   IS 'Legal structure (LLC, Ltd, PLC, GmbH etc.). Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.parent_tenant_id    IS 'Self-reference for enterprise org trees and reseller chains. NULL = root tenant. Max depth 3 (application-enforced). Circular refs blocked by trigger.';
COMMENT ON COLUMN tenants.created_by          IS 'UUID of the actor (user or service account) that created this tenant. NULL if seeded by migration.';
COMMENT ON COLUMN tenants.deleted_by          IS 'UUID of the actor that issued the soft-delete. Set alongside deleted_at. NULL when not deleted.';
COMMENT ON COLUMN tenants.created_at          IS 'Row creation timestamp.';
COMMENT ON COLUMN tenants.updated_at          IS 'Last row modification timestamp. Maintained by trigger.';
COMMENT ON COLUMN tenants.deleted_at          IS 'Soft-delete timestamp. NULL = tenant is active. Never hard-DELETE a tenant row.';





-- -- =====================================================
-- -- EXTENSIONS
-- -- =====================================================
-- -- Enable UUID generation for unique identifiers
-- -- Enable required extensions
--
--
-- -- CREATE SCHEMA IF NOT EXISTS ledger;
--
-- -- SET search_path TO ledger;
--
-- -- Enable Row Level Security globally
-- SET
--   row_security = ON;
--
-- -- =====================================================
-- -- ROLES AND PERMISSIONS
-- -- =====================================================
-- -- Create application role if it doesn't exist
-- DO
-- $$
-- BEGIN
-- IF NOT EXISTS (
--   SELECT
--     1
--   FROM
--     pg_roles
--   WHERE
--     rolname = 'application_role'
-- ) THEN CREATE ROLE application_role;
--
-- END IF;
--
-- END
-- $$
-- ;
--
-- -- Create admin role if it doesn't exist
-- DO
-- $$
-- BEGIN
-- IF NOT EXISTS (
--   SELECT
--     1
--   FROM
--     pg_roles
--   WHERE
--     rolname = 'admin_role'
-- ) THEN CREATE ROLE admin_role;
--
-- END IF;
--
-- END
-- $$
-- ;
--
-- DO
-- $$
-- BEGIN
-- IF NOT EXISTS (
--   SELECT
--     1
--   FROM
--     pg_roles
--   WHERE
--     rolname = 'readonly_role'
-- ) THEN CREATE ROLE readonly_role;
--
-- END IF;
--
-- END
-- $$
-- ;
--
-- -- =====================================================
-- -- CORE TENANT MANAGEMENT
-- -- =====================================================
-- -- -----------------------------------------------------
-- -- TENANTS TABLE
-- -- -----------------------------------------------------
-- -- Primary table for multi-tenant SaaS architecture
-- -- Stores tenant information, business details, and configuration
-- CREATE TABLE tenants (
--   -- Primary identifiers
--   id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
--   slug VARCHAR(50) NOT NULL,
--   name VARCHAR(255) UNIQUE NOT NULL,
--   -- Contact and access information
--   email VARCHAR(255) NOT NULL,
--   subdomain VARCHAR(63) UNIQUE,
--   -- Status and operational settings
--   Status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (Status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),
--   timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
--   currency_code CHAR(3) NOT NULL DEFAULT 'USD',
--   -- Flexible metadata storage
--   metadata JSONB DEFAULT '{}',
--   -- Business classification
--   industry VARCHAR(50),  -- For future industry-specific modules
--   company_size VARCHAR(20) CHECK (
--     company_size IN (
--       'Startup',
--       'Small',
--       'Medium',
--       'Large',
--       'Enterprise'
--     )
--   ),
--   -- Compliance and legal information
--   tax_id VARCHAR(50),
--   registration_number VARCHAR(50),
--   legal_entity_type VARCHAR(50),
--   -- Tenant-specific settings
--   last_activity_at TIMESTAMPTZ DEFAULT NOW(),
--   settings JSONB NOT NULL DEFAULT '{}',
--   -- Audit timestamps
--   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--   deleted_at TIMESTAMPTZ -- Soft delete support
-- );
--
-- -- Create indexes for performance
-- CREATE INDEX idx_tenants_slug ON tenants(slug);
--
-- CREATE INDEX idx_tenants_status ON tenants(Status);
--
-- CREATE INDEX idx_tenants_subdomain ON tenants(subdomain)
-- WHERE
--   subdomain IS NOT NULL;
--
-- CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at)
-- WHERE
--   deleted_at IS NOT NULL;
--
-- -- Add comments for documentation
-- COMMENT ON TABLE tenants IS 'Core tenant management table for multi-tenant SaaS architecture';
--
-- COMMENT ON COLUMN tenants.id IS 'Universal unique identifier for external API references';
--
-- COMMENT ON COLUMN tenants.slug IS 'URL-friendly tenant identifier';
--
-- COMMENT ON COLUMN tenants.metadata IS 'Flexible JSONB storage for additional tenant metadata';
--
-- COMMENT ON COLUMN tenants.settings IS 'Tenant-specific configuration settings';
--
-- COMMENT ON COLUMN tenants.deleted_at IS 'Soft delete timestamp - NULL means active';
--
-- -- =====================================================
-- -- UTILITY FUNCTIONS
-- -- =====================================================
-- -- -----------------------------------------------------
-- -- TENANT CONTEXT MANAGEMENT
-- -- -----------------------------------------------------
-- -- Function to set tenant context for the current session
-- CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID, user_role TEXT DEFAULT 'application_role') 
--   RETURNS VOID AS $$
-- DECLARE
--   tenant_status TEXT;
-- BEGIN
--   -- Get tenant status in one query
--   SELECT status INTO tenant_status 
--   FROM tenants 
--   WHERE id = tenant_id AND deleted_at IS NULL;
--
--   IF NOT FOUND THEN
--     RAISE EXCEPTION 'Tenant not found: %', tenant_id;
--   END IF;
--
--   IF tenant_status != 'ACTIVE' THEN
--     RAISE EXCEPTION 'Tenant is not active: % (status: %)', tenant_id, tenant_status;
--   END IF;
--
--   -- Set multiple context variables
--   PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
--   PERFORM set_config('app.tenant_status', tenant_status, true);
--   PERFORM set_config('app.context_set_at', NOW()::text, true);
--
-- END;
-- $$ LANGUAGE plpgsql SECURITY DEFINER;
--
-- -- Add function comment
-- COMMENT ON FUNCTION set_tenant_context(UUID,TEXT) IS 'Sets the current tenant context for the session with validation';
--
-- -- -----------------------------------------------------
-- -- GET CURRENT TENANT FUNCTION
-- -- -----------------------------------------------------
-- -- Utility function to retrieve current tenant ID from session
-- CREATE
-- OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS
-- $$
-- BEGIN
-- -- Return current tenant ID from session variable, default to NULL if not set
-- RETURN COALESCE(
--   nullif(
--     current_setting('app.current_tenant_id', FALSE),
--     ''
--   ),
--   NULL
-- )::UUID;
--
-- EXCEPTION
-- WHEN OTHERS THEN
-- -- Return NULL if any error occurs (e.g., invalid cast)
-- RETURN NULL;
--
-- END;
--
-- $$
-- LANGUAGE plpgsql;
--
-- -- Add function comment
-- COMMENT ON FUNCTION current_tenant_id() IS 'Retrieves the current tenant ID from session context';
--
-- -- Function to clear tenant context (important for connection pooling)
-- CREATE OR REPLACE FUNCTION clear_tenant_context() 
--   RETURNS VOID AS $$
-- BEGIN
--   PERFORM set_config('app.current_tenant_id', NULL, true);
--   PERFORM set_config('app.tenant_status', NULL, true);
--   PERFORM set_config('app.context_set_at', NULL, true);
-- END;
-- $$ LANGUAGE plpgsql;
--
-- COMMENT ON FUNCTION clear_tenant_context() IS 'clear tenant context (important for connection pooling)';
--
-- -- =====================================================
-- -- ROW LEVEL SECURITY (RLS)
-- -- =====================================================
-- -- -----------------------------------------------------
-- -- ENABLE RLS ON TENANT TABLE
-- -- -----------------------------------------------------
-- -- Enable Row Level Security on tenants table
-- ALTER TABLE
--   tenants ENABLE ROW LEVEL SECURITY;
--
-- -- Create policy for tenant isolation
-- -- Only allow access to tenant data based on current session context
-- -- FIXME: I am not sure if the tenants table can take this policy
-- CREATE POLICY tenant_isolation_policy ON tenants FOR ALL TO application_role USING (
--   id = current_tenant_id()
--   OR current_tenant_id() IS NULL
-- );
--
-- CREATE POLICY admin_full_access_policy ON tenants FOR ALL TO admin_role USING (true);
--
-- CREATE POLICY readonly_access_policy ON tenants FOR SELECT TO readonly_role USING (true);
--
-- --
-- -- Add policy comment
-- COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Ensures tenant data isolation based on session context';
-- COMMENT ON POLICY admin_full_access_policy ON tenants IS 'Allows admin_role full access to all tenant data';
--
-- -- =====================================================
-- -- PERMISSIONS AND GRANTS
-- -- =====================================================
-- -- Grant necessary permissions to application role
-- GRANT SELECT , INSERT ,UPDATE , DELETE ON tenants TO application_role;
--
-- -- Grant necessary permissions to admin_role
-- GRANT ALL PRIVILEGES ON tenants TO admin_role;
--
--
-- -- Grant necessary permissions t readonly_role
-- GRANT 
--   SELECT
--    ON tenants TO readonly_role;
--
--
-- -- Grant execute permissions on functions
-- GRANT EXECUTE ON FUNCTION set_tenant_context(UUID,TEXT) TO application_role;
--
-- GRANT EXECUTE ON FUNCTION current_tenant_id() TO application_role;
-- GRANT EXECUTE ON FUNCTION current_tenant_id() TO readonly_role;
--
-- -- =====================================================
-- -- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- -- =====================================================
-- -- -----------------------------------------------------
-- -- UPDATED_AT TRIGGER FUNCTION
-- -- -----------------------------------------------------
-- -- Generic function to update the updated_at timestamp
-- CREATE
-- OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS
-- $$
-- BEGIN
-- NEW.updated_at = NOW();
--
-- RETURN NEW;
--
-- END;
--
-- $$
-- LANGUAGE plpgsql;
--
-- -- Add function comment
-- COMMENT ON FUNCTION update_updated_at_column() IS 'Generic trigger function to update updated_at timestamp';
--
-- -- -----------------------------------------------------
-- -- APPLY TRIGGER TO TENANTS TABLE
-- -- -----------------------------------------------------
-- -- Trigger for tenants table
-- CREATE TRIGGER update_tenants_updated_at BEFORE
-- UPDATE
--   ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
--
-- -- -----------------------------------------------------
-- -- SLUG GENERATION TRIGGER
-- -- -----------------------------------------------------
-- CREATE OR REPLACE FUNCTION generate_unique_slug_from_name() 
--   RETURNS TRIGGER AS $$
-- DECLARE
--   base_slug TEXT;
--   final_slug TEXT;
--   counter INTEGER := 1;
-- BEGIN
--   IF NEW.slug IS NULL THEN
--     -- Create base slug with better sanitization
--     base_slug := lower(trim(both '-' from 
--       regexp_replace(
--         regexp_replace(NEW.name, '[^\w\s-]', '', 'g'),
--         '\s+', '-', 'g'
--       )
--     ));
--
--     -- Ensure slug is not empty
--     IF base_slug = '' THEN
--       base_slug := 'tenant';
--     END IF;
--
--     final_slug := base_slug;
--
--     -- Handle slug collisions
--     WHILE EXISTS(SELECT 1 FROM tenants WHERE slug = final_slug AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID)) LOOP
--       final_slug := base_slug || '-' || counter;
--       counter := counter + 1;
--     END LOOP;
--
--     NEW.slug := final_slug;
--   END IF;
--
--   RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER tenant_slug_trigger BEFORE
-- INSERT
--   ON tenants FOR EACH ROW EXECUTE FUNCTION generate_unique_slug_from_name();
--
--
-- -- Better email validation
-- -- ALTER TABLE tenants ADD CONSTRAINT valid_email
-- --   CHECK ( email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
--
-- -- Subdomain validation
-- ALTER TABLE tenants ADD CONSTRAINT valid_subdomain 
--   CHECK (subdomain IS NULL OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$');
--
-- -- Slug validation
-- ALTER TABLE tenants ADD CONSTRAINT valid_slug 
--   CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$');
--
-- -- Currency code validation (ISO 4217)
-- ALTER TABLE tenants ADD CONSTRAINT valid_currency 
--   CHECK (currency_code ~* '^[A-Z]{3}$');
