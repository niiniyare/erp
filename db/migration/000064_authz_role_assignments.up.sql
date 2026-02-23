CREATE TABLE role_assignments (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  subject      VARCHAR(256) NOT NULL,
  role_name    VARCHAR(100) NOT NULL,
  domain       VARCHAR(256) NOT NULL,
  assigned_by  VARCHAR(256),
  delegated_by VARCHAR(256),
  expires_at   TIMESTAMPTZ,
  is_active    BOOLEAN      DEFAULT TRUE,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT role_assignments_unique UNIQUE (subject, role_name, domain)
);

CREATE INDEX idx_role_assignments_subject ON role_assignments(subject, domain);
CREATE INDEX idx_role_assignments_expires ON role_assignments(expires_at) WHERE expires_at IS NOT NULL;

ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY ra_tenant ON role_assignments FOR ALL TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
CREATE POLICY ra_admin ON role_assignments FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
GRANT SELECT, INSERT, UPDATE, DELETE ON role_assignments TO application_role;
GRANT ALL ON role_assignments TO admin_role;
