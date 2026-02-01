-- =====================================================
-- USER ACTIVITIES TABLE FOR ABAC BEHAVIORAL ANALYTICS
-- =====================================================
-- Advanced user activity tracking for behavioral analytics
-- Supports ABAC evaluation with rich security context
-- Partitioned by timestamp for performance at scale
-- Main partitioned table for user activity tracking
CREATE TABLE user_activities (
  id UUID DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  session_id UUID REFERENCES user_sessions(id) ON DELETE
  SET
    NULL,
    -- Activity classification
    activity_type VARCHAR(50) NOT NULL,
    module VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id UUID,
    action_performed VARCHAR(50),
    -- Security context for ABAC evaluation
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    location_data JSONB DEFAULT '{}'::jsonb,
    -- Performance and request metrics
    request_method VARCHAR(10),
    request_path TEXT,
    request_params JSONB DEFAULT '{}'::jsonb,
    response_status INTEGER,
    response_time_ms INTEGER,
    -- Risk assessment data
    risk_indicators JSONB DEFAULT '{}'::jsonb,
    anomaly_score DECIMAL(5, 2) DEFAULT 0.00,
    -- Activity metadata and context
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    additional_data JSONB DEFAULT '{}'::jsonb,
    -- Constraints
    CONSTRAINT user_activities_anomaly_score_range CHECK (
      anomaly_score >= 0.00
      AND anomaly_score <= 100.00
    ),
    -- Composite primary key including partition column
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create initial partitions (last 3 months + next 3 months)
-- Current month partition
-- CREATE TABLE user_activities_current PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE))
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month');
--
-- -- Previous 2 months partitions
-- CREATE TABLE user_activities_prev1 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) - INTERVAL '1 month')
--     TO (date_trunc('month', CURRENT_DATE));
--
-- CREATE TABLE user_activities_prev2 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) - INTERVAL '2 months')
--     TO (date_trunc('month', CURRENT_DATE) - INTERVAL '1 month');
--
-- -- Next 2 months partitions
-- CREATE TABLE user_activities_next1 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '2 months');
--
-- CREATE TABLE user_activities_next2 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) + INTERVAL '2 months')
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '3 months');
--
-- Performance indexes
CREATE INDEX idx_user_activities_user_timestamp ON user_activities (user_id, timestamp DESC);

CREATE INDEX idx_user_activities_tenant_timestamp ON user_activities (tenant_id, timestamp DESC);

CREATE INDEX idx_user_activities_activity_type ON user_activities (activity_type, timestamp DESC);

CREATE INDEX idx_user_activities_session ON user_activities (session_id)
WHERE
  session_id IS NOT NULL;

CREATE INDEX idx_user_activities_resource ON user_activities (resource_type, resource_id)
WHERE
  resource_id IS NOT NULL;

-- JSONB indexes for ABAC attribute queries
CREATE INDEX idx_user_activities_location_data ON user_activities USING GIN (location_data);

CREATE INDEX idx_user_activities_risk_indicators ON user_activities USING GIN (risk_indicators);

CREATE INDEX idx_user_activities_additional_data ON user_activities USING GIN (additional_data);

-- Risk and anomaly detection indexes
CREATE INDEX idx_user_activities_anomaly_score ON user_activities (anomaly_score DESC)
WHERE
  anomaly_score > 0;

CREATE INDEX idx_user_activities_high_risk ON user_activities (user_id, timestamp DESC)
WHERE
  anomaly_score > 50.0;

-- Enable Row Level Security
ALTER TABLE
  user_activities ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only access activities within their tenant
CREATE POLICY user_activities_tenant_isolation ON user_activities FOR ALL TO application_role USING (
  tenant_id = current_setting('app.current_tenant_id')::uuid
);

-- RLS Policy: Users can view their own activities (for self-service features)
CREATE POLICY user_activities_self_access ON user_activities FOR
SELECT
  TO application_role USING (
    user_id = current_setting('app.current_user_id')::uuid
    AND tenant_id = current_setting('app.current_tenant_id')::uuid
  );

-- RLS Policy: Admin bypass - system administrators can access all activities within tenant
CREATE POLICY user_activities_admin_bypass ON user_activities FOR ALL TO admin_role USING (
  tenant_id = current_setting('app.current_tenant_id')::uuid
);

-- Grant permissions
GRANT
SELECT
,
INSERT
,
UPDATE
  ON user_activities TO application_role;

GRANT ALL PRIVILEGES ON user_activities TO admin_role;

-- Table and column comments for documentation
COMMENT ON TABLE user_activities IS 'Partitioned table for user activity tracking and behavioral analytics supporting ABAC evaluation';

COMMENT ON COLUMN user_activities.id IS 'Unique identifier for the activity record';

COMMENT ON COLUMN user_activities.user_id IS 'Reference to the user who performed the activity';

COMMENT ON COLUMN user_activities.tenant_id IS 'Tenant isolation for multi-tenant architecture';

COMMENT ON COLUMN user_activities.session_id IS 'Reference to the user session when activity occurred';

COMMENT ON COLUMN user_activities.activity_type IS 'Classification of the activity (login, access, modification, etc.)';

COMMENT ON COLUMN user_activities.location_data IS 'JSONB containing geographic and network location information';

COMMENT ON COLUMN user_activities.risk_indicators IS 'JSONB containing calculated risk factors for the activity';

COMMENT ON COLUMN user_activities.anomaly_score IS 'Calculated anomaly score from 0.00 to 100.00 for behavioral analysis';

COMMENT ON COLUMN user_activities.additional_data IS 'Flexible JSONB storage for activity-specific metadata';

-- Function to automatically create monthly partitions
CREATE
OR REPLACE FUNCTION create_monthly_user_activities_partition(partition_date DATE) RETURNS TEXT AS
$$
DECLARE
partition_name TEXT;

start_date DATE;

end_date DATE;

BEGIN
-- Generate partition name
partition_name := 'user_activities_' || to_char(partition_date, 'YYYY_MM');

-- Calculate partition boundaries
start_date := date_trunc('month', partition_date)::DATE;

end_date := (
  date_trunc('month', partition_date) + INTERVAL '1 month'
)::DATE;

-- Create partition
EXECUTE format(
  'CREATE TABLE %I PARTITION OF user_activities 
                    FOR VALUES FROM (%L) TO (%L)',
  partition_name,
  start_date,
  end_date
);

RETURN 'Created partition: ' || partition_name;

END;

$$
LANGUAGE plpgsql;

-- Function to drop old partitions (data retention)
CREATE
OR REPLACE FUNCTION drop_old_user_activities_partitions(retention_months INTEGER DEFAULT 12) RETURNS TEXT AS
$$
DECLARE
partition_name TEXT;

cutoff_date DATE;

dropped_partitions TEXT [] := '{}';

partition_record RECORD;

BEGIN
cutoff_date := (
  date_trunc('month', CURRENT_DATE) - (retention_months || ' months')::INTERVAL
)::DATE;

-- Find partitions older than retention period
FOR partition_record IN
SELECT
  schemaname,
  tablename
FROM
  pg_tables
WHERE
  schemaname = 'public'
  AND tablename LIKE 'user_activities_%'
  AND tablename ~ '^user_activities_[0-9]{4}_[0-9]{2}$' LOOP
  -- Extract date from partition name and check if it's old enough
BEGIN
DECLARE
partition_date DATE;

BEGIN
partition_date := to_date(
  substring(
    partition_record.tablename
    FROM
      'user_activities_([0-9]{4}_[0-9]{2})$'
  ),
  'YYYY_MM'
);

IF partition_date < cutoff_date THEN EXECUTE format(
  'DROP TABLE IF EXISTS %I',
  partition_record.tablename
);

dropped_partitions := array_append(dropped_partitions, partition_record.tablename);

END IF;

END;

EXCEPTION
WHEN OTHERS THEN
-- Skip invalid partition names
CONTINUE;

END;

END LOOP;

IF array_length(dropped_partitions, 1) > 0 THEN RETURN 'Dropped partitions: ' || array_to_string(dropped_partitions, ', ');

ELSE RETURN 'No old partitions found to drop';

END IF;

END;

$$
LANGUAGE plpgsql;

-- Grant execute permissions on utility functions
GRANT EXECUTE ON FUNCTION create_monthly_user_activities_partition(DATE) TO application_role;

GRANT EXECUTE ON FUNCTION drop_old_user_activities_partitions(INTEGER) TO application_role;

GRANT EXECUTE ON FUNCTION create_monthly_user_activities_partition(DATE) TO admin_role;

GRANT EXECUTE ON FUNCTION drop_old_user_activities_partitions(INTEGER) TO admin_role;
