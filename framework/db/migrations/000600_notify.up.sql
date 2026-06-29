-- ============================================================
-- 000600_notify — outbound notifications queue
-- ============================================================

CREATE TABLE IF NOT EXISTS notifications (
    id           uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id    uuid        NOT NULL DEFAULT current_tenant_id(),
    recipient_id text        NOT NULL,
    channel      text        NOT NULL CHECK (channel IN ('email','sms','push','webhook')),
    subject      text        NOT NULL DEFAULT '',
    body         text        NOT NULL DEFAULT '',
    status       text        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending','sent','failed')),
    error_msg    text,
    sent_at      timestamptz,
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS notifications_tenant_status_idx
    ON notifications (tenant_id, status, created_at);

SELECT apply_tenant_rls('notifications');
