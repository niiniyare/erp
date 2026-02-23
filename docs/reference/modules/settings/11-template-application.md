# Chapter 11: Template Application & Workflows

[← Template Management](./10-template-management.md) | [Next: Bulk Operations →](./12-bulk-operations.md)

---

## Applying a Template

Template application writes the template's configuration values to the target (tenant or entity), respecting the conflict resolution strategy and recording the operation in the audit trail.

### HTTP API

```http
POST /api/v1/settings/templates/{template_id}/apply
Authorization: Bearer <token>
X-Tenant-ID: <tenant-id>

{
  "target_type": "tenant",
  "target_id": "tenant-uuid",
  "options": {
    "preserve_existing": true,
    "conflict_resolution": "merge",
    "dry_run": false
  },
  "selective_application": {
    "modules": ["finance", "inventory"],
    "exclude_keys": ["finance.bank_account"]
  }
}
```

### Response

```json
{
  "operation_id": "op-uuid-123",
  "status": "completed",
  "summary": {
    "total_configs": 45,
    "applied": 40,
    "skipped": 3,
    "conflicts": 2
  },
  "applied_configurations": [
    { "module": "finance", "config_key": "invoice_prefix",
      "old_value": "INV-", "new_value": "MFG-INV-", "action": "updated" }
  ],
  "conflicts": [
    { "module": "inventory", "config_key": "reorder_threshold",
      "template_value": 10, "existing_value": 15,
      "resolution": "kept_existing", "reason": "preserve_existing_policy" }
  ],
  "applied_at": "2025-09-01T16:00:00Z"
}
```

---

## Application Workflow with Validation

Production template applications should follow a multi-phase workflow to support rollback:

```go
func (t *TemplateApplicator) ApplyTemplateWithValidation(
    ctx context.Context, req TemplateApplicationRequest,
) (*TemplateApplicationResult, error) {

    // Phase 1: Pre-application validation
    if err := t.validateTemplateApplication(ctx, req); err != nil {
        return nil, err
    }

    // Phase 2: Create rollback snapshot
    rollbackPoint, err := t.createRollbackPoint(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create rollback point: %w", err)
    }

    // Phase 3: Apply template
    application, err := t.settingsService.ApplyTemplate(ctx, req)
    if err != nil {
        t.rollback(ctx, rollbackPoint)
        return nil, err
    }

    // Phase 4: Post-application validation
    if err := t.validatePostApplication(ctx, req, application); err != nil {
        t.rollback(ctx, rollbackPoint)
        return nil, err
    }

    // Phase 5: Trigger dependent workflows
    t.triggerDependentWorkflows(ctx, req, application)

    return &TemplateApplicationResult{
        Status:           "completed",
        ApplicationResult: application,
        RollbackPointID:  rollbackPoint.ID,
    }, nil
}
```

---

## Dry Run

Set `dry_run: true` in the request to preview what would be applied without making changes. The response shows the same structure but no database writes occur. Use this to review conflicts before committing.

---

## Selective Application

Use `selective_application` to apply only a subset of a template's configurations:

```json
{
  "selective_application": {
    "modules": ["finance"],           // Only apply finance keys
    "exclude_keys": ["finance.bank_account", "finance.tax_id"]  // Skip sensitive keys
  }
}
```

---

## Industry Template Ordering

Some industries require multiple templates applied in dependency order:

```go
func (t *TemplateApplicator) ApplyIndustryTemplate(
    ctx context.Context, industry IndustryType, tenantID uuid.UUID,
) error {
    templatesByIndustry := map[IndustryType][]string{
        IndustryManufacturing: {
            "base-erp-foundation",       // Priority 1: Foundation
            "manufacturing-core",         // Priority 2: Core
            "inventory-manufacturing",    // Priority 3: Inventory
            "finance-manufacturing",      // Priority 4: Finance
        },
        IndustryRetail: {
            "base-erp-foundation",
            "retail-core",
            "pos-integration",
            "inventory-retail",
            "finance-retail",
        },
    }

    for _, templateID := range templatesByIndustry[industry] {
        if err := t.applyWithValidation(ctx, templateID, tenantID); err != nil {
            return fmt.Errorf("failed at template %s: %w", templateID, err)
        }
    }
    return nil
}
```

---

## Application History

Every template application is recorded in `template_applications`:

```sql
-- name: CreateTemplateApplication :one
INSERT INTO template_applications (
    template_id, tenant_id, entity_id, target_type,
    applied_configs, skipped_configs, conflict_count,
    application_summary, applied_by, correlation_id
) VALUES (...) RETURNING *;
```

And in `tenant_configurations`:

```sql
-- name: UpdateTenantTemplateInfo :exec
UPDATE tenant_configurations
SET
    last_template_applied = $1,
    template_applied_at   = NOW()
WHERE tenant_id = current_tenant_id();
```

Query history with the `GetTemplateApplicationHistory` query, which joins `template_applications` with `configuration_templates` to include template name and version.

---

## Custom Template Creation

Tenants can create their own templates by capturing their current configuration:

```go
func (c *CustomTemplateBuilder) CreateCustomTemplate(
    ctx context.Context, req CustomTemplateRequest,
) (*CustomTemplate, error) {
    // Extract current configs from source entity/tenant
    currentConfigs, _ := c.extractConfigurations(ctx, req.SourceTenantID, req.SourceEntityID)

    // Filter to relevant modules/keys
    filteredConfigs := c.filterConfigurations(currentConfigs, req.IncludeModules, req.ExcludeKeys)

    // Save as TENANT-scope template
    return c.templateService.CreateTemplate(ctx, &domain.Template{
        Scope:              "TENANT",
        Name:               req.Name,
        Category:           "custom",
        Version:            "1.0",
        Configurations:     c.buildTemplateConfigurations(filteredConfigs),
        ConflictResolution: req.ConflictResolution,
    })
}
```

---

[← Template Management](./10-template-management.md) | [Next: Bulk Operations →](./12-bulk-operations.md)
