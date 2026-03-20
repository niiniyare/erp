-- =====================================================================
-- ENTITIES, HIERARCHY PATHS, AND ENTITY STATE
-- Migrations: 000201 (entities), 000202 (hierarchy_paths), 000203 (entitystate)
-- Notes:
--   - entities.uuid is the PRIMARY KEY with NO DEFAULT — must be supplied
--   - hierarchy_paths is a closure table: stores ALL ancestor/descendant pairs
--     including self-references (depth=0)
--   - entitystate.uuid is the PK with NO DEFAULT — must be supplied
--   - accrual_method (boolean) and fy_start_month (int 1-12) are NOT NULL
-- =====================================================================
DO $$
DECLARE
  -- Tenant IDs (resolved by slug)
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;

  -- ACME entity UUIDs (fixed so other seed files can reference them)
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001'; -- COMPANY
  v_us_ops    UUID := 'a0000000-0000-0000-0000-000000000002'; -- REGION
  v_sales     UUID := 'a0000000-0000-0000-0000-000000000003'; -- DEPARTMENT
  v_eng       UUID := 'a0000000-0000-0000-0000-000000000004'; -- DEPARTMENT
  v_eu_hq     UUID := 'a0000000-0000-0000-0000-000000000005'; -- REGION

  -- Globex entity UUIDs
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001'; -- COMPANY
  v_asia_div  UUID := 'b0000000-0000-0000-0000-000000000002'; -- REGION
  v_mfg       UUID := 'b0000000-0000-0000-0000-000000000003'; -- DEPARTMENT

  -- Stark entity UUIDs
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001'; -- COMPANY
  v_rnd       UUID := 'c0000000-0000-0000-0000-000000000002'; -- DEPARTMENT
  v_weapons   UUID := 'c0000000-0000-0000-0000-000000000003'; -- COST_CENTER

BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  -- -----------------------------------------------------------------------
  -- ACME CORPORATION ENTITIES
  -- -----------------------------------------------------------------------
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month)
  VALUES
    (v_acme_co, v_acme_id, 'ACME Corporation', 'ACME',   'COMPANY',     NULL,       TRUE, 1),
    (v_us_ops,  v_acme_id, 'US Operations',    'US-OPS', 'REGION',      v_acme_co,  TRUE, 1),
    (v_sales,   v_acme_id, 'Sales Department', 'SALES',  'DEPARTMENT',  v_us_ops,   TRUE, 1),
    (v_eng,     v_acme_id, 'Engineering',      'ENG',    'DEPARTMENT',  v_us_ops,   TRUE, 1),
    (v_eu_hq,   v_acme_id, 'Europe HQ',        'EU-HQ',  'REGION',      v_acme_co,  FALSE,1)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- ACME hierarchy paths (closure table — all ancestor/descendant pairs)
  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    -- Self-references
    (v_acme_id, v_acme_co, v_acme_co, 0),
    (v_acme_id, v_us_ops,  v_us_ops,  0),
    (v_acme_id, v_sales,   v_sales,   0),
    (v_acme_id, v_eng,     v_eng,     0),
    (v_acme_id, v_eu_hq,   v_eu_hq,   0),
    -- Depth 1: direct parent → child
    (v_acme_id, v_acme_co, v_us_ops,  1),
    (v_acme_id, v_acme_co, v_eu_hq,   1),
    (v_acme_id, v_us_ops,  v_sales,   1),
    (v_acme_id, v_us_ops,  v_eng,     1),
    -- Depth 2: grandparent → grandchild
    (v_acme_id, v_acme_co, v_sales,   2),
    (v_acme_id, v_acme_co, v_eng,     2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  -- ACME entity state (document sequence tracking)
  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config)
  VALUES
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_acme_co,
     '{"prefix":"ACME-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_acme_co,
     '{"prefix":"ACME-PO-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'SO',  1000, v_acme_co,
     '{"prefix":"ACME-SO-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_sales,
     '{"prefix":"SALES-INV-","pad_length":5,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_eu_hq,
     '{"prefix":"EU-INV-","pad_length":5,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;

  -- -----------------------------------------------------------------------
  -- GLOBEX CORPORATION ENTITIES
  -- -----------------------------------------------------------------------
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month)
  VALUES
    (v_globex_co, v_globex_id, 'Globex Corp',   'GLOBEX', 'COMPANY',    NULL,       FALSE, 4),
    (v_asia_div,  v_globex_id, 'Asia Division', 'ASIA',   'REGION',     v_globex_co,FALSE, 4),
    (v_mfg,       v_globex_id, 'Manufacturing', 'MFG',    'DEPARTMENT', v_asia_div,  FALSE, 4)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- Globex hierarchy paths
  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    (v_globex_id, v_globex_co, v_globex_co, 0),
    (v_globex_id, v_asia_div,  v_asia_div,  0),
    (v_globex_id, v_mfg,       v_mfg,       0),
    (v_globex_id, v_globex_co, v_asia_div,  1),
    (v_globex_id, v_asia_div,  v_mfg,       1),
    (v_globex_id, v_globex_co, v_mfg,       2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  -- Globex entity state
  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config)
  VALUES
    (gen_random_uuid(), v_globex_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_globex_co,
     '{"prefix":"GX-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_globex_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_globex_co,
     '{"prefix":"GX-PO-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_globex_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'SO',  1000, v_globex_co,
     '{"prefix":"GX-SO-","pad_length":6,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;

  -- -----------------------------------------------------------------------
  -- STARK INDUSTRIES ENTITIES
  -- -----------------------------------------------------------------------
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month)
  VALUES
    (v_stark_co, v_stark_id, 'Stark Industries', 'STARK',   'COMPANY',     NULL,       TRUE, 10),
    (v_rnd,      v_stark_id, 'R&D Division',     'RND',     'DEPARTMENT',  v_stark_co, TRUE, 10),
    (v_weapons,  v_stark_id, 'Advanced Weapons', 'WEAPONS', 'COST_CENTER', v_rnd,      TRUE, 10)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- Stark hierarchy paths
  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    (v_stark_id, v_stark_co, v_stark_co, 0),
    (v_stark_id, v_rnd,      v_rnd,      0),
    (v_stark_id, v_weapons,  v_weapons,  0),
    (v_stark_id, v_stark_co, v_rnd,      1),
    (v_stark_id, v_rnd,      v_weapons,  1),
    (v_stark_id, v_stark_co, v_weapons,  2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  -- Stark entity state
  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config)
  VALUES
    (gen_random_uuid(), v_stark_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_stark_co,
     '{"prefix":"SI-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_stark_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_stark_co,
     '{"prefix":"SI-PO-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_stark_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'SO',  1000, v_stark_co,
     '{"prefix":"SI-SO-","pad_length":6,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;

END $$;
