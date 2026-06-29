-- ============================================================
-- 000700_reporting — saved reports and async run tracking
-- ============================================================

CREATE TABLE IF NOT EXISTS reports (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id  uuid        NOT NULL DEFAULT current_tenant_id(),
    name       text        NOT NULL,
    entity     text        NOT NULL,
    filters    jsonb       NOT NULL DEFAULT '{}',
    columns    jsonb       NOT NULL DEFAULT '[]',
    created_by text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS report_runs (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id   uuid        NOT NULL DEFAULT current_tenant_id(),
    report_id   uuid        NOT NULL REFERENCES reports (id) ON DELETE CASCADE,
    status      text        NOT NULL DEFAULT 'running'
                            CHECK (status IN ('running','done','failed')),
    started_by  text        NOT NULL DEFAULT '',
    result_url  text,
    error_msg   text,
    started_at  timestamptz NOT NULL DEFAULT NOW(),
    finished_at timestamptz
);

CREATE INDEX IF NOT EXISTS reports_tenant_idx    ON reports (tenant_id, name);
CREATE INDEX IF NOT EXISTS report_runs_report_idx ON report_runs (report_id, started_at DESC);

SELECT apply_tenant_rls('reports');
SELECT apply_tenant_rls('report_runs');
