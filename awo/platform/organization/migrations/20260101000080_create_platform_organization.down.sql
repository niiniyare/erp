-- Reverse: drop platform_organization and all its indexes + policies.
-- WARNING: This destroys all org hierarchy data. Ensure a backup exists before
-- running down migrations in any environment with real tenant data.

DROP POLICY IF EXISTS tenant_isolation ON platform_organization;
DROP TABLE IF EXISTS platform_organization;
