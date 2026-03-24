-- ------------------------------------------------------------------------------------------------
-- ACCESS_REQUESTS
-- ------------------------------------------------------------------------------------------------
-- Approval workflow for access requests with business justification and lifecycle management.
-- Covers role assignments, permission grants, resource access, and privilege elevations.
-- request_type IN ('ROLE_ASSIGNMENT','PERMISSION_GRANT','RESOURCE_ACCESS','ELEVATION').
-- approval_status IN ('PENDING','APPROVED','REJECTED','EXPIRED','REVOKED').
--
-- NOTE: FKs reference tenants (000001), users (000303), entities (000010), roles (000405),
--       permissions and resources defined in platform-IAM migrations.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS access_requests (
  id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  requester_id     UUID          NOT NULL REFERENCES users(id),
  target_user_id   UUID          REFERENCES users(id),           -- User receiving access if different from requester
  entity_id        UUID          NOT NULL REFERENCES entities(uuid),
  request_type     VARCHAR(20)   NOT NULL CHECK (
    request_type IN (
      'ROLE_ASSIGNMENT',
      'PERMISSION_GRANT',
      'RESOURCE_ACCESS',
      'ELEVATION'
    )
  ),
  role_id          UUID          REFERENCES roles(id),
  permission_id    UUID          REFERENCES permissions(id),
  resource_id      UUID          REFERENCES resources(id),
  justification    TEXT          NOT NULL,                        -- Required business justification
  business_reason  VARCHAR(500),
  duration_hours   INTEGER,                                       -- Requested duration for temporary access
  approval_status  VARCHAR(20)   DEFAULT 'PENDING' CHECK (
    approval_status IN (
      'PENDING',
      'APPROVED',
      'REJECTED',
      'EXPIRED',
      'REVOKED'
    )
  ),
  approved_by      UUID          REFERENCES users(id),
  approved_at      TIMESTAMPTZ,
  approval_comments TEXT,
  expires_at       TIMESTAMPTZ,
  auto_revoke      BOOLEAN       DEFAULT TRUE,                    -- Auto-revoke access when expires_at is reached
  created_at       TIMESTAMPTZ   DEFAULT NOW(),
  updated_at       TIMESTAMPTZ   DEFAULT NOW()
);

COMMENT ON TABLE access_requests IS 'Access request approval workflow with business justification, lifecycle management, and automatic revocation for governance and compliance.';

COMMENT ON COLUMN access_requests.request_type IS 'Type of access request: ROLE_ASSIGNMENT, PERMISSION_GRANT, RESOURCE_ACCESS, ELEVATION';
COMMENT ON COLUMN access_requests.target_user_id IS 'User receiving the access (if different from requester)';
COMMENT ON COLUMN access_requests.duration_hours IS 'Requested access duration in hours for temporary access';
COMMENT ON COLUMN access_requests.auto_revoke IS 'Whether to automatically revoke access when it expires';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE access_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY access_requests_tenant_isolation ON access_requests FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
