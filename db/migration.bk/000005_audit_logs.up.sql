

-- -- Audit logging (enhanced)
-- CREATE TABLE audit_logs (
--     id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     user_id UUID REFERENCES users(id) ON DELETE SET NULL,
--     entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
--     action VARCHAR(50) NOT NULL,
--     resource_type VARCHAR(50), -- What was changed
--     resource_id BIGINT, -- ID of the changed resource
--     old_values JSONB,
--     new_values JSONB,
--     ip_address INET,
--     user_agent TEXT,
--     session_id VARCHAR(255),
--     module VARCHAR(50), -- Which module generated the log
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );
--
-- CREATE TABLE IF NOT EXISTS rls_change_log (
--     id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
--     schema_name TEXT NOT NULL,
--     table_name TEXT NOT NULL,
--     action TEXT NOT NULL,           -- 'apply' or 'remove'
--     policy_name TEXT,
--     policy_type TEXT,
--     command TEXT,
--     dry_run BOOLEAN DEFAULT false,
--     changed_at TIMESTAMPTZ DEFAULT now()
-- );


/*
CREATE OR REPLACE FUNCTION apply_rls_to_tenant_tables(
    schema_name TEXT,
    policy_name TEXT DEFAULT 'tenant_isolation_policy',
    policy_type TEXT DEFAULT 'all',
    force_rls BOOLEAN DEFAULT true
)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    tbl TEXT;
    current_tenant TEXT;
    rls_clause TEXT := 'USING (tenant_id = current_setting(''app.current_tenant'')::INT)';
    commands TEXT[];
BEGIN
    BEGIN
        current_tenant := current_setting('app.current_tenant', true);
        IF current_tenant IS NULL THEN
            RAISE EXCEPTION 'app.current_tenant is not set for this session';
        END IF;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'app.current_tenant setting is missing or inaccessible';
    END;

    FOR tbl IN
        SELECT table_name
        FROM information_schema.columns
        WHERE column_name = 'tenant_id'
          AND table_schema = schema_name
    LOOP
        EXECUTE format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', schema_name, tbl);
        EXECUTE format('DROP POLICY IF EXISTS %I ON %I.%I', policy_name, schema_name, tbl);

        IF policy_type = 'read' THEN
            commands := ARRAY['SELECT'];
            EXECUTE format('CREATE POLICY %I ON %I.%I FOR SELECT %s',
                policy_name, schema_name, tbl, rls_clause);
        ELSIF policy_type = 'write' THEN
            commands := ARRAY['INSERT', 'UPDATE', 'DELETE'];
            EXECUTE format('CREATE POLICY %I ON %I.%I FOR INSERT, UPDATE, DELETE %s',
                policy_name, schema_name, tbl, rls_clause);
        ELSIF policy_type = 'all' THEN
            commands := ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE'];
            EXECUTE format('CREATE POLICY %I ON %I.%I FOR ALL %s',
                policy_name, schema_name, tbl, rls_clause);
        ELSE
            RAISE EXCEPTION 'Invalid policy_type: %, expected read/write/all', policy_type;
        END IF;

        IF force_rls THEN
            EXECUTE format('ALTER TABLE %I.%I FORCE ROW LEVEL SECURITY', schema_name, tbl);
        END IF;

        -- Log each command individually
        FOREACH rls_cmd IN ARRAY commands LOOP
            INSERT INTO rls_change_log(schema_name, table_name, action, policy_name, policy_type, command, dry_run)
            VALUES (schema_name, tbl, 'apply', policy_name, policy_type, rls_cmd, false);
        END LOOP;
    END LOOP;
END;
$$;


CREATE OR REPLACE FUNCTION apply_rls_to_tenant_tables()
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    tbl TEXT;
    rls_policy_name TEXT := 'tenant_isolation_policy';
BEGIN
    FOR tbl IN
        SELECT table_name
        FROM information_schema.columns
        WHERE column_name = 'tenant_id'
          AND table_schema = 'public'
    LOOP
        -- Enable RLS on the table
        EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', tbl);

        -- Drop policy if it already exists
        EXECUTE format('DROP POLICY IF EXISTS %I ON public.%I', rls_policy_name, tbl);

        -- Create RLS policy for tenant isolation
        EXECUTE format($sql$
            CREATE POLICY %I ON public.%I
            USING (tenant_id = current_setting('app.current_tenant')::INT)
        $sql$, rls_policy_name, tbl);
    END LOOP;
END;
$$;
COMMENT ON FUNCTION apply_rls_to_tenant_tables IS '
apply_rls_to_tenant_tables(). This function will apply Row-Level Security (RLS) to all public schema tables that have a tenant_id column, dropping any existing policy named tenant_isolation_policy and recreating it.'

*/
