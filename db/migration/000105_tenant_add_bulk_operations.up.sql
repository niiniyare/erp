-- ------------------------------------------------------------------------------------------------
-- TENANT_BULK_OPERATIONS
-- ------------------------------------------------------------------------------------------------
-- Master table tracking bulk tenant management operations (mass suspend, reactivate, archive,
-- and configuration updates) with full progress monitoring and audit trail.
-- operation_type IN ('SUSPEND','REACTIVATE','ARCHIVE','UPDATE_LIMITS','UPDATE_FEATURES').
-- status IN ('IN_PROGRESS','COMPLETED','FAILED','PARTIAL_SUCCESS','CANCELLED').
--
-- NOTE: Depends on tenants(id) FK and the update_updated_at_column() trigger function
--       defined in earlier platform migrations.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE tenant_bulk_operations (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  operation_type    VARCHAR(50)  NOT NULL CHECK (operation_type IN (
                                   'SUSPEND',         -- Mass suspend tenants
                                   'REACTIVATE',      -- Mass reactivate suspended tenants
                                   'ARCHIVE',         -- Mass archive tenants with retention policies
                                   'UPDATE_LIMITS',   -- Mass update tenant resource limits
                                   'UPDATE_FEATURES'  -- Mass enable/disable tenant features
                                 )),
  actor_id          UUID         NOT NULL,             -- User ID who initiated the operation
  actor_name        VARCHAR(255),                      -- User name for audit display
  total_tenants     INT          NOT NULL DEFAULT 0,   -- Total tenants targeted by this operation
  successful_count  INT          NOT NULL DEFAULT 0,   -- Successfully processed tenant count
  failed_count      INT          NOT NULL DEFAULT 0,   -- Failed tenant operation count
  status            VARCHAR(20)  NOT NULL DEFAULT 'IN_PROGRESS' CHECK (status IN (
                                   'IN_PROGRESS',     -- Operation is currently running
                                   'COMPLETED',       -- All operations completed successfully
                                   'FAILED',          -- All operations failed
                                   'PARTIAL_SUCCESS', -- Some succeeded, some failed
                                   'CANCELLED'        -- Operation was cancelled by user
                                 )),
  started_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  completed_at      TIMESTAMPTZ,                       -- Set when the operation finishes
  parameters        JSONB        DEFAULT '{}'::jsonb,  -- Operation-specific data (reason, limits, etc.)
  error_summary     TEXT,                              -- High-level error description if applicable
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  tenant_bulk_operations                IS 'Master table tracking bulk tenant management operations with progress monitoring and audit trail.';
COMMENT ON COLUMN tenant_bulk_operations.operation_type IS 'Type of bulk operation: SUSPEND, REACTIVATE, ARCHIVE, UPDATE_LIMITS, UPDATE_FEATURES.';
COMMENT ON COLUMN tenant_bulk_operations.actor_id       IS 'UUID of the administrator who initiated the bulk operation.';
COMMENT ON COLUMN tenant_bulk_operations.parameters     IS 'JSON parameters specific to operation type (reason, limits, features, retention policies, etc.).';
COMMENT ON COLUMN tenant_bulk_operations.error_summary  IS 'High-level summary of errors if the operation had failures.';

-- ------------------------------------------------------------------------------------------------
-- TENANT_BULK_OPERATION_RESULTS
-- ------------------------------------------------------------------------------------------------
-- Tracks the individual result for each tenant within a bulk operation.
-- status IN ('PENDING','PROCESSING','COMPLETED','FAILED','SKIPPED').
--
-- NOTE: Composite PK on (operation_id, tenant_id). Cascades on delete from both parent tables.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE tenant_bulk_operation_results (
  operation_id  UUID         NOT NULL REFERENCES tenant_bulk_operations(id) ON DELETE CASCADE,
  tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  status        VARCHAR(20)  NOT NULL DEFAULT 'PENDING' CHECK (status IN (
                               'PENDING',    -- Waiting to be processed
                               'PROCESSING', -- Currently being processed
                               'COMPLETED',  -- Successfully completed
                               'FAILED',     -- Operation failed for this tenant
                               'SKIPPED'     -- Skipped (e.g., tenant already in target state)
                             )),
  message       TEXT,                      -- Success or informational message
  error_details TEXT,                      -- Detailed error information if failed
  started_at    TIMESTAMPTZ,               -- Per-tenant processing start time
  completed_at  TIMESTAMPTZ,               -- Per-tenant processing end time
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (operation_id, tenant_id)
);

COMMENT ON TABLE  tenant_bulk_operation_results              IS 'Individual operation results for each tenant within a bulk operation.';
COMMENT ON COLUMN tenant_bulk_operation_results.status       IS 'Individual tenant operation status: PENDING, PROCESSING, COMPLETED, FAILED, SKIPPED.';
COMMENT ON COLUMN tenant_bulk_operation_results.error_details IS 'Detailed error information specific to this tenant if the operation failed.';
COMMENT ON COLUMN tenant_bulk_operation_results.message      IS 'Success message or additional context for this tenant operation.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_tenant_bulk_operations_actor      ON tenant_bulk_operations(actor_id);            -- Look up operations by initiating user
CREATE INDEX idx_tenant_bulk_operations_status     ON tenant_bulk_operations(status);              -- Filter by overall operation status
CREATE INDEX idx_tenant_bulk_operations_type       ON tenant_bulk_operations(operation_type);       -- Filter by operation type
CREATE INDEX idx_tenant_bulk_operations_created_at ON tenant_bulk_operations(created_at);           -- Chronological ordering and range scans

CREATE INDEX idx_tenant_bulk_operation_results_tenant ON tenant_bulk_operation_results(tenant_id); -- Find all operations affecting a tenant
CREATE INDEX idx_tenant_bulk_operation_results_status ON tenant_bulk_operation_results(status);    -- Filter results by processing status

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE tenant_bulk_operations        ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_bulk_operation_results ENABLE ROW LEVEL SECURITY;

-- Admin role: full access to all bulk operations
CREATE POLICY tenant_bulk_operations_admin_policy
  ON tenant_bulk_operations
  FOR ALL TO admin_role
  USING (true);

CREATE POLICY tenant_bulk_operation_results_admin_policy
  ON tenant_bulk_operation_results
  FOR ALL TO admin_role
  USING (true);

-- Application role: full access for API operations
CREATE POLICY tenant_bulk_operations_app_policy
  ON tenant_bulk_operations
  FOR ALL TO application_role
  USING (true);

CREATE POLICY tenant_bulk_operation_results_app_policy
  ON tenant_bulk_operation_results
  FOR ALL TO application_role
  USING (true);

-- Readonly role: select-only for monitoring and reporting
CREATE POLICY tenant_bulk_operations_readonly_policy
  ON tenant_bulk_operations
  FOR SELECT TO readonly_role
  USING (true);

CREATE POLICY tenant_bulk_operation_results_readonly_policy
  ON tenant_bulk_operation_results
  FOR SELECT TO readonly_role
  USING (true);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operations        TO admin_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operation_results TO admin_role;

GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operations        TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operation_results TO application_role;

GRANT SELECT ON tenant_bulk_operations        TO readonly_role;
GRANT SELECT ON tenant_bulk_operation_results TO readonly_role;

-- ------------------------------------------------------------------------------------------------
-- TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_tenant_bulk_operations_updated_at
  BEFORE UPDATE ON tenant_bulk_operations
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_bulk_operation_results_updated_at
  BEFORE UPDATE ON tenant_bulk_operation_results
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- GET_BULK_OPERATION_SUMMARY
-- ------------------------------------------------------------------------------------------------
-- Returns comprehensive real-time summary statistics for a bulk operation.
-- Joins the master record with individual results to compute in-progress count and
-- elapsed/total duration without requiring the caller to aggregate manually.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION get_bulk_operation_summary(p_operation_id UUID)
RETURNS TABLE (
  operation_id      UUID,
  operation_type    VARCHAR(50),
  status            VARCHAR(20),
  total_tenants     INT,
  successful_count  INT,
  failed_count      INT,
  in_progress_count INT,
  duration_seconds  INT
) AS $$
BEGIN
  RETURN QUERY
  SELECT
    bo.id,
    bo.operation_type,
    bo.status,
    bo.total_tenants,
    bo.successful_count,
    bo.failed_count,
    -- Count in-progress items dynamically from results table
    COUNT(CASE WHEN br.status IN ('PENDING', 'PROCESSING') THEN 1 END)::INT AS in_progress_count,
    -- Completed operations use actual duration; in-progress use current time
    CASE
      WHEN bo.completed_at IS NOT NULL
      THEN EXTRACT(EPOCH FROM (bo.completed_at - bo.started_at))::INT
      ELSE EXTRACT(EPOCH FROM (NOW() - bo.started_at))::INT
    END AS duration_seconds
  FROM tenant_bulk_operations bo
  LEFT JOIN tenant_bulk_operation_results br ON bo.id = br.operation_id
  WHERE bo.id = p_operation_id
  GROUP BY bo.id, bo.operation_type, bo.status, bo.total_tenants,
           bo.successful_count, bo.failed_count, bo.started_at, bo.completed_at;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_bulk_operation_summary(UUID) IS 'Returns comprehensive real-time summary statistics for a bulk operation including progress and timing.';

GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO readonly_role;

-- ------------------------------------------------------------------------------------------------
-- UPDATE_BULK_OPERATION_COUNTS
-- ------------------------------------------------------------------------------------------------
-- Recalculates success/failure counts and derives the overall status for a bulk operation.
-- Status derivation: IN_PROGRESS if any items still pending/processing; COMPLETED if all
-- succeeded; FAILED if all failed; PARTIAL_SUCCESS if mixed results with none in progress.
-- Sets completed_at the first time all items finish.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_bulk_operation_counts(p_operation_id UUID)
RETURNS VOID AS $$
DECLARE
  v_successful  INT; -- Count of successful operations
  v_failed      INT; -- Count of failed operations
  v_total       INT; -- Total operations
  v_in_progress INT; -- Count of pending/processing operations
BEGIN
  SELECT
    COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END),
    COUNT(CASE WHEN status = 'FAILED'    THEN 1 END),
    COUNT(*),
    COUNT(CASE WHEN status IN ('PENDING', 'PROCESSING') THEN 1 END)
  INTO v_successful, v_failed, v_total, v_in_progress
  FROM tenant_bulk_operation_results
  WHERE operation_id = p_operation_id;

  UPDATE tenant_bulk_operations SET
    successful_count = v_successful,
    failed_count     = v_failed,
    status = CASE
      WHEN v_in_progress > 0 THEN 'IN_PROGRESS'    -- Still processing
      WHEN v_failed = 0      THEN 'COMPLETED'       -- All succeeded
      WHEN v_successful = 0  THEN 'FAILED'          -- All failed
      ELSE 'PARTIAL_SUCCESS'                        -- Mixed results
    END,
    -- Set completion timestamp the first time all items finish
    completed_at = CASE
      WHEN v_in_progress = 0 AND completed_at IS NULL THEN NOW()
      ELSE completed_at
    END
  WHERE id = p_operation_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_bulk_operation_counts(UUID) IS 'Automatically updates bulk operation counts and status based on individual result statuses.';

GRANT EXECUTE ON FUNCTION update_bulk_operation_counts(UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION update_bulk_operation_counts(UUID) TO application_role;

-- ------------------------------------------------------------------------------------------------
-- TRIGGER_UPDATE_BULK_OPERATION_COUNTS
-- ------------------------------------------------------------------------------------------------
-- Trigger function that automatically keeps bulk operation counts in sync whenever an
-- individual result row is inserted, updated, or deleted.
--
-- NOTE: Handles all three DML events by coalescing NEW/OLD to find the affected operation_id.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION trigger_update_bulk_operation_counts()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM update_bulk_operation_counts(COALESCE(NEW.operation_id, OLD.operation_id));
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenant_bulk_operation_results_update_counts
  AFTER INSERT OR UPDATE OR DELETE ON tenant_bulk_operation_results
  FOR EACH ROW EXECUTE FUNCTION trigger_update_bulk_operation_counts();

COMMENT ON TRIGGER tenant_bulk_operation_results_update_counts ON tenant_bulk_operation_results
  IS 'Automatically updates bulk operation counts and status whenever individual results change.';
