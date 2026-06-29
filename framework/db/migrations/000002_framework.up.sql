-- ============================================================
-- 000002_framework — platform table DDL
-- ============================================================
-- Installs the three shared tables used by the framework's
-- built-in platform modules: naming series, audit log, and
-- custom field definitions.
-- ============================================================

-- ── Naming sequences ─────────────────────────────────────────
-- One row per (tenant, entity). current_seq is atomically
-- incremented by platform/naming.Next via INSERT … ON CONFLICT.

CREATE TABLE IF NOT EXISTS awo_naming_sequences (
    tenant_id   uuid   NOT NULL,
    entity      text   NOT NULL,
    current_seq bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, entity)
);

SELECT apply_tenant_rls('awo_naming_sequences');

-- ── Audit log ────────────────────────────────────────────────
-- Append-only. tenant_id comes from current_tenant_id();
-- INSERTs must not pass it as a parameter.

CREATE TABLE IF NOT EXISTS awo_audit_log (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id  uuid        NOT NULL DEFAULT current_tenant_id(),
    entity     text        NOT NULL,
    record_id  uuid        NOT NULL,
    op         text        NOT NULL,  -- "create" | "update" | "delete"
    actor_id   text        NOT NULL DEFAULT '',
    changes    jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS awo_audit_log_record_idx
    ON awo_audit_log (entity, record_id);

CREATE INDEX IF NOT EXISTS awo_audit_log_tenant_idx
    ON awo_audit_log (tenant_id, created_at DESC);

SELECT apply_tenant_rls('awo_audit_log');

-- ── Custom fields ─────────────────────────────────────────────
-- Stores JSONB field definitions per (tenant, entity).
-- Managed by platform/customfield.Registry.

CREATE TABLE IF NOT EXISTS awo_custom_fields (
    tenant_id  uuid        NOT NULL,
    entity     text        NOT NULL,
    fields     jsonb       NOT NULL DEFAULT '[]',
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, entity)
);

SELECT apply_tenant_rls('awo_custom_fields');
