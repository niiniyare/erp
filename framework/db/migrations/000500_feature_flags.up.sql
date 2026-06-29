-- ============================================================
-- 000500_feature_flags — per-tenant + global feature flags
-- ============================================================

CREATE TABLE IF NOT EXISTS feature_flags (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id   uuid,  -- NULL = global flag
    key         text        NOT NULL,
    enabled     boolean     NOT NULL DEFAULT false,
    rollout_pct smallint    NOT NULL DEFAULT 100
                            CHECK (rollout_pct BETWEEN 0 AND 100),
    metadata    jsonb       NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, key)
);

CREATE INDEX IF NOT EXISTS feature_flags_key_idx ON feature_flags (key);

-- Global flags (tenant_id IS NULL) are readable by all.
-- Tenant-scoped flags use RLS.
SELECT apply_global_rls('feature_flags');
