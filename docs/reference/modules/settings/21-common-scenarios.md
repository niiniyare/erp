# Chapter 21: Common Configuration Scenarios

[← Search & Analysis](./20-search-and-analysis.md) | [Next: Troubleshooting →](./22-troubleshooting.md)

---

## Scenario 1: Onboarding a New Manufacturing Tenant

**Goal**: A new manufacturing company signs up and needs standard manufacturing configuration applied immediately.

**Steps**:

1. Create tenant (Tenant module)
2. Apply base ERP foundation template
3. Apply manufacturing industry template
4. Apply finance-manufacturing template

```http
POST /api/v1/settings/templates/base-erp-foundation/apply
{ "target_type": "tenant", "target_id": "new-tenant-uuid" }

POST /api/v1/settings/templates/manufacturing-core/apply
{ "target_type": "tenant", "target_id": "new-tenant-uuid" }

POST /api/v1/settings/templates/finance-manufacturing/apply
{
  "target_type": "tenant",
  "target_id": "new-tenant-uuid",
  "options": { "conflict_resolution": "merge", "preserve_existing": true }
}
```

**Result**: Tenant gets ~90 standard configurations applied in seconds, with full audit trail.

---

## Scenario 2: Opening a New Branch

**Goal**: A retail chain opens a new NYC branch that needs branch-specific invoice prefixes and a lower approval limit than the tenant default.

```http
# Set branch-specific invoice prefix
PUT /api/v1/settings/config/finance/invoice_prefix
{
  "value": "NYC-INV-",
  "target_type": "entity",
  "target_id": "nyc-branch-uuid"
}

# Set branch-specific approval limit
PUT /api/v1/settings/config/finance/auto_approval_limit
{
  "value": 500,
  "target_type": "entity",
  "target_id": "nyc-branch-uuid"
}
```

The branch inherits all other configurations from the tenant level.

---

## Scenario 3: Rolling Out a New Approval Limit Org-Wide

**Goal**: Finance director raises the auto-approval limit from $1,000 to $2,500 for all branches except headquarters.

**Step 1**: Update tenant-level default:

```http
PUT /api/v1/settings/config/finance/auto_approval_limit
{
  "value": 2500,
  "target_type": "tenant",
  "target_id": "tenant-uuid"
}
```

All branches without an explicit override now resolve $2,500. Headquarters already has a $10,000 entity override — it is unaffected.

---

## Scenario 4: Bulk Update — Enable Reorder for Retail Locations

**Goal**: Enable inventory auto-reorder for all 150 retail locations (tagged "retail").

```http
POST /api/v1/settings/bulk-update
{
  "targets": {
    "type": "entity",
    "filters": { "tags": ["retail"] }
  },
  "configurations": [
    { "module": "inventory", "config_key": "auto_reorder_enabled", "value": true }
  ],
  "options": {
    "preserve_explicit_overrides": false,
    "batch_size": 50,
    "continue_on_error": true
  }
}
```

Poll `/api/v1/settings/operations/{id}` until complete.

---

## Scenario 5: Investigating a Configuration Change

**Goal**: A branch manager reports that invoice numbers changed format unexpectedly. Investigate when and who changed it.

```http
GET /api/v1/settings/config/finance/invoice_prefix
  ?entity_id=nyc-branch-uuid
  &include_metadata=true
```

Then query audit history:

```sql
SELECT module_name, config_key_name, old_value, new_value, user_id, applied_at
FROM configuration_audit
WHERE tenant_id = current_tenant_id()
  AND module_name = 'finance'
  AND config_key_name = 'invoice_prefix'
  AND entity_id = 'nyc-branch-uuid'
ORDER BY applied_at DESC;
```

---

## Scenario 6: Creating a Custom Branch Template

**Goal**: After configuring a model branch, save that configuration as a reusable template for future branches.

```go
// Extract current model branch configuration
currentConfigs, _ := extractConfigurations(ctx, tenantID, modelBranchID)

// Filter to relevant keys
filteredConfigs := filter(currentConfigs, []string{"finance", "inventory"}, []string{"finance.bank_account"})

// Save as TENANT template
store.CreateConfigurationTemplate(ctx, db.CreateConfigurationTemplateParams{
    TenantID:           &tenantID,
    Scope:              "TENANT",
    Name:               "Retail Branch Standard v2",
    Category:           "operational",
    Version:            "2.0",
    Configurations:     filteredConfigsJSON,
    ConflictResolution: "preserve",
    CreatedBy:          adminUserID,
})
```

---

## Scenario 7: Switching Inventory Valuation Method

**Goal**: Accounting department wants to switch from FIFO to AVCO for two specific warehouses.

```http
PUT /api/v1/settings/config/inventory/default_valuation_method
{
  "value": "AVCO",
  "target_type": "entity",
  "target_id": "warehouse-1-uuid"
}

PUT /api/v1/settings/config/inventory/default_valuation_method
{
  "value": "AVCO",
  "target_type": "entity",
  "target_id": "warehouse-2-uuid"
}
```

The inventory service receives `ConfigurationChangedEvent` for each warehouse and schedules a forward-looking valuation recalculation.

---

[← Search & Analysis](./20-search-and-analysis.md) | [Next: Troubleshooting →](./22-troubleshooting.md)
