WITH tables AS (
    SELECT 
        c.table_schema,
        c.table_name,
        json_agg(json_build_object(
            'column_name', c.column_name,
            'data_type', c.data_type,
            'is_nullable', c.is_nullable
        )) AS columns
    FROM information_schema.columns c
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema')
    GROUP BY c.table_schema, c.table_name
),

tenant_tables AS (
    -- Only include tables that have tenant_id column (multi-tenant safe)
    SELECT t.*
    FROM tables t
    JOIN information_schema.columns c
      ON t.table_name = c.table_name
     AND t.table_schema = c.table_schema
    WHERE c.column_name = 'tenant_id'
),

foreign_keys AS (
    SELECT json_agg(json_build_object(
        'source_table', tc.table_name,
        'source_column', kcu.column_name,
        'target_table', ccu.table_name,
        'target_column', ccu.column_name
    )) AS fk_json
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage ccu
      ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
),

views AS (
    SELECT json_agg(json_build_object(
        'schema', table_schema,
        'name', table_name,
        'definition', view_definition
    )) AS views_json
    FROM information_schema.views
    WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
),

functions AS (
    SELECT json_agg(json_build_object(
        'schema', n.nspname,
        'name', p.proname,
        'args', pg_get_function_arguments(p.oid),
        'return_type', pg_get_function_result(p.oid)
    )) AS fn_json
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
),

triggers AS (
    SELECT json_agg(json_build_object(
        'table', event_object_table,
        'name', trigger_name,
        'event', event_manipulation,
        'timing', action_timing
    )) AS trigger_json
    FROM information_schema.triggers
)

SELECT json_build_object(
    'tables', (
        SELECT json_agg(json_build_object(
            'schema', table_schema,
            'name', table_name,
            'columns', columns
        ))
        FROM tenant_tables
    ),
    'relationships', (SELECT fk_json FROM foreign_keys),
    'views', (SELECT views_json FROM views),
    'functions', (SELECT fn_json FROM functions),
    'triggers', (SELECT trigger_json FROM triggers)
);
