
-- Audit logging (enhanced)
CREATE TABLE audit_logs (
    uuid UUID PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50), -- What was changed
    resource_id BIGINT, -- ID of the changed resource
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    module VARCHAR(50), -- Which module generated the log
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

