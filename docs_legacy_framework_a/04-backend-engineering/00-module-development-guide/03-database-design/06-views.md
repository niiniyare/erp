> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Views and Materialised Views
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./05-child-tables.md
    title: Child Tables
  - path: ../04-sqlc-queries/01-query-overview.md
    title: SQLC Query Overview
---

# Views and Materialised Views

The third migration file creates read-only views that simplify complex reporting queries. Views are optional — only create them when they provide clear value to SQLC queries or external tools.

## Complete Migration: 011003_create_contracts_views.sql

```sql
-- +migrate Up

-- ============================================================
-- v_contract_summary
-- ============================================================
-- Aggregates contract status counts per tenant for dashboard widgets.
-- Used by the contracts list header to show "5 pending approval" badges.

CREATE VIEW v_contract_summary AS
SELECT
    tenant_id,
    entity_id,
    status,
    COUNT(*)                          AS contract_count,
    COALESCE(SUM(total_value), 0)     AS total_value,
    currency,
    MIN(start_date)                   AS earliest_start,
    MAX(end_date)                     AS latest_end
FROM contracts
WHERE deleted_at IS NULL
GROUP BY tenant_id, entity_id, status, currency;

COMMENT ON VIEW v_contract_summary IS 'Dashboard summary: contract counts and total values grouped by tenant, entity, status, and currency.';

-- ============================================================
-- v_contracts_with_lines
-- ============================================================
-- Pre-joins contracts with their line counts.
-- Used by list views that show line count in the table.

CREATE VIEW v_contracts_with_lines AS
SELECT
    c.*,
    COALESCE(l.line_count, 0)   AS line_count,
    COALESCE(l.active_lines, 0) AS active_lines
FROM contracts c
LEFT JOIN (
    SELECT
        contract_id,
        COUNT(*)                                         AS line_count,
        COUNT(*) FILTER (WHERE status = 'active')        AS active_lines
    FROM contract_lines
    WHERE deleted_at IS NULL
    GROUP BY contract_id
) l ON l.contract_id = c.id
WHERE c.deleted_at IS NULL;

COMMENT ON VIEW v_contracts_with_lines IS 'Contracts with aggregated line counts. Use for list views that need line count without fetching lines.';

-- ============================================================
-- mv_contract_monthly_totals
-- ============================================================
-- Materialised view: monthly contract value totals per tenant.
-- Refreshed on demand (scheduled or triggered after contract activations).
-- Used by finance dashboard charts.

CREATE MATERIALIZED VIEW mv_contract_monthly_totals AS
SELECT
    tenant_id,
    entity_id,
    currency,
    DATE_TRUNC('month', start_date) AS month,
    COUNT(*)                         AS contract_count,
    SUM(total_value)                 AS total_value
FROM contracts
WHERE deleted_at IS NULL
  AND status = 'active'
GROUP BY tenant_id, entity_id, currency, DATE_TRUNC('month', start_date)
WITH DATA;

COMMENT ON MATERIALIZED VIEW mv_contract_monthly_totals IS 'Monthly contract value totals for active contracts. Refresh after bulk activations.';

-- Index on the materialised view for tenant + month queries
CREATE UNIQUE INDEX idx_mv_contract_monthly_tenant_month
    ON mv_contract_monthly_totals (tenant_id, entity_id, currency, month);

-- +migrate Down

DROP MATERIALIZED VIEW IF EXISTS mv_contract_monthly_totals;
DROP VIEW IF EXISTS v_contracts_with_lines;
DROP VIEW IF EXISTS v_contract_summary;
```

## Views vs Materialised Views

| Feature | View | Materialised View |
|---------|------|-------------------|
| Data freshness | Always current (re-queries on access) | Snapshot at last refresh |
| Performance | As slow as the underlying query | Fast (reads cached result) |
| Storage | None | Disk space for cached result |
| Index support | No | Yes — can be indexed |
| Refresh | Automatic | Manual: `REFRESH MATERIALIZED VIEW CONCURRENTLY ...` |
| RLS | Inherited from underlying tables | Does not inherit — add separate policy |

Use views for:
- Joins used in multiple SQLC queries (avoid repeating the join)
- Filters applied everywhere (e.g., always `deleted_at IS NULL`)

Use materialised views for:
- Expensive aggregations (count, sum across large tables)
- Dashboard widgets that tolerate slight staleness
- Cross-table reports that would otherwise require multiple round-trips

## RLS on Materialised Views

Regular views inherit RLS from their underlying tables. Materialised views do not — they are independent tables of cached data. Add an explicit RLS policy:

```sql
ALTER MATERIALIZED VIEW mv_contract_monthly_totals ENABLE ROW LEVEL SECURITY;

CREATE POLICY mv_contract_monthly_totals_isolation ON mv_contract_monthly_totals
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

## Refreshing Materialised Views

Refresh after bulk operations:

```go
// In the service, after activating contracts in bulk:
func (s *contractService) BulkActivate(ctx context.Context, ...) error {
    // ... activate contracts ...

    // Schedule or trigger MV refresh (fire-and-forget)
    go func() {
        if _, err := s.store.ExecRaw(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_contract_monthly_totals"); err != nil {
            s.logger.Warn().Err(err).Msg("failed to refresh mv_contract_monthly_totals")
        }
    }()

    return nil
}
```

`CONCURRENTLY` allows reads during refresh. Without it, the view is locked for reads during the refresh — acceptable for small views during off-peak hours, not for large views in production.

## Naming Conventions

| Object | Prefix | Example |
|--------|--------|---------|
| View | `v_` | `v_contract_summary` |
| Materialised view | `mv_` | `mv_contract_monthly_totals` |
| View index | `idx_mv_<view>_<columns>` | `idx_mv_contract_monthly_tenant_month` |
