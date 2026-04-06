-- ------------------------------------------------------------------------------------------------
-- POLICIES TABLE
-- ------------------------------------------------------------------------------------------------
-- ABAC policies with advanced rule engine supporting multiple policy types, priorities,
-- obligations, and compliance tracking.
-- policy_type IN ('ABAC','RBAC','HYBRID','TIME_BASED','LOCATION_BASED').
-- category IN ('ACCESS','DATA_FILTER','FIELD_MASK','AUDIT','COMPLIANCE').
-- effect IN ('ALLOW','DENY'): higher-priority DENY overrides ALLOW.
--
-- NOTE: Depends on tenants(id), entities(uuid), and users(id). RLS uses current_tenant_id().
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policies (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id    UUID         REFERENCES entities(uuid),   -- Entity scope for policy
  name         VARCHAR(150) NOT NULL,
  display_name VARCHAR(200),
  description  TEXT,
  policy_type  VARCHAR(50)  DEFAULT 'ABAC' CHECK (
    policy_type IN ('ABAC', 'RBAC', 'HYBRID', 'TIME_BASED', 'LOCATION_BASED')
  ),
  effect       VARCHAR(5)   DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
  priority     INTEGER      DEFAULT 100,                 -- Higher numbers = higher priority
  category     VARCHAR(50)  DEFAULT 'ACCESS' CHECK (
    category IN ('ACCESS', 'DATA_FILTER', 'FIELD_MASK', 'AUDIT', 'COMPLIANCE')
  ),
  target       JSONB        NOT NULL,                    -- When policy applies (conditions)
  rule         JSONB        NOT NULL,                    -- Policy logic/evaluation rules
  obligations  JSONB        DEFAULT '{}'::jsonb,         -- Required actions when policy fires
  advice       JSONB        DEFAULT '{}'::jsonb,         -- Optional actions/logging recommendations
  is_active    BOOLEAN      DEFAULT TRUE,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  DEFAULT NOW(),
  created_by   UUID         REFERENCES users(id),
  deleted_at   TIMESTAMPTZ,
  CONSTRAINT policies_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE policies IS 'ABAC policies with advanced rule engine supporting multiple policy types, priorities, obligations, and compliance tracking.';

COMMENT ON COLUMN policies.policy_type  IS 'Policy type: ABAC (attribute-based), RBAC (role-based), HYBRID (combined), TIME_BASED (temporal), LOCATION_BASED (geographic)';
COMMENT ON COLUMN policies.priority     IS 'Policy priority for conflict resolution (higher numbers processed first)';
COMMENT ON COLUMN policies.category     IS 'Policy category: ACCESS (authorization), DATA_FILTER (row-level), FIELD_MASK (column-level), AUDIT (logging), COMPLIANCE (regulatory)';
COMMENT ON COLUMN policies.target       IS 'JSONB defining when policy applies (subjects, resources, actions, conditions)';
COMMENT ON COLUMN policies.rule         IS 'JSONB containing policy evaluation logic and conditions';
COMMENT ON COLUMN policies.obligations  IS 'JSONB defining required actions when policy fires (logging, notifications, etc.)';
COMMENT ON COLUMN policies.advice       IS 'JSONB defining optional actions and recommendations';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY policies_tenant_isolation ON policies FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY policies_admin ON policies FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
