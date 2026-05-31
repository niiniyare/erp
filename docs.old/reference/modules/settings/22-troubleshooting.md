# Chapter 22: Troubleshooting Guide

[← Common Scenarios](./21-common-scenarios.md) | [Next: Business Rules →](./23-business-rules.md)

---

## Configuration Not Resolving Expected Value

**Symptom**: `GET /config/finance/invoice_prefix` returns the system default even though a tenant override was set.

**Checklist**:
1. Verify `X-Tenant-ID` header is present and correct
2. Check that `current_tenant_id()` returns the expected UUID in the database session
3. Query `tenant_configurations` directly:
   ```sql
   SELECT settings->'finance.invoice_prefix' FROM tenant_configurations
   WHERE tenant_id = '<your-tenant-id>';
   ```
4. Confirm the key is stored as `finance.invoice_prefix` (dot notation), not `invoice_prefix`
5. Check `required_feature_flag` on `config_definitions` — the feature may not be enabled

---

## Entity Override Not Taking Effect

**Symptom**: An entity-level override was set but the effective value still shows the tenant value.

**Checklist**:
1. Query entity settings directly:
   ```sql
   SELECT settings FROM entities
   WHERE uuid = '<entity-id>' AND tenant_id = current_tenant_id();
   ```
2. Verify the key format — it must be `module.config_key` exactly as registered in `config_definitions`
3. Confirm `entity_id` is being passed to the resolution query
4. Check that the entity is not soft-deleted (`deleted_at IS NULL`)
5. Check cache — stale cache may serve an old value for up to the TTL

---

## Template Application Partially Failed

**Symptom**: Template apply returned `conflicts` or `skipped` entries.

**Checklist**:
1. Review the `conflicts` array in the response — each entry explains why it was skipped
2. If `preserve_existing: true`, keys already set at the target level are skipped by design
3. Check `required_feature_flags` on the template — if a flag is missing, the whole template may be rejected
4. Query `template_applications` to see the full history:
   ```sql
   SELECT applied_configs, skipped_configs, conflict_count, application_summary
   FROM template_applications
   WHERE tenant_id = current_tenant_id()
   ORDER BY applied_at DESC LIMIT 5;
   ```
5. Use `dry_run: true` to preview before applying again

---

## Version Conflict Error (409)

**Symptom**: Update returns `409 version_conflict`.

**Cause**: Another write happened between the read and the write (optimistic locking).

**Fix**: Re-read the configuration, apply your change to the latest value, and retry the write.

```go
for retries := 0; retries < 3; retries++ {
    current, _ := store.GetTenantConfigurations(ctx)
    err := store.UpdateTenantConfigurationSettings(ctx, params)
    if err == nil {
        break
    }
    if !isVersionConflict(err) {
        return err
    }
    // Back off and retry
    time.Sleep(time.Duration(retries+1) * 100 * time.Millisecond)
}
```

---

## Bulk Operation Stuck / Not Completing

**Symptom**: Bulk operation shows `in_progress` indefinitely.

**Checklist**:
1. Check server logs for panics or errors in the batch processor
2. Query `configuration_audit` with the `correlation_id` to see how far it got:
   ```sql
   SELECT COUNT(*), MAX(applied_at) FROM configuration_audit
   WHERE correlation_id = '<bulk-op-id>';
   ```
3. If `continue_on_error: false` was set, one failure may have stopped the whole operation
4. Check Redis memory — the progress key may have expired

---

## Cache Returning Stale Value

**Symptom**: After updating a configuration, the old value is still returned for up to 30 minutes.

**Cause**: Cache was not invalidated after the write, or the write was made directly to the database bypassing the service layer.

**Fix**:
- Always write through the Settings service — never directly to the database
- If needed, force cache invalidation: delete the Redis key `config:{tenantID}:{entityID}:{module}:{key}`
- Check event bus — if Redis Streams are backed up, cache invalidation events may be delayed

---

## Audit Records Missing

**Symptom**: A configuration changed but there's no record in `configuration_audit`.

**Cause**: Either the change was made directly to the database (bypassing the service layer) or the audit write failed and was swallowed.

**Checklist**:
1. Confirm the write went through the Settings service, not a direct SQL update
2. Check service logs for audit write errors
3. Confirm `configuration_audit` table has the correct RLS policies for `application_role`

---

## Performance Issues

**Symptom**: `GetEffectiveConfiguration` takes >100ms.

**Checklist**:
1. Confirm indexes exist: `idx_tenant_configurations_tenant_id`, `idx_entities_tenant_settings`
2. Check if `EXPLAIN ANALYZE` shows a sequential scan on `entities`
3. Enable caching — single key lookups should almost always hit the cache
4. Use batch resolution (`ListTenantEffectiveConfigurations`) instead of N individual calls
5. Check for table bloat on `configuration_audit` — run `VACUUM ANALYZE`

---

[← Common Scenarios](./21-common-scenarios.md) | [Next: Business Rules →](./23-business-rules.md)
