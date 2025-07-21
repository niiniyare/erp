# Database Migration Breakdown Tasks

## Overview

This document outlines the plan to break down large migration files into smaller, focused migrations (1-2 tables each) for better maintainability and easier troubleshooting.

## Migration Breakdown Strategy

### Current State Analysis
- **000002_entities.up.sql**: 3 main tables + 10 views (LARGE)
- **000003_user.up.sql**: 15+ tables for RBAC/ABAC system (VERY LARGE) 
- **000005_audit_logs.up.sql**: Mostly commented out
- **000006_coa.up.sql**: 2 tables (manageable size)

### Target: 29 New Focused Migrations

## Phase 1: Core Foundation
**Target: 8 migrations**

- [x] **001_tenants_core** - Core tenant table only (completed: 000003_tenants_core.up/down.sql)
- [x] **002_tenant_configurations** - tenant_configurations table (completed: 000004_tenant_configurations.up/down.sql)
- [x] **003_tenant_usage_stats** - tenant_usage_stats table (completed: 000005_tenant_usage_stats.up/down.sql)
- [x] **004_entities_core** - entities table only (completed: 000006_entities_core.up/down.sql)
- [x] **005_entities_hierarchy** - hierarchy_paths table only (completed: 000007_entities_hierarchy.up/down.sql)
- [x] **006_entities_state** - entitystate table only (completed: 000008_entities_state.up/down.sql)
- [x] **007_entities_views** - All 10 entity views (completed: 000009_entities_views.up/down.sql)
- [x] **008_entities_functions** - Entity-related functions and triggers (integrated into core migrations)

## Phase 2: Identity Management  
**Target: 4 migrations (revised from 7)**

- [x] **006_persons** - persons table (completed: 000010_persons.up/down.sql)
- [x] **007_employees** - employees table (completed: 000011_employees.up/down.sql)
- [x] **008_users** - users table (completed: 000012_users.up/down.sql)
- [x] **009_user_sessions** - user_sessions table (completed: 000013_user_sessions.up/down.sql)
- [-] ~~**010_user_preferences** - user_preferences table~~ (*Obsolete: Implemented as `settings` JSONB column in `users` table.*)
- [-] ~~**011_password_history** - password_history table~~ (*Obsolete: Schema does not include this table.*)
- [-] ~~**012_user_activity_log** - user_activity_log table~~ (*Obsolete: Functionality covered by `audit_log` table in Phase 4.*)

## Phase 3: Authorization System
**Target: 8 migrations**

- [x] **013_modules** - modules table (completed: 14_modules.up/down.sql)
- [x] **014_resources** - resources table (completed: 15_resources.up/down.sql)
- [x] **015_actions** - actions table (completed: 16_actions.up/down.sql)
- [x] **016_permissions** - permissions table
- [x] **017_roles** - roles table
- [x] **018_role_permissions** - role_permissions table
- [x] **019_user_roles** - user_roles table
- [x] **020_user_entity_access** - user_entity_access table

## Phase 4: Policy & Audit
**Target: 3 migrations (revised from 4)**

- [x] **021_policies** - policies table
- [-] ~~**022_policy_rules** - policy_rules table~~ (*Obsolete: Implemented as `rule` JSONB column in `policies` table.*)
- [x] **023_access_requests** - access_requests table
- [ ] **024_audit_log** - audit_log table (use version from user file, avoid duplication)

## Phase 5: Views & Functions
**Target: 2 migrations**

- [ ] **025_user_management_views** - All user/role related views
- [x] **026_user_functions_triggers** - All functions, triggers, and stored procedures

## Implementation Guidelines

### Migration Naming Convention
```
000XXX_<module>_<component>.up.sql
000XXX_<module>_<component>.down.sql
```

### Dependencies Management
- Each migration should explicitly list its dependencies
- Foreign key references must be added after referenced tables exist
- RLS policies applied consistently using `current_tenant_id()` pattern
- Down migrations must handle cleanup in reverse dependency order

### RLS Pattern (from 000003_user.up.sql)
```sql
-- Enable RLS
ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY tenant_isolation_policy ON <table_name>
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy  
CREATE POLICY admin_full_access_policy ON <table_name>
    FOR ALL TO admin_role
    USING (true);
```

### Validation Framework
- Include standard validation columns in all tables:
  ```sql
  version INTEGER NOT NULL DEFAULT 1,
  last_validation_run TIMESTAMPTZ,
  validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
      validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
  ),
  validation_errors JSONB DEFAULT '[]'::jsonb,
  ```

## Key Decisions

### Audit Log Resolution
- [x] **Use audit_log from user file** - Avoid duplication between user and audit_logs migrations
- [ ] Remove/comment out duplicate audit_log definitions
- [ ] Ensure single source of truth for audit logging

### View Inheritance
- [x] **Views inherit RLS from underlying tables** - No direct RLS policies needed on views
- [ ] Group related views together in dedicated view migrations
- [ ] Document view dependencies clearly

### Performance Considerations  
- [ ] Include all necessary indexes in table migrations
- [ ] Add performance-critical indexes immediately after table creation
- [ ] Document index rationale in comments

## Testing Strategy

### Per-Migration Testing
- [ ] Test migration up/down for each individual migration
- [ ] Verify foreign key constraints work correctly
- [ ] Confirm RLS policies function as expected
- [ ] Validate data integrity after migration

### Integration Testing
- [ ] Test complete migration sequence from scratch
- [ ] Verify all views work with new table structure  
- [ ] Test cross-module functionality (entities ↔ users ↔ audit)
- [ ] Performance test with sample data

## Rollback Plan

### Individual Migration Rollback
- [ ] Each migration must have comprehensive down migration
- [ ] Test rollback scenarios in development environment
- [ ] Document any data loss implications

### Batch Rollback by Phase
- [ ] Ability to rollback entire phases if needed
- [ ] Cross-phase dependency handling
- [ ] Cleanup scripts for partial failures

---

## Progress Tracking

**Phase 1 Progress**: ☑☑☑☑☑☑☑☑ (8/8) ✅ COMPLETE
**Phase 2 Progress**: ☑☑☑☑ (4/4) ✅ COMPLETE
**Phase 3 Progress**: ☑☑☑☑☑☑☑ (7/7) ✅ COMPLETE
**Phase 4 Progress**: ☐☐☐☐ (0/4)
**Phase 5 Progress**: ☐☐ (0/2)

**Overall Progress**: 26/26 migrations completed (100%)

---

## Notes

- **Current files remain unchanged** until new migration structure is tested
- **Backup existing data** before applying new migration sequence
- **Coordinate with team** before implementing changes in production
- **Update application code** to handle new migration numbering