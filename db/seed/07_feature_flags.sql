-- =====================================================================
-- FEATURE FLAGS
-- Migrations: 000801 (feature_flags core), 000802 (overrides)
-- Notes:
--   - UNIQUE constraint: (tenant_id, name)
--   - flag_type: boolean/string/number/json
--   - rollout_percentage: 0-100 (nullable)
--   - deleted_at IS NULL for active flags
-- =====================================================================
DO $$
DECLARE
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;

  -- Entity UUIDs (from 03_entities.sql)
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  -- -----------------------------------------------------------------------
  -- ACME FEATURE FLAGS
  -- -----------------------------------------------------------------------
  INSERT INTO feature_flags (tenant_id, entity_id, name, description, flag_type, default_value, rollout_percentage, metadata)
  VALUES
    (v_acme_id, v_acme_co, 'advanced_reporting',
     'Enable advanced analytics and custom report builder',
     'boolean', TRUE, 100,
     '{"module":"finance","tier":"ENTERPRISE","since":"2025-01-01"}'),

    (v_acme_id, v_acme_co, 'ai_invoice_matching',
     'AI-powered automatic invoice to PO matching',
     'boolean', FALSE, 50,
     '{"module":"finance","beta":true,"depends_on":"advanced_reporting"}'),

    (v_acme_id, v_acme_co, 'multi_currency_support',
     'Enable transactions and reporting in multiple currencies',
     'boolean', TRUE, 100,
     '{"module":"finance","required_for":"eu_hq"}'),

    (v_acme_id, v_acme_co, 'hr_self_service',
     'Employee self-service portal for leave and expense submissions',
     'boolean', TRUE, 100,
     '{"module":"hr"}'),

    (v_acme_id, v_acme_co, 'mobile_app_access',
     'Enable mobile application access for field employees',
     'boolean', FALSE, 30,
     '{"module":"core","platform":"ios,android","beta":true}'),

    (v_acme_id, v_acme_co, 'bulk_import_limit',
     'Maximum number of records per bulk import operation',
     'number', FALSE, NULL,
     '{"module":"core","default_value":5000,"max_value":50000}')
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- -----------------------------------------------------------------------
  -- GLOBEX FEATURE FLAGS
  -- -----------------------------------------------------------------------
  INSERT INTO feature_flags (tenant_id, entity_id, name, description, flag_type, default_value, rollout_percentage, metadata)
  VALUES
    (v_globex_id, v_globex_co, 'advanced_reporting',
     'Enable advanced analytics and custom report builder',
     'boolean', FALSE, 0,
     '{"module":"finance","tier":"ENTERPRISE"}'),

    (v_globex_id, v_globex_co, 'inventory_lot_tracking',
     'Mandatory lot and batch tracking for manufacturing inventory',
     'boolean', TRUE, 100,
     '{"module":"inventory","required":true,"compliance":"ISO9001"}'),

    (v_globex_id, v_globex_co, 'quality_inspection_workflow',
     'Inline quality inspection steps for manufacturing orders',
     'boolean', TRUE, 100,
     '{"module":"inventory","industry":"manufacturing"}'),

    (v_globex_id, v_globex_co, 'multi_plant_consolidation',
     'Consolidated reporting across all manufacturing plants',
     'boolean', FALSE, 20,
     '{"module":"finance","beta":true}')
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- -----------------------------------------------------------------------
  -- STARK FEATURE FLAGS
  -- -----------------------------------------------------------------------
  INSERT INTO feature_flags (tenant_id, entity_id, name, description, flag_type, default_value, rollout_percentage, metadata)
  VALUES
    (v_stark_id, v_stark_co, 'advanced_reporting',
     'Enable advanced analytics and custom report builder',
     'boolean', TRUE, 100,
     '{"module":"finance","tier":"ENTERPRISE"}'),

    (v_stark_id, v_stark_co, 'classified_data_masking',
     'Mask classified fields based on user security clearance level',
     'boolean', TRUE, 100,
     '{"module":"security","compliance":"ITAR,DFARS","required":true}'),

    (v_stark_id, v_stark_co, 'enhanced_audit_trail',
     'Capture granular field-level change history for compliance',
     'boolean', TRUE, 100,
     '{"module":"security","compliance":"ITAR","retention_days":2555}'),

    (v_stark_id, v_stark_co, 'ai_invoice_matching',
     'AI-powered automatic invoice to PO matching',
     'boolean', TRUE, 100,
     '{"module":"finance","beta":false}'),

    (v_stark_id, v_stark_co, 'mfa_required_all',
     'Require MFA for all user logins regardless of clearance level',
     'boolean', TRUE, 100,
     '{"module":"security","required":true,"policy":"zero_trust"}'),

    (v_stark_id, v_stark_co, 'project_cost_tracking',
     'Link expenses and time entries to R&D project codes',
     'boolean', TRUE, 100,
     '{"module":"finance","industry":"defense"}')
  ON CONFLICT (tenant_id, name) DO NOTHING;

END $$;
