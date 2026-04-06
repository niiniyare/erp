-- ------------------------------------------------------------------------------------------------
-- CASBIN_RULE
-- ------------------------------------------------------------------------------------------------
-- Stores raw Casbin policy and role-inheritance rows used by the enforcer at request time.
-- ptype distinguishes row kinds: 'p' = policy rule, 'g' = role-inheritance (grouping) rule.
-- v0–v5 are positional fields whose meaning depends on the Casbin model (sub, dom, obj, act, …).
--
-- NOTE: Tenant-level isolation is NOT enforced via RLS here; the domain value carried in v1
--       enforces multi-tenant isolation entirely at the application layer.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE casbin_rule (
  id    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  ptype VARCHAR(10)  NOT NULL,           -- 'p' (policy) or 'g' (grouping/role-inheritance)
  v0    VARCHAR(256) NOT NULL DEFAULT '', -- subject (user/role)
  v1    VARCHAR(256) NOT NULL DEFAULT '', -- domain (tenant scope)
  v2    VARCHAR(256) NOT NULL DEFAULT '', -- object (resource)
  v3    VARCHAR(256) NOT NULL DEFAULT '', -- action
  v4    VARCHAR(256) NOT NULL DEFAULT '', -- effect or extra field
  v5    VARCHAR(256) NOT NULL DEFAULT ''  -- extra field
);

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE UNIQUE INDEX idx_casbin_rule_unique
  ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5); -- deduplicate: Casbin assumes unique rows
CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype); -- fast filter by row kind

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE casbin_rule ENABLE ROW LEVEL SECURITY;
CREATE POLICY casbin_rule_app   ON casbin_rule FOR ALL TO application_role USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY casbin_rule_admin ON casbin_rule FOR ALL TO admin_role        USING (TRUE) WITH CHECK (TRUE);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON casbin_rule TO application_role;
GRANT ALL ON casbin_rule TO admin_role;
