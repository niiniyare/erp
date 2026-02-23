-- =============================================================================
-- MIGRATION 001 DOWN: Extensions and Roles
-- =============================================================================
-- WARNING: Dropping roles will fail if they own objects or have active grants.
-- Run this only after all dependent migrations have been rolled back.
--
-- Dropping pg_trgm or pgcrypto will fail if any indexes or functions still
-- depend on them. The ordering of the dependent migrations' DOWN files must
-- be run first (009 → 008 → ... → 002) before this file.
-- =============================================================================

-- Drop roles only if they exist and have no dependent objects.
-- In production it is safer to revoke all grants first (done in 009.down)
-- then drop here.
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'readonly_role') THEN
    DROP ROLE readonly_role;
  END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'admin_role') THEN
    DROP ROLE admin_role;
  END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'application_role') THEN
    DROP ROLE application_role;
  END IF;
END $$;

-- Extensions: only drop if no other migrations depend on them.
-- In most deployments these should be left in place — other schemas
-- in the same database cluster may rely on them.
-- Uncomment deliberately if this is a full teardown:
-- DROP EXTENSION IF EXISTS pg_trgm;
-- DROP EXTENSION IF EXISTS pgcrypto;
