# 🔧 ERP Migration Validation System Guide

## 📋 **For New Developers**

This guide shows you how to apply the standardized 3-tier validation system to any ERP migration following the pattern used in **`000002_entities.up.sql`** and **`000003_user.up.sql`**.

## 🎯 **The 3-Tier Validation System**

### **Tier 1: Structural Validation** ✅
- Standard validation columns on every table
-  RLS policies with NULL protection
- Performance indexes for validation queries
- Proper documentation and comments

### **Tier 2: Business Logic Validation** 🧠
- Financial data integrity checks
- Workflow validation (status transitions)
- Currency and exchange rate validation
- Date logic and temporal consistency
- Compliance and audit requirements

### **Tier 3: Custom Business Rules** ⚙️
- Tenant-specific validation rules
- Configurable severity levels
- Dynamic SQL execution engine
- Rule management system

## 🛠️ **Standard Implementation Pattern**

<!-- ### **Step 1: Add Validation Columns** -->
<!---->
<!-- Add these to **every table** in your migration: -->
<!---->
<!-- ```sql -->
<!-- -- Standard validation columns (REQUIRED for all tables) -->
<!-- version INTEGER NOT NULL DEFAULT 1, -->
<!-- last_validation_run TIMESTAMPTZ, -->
<!-- validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK ( -->
<!--     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR') -->
<!-- ), -->
<!-- validation_errors JSONB DEFAULT '[]'::jsonb, -->
<!-- ``` -->
<!---->
### **Step 2:  RLS Policies**

Apply this pattern to **every table**:

```sql
-- Enable RLS
ALTER TABLE your_table ENABLE ROW LEVEL SECURITY;

-- Application role policy with NULL context protection
CREATE POLICY your_table_tenant_isolation ON your_table
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy for maintenance
CREATE POLICY your_table_admin_access ON your_table
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
```

### **Step 3: Required Comments**

Add  documentation:

```sql
COMMENT ON TABLE your_table IS 
'[Purpose and description of table with validation capabilities]';

COMMENT ON COLUMN your_table.version IS 
'Record version for optimistic locking and change tracking';

COMMENT ON COLUMN your_table.last_validation_run IS 
'Timestamp of last validation check execution';

COMMENT ON COLUMN your_table.validation_status IS 
'Current validation state - PENDING/VALID/WARNING/ERROR';

COMMENT ON COLUMN your_table.validation_errors IS 
'JSON array of current validation errors and warnings';
```

### **Step 4: Performance Indexes**

Add essential indexes:

```sql
-- Tenant isolation index (CRITICAL for RLS performance)
CREATE INDEX idx_your_table_tenant ON your_table(tenant_id);

-- Validation status index for monitoring
CREATE INDEX idx_your_table_validation_status ON your_table(validation_status) 
    WHERE validation_status != 'VALID';

-- JSONB validation errors index for error analysis  
CREATE INDEX idx_your_table_validation_errors ON your_table USING GIN(validation_errors)
    WHERE validation_errors != '[]'::jsonb;
```

## 📝 **Down Migration Pattern**

**Always** update your `.down.sql` with proper cleanup order:

```sql
-- 1. Drop RLS policies first
DROP POLICY IF EXISTS your_table_admin_access ON your_table;
DROP POLICY IF EXISTS your_table_tenant_isolation ON your_table;

-- 2. Disable RLS
ALTER TABLE your_table DISABLE ROW LEVEL SECURITY;

-- 3. Drop triggers (if any)
DROP TRIGGER IF EXISTS your_trigger_name ON your_table;

-- 4. Drop functions (if any) 
DROP FUNCTION IF EXISTS your_function_name() CASCADE;

-- 5. Drop tables last
DROP TABLE IF EXISTS your_table CASCADE;
```

## ⚡ **Performance Monitoring**

### **Check RLS Performance**
```sql
-- Monitor RLS policy overhead
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM your_table WHERE tenant_id = current_tenant_id();

-- Should use tenant index, not sequential scan
```

<!-- ### **Check Validation System Health** -->
<!-- ```sql -->
<!-- -- Daily validation health check -->
<!-- SELECT  -->
<!--     table_name, -->
<!--     COUNT(*) as total_records, -->
<!--     COUNT(*) FILTER (WHERE validation_status = 'VALID') as valid_records, -->
<!--     COUNT(*) FILTER (WHERE validation_status = 'ERROR') as error_records, -->
<!--     ROUND(100.0 * COUNT(*) FILTER (WHERE validation_status = 'VALID') / COUNT(*), 2) as health_percentage -->
<!-- FROM information_schema.tables t -->
<!-- JOIN your_table yt ON true  -- Replace with actual table -->
<!-- WHERE t.table_schema = 'public' -->
<!-- GROUP BY table_name; -->
<!-- ``` -->

### **Monitor Index Usage**
```sql
-- Check if validation indexes are being used
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as times_used,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE tablename LIKE '%your_table%'
ORDER BY idx_scan DESC;
```

## ✅ **Testing Checklist**

Before committing your migration, verify:

### **Migration Tests:**
- [ ] **Up Migration**: `psql -f 000XXX_your_module.up.sql` runs successfully
- [ ] **Validation Columns**: `\d your_table` shows all 4 validation columns
- [ ] **RLS Policies**: `\d+ your_table` shows both tenant and admin policies
- [ ] **Constraints**: validation_status constraint accepts only valid values
- [ ] **Down Migration**: `psql -f 000XXX_your_module.down.sql` cleans up completely
- [ ] **No Orphans**: `\dt` shows only expected tables after down migration

### **Performance Tests:**
- [ ] **Index Usage**: `EXPLAIN ANALYZE` shows index scans, not sequential
- [ ] **RLS Overhead**: Queries with tenant filter perform acceptably
- [ ] **Validation Queries**: Error status queries use validation indexes

### **Security Tests:**
- [ ] **Tenant Isolation**: Cannot access other tenant's data
- [ ] **NULL Context**: Queries return empty when tenant context not set
- [ ] **Admin Access**: admin_role can access all tenant data

## 🚀 **CI/CD Integration**

### **Pre-Deployment Validation**
```bash
#!/bin/bash
# validate-migration.sh

echo "🔍 Running migration validation..."

# Test migration up
if ! psql -d $TEST_DB -f db/migration/000XXX_your_module.up.sql; then
    echo "❌ Up migration failed"
    exit 1
fi

# Verify validation columns exist
VALIDATION_COLUMNS=$(psql -d $TEST_DB -t -c "
    SELECT COUNT(*) FROM information_schema.columns 
    WHERE table_name = 'your_table' 
    AND column_name IN ('version', 'last_validation_run', 'validation_status', 'validation_errors')
")

if [ "$VALIDATION_COLUMNS" -ne 4 ]; then
    echo "❌ Missing validation columns"
    exit 1
fi

# Test migration down
if ! psql -d $TEST_DB -f db/migration/000XXX_your_module.down.sql; then
    echo "❌ Down migration failed"
    exit 1
fi

echo "✅ Migration validation passed"
```

## 📊 **Business Logic Integration**

### **Custom Validation Rules**
Once your structural validation is in place, add business rules:

```sql
-- Example: Invoice validation rule
INSERT INTO business_validation_rules (
    tenant_id, rule_name, rule_description, validation_sql, severity, rule_category
) VALUES (
    current_tenant_id(),
    'Invoice Total Validation',
    'Invoice totals must match line item sums',
    'SELECT 1 FROM invoices i WHERE i.tenant_id = current_tenant_id() 
     AND ABS(i.total_amount - (SELECT COALESCE(SUM(li.amount), 0) 
     FROM invoice_line_items li WHERE li.invoice_id = i.id)) > 0.01',
    'CRITICAL',
    'FINANCIAL_INTEGRITY'
);
```

### **Run Validation Checks**
```sql
-- Set tenant context first
SELECT set_tenant_context('your-tenant-uuid-here');

-- Run  validation
SELECT * FROM run__validation();

-- Check specific table validation
SELECT * FROM check_all_business_logic() 
WHERE table_name = 'your_table' AND severity = 'CRITICAL';
```

## 📚 **Migration Examples**

Study these working implementations:

### **Reference Migrations:**
- **`000002_entities.up.sql`** - Entity management with full validation
- **`000003_user.up.sql`** - User RBAC system with ABAC validation

### **Naming Convention:**
- `000XXX_module_name.up.sql` - Implementation 
- `000XXX_module_name.down.sql` - Cleanup
- Always include validation system in both files

## 🚨 **Critical Requirements**

1. **ALWAYS** add the 4 standard validation columns to every table
2. **ALWAYS** implement both application and admin RLS policies
3. **ALWAYS** add WITH CHECK clauses (not just USING)
4. **ALWAYS** include proper down migration cleanup
5. **ALWAYS** add tenant isolation indexes for performance
6. **ALWAYS** test both up and down migrations thoroughly

## 🔍 **Troubleshooting Common Issues**

### **RLS Policy Not Working**
```sql
-- Check tenant context is set
SELECT current_tenant_id(); -- Should return UUID, not NULL

-- Verify policy exists
SELECT policyname, cmd, qual FROM pg_policies WHERE tablename = 'your_table';
```

### **Performance Problems**
```sql
-- Check for missing tenant index
SELECT indexname FROM pg_indexes 
WHERE tablename = 'your_table' AND indexdef LIKE '%tenant_id%';

-- Should show: idx_your_table_tenant
```

### **Validation Status Issues**
```sql
-- Check constraint violations
SELECT validation_status, COUNT(*) 
FROM your_table 
GROUP BY validation_status;

-- All should be valid status values
```

## 💡 **Pro Tips**

1. **Copy & Adapt**: Start by copying validation patterns from entities/user migrations
2. **Test Early**: Run validation tests on small datasets first  
3. **Monitor Performance**: Watch for RLS policy overhead in production
4. **Document Changes**: Update this guide when you discover new patterns
5. **Business Rules**: Add business logic validation after structural validation works

---

**🎉 You now have everything needed to implement consistent validation across all ERP migrations!**

**Next Steps:** Apply this pattern to your next migration, test thoroughly, and help maintain this guide for future developers.
