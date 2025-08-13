CREATE OR REPLACE FUNCTION provision_tenant_complete(
    p_name VARCHAR(255),
    p_email VARCHAR(255),
    p_subdomain VARCHAR(63) DEFAULT NULL,
    p_industry VARCHAR(50) DEFAULT NULL,
    p_company_size VARCHAR(20) DEFAULT 'small',
    p_currency_code CHAR(3) DEFAULT 'USD',
    p_timezone VARCHAR(50) DEFAULT 'UTC',
    p_settings JSONB DEFAULT '{}'
)
RETURNS TABLE(
    id UUID
) AS $body$
DECLARE
    v_tenant_id UUID;
    v_slug VARCHAR(50);
BEGIN
    -- Generate UUID and slug
    v_tenant_id := uuid_generate_v4();
    v_slug := lower(regexp_replace(p_name, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Ensure slug uniqueness
    WHILE EXISTS (SELECT 1 FROM tenants WHERE slug = v_slug AND deleted_at IS NULL) LOOP
        v_slug := v_slug || '-' || substring(v_tenant_id::text, 1, 8);
    END LOOP;

    -- Create tenant record
    INSERT INTO tenants (
        id, slug, name, email, subdomain, status, industry,
        company_size, currency_code, timezone, settings
    ) VALUES (
        v_tenant_id, v_slug, p_name, p_email, p_subdomain, 'pending',
        p_industry, p_company_size, p_currency_code, p_timezone, p_settings
    );

    -- Create default configuration
    INSERT INTO tenant_configurations (
        tenant_id, default_currency
    ) VALUES (
        v_tenant_id, p_currency_code
    );

    -- Initialize usage statistics for current month
    INSERT INTO tenant_usage_stats (tenant_id, period_start, period_end)
    VALUES (
        v_tenant_id,
        date_trunc('month', CURRENT_DATE)::DATE,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::DATE
    );

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_id AS id;
END;
$body$ LANGUAGE plpgsql;