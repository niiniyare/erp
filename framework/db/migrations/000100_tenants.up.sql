-- ============================================================
-- 000100_tenants — global tenants table
-- ============================================================

CREATE TABLE IF NOT EXISTS tenants (
    id           uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name         text        NOT NULL,
    slug         text        NOT NULL UNIQUE,
    status       text        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending','active','suspended','archived')),
    plan         text        NOT NULL DEFAULT 'free',
    company_size text,
    country      text,
    currency     text        NOT NULL DEFAULT 'KES',
    timezone     text        NOT NULL DEFAULT 'Africa/Nairobi',
    locale       text        NOT NULL DEFAULT 'en',
    metadata     jsonb       NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS tenants_slug_idx ON tenants (slug);
CREATE INDEX IF NOT EXISTS tenants_status_idx ON tenants (status);

SELECT apply_global_rls('tenants');
