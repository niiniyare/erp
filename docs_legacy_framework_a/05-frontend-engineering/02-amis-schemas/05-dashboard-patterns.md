> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Dashboard Patterns
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[AMIS Schema Overview](01-amis-schema-overview.md)"
  - "[CRUD Patterns](02-crud-patterns.md)"
  - "[Dark Mode](../03-theming/01-dark-mode.md)"
---

# Dashboard Patterns

Dashboards use stat cards, charts, and filtered lists. All data is loaded from API endpoints.

## Stat Cards Row

```json
{
  "type": "grid",
  "columns": [
    {
      "md": 3,
      "body": {
        "type": "service",
        "api": "GET /api/v1/contracts/stats",
        "body": {
          "type": "stat",
          "title": "Active Contracts",
          "value": "${active_count}",
          "icon": "fa fa-file-contract",
          "tpl": "${active_count}",
          "subValue": "${active_change_pct}%",
          "subTpl": "${active_change_pct >= 0 ? '+' : ''}${active_change_pct}% vs last month"
        }
      }
    },
    {
      "md": 3,
      "body": {
        "type": "service",
        "api": "GET /api/v1/contracts/stats",
        "body": {
          "type": "stat",
          "title": "Pending Review",
          "value": "${pending_count}",
          "icon": "fa fa-clock",
          "level": "${pending_count > 10 ? 'warning' : 'info'}"
        }
      }
    },
    {
      "md": 3,
      "body": {
        "type": "service",
        "api": "GET /api/v1/contracts/stats",
        "body": {
          "type": "stat",
          "title": "Expiring (30d)",
          "value": "${expiring_soon_count}",
          "icon": "fa fa-exclamation-triangle",
          "level": "${expiring_soon_count > 0 ? 'danger' : 'success'}"
        }
      }
    },
    {
      "md": 3,
      "body": {
        "type": "service",
        "api": "GET /api/v1/finance/stats",
        "body": {
          "type": "stat",
          "title": "Total Contract Value",
          "value": "${total_value_formatted}",
          "icon": "fa fa-dollar-sign"
        }
      }
    }
  ]
}
```

## Bar Chart — Contracts by Status

```json
{
  "type": "panel",
  "title": "Contracts by Status",
  "body": {
    "type": "chart",
    "api": "GET /api/v1/contracts/stats/by-status",
    "config": {
      "xAxis": {
        "type": "category",
        "data": "${data.labels}"
      },
      "yAxis": { "type": "value" },
      "series": [
        {
          "type": "bar",
          "data": "${data.values}",
          "itemStyle": {
            "color": {
              "type": "function",
              "fn": "function(params) { var colors = {draft:'#8c8c8c',submitted:'#1890ff',under_review:'#faad14',approved:'#52c41a',active:'#13c2c2',terminated:'#ff4d4f'}; return colors[params.name] || '#1890ff'; }"
            }
          }
        }
      ],
      "tooltip": { "trigger": "axis" }
    }
  }
}
```

## Line Chart — Monthly Activity

```json
{
  "type": "panel",
  "title": "Contract Activity (12 months)",
  "body": {
    "type": "chart",
    "api": "GET /api/v1/contracts/stats/monthly?months=12",
    "config": {
      "xAxis": {
        "type": "category",
        "data": "${data.months}"
      },
      "yAxis": { "type": "value" },
      "legend": { "data": ["Created", "Approved", "Terminated"] },
      "series": [
        {
          "name": "Created",
          "type": "line",
          "data": "${data.created}",
          "smooth": true
        },
        {
          "name": "Approved",
          "type": "line",
          "data": "${data.approved}",
          "smooth": true
        },
        {
          "name": "Terminated",
          "type": "line",
          "data": "${data.terminated}",
          "smooth": true
        }
      ],
      "tooltip": { "trigger": "axis" }
    }
  }
}
```

## Pie Chart — Contracts by Type

```json
{
  "type": "panel",
  "title": "Contracts by Type",
  "body": {
    "type": "chart",
    "api": "GET /api/v1/contracts/stats/by-type",
    "config": {
      "series": [
        {
          "type": "pie",
          "radius": ["40%", "70%"],
          "data": "${data.items}",
          "label": {
            "formatter": "{b}: {d}%"
          }
        }
      ],
      "tooltip": {
        "trigger": "item",
        "formatter": "{b}: {c} ({d}%)"
      },
      "legend": { "orient": "vertical", "right": 10 }
    }
  }
}
```

## Recent Activity Feed

```json
{
  "type": "panel",
  "title": "Recent Activity",
  "body": {
    "type": "service",
    "api": "GET /api/v1/audit-log?limit=10&resource_type=contract",
    "body": {
      "type": "timeline",
      "items": "${items}",
      "itemTpl": "<div><strong>${action | replace:'contracts.contract.':''}</strong> — ${resource_id_short} <span style='color:#8c8c8c;float:right'>${created_at | date:'HH:mm'}</span><br><small style='color:#8c8c8c'>${actor_name}</small></div>"
    }
  }
}
```

## Quick Filter + List

Dashboard with filter bar above a compact list:

```json
{
  "type": "panel",
  "title": "Contracts Requiring Action",
  "body": {
    "type": "crud",
    "api": "GET /api/v1/contracts?status=under_review",
    "syncLocation": false,
    "perPage": 5,
    "footerToolbar": ["pagination"],
    "columns": [
      {
        "name": "contract_number",
        "label": "Contract",
        "type": "link",
        "href": "/contracts/${id}"
      },
      { "name": "vendor_name", "label": "Vendor" },
      {
        "name": "submitted_at",
        "label": "Waiting Since",
        "tpl": "${submitted_at | date:'YYYY-MM-DD'}"
      },
      {
        "name": "total_value",
        "label": "Value",
        "tpl": "${currency} ${total_value}"
      },
      {
        "type": "operation",
        "label": "Actions",
        "buttons": [
          {
            "label": "Review",
            "size": "sm",
            "actionType": "link",
            "link": "/contracts/${id}"
          }
        ]
      }
    ]
  }
}
```

## Date Range Filter on Dashboard

```json
{
  "type": "form",
  "wrapWithPanel": false,
  "mode": "inline",
  "target": "stats-service",
  "body": [
    {
      "type": "input-date-range",
      "name": "date_range",
      "label": "Period",
      "value": "thisMonth",
      "format": "YYYY-MM-DD",
      "inputFormat": "YYYY-MM-DD"
    },
    {
      "type": "select",
      "name": "entity_id",
      "label": "Division",
      "clearable": true,
      "source": "GET /api/v1/entities?type=division",
      "labelField": "name",
      "valueField": "id"
    },
    {
      "type": "submit",
      "label": "Apply",
      "level": "primary"
    }
  ]
}
```

## Full Dashboard Page Assembly

```json
{
  "type": "page",
  "title": "Contracts Dashboard",
  "body": [
    {
      "type": "grid",
      "className": "mb-4",
      "columns": [
        { "md": 3, "body": { "/* stat card 1 */": "" } },
        { "md": 3, "body": { "/* stat card 2 */": "" } },
        { "md": 3, "body": { "/* stat card 3 */": "" } },
        { "md": 3, "body": { "/* stat card 4 */": "" } }
      ]
    },
    {
      "type": "grid",
      "columns": [
        { "md": 8, "body": { "/* monthly line chart */": "" } },
        { "md": 4, "body": { "/* by-type pie chart */": "" } }
      ]
    },
    {
      "type": "grid",
      "columns": [
        { "md": 7, "body": { "/* pending review list */": "" } },
        { "md": 5, "body": { "/* recent activity feed */": "" } }
      ]
    }
  ]
}
```

## Dashboard Checklist

- Each stat card uses own `service` component — independent refresh, isolated errors
- Charts: always provide `api` not inline `data` — dashboards must reflect live data
- `syncLocation: false` on all dashboard CRUDs — prevent URL pollution
- Date range filter: use `target` pointing to service `id` to refresh on submit
- Avoid loading all data in one giant API call — parallel smaller calls are faster and more resilient
- ECharts dark mode: apply `darkTheme` in `window.EChartsConfig` — see Dark Mode doc
