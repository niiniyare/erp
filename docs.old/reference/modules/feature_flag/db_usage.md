# Feature Flag System - Usage Guide

A comprehensive guide for using the feature flag system, including examples, best practices, and troubleshooting tips.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Basic Operations](#basic-operations)
3. [Feature Flag Evaluation](#feature-flag-evaluation)
4. [Monitoring & Analytics](#monitoring--analytics)
5. [Bulk Operations](#bulk-operations)
6. [Maintenance](#maintenance)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

Before working with feature flags, you must set the tenant context. This ensures all operations are scoped to the correct tenant:

```sql
-- Required: Set tenant context for all operations
SET app.current_tenant_id = 'your-tenant-uuid';
SET app.current_user_id = 'your-user-uuid';
SET app.current_session_id = 'your-session-uuid';
```

> **Note:** Replace the UUID values with your actual tenant, user, and session identifiers.

---

## Basic Operations

### Creating Feature Flags

Feature flags can be created with various types (boolean, number, string) and optional configuration:

```sql
-- Create multiple feature flags with different purposes
INSERT INTO feature_flags (tenant_id, name, description, flag_type, default_value, metadata) VALUES
  (
    'your-tenant-uuid', 
    'advanced_reporting', 
    'Enable advanced reporting features', 
    'boolean', 
    false, 
    '{"category": "reporting", "priority": "high"}'
  ),
  (
    'your-tenant-uuid', 
    'new_dashboard', 
    'Enable new dashboard UI', 
    'boolean', 
    false, 
    '{"category": "ui", "priority": "medium"}'
  ),
  (
    'your-tenant-uuid', 
    'api_rate_limit', 
    'API rate limit configuration', 
    'number', 
    true, 
    '{"category": "performance", "default_limit": 1000}'
  ),
  (
    'your-tenant-uuid', 
    'feature_rollout_test', 
    'Test gradual rollout', 
    'boolean', 
    false, 
    '{"category": "testing"}'
  );
```

**Key Components:**
- **name**: Unique identifier for the feature flag
- **flag_type**: `boolean`, `number`, or `string`
- **default_value**: Default state when no overrides exist
- **metadata**: JSON field for additional configuration and categorization

### Configuring Gradual Rollouts

For controlled feature releases, set a rollout percentage. The system will enable the feature for approximately that percentage of users:

```sql
-- Enable feature for 25% of users (useful for phased rollouts)
UPDATE feature_flags 
SET rollout_percentage = 25 
WHERE name = 'feature_rollout_test' 
  AND tenant_id = 'your-tenant-uuid';
```

**Use Cases:**
- Initial testing: 5-10%
- Controlled expansion: 25-50%
- Near-full rollout: 75-95%
- Full release: 100%

### Creating Tenant Overrides

Override default behavior for specific tenants, useful for premium features or special access:

```sql
-- Enable advanced reporting for a specific premium tenant
INSERT INTO tenant_feature_overrides (tenant_id, feature_flag_id, feature_flag_name, enabled, reason) 
SELECT 
  'your-tenant-uuid',
  ff.id,
  ff.name,
  true,
  'Enable advanced reporting for premium tenant'
FROM feature_flags ff 
WHERE ff.name = 'advanced_reporting' 
  AND ff.tenant_id = 'your-tenant-uuid';
```

**Benefits:**
- Override rollout percentages for specific tenants
- Provide early access to beta testers
- Disable problematic features for affected tenants
- Document reasoning for audit purposes

---

## Feature Flag Evaluation

### Standard Evaluation

Evaluate a feature flag with full audit logging (recommended for critical features):

```sql
-- Evaluate with complete audit trail
SELECT * FROM evaluate_feature_flag('advanced_reporting');
```

**Returns:** Flag status, evaluation source (override, rollout, default), and metadata.

### Fast Evaluation

Skip audit logging for performance-critical paths where you don't need historical tracking:

```sql
-- Faster evaluation without audit logging (use for high-frequency checks)
SELECT * FROM evaluate_feature_flag_fast('new_dashboard');
```

**When to Use:**
- High-frequency API endpoints
- Real-time UI rendering
- Performance-critical operations

### Batch Evaluation

Evaluate all feature flags at once, useful for application initialization:

```sql
-- Get all feature flags for the current tenant in one query
SELECT * FROM evaluate_all_feature_flags();
```

**Common Use Case:** Load all feature flags when a user logs in, cache them client-side.

### Cached Evaluation

Use materialized view cache for maximum performance (best for very high-frequency access):

```sql
-- Use pre-computed cache (fastest option)
SELECT * FROM evaluate_feature_flag_cached('advanced_reporting');

-- Get all flags from cache
SELECT * FROM evaluate_all_feature_flags_cached();
```

**Performance Benefit:** 10-100x faster than standard evaluation for read-heavy workloads.

**Cache Refresh:** The cache is automatically refreshed when flags are modified.

---

## Monitoring & Analytics

### System Health Metrics

Get an overview of the feature flag system's health and performance:

```sql
-- Check overall system health and usage statistics
SELECT * FROM get_feature_flags_health_metrics();
```

**Metrics Included:**
- Total feature flags
- Active/inactive counts
- Average rollout percentages
- Cache hit rates

### Tenant-Specific Statistics

Analyze feature flag usage for a specific tenant:

```sql
-- Get detailed statistics for a tenant
SELECT * FROM get_tenant_feature_flag_stats('your-tenant-uuid');
```

**Information Provided:**
- Number of flags by type
- Override counts
- Evaluation frequency
- Enabled/disabled ratios

### Cache Performance

Monitor the effectiveness of the caching layer:

```sql
-- Check cache statistics and hit rates
SELECT * FROM get_feature_flags_cache_stats();

-- Verify cache freshness for a tenant
SELECT * FROM check_cache_freshness('your-tenant-uuid');
```

**Why Monitor Cache:**
- Identify stale data issues
- Optimize refresh strategies
- Ensure performance targets are met

### Data Integrity Checks

Verify the consistency and correctness of feature flag data:

```sql
-- Run comprehensive integrity checks
SELECT * FROM check_feature_flags_integrity();
```

**Detects:**
- Orphaned overrides
- Invalid flag references
- Inconsistent tenant associations
- Duplicate entries

### Advanced Analytics

#### Feature Flag Usage by Type

```sql
-- Analyze flag distribution and enabled rates by type
SELECT 
  flag_type,
  evaluation_source,
  COUNT(*) as count,
  ROUND(AVG(CASE WHEN enabled THEN 1 ELSE 0 END) * 100, 2) as enabled_percentage
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = 'your-tenant-uuid'
GROUP BY flag_type, evaluation_source
ORDER BY flag_type, evaluation_source;
```

**Insights:** Understand which types of flags are most commonly enabled and their evaluation sources.

#### Most Evaluated Features

```sql
-- Identify the most frequently checked features (last 7 days)
SELECT 
  context->>'feature_flag_name' as feature_name,
  COUNT(*) as evaluation_count,
  COUNT(*) FILTER (WHERE decision = 'ALLOW') as enabled_count,
  ROUND(COUNT(*) FILTER (WHERE decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as enabled_percentage
FROM audit_log
WHERE tenant_id = 'your-tenant-uuid'
  AND event_type = 'FEATURE_FLAG_EVALUATED'
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY context->>'feature_flag_name'
ORDER BY evaluation_count DESC
LIMIT 10;
```

**Use Cases:**
- Identify candidates for caching
- Find features to optimize
- Understand usage patterns

#### Rollout Effectiveness

```sql
-- Verify that rollout percentages are working as expected
SELECT 
  ff.name,
  ff.rollout_percentage,
  COUNT(DISTINCT al.session_id) as unique_evaluations,
  COUNT(*) FILTER (WHERE al.decision = 'ALLOW') as enabled_evaluations,
  ROUND(COUNT(*) FILTER (WHERE al.decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as actual_enabled_percentage
FROM feature_flags ff
JOIN audit_log al ON al.context->>'feature_flag_name' = ff.name
WHERE ff.tenant_id = 'your-tenant-uuid'
  AND ff.rollout_percentage IS NOT NULL
  AND al.event_type = 'FEATURE_FLAG_EVALUATED'
  AND al.created_at >= NOW() - INTERVAL '24 hours'
GROUP BY ff.name, ff.rollout_percentage
ORDER BY ff.rollout_percentage DESC;
```

**Validation:** Ensure actual enabled percentage roughly matches the configured rollout percentage.

---

## Bulk Operations

### Bulk Rollout Updates

Update rollout percentages for multiple features simultaneously:

```sql
-- Increase rollout to 50% for multiple features at once
SELECT bulk_update_rollout_percentage(
  ARRAY['new_dashboard', 'feature_rollout_test'], 
  50, 
  'your-tenant-uuid'
);
```

**Scenario:** Coordinated feature releases that should roll out together.

### Bulk Feature Creation

Create multiple feature flags from JSON configuration:

```sql
-- Create multiple experimental features in one operation
SELECT bulk_create_feature_flags('[
  {
    "name": "experimental_feature_a",
    "description": "Experimental feature A",
    "flag_type": "boolean",
    "default_value": false,
    "rollout_percentage": 10,
    "metadata": {"category": "experimental"}
  },
  {
    "name": "experimental_feature_b",
    "description": "Experimental feature B",
    "flag_type": "boolean",  
    "default_value": false,
    "metadata": {"category": "experimental"}
  }
]'::jsonb, 'your-tenant-uuid');
```

**Benefits:**
- Faster than individual inserts
- Atomic operation (all or nothing)
- Easier configuration management

### Configuration Export

Export all feature flag configurations for a tenant:

```sql
-- Export complete feature flag configuration as JSON
SELECT export_tenant_feature_flags('your-tenant-uuid');
```

**Use Cases:**
- Backup configurations
- Replicate settings across environments
- Documentation and auditing

---

## Maintenance

### Cache Refresh

Manually refresh the materialized view cache after bulk operations:

```sql
-- Force cache refresh (usually automatic, but useful after bulk changes)
SELECT refresh_feature_flags_cache();
```

**When to Use:**
- After bulk updates
- When cache freshness checks fail
- During maintenance windows

### Audit Log Cleanup

Remove old audit logs to manage database size:

```sql
-- Delete audit logs older than 90 days
SELECT cleanup_old_feature_flag_audit_logs(90);
```

**Recommendation:** Run monthly, adjust retention period based on compliance requirements.

### Soft-Deleted Flag Cleanup

Permanently remove feature flags that were soft-deleted:

```sql
-- Hard delete flags that have been soft-deleted for 30+ days
SELECT cleanup_soft_deleted_feature_flags(30);
```

**Best Practice:** Keep a grace period (30-90 days) to allow for recovery if needed.

### Fix Orphaned Data

Repair data integrity issues automatically:

```sql
-- Clean up orphaned overrides and fix inconsistencies
SELECT fix_orphaned_overrides();
```

**Run After:** Bulk deletions or when integrity checks report issues.

---

## Best Practices

### Naming Conventions

Use clear, hierarchical naming to improve organization and discoverability:

```
✅ Good Examples:
- ui.new_dashboard
- api.v2_endpoints
- reporting.advanced_analytics
- performance.cache_enabled

❌ Avoid:
- feature1
- test flag
- NewFeature
```

**Guidelines:**
- Use lowercase with underscores or dots
- Prefix with feature area (ui, api, reporting)
- Be descriptive and specific
- Avoid abbreviations unless widely understood

### Rollout Strategy

Implement gradual, measured rollouts to minimize risk:

1. **Initial Testing (5-10%)**: Limited internal or beta users
2. **Controlled Expansion (25-50%)**: Monitor metrics closely
3. **Broad Rollout (75-95%)**: Watch for any issues at scale
4. **Full Release (100%)**: Complete deployment

**Key Actions:**
- Monitor error rates and performance metrics
- Use tenant overrides for early access testers
- Have rollback procedures ready
- Communicate clearly with users

### Cleanup & Lifecycle Management

Prevent technical debt from accumulating:

```sql
-- Add expiration metadata when creating flags
INSERT INTO feature_flags (tenant_id, name, description, flag_type, default_value, metadata)
VALUES (
  'your-tenant-uuid',
  'temporary_holiday_theme',
  'Special holiday UI theme',
  'boolean',
  false,
  '{"category": "ui", "expires": "2025-01-15", "owner": "design-team"}'
);
```

**Regular Reviews:**
- Monthly: Review flags with expiration dates
- Quarterly: Audit all flags at 100% rollout
- Annually: Remove flags fully integrated into codebase

### Monitoring & Alerting

Set up proactive monitoring for the feature flag system:

```sql
-- Create a daily health check job
SELECT * FROM get_feature_flags_health_metrics();
SELECT * FROM check_feature_flags_integrity();
```

**Alert On:**
- Integrity check failures
- Cache staleness exceeding threshold
- Abnormal evaluation patterns
- Audit log size growth

### Documentation

Maintain clear documentation for each feature flag:

```sql
-- Well-documented feature flag example
INSERT INTO feature_flags (tenant_id, name, description, flag_type, default_value, metadata)
VALUES (
  'your-tenant-uuid',
  'api.rate_limiting_v2',
  'New rate limiting algorithm with burst support. See RFC-2024-007 for details.',
  'boolean',
  false,
  '{
    "category": "performance",
    "owner": "platform-team",
    "created_date": "2025-01-15",
    "target_completion": "2025-03-01",
    "jira_ticket": "PLAT-1234",
    "documentation_url": "https://docs.internal.com/rate-limiting-v2"
  }'
);
```

**Documentation Elements:**
- Purpose and expected outcome
- Responsible team or individual
- Timeline and milestones
- Related tickets or documentation
- Dependencies and prerequisites

### Security & Compliance

Maintain audit trails and access controls:

- **Always Use Reason Field**: Document why overrides are created
- **Regular Access Reviews**: Audit who can modify flags
- **Monitor Admin Actions**: Review all flag modifications
- **Data Retention**: Follow company policies for audit logs

```sql
-- Example: Override with clear justification
INSERT INTO tenant_feature_overrides (tenant_id, feature_flag_id, feature_flag_name, enabled, reason)
VALUES (
  'tenant-123',
  (SELECT id FROM feature_flags WHERE name = 'advanced_features'),
  'advanced_features',
  true,
  'Approved by CSM for enterprise customer. Ticket: SUP-9876. Approved by: j.smith@company.com'
);
```

### Performance Optimization

Choose the right evaluation method for your use case:

| Method | Speed | Audit Trail | Use Case |
|--------|-------|-------------|----------|
| `evaluate_feature_flag()` | Normal | Yes | Critical business logic |
| `evaluate_feature_flag_fast()` | Fast | No | High-frequency checks |
| `evaluate_feature_flag_cached()` | Fastest | No | Read-heavy workloads |
| `evaluate_all_feature_flags_cached()` | Fastest | No | Application initialization |

**Tips:**
- Use cached evaluation for features checked on every request
- Batch evaluate flags during user login
- Monitor query performance regularly
- Consider caching results in application layer

### Testing

Comprehensive testing ensures feature flags work as intended:

**Test Cases:**
1. **Both States**: Verify feature works when enabled AND disabled
2. **Rollout Percentages**: Confirm approximate distribution
3. **Overrides**: Test tenant-specific overrides
4. **Edge Cases**: Missing context, invalid flags, network issues
5. **Performance**: Load test evaluation under high concurrency

```sql
-- Testing example: Verify rollout percentage distribution
-- Run this multiple times with different session IDs to verify randomness
SET app.current_session_id = gen_random_uuid()::text;
SELECT * FROM evaluate_feature_flag('feature_rollout_test');
```

---

## Troubleshooting

### Error: "No tenant context set"

**Problem:** Functions require tenant context but it hasn't been set.

**Solution:**
```sql
-- Set the tenant context before any operations
SET app.current_tenant_id = 'your-tenant-uuid';
```

**Prevention:** Include context setup in application initialization code.

---

### Error: "Feature flag not found"

**Problem:** The requested feature flag doesn't exist or has been deleted.

**Diagnosis:**
```sql
-- Check if flag exists and its status
SELECT * FROM feature_flags 
WHERE name = 'flag_name' 
  AND deleted_at IS NULL;
```

**Common Causes:**
- Typo in flag name
- Flag was soft-deleted
- Wrong tenant context
- Flag not created yet

---

### Stale Cache Issues

**Problem:** Cached values don't reflect recent changes.

**Diagnosis:**
```sql
-- Check cache freshness
SELECT * FROM check_cache_freshness('your-tenant-uuid');
```

**Solution:**
```sql
-- Manually refresh cache
SELECT refresh_feature_flags_cache();

-- Listen for automatic refresh notifications
LISTEN feature_flags_cache_refresh;
```

**Prevention:** Cache should auto-refresh on changes. Check notification system if this isn't working.

---

### Performance Degradation

**Problem:** Feature flag evaluations are slow.

**Diagnosis:**
```sql
-- Check index usage
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan as index_scans,
  idx_tup_read as tuples_read
FROM pg_stat_user_indexes 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides')
ORDER BY idx_scan DESC;

-- Check slow queries
SELECT 
  query,
  calls,
  mean_time,
  total_time
FROM pg_stat_statements 
WHERE query ILIKE '%feature_flag%'
ORDER BY mean_time DESC
LIMIT 10;
```

**Solutions:**
1. Use cached evaluation functions
2. Batch evaluate flags at initialization
3. Check if indexes are being used
4. Consider database-level query optimization

---

### Data Integrity Issues

**Problem:** Inconsistent or orphaned data detected.

**Diagnosis:**
```sql
-- Run integrity checks
SELECT * FROM check_feature_flags_integrity();
```

**Solution:**
```sql
-- Automatically fix common issues
SELECT fix_orphaned_overrides();
```

**Prevention:**
- Use foreign key constraints (should be in schema)
- Regular integrity check scheduling
- Proper transaction handling in bulk operations

---

### Audit Log Size Growth

**Problem:** Audit log table becoming too large, affecting performance.

**Solutions:**

1. **Regular Cleanup:**
```sql
-- Remove logs older than retention period
SELECT cleanup_old_feature_flag_audit_logs(90);
```

2. **Partitioning:** For very large deployments, consider time-based partitioning:
```sql
-- Example partitioning strategy (requires schema modifications)
-- Partition by month for efficient querying and cleanup
```

3. **Selective Logging:** Use `evaluate_feature_flag_fast()` for high-frequency, low-priority checks.

---

### Row Level Security (RLS) Issues

**Problem:** Users can't access feature flags they should be able to see.

**Diagnosis:**
```sql
-- Verify RLS policies
SELECT * FROM pg_policies 
WHERE tablename = 'feature_flags';

-- Check current tenant context
SHOW app.current_tenant_id;
```

**Common Issues:**
- Tenant context not set correctly
- User doesn't have required role
- RLS policies too restrictive

**Solution:** Review RLS policies and ensure proper context setup.

---

## Performance Monitoring

### Index Usage Statistics

Monitor which indexes are being used effectively:

```sql
-- Verify indexes are being utilized
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan as index_scans,
  idx_tup_read as tuples_read,
  idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'mv_tenant_feature_flags_cache')
ORDER BY idx_scan DESC;
```

**Red Flags:**
- Indexes with 0 scans
- Sequential scans on large tables
- High tuples_read vs tuples_fetched ratio

### Table Size & Growth

Track storage usage and plan capacity:

```sql
-- Monitor table sizes and statistics
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
  n_tup_ins as inserts,
  n_tup_upd as updates,
  n_tup_del as deletes,
  n_live_tup as live_tuples,
  n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'audit_log')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Action Items:**
- High dead_tuples: Schedule VACUUM
- Rapid growth: Review retention policies
- Large audit_log: Implement cleanup job

### Query Performance Analysis

Identify slow queries that need optimization:

```sql
-- Find slow feature flag operations (requires pg_stat_statements extension)
SELECT 
  query,
  calls,
  total_time,
  mean_time,
  rows,
  100.0 * shared_blks_hit / NULLIF(shared_blks_hit + shared_blks_read, 0) AS cache_hit_percent
FROM pg_stat_statements 
WHERE query ILIKE '%feature_flag%' 
  OR query ILIKE '%tenant_feature_overrides%'
ORDER BY mean_time DESC
LIMIT 10;
```

**Optimization Targets:**
- Queries with high mean_time
- Low cache hit percentages
- High call counts with moderate times (good candidates for caching)

---

## Summary

This feature flag system provides a robust, performant solution for managing feature releases across multi-tenant applications. Key takeaways:

- **Start Simple**: Use basic boolean flags for most use cases
- **Scale Gradually**: Implement rollout percentages for controlled releases
- **Monitor Actively**: Use built-in health checks and analytics
- **Optimize Strategically**: Use cached evaluation for high-frequency access
- **Clean Regularly**: Remove deprecated flags and old audit logs
- **Document Thoroughly**: Make flag purpose and lifecycle clear

By following these best practices and using the provided tools, you can safely and efficiently manage feature releases while maintaining system performance and data integrity.
