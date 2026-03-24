-- ------------------------------------------------------------------------------------------------
-- TENANTS TABLE — INDEXES
-- ------------------------------------------------------------------------------------------------
-- Indexes are maintained separately from the table DDL so that index strategy
-- can evolve independently of the table structure, and CONCURRENTLY builds can
-- be added for live deployments without blocking a monolithic migration.
-- Note: CONCURRENTLY cannot run inside a transaction block — use psql directly
-- for those builds in production.
--
-- Index naming convention: idx_{table}_{columns}[_{qualifier}]
--   qualifier = 'active'  for partial indexes on deleted_at IS NULL
--               'gin'     for GIN indexes on JSONB
--               'trgm'    for trigram text search indexes
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- STANDARD B-TREE INDEXES
-- ------------------------------------------------------------------------------------------------

-- slug: primary lookup key in URL routing and API path resolution.
-- The UNIQUE constraint already provides an implicit index but naming it
-- explicitly makes query plans easier to read in EXPLAIN output.
CREATE INDEX IF NOT EXISTS idx_tenants_slug
  ON tenants(slug);

-- Status: filtered heavily in tenant admin dashboards (show active, show suspended).
CREATE INDEX IF NOT EXISTS idx_tenants_status
  ON tenants("Status");

-- plan_tier: queried by FeatureFlag and Settings services to resolve
-- tier-gated configuration values. Expect constant use in the hot path.
CREATE INDEX IF NOT EXISTS idx_tenants_plan_tier
  ON tenants(plan_tier);

-- email: standard auth lookup path (login by email, admin tenant search).
CREATE INDEX IF NOT EXISTS idx_tenants_email
  ON tenants(email);

-- billing_email: queried by the Billing module when sending invoices and
-- by accounts-receivable tooling. Separate from email to allow independent lookups.
CREATE INDEX IF NOT EXISTS idx_tenants_billing_email
  ON tenants(billing_email)
  WHERE billing_email IS NOT NULL;

-- subdomain: used on every inbound HTTP request for vanity subdomain routing.
-- Partial index on non-NULL only — most tenants use path-based routing.
CREATE INDEX IF NOT EXISTS idx_tenants_subdomain
  ON tenants(subdomain)
  WHERE subdomain IS NOT NULL;

-- parent_tenant_id: queried when listing child tenants for an enterprise
-- parent or traversing the reseller hierarchy.
CREATE INDEX IF NOT EXISTS idx_tenants_parent
  ON tenants(parent_tenant_id)
  WHERE parent_tenant_id IS NOT NULL;

-- last_activity_at: used in churn analysis queries and by the job that
-- auto-archives dormant tenants after a configurable inactivity window.
CREATE INDEX IF NOT EXISTS idx_tenants_last_activity
  ON tenants(last_activity_at DESC);

-- deleted_at: queried by the hard-delete retention job that purges rows
-- past the compliance retention window. Partial index on deleted rows only.
CREATE INDEX IF NOT EXISTS idx_tenants_deleted
  ON tenants(deleted_at)
  WHERE deleted_at IS NOT NULL;

-- ------------------------------------------------------------------------------------------------
-- COMPOSITE PARTIAL INDEX — active tenant fast path
-- ------------------------------------------------------------------------------------------------
-- The three most frequent application-layer lookups are by id, slug, and email.
-- A single composite partial index on active rows (deleted_at IS NULL) covers
-- all three in one index scan without touching deleted-row bloat. This is the
-- hot path for every authenticated request.
-- The partial predicate (deleted_at IS NULL) keeps the index compact by
-- excluding deleted rows entirely.
-- ------------------------------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_tenants_active
  ON tenants(id, slug, email)
  WHERE deleted_at IS NULL;

-- ------------------------------------------------------------------------------------------------
-- GIN INDEXES — JSONB columns
-- ------------------------------------------------------------------------------------------------
-- Both metadata and settings are queried with JSONB containment operators
-- (@>, ?, ?|, ?&) by integration services and the Settings module.
-- Without GIN indexes these are full table scans.
-- jsonb_path_ops operator class: smaller, faster indexes for @> containment
-- queries. Tradeoff: does not support ? / ?| / ?& key-existence queries.
-- If key-existence queries are needed, switch to jsonb_ops (larger index).
-- ------------------------------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_tenants_metadata_gin
  ON tenants USING gin(metadata jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_tenants_settings_gin
  ON tenants USING gin(settings jsonb_path_ops);

-- ------------------------------------------------------------------------------------------------
-- TRIGRAM INDEXES — text search
-- ------------------------------------------------------------------------------------------------
-- Admin search boxes do ILIKE '%term%' queries on name and email.
-- Without trigram indexes, ILIKE with a leading wildcard forces a sequential scan.
-- pg_trgm GIN indexes support both ILIKE and SIMILAR TO patterns efficiently.
-- Requires the pg_trgm extension (created in migration 000051).
-- ------------------------------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_tenants_name_trgm
  ON tenants USING gin(name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_tenants_email_trgm
  ON tenants USING gin(email gin_trgm_ops);
