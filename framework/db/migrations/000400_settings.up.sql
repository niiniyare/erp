-- ============================================================
-- 000400_settings — per-tenant key/value settings store
-- ============================================================

CREATE TABLE IF NOT EXISTS settings (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id  uuid        NOT NULL DEFAULT current_tenant_id(),
    key        text        NOT NULL,
    value      jsonb       NOT NULL DEFAULT 'null',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, key)
);

SELECT apply_tenant_rls('settings');
