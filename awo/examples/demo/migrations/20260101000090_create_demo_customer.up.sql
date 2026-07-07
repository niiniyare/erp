-- demo_customer: minimal customer entity for framework end-to-end validation.
-- This table exercises: RLS, tenant_id FK, org_id FK, unique constraint,
-- GIN trigram index, active flag, timestamps, custom_fields JSONB.

CREATE TABLE IF NOT EXISTS demo_customer (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES platform_tenant(id),
    org_id          UUID NOT NULL REFERENCES platform_organization(id),
    customer_code   VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(254),
    phone           VARCHAR(30),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    custom_fields   JSONB NOT NULL DEFAULT '{}',

    CONSTRAINT demo_customer_code_unique UNIQUE (tenant_id, customer_code),
    CONSTRAINT demo_customer_email_unique UNIQUE (tenant_id, email)
);

-- Tenant-scoped code lookup.
CREATE INDEX IF NOT EXISTS idx_demo_customer_tenant_code
    ON demo_customer (tenant_id, customer_code);

-- Trigram index for full-text search on name.
CREATE INDEX IF NOT EXISTS idx_demo_customer_name_trgm
    ON demo_customer USING GIN (name gin_trgm_ops);

-- Trigram index for full-text search on customer_code.
CREATE INDEX IF NOT EXISTS idx_demo_customer_code_trgm
    ON demo_customer USING GIN (customer_code gin_trgm_ops);

-- Org-scoped queries (Stage 2 org scope — no RLS, explicit IN predicate).
CREATE INDEX IF NOT EXISTS idx_demo_customer_tenant_org
    ON demo_customer (tenant_id, org_id);

-- Active filter.
CREATE INDEX IF NOT EXISTS idx_demo_customer_active
    ON demo_customer (tenant_id, active);

-- GIN for custom_fields.
CREATE INDEX IF NOT EXISTS idx_demo_customer_custom_fields
    ON demo_customer USING GIN (custom_fields);

-- Stage 1: standard tenant RLS.
ALTER TABLE demo_customer ENABLE ROW LEVEL SECURITY;
ALTER TABLE demo_customer FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON demo_customer
    USING (tenant_id = current_tenant_id());
