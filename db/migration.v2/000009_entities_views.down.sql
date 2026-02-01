-- =====================================================================
-- ENTITIES VIEWS DOWN MIGRATION
-- =====================================================================
-- Drop all entity views in reverse order of creation
DROP VIEW IF EXISTS v_tenant_resource_utilization;

DROP VIEW IF EXISTS v_entity_paths;

DROP VIEW IF EXISTS v_entity_changes;

DROP VIEW IF EXISTS v_active_entities;

DROP VIEW IF EXISTS v_tenant_entity_summary;

DROP VIEW IF EXISTS v_company_structure;

DROP VIEW IF EXISTS v_department_summary;

DROP VIEW IF EXISTS v_cost_center_info;

DROP VIEW IF EXISTS v_entity_structure;

DROP VIEW IF EXISTS v_tenant_hierarchy;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES VIEWS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 10 entity views for reporting and analytics';

RAISE NOTICE '- v_tenant_hierarchy (recursive hierarchy view)';

RAISE NOTICE '- v_entity_structure (organizational structure)';

RAISE NOTICE '- v_cost_center_info (cost center details)';

RAISE NOTICE '- v_department_summary (department aggregation)';

RAISE NOTICE '- v_company_structure (company organization)';

RAISE NOTICE '- v_tenant_entity_summary (tenant statistics)';

RAISE NOTICE '- v_active_entities (active entity report)';

RAISE NOTICE '- v_entity_changes (change tracking)';

RAISE NOTICE '- v_entity_paths (hierarchy paths)';

RAISE NOTICE '- v_tenant_resource_utilization (resource usage)';

RAISE NOTICE '===================================================================';

END;

$$
;
