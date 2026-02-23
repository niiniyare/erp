-- =============================================================================
-- MIGRATION 001 UP: Extensions and Roles
-- =============================================================================
-- Architecture Decision (ADR-001):
--   Extensions and roles are created first as foundational infrastructure.
--   Roles are created idempotently using DO blocks rather than IF NOT EXISTS
--   (which is only available for roles in PG 9.x). This makes the migration
--   safe to re-run in CI and staging pipelines without manual cleanup.
--
-- Roles defined here:
--   application_role  — runtime role for the web application / API servers.
--                       Governed by Row Level Security. Sees only its own tenant.
--   admin_role        — internal tooling, ops dashboards, migration runners.
--                       Bypasses RLS via policy; never granted to end users.
--   readonly_role     — read replicas, reporting services, data exports.
--                       SELECT-only on active (non-deleted) rows.
--
-- NEVER grant admin_role to a long-lived application credential.
-- =============================================================================

-- pgcrypto provides gen_random_uuid() for UUID primary keys.
-- uuid-ossp is deliberately NOT used here; pgcrypto is sufficient and
-- avoids the extra extension dependency.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- pg_trgm enables GIN trigram indexes used for fast ILIKE searches on
-- tenant name, slug, and email fields (added in 004).
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- -------------------------------------------------------------------------
-- ROLES
-- -------------------------------------------------------------------------
-- Idempotent role creation via DO blocks — safe for re-runs.
-- Production deployments should ensure these roles exist in the cluster
-- before running this migration (e.g., via Terraform/Pulumi provisioning).
-- -------------------------------------------------------------------------

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'application_role') THEN
    CREATE ROLE application_role;
  END IF;
END $$;

COMMENT ON ROLE application_role IS
  'Runtime role for API servers and background workers. '
  'Governed by RLS — sees only its own tenant row. '
  'NEVER grant SUPERUSER or BYPASSRLS to this role.';

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'admin_role') THEN
    CREATE ROLE admin_role;
  END IF;
END $$;

COMMENT ON ROLE admin_role IS
  'Internal operations role for migration runners and admin tooling. '
  'Bypasses RLS via explicit policies — restrict credential issuance tightly. '
  'Do not embed in application connection pools.';

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'readonly_role') THEN
    CREATE ROLE readonly_role;
  END IF;
END $$;

COMMENT ON ROLE readonly_role IS
  'Read-only role for reporting replicas and data export jobs. '
  'May SELECT active (non-deleted) rows only. '
  'Explicitly excluded from INSERT / UPDATE / DELETE.';
