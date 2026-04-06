-- ------------------------------------------------------------------------------------------------
-- TENANTS TABLE
-- ------------------------------------------------------------------------------------------------
-- Core tenant registry for the multi-tenant ERP. One row per customer organisation.
-- Intentionally flat — new capabilities are added as columns, not new tables, while
-- the domain is still evolving. The only exception is config_definitions (migration
-- 000058) which belongs to the Settings module and has its own ownership boundary.
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
-- metadata vs. settings JSONB columns (TWO columns are intentional):
--   • metadata  = integration data owned by third-party systems and ops
--                 (webhook URLs, Stripe customer IDs, Salesforce org IDs)
--   • settings  = product-owned configuration managed by the Settings module's
--                 ConfigurationService. Other services must use GetEffectiveConfiguration();
--                 they must NOT write directly to this column.
--
-- plan_tier is denormalized — the Billing module will own the authoritative subscription
-- record and write back to this column via event when built.
--
-- parent_tenant_id supports enterprise hierarchies and white-label reseller chains.
-- NULL means a root-level tenant. Max depth 5 (DB-enforced), 3 recommended (UI).
-- Circular reference prevention is handled by trigger (migration 000057).
--
-- created_by / deleted_by store the UUID of the actor that performed the action.
-- The application layer is responsible for populating them.
--
-- NOTE: Indexes are in migration 000054. Tenant context functions in 000055.
--       RLS policies in 000056. Triggers in 000057.
-- ------------------------------------------------------------------------------------------------

SET row_security = ON;

CREATE TABLE IF NOT EXISTS tenants (

  -- ------------------------------------------------------------------------------------------------
  -- 1. IDENTITY
  -- ------------------------------------------------------------------------------------------------

  id    UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,

  -- URL-safe tenant identifier used in subdirectory routing and API paths.
  -- Auto-generated from name on INSERT by trigger if not supplied.
  -- IMMUTABLE after creation — changing it breaks bookmarked URLs and
  -- stored API paths. Immutability enforced by trigger (migration 000057).
  -- UNIQUE constraint is the authoritative collision guard for the slug
  -- generation trigger (see migration 000057).
  slug  VARCHAR(50)  NOT NULL
        CONSTRAINT tenants_slug_key UNIQUE
        CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),

  name  VARCHAR(255) NOT NULL CONSTRAINT tenants_name_key UNIQUE,

  -- ------------------------------------------------------------------------------------------------
  -- 2. CONTACT
  -- ------------------------------------------------------------------------------------------------

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
  -- Format enforced by CHECK; reserved names blocked by trigger (000057).
  subdomain  VARCHAR(63) UNIQUE
             CHECK (
               subdomain IS NULL
               OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'
             ),

  -- ------------------------------------------------------------------------------------------------
  -- 3. LIFECYCLE
  -- ------------------------------------------------------------------------------------------------

  -- Status column retains the quoted "Status" name from the original schema
  -- to preserve ORM mappings that quote the identifier.
  "Status"   VARCHAR(20) NOT NULL DEFAULT 'PENDING'
             CHECK ("Status" IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),

  -- Denormalized subscription tier for FeatureFlag and Settings services.
  -- The Billing module will own the authoritative subscription record and
  -- write back to this column via event when built.
  plan_tier  VARCHAR(20) NOT NULL DEFAULT 'STARTER'
             CHECK (plan_tier IN ('STARTER', 'GROWTH', 'PROFESSIONAL', 'ENTERPRISE')),

  -- Tracks the most recent meaningful interaction (login, API call, row change).
  -- Updated by trigger on any row change AND by the application on significant
  -- events that do not touch other columns. Trigger only sets this when the
  -- application has NOT already changed it in the same statement (migration 000057).
  last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- ------------------------------------------------------------------------------------------------
  -- 4. LOCALE
  -- ------------------------------------------------------------------------------------------------

  timezone      VARCHAR(50) NOT NULL DEFAULT 'UTC',
  currency_code CHAR(3)     NOT NULL DEFAULT 'USD'
                CHECK (currency_code ~* '^[A-Z]{3}$'),

  -- ------------------------------------------------------------------------------------------------
  -- 5. FLEXIBLE DATA
  -- ------------------------------------------------------------------------------------------------
  -- NOT NULL DEFAULT '{}' — v1 allowed NULL here which caused silent
  -- JSON merge failures in integration code. Empty object is correct sentinel.
  metadata  JSONB NOT NULL DEFAULT '{}',  -- ops/integration-owned: Stripe IDs, webhook URLs, etc.
  settings  JSONB NOT NULL DEFAULT '{}',  -- Settings module-owned: do not write from other services

  -- ------------------------------------------------------------------------------------------------
  -- 6. CLASSIFICATION
  -- ------------------------------------------------------------------------------------------------

  industry      VARCHAR(50),
  company_size  VARCHAR(20)
                CHECK (company_size IN
                  ('STARTUP','SMALL','MEDIUM','LARGE','ENTERPRISE')),

  -- ------------------------------------------------------------------------------------------------
  -- 7. COMPLIANCE
  -- ------------------------------------------------------------------------------------------------
  -- These three columns live here temporarily. When a Compliance module is built,
  -- extract them to a tenant_legal_profiles table owned by that module. Until then,
  -- keeping them here avoids premature normalization for data that 80%+ of tenants
  -- will leave NULL.

  tax_id               VARCHAR(50),
  registration_number  VARCHAR(50),
  legal_entity_type    VARCHAR(50),

  -- ------------------------------------------------------------------------------------------------
  -- 8. HIERARCHY
  -- ------------------------------------------------------------------------------------------------

  -- Self-referential FK for enterprise org trees and reseller chains.
  -- Deferrable INITIALLY DEFERRED allows bulk-inserting a parent and its
  -- children in the same transaction without FK ordering constraints.
  -- Circular reference prevention is handled by the check_tenant_hierarchy
  -- trigger in migration 000057.
  parent_tenant_id  UUID REFERENCES tenants(id)
                    ON DELETE RESTRICT          -- never orphan children silently
                    DEFERRABLE INITIALLY DEFERRED,

  -- ------------------------------------------------------------------------------------------------
  -- 9. AUDIT PROVENANCE
  -- ------------------------------------------------------------------------------------------------

  -- UUID of the actor (user or service account) that created this tenant.
  -- Application layer is responsible for populating on INSERT.
  -- NULL means created by a migration or seed script.
  created_by  UUID,

  -- UUID of the actor that issued the soft-delete (set deleted_at != NULL).
  -- Application layer sets both deleted_at and deleted_by together.
  -- NULL when the tenant is not deleted.
  deleted_by  UUID,

  -- Standard audit timestamps — both maintained by triggers in migration 000057.
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- Soft-delete: NULL means the tenant is active.
  -- Hard deletes are prohibited at the database level (no cascade DELETE).
  -- Archive workflow: set Status = 'ARCHIVED' first, then deleted_at after
  -- a configurable retention window.
  deleted_at  TIMESTAMPTZ
);

-- ------------------------------------------------------------------------------------------------
-- TABLE AND COLUMN COMMENTS
-- ------------------------------------------------------------------------------------------------

COMMENT ON TABLE tenants IS
  'Core tenant registry for the multi-tenant ERP. '
  'One row per customer organisation. '
  'The Settings module reads/writes the settings JSONB column via its '
  'ConfigurationService — other services must not write directly to it. '
  'Soft-delete only: set deleted_at + deleted_by; never hard-DELETE.';

COMMENT ON COLUMN tenants.id                   IS 'Immutable UUID — used in all external API references and foreign keys.';
COMMENT ON COLUMN tenants.slug                 IS 'URL-safe identifier. Auto-generated from name on INSERT. IMMUTABLE after creation (trigger-enforced in migration 000057).';
COMMENT ON COLUMN tenants.name                 IS 'Human-readable display name. Must be globally unique.';
COMMENT ON COLUMN tenants.email                IS 'Primary account email — used for auth and account notifications.';
COMMENT ON COLUMN tenants.billing_email        IS 'Billing contact email — receives invoices and payment alerts. Falls back to email when NULL.';
COMMENT ON COLUMN tenants.billing_contact_name IS 'Billing contact name printed on invoices. Optional.';
COMMENT ON COLUMN tenants.subdomain            IS 'Vanity subdomain for tenant routing. NULL = use path-based routing. Reserved names blocked by trigger (000057).';
COMMENT ON COLUMN tenants."Status"             IS 'Lifecycle state. New tenants start PENDING until email verification + payment setup complete.';
COMMENT ON COLUMN tenants.plan_tier            IS 'Denormalized subscription tier for FeatureFlag and Settings services. Billing module is source of truth; writes back here via event.';
COMMENT ON COLUMN tenants.last_activity_at     IS 'Last meaningful interaction (login, API call, row change). Updated by trigger unless application already set it in this transaction.';
COMMENT ON COLUMN tenants.timezone             IS 'IANA timezone string (e.g. America/New_York). Used for date-relative operations in the ERP.';
COMMENT ON COLUMN tenants.currency_code        IS 'ISO 4217 three-letter currency code. Default is USD.';
COMMENT ON COLUMN tenants.metadata             IS 'Integration metadata owned by ops/third-party systems: Stripe IDs, webhook URLs, external system refs. NOT managed by Settings module.';
COMMENT ON COLUMN tenants.settings             IS 'Application-owned config managed exclusively by the Settings module (ConfigurationService). Other services must use GetEffectiveConfiguration() — do not write here directly.';
COMMENT ON COLUMN tenants.industry             IS 'Industry vertical for analytics and template selection (e.g. Manufacturing, Retail, Services).';
COMMENT ON COLUMN tenants.company_size         IS 'Approximate headcount band. Used for template selection and capacity planning.';
COMMENT ON COLUMN tenants.tax_id               IS 'Government tax identifier (VAT, EIN, GST etc.). Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.registration_number  IS 'Company registry number. Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.legal_entity_type    IS 'Legal structure (LLC, Ltd, PLC, GmbH etc.). Migrate to Compliance module when built.';
COMMENT ON COLUMN tenants.parent_tenant_id     IS 'Self-reference for enterprise org trees and reseller chains. NULL = root tenant. Max depth 5 (DB-enforced), 3 recommended (UI). Circular refs blocked by trigger (000057).';
COMMENT ON COLUMN tenants.created_by           IS 'UUID of the actor (user or service account) that created this tenant. NULL if seeded by migration.';
COMMENT ON COLUMN tenants.deleted_by           IS 'UUID of the actor that issued the soft-delete. Set alongside deleted_at. NULL when not deleted.';
COMMENT ON COLUMN tenants.created_at           IS 'Row creation timestamp.';
COMMENT ON COLUMN tenants.updated_at           IS 'Last row modification timestamp. Maintained by trigger (migration 000057).';
COMMENT ON COLUMN tenants.deleted_at           IS 'Soft-delete timestamp. NULL = tenant is active. Never hard-DELETE a tenant row.';
