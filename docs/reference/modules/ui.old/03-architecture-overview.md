[<-- Back to Index](README.md)

## Architecture Overview

### Full Request Flow

```markdown
USER OPENS BROWSER → GET / → Go serves web/index.html
        │
        ▼
Browser parses index.html:
├── Loads /sdk/sdk.js  (AMIS runtime, ~2MB)
├── Loads /sdk/sdk.css (AMIS styles)
├── Loads /sdk/charts.js (ECharts)
├── Loads /utils/schema-loader.js (custom)
└── Renders shell: sidebar + topbar

        │
        ▼
Shell JS runs:
├── Reads localStorage for theme → applies html.dark if needed
├── Renders sidebar from menuConfig
├── Detects URL hash (e.g. #dashboard)
└── Calls navigate('#dashboard')

        │
        ▼
navigate() function:
├── Updates active sidebar item
├── Updates breadcrumb
└── Loads schema from web/schemas/pages/dashboard.json

        │
        ▼
AMIS embed(#content, schema, {}, amisEnv):
├── Renders AMIS components defined in the schema
├── Each component with an api: field calls the Go backend
└── AMIS manages all data state, form state, filter state locally

        │
        ▼
AMIS fetcher → /api/v1/...  (Go backend)
├── Translates page/perPage → offset/limit
├── Translates backend envelope {success, data, meta}
└── Returns AMIS envelope {status: 0, data: {items, count}}
```

### File Roles

```markdown
web/
├── index.html          ← The entire shell: CSS + HTML + JS in one file
├── schemas/
│   └── pages/
│       ├── dashboard.json   ← AMIS schema for each page
│       ├── users.json
│       └── ...
├── sdk/
│   ├── sdk.js          ← AMIS runtime (pre-built, ~2MB)
│   ├── sdk.css         ← AMIS base styles
│   ├── charts.js       ← ECharts (for chart component)
│   └── rest.js         ← AMIS REST adapter
├── utils/
│   ├── schema-loader.js ← Fetches + caches JSON schema files
│   └── locale-en.js     ← AMIS English locale strings
└── public/             ← Static assets (favicon, logo, etc.)
```

### Two Envelope Formats — The Bridge

The Go backend uses its own envelope. AMIS expects its own. The `fetcher` function bridges them:

```markdown
GO BACKEND RESPONSE:
{
  "success": true,
  "data": [...],
  "meta": {
    "pagination": {
      "total_records": 142,
      "offset": 0,
      "limit": 20
    }
  }
}

AMIS FETCHER TRANSLATES TO:
{
  "status": 0,
  "data": {
    "items": [...],
    "count": 142
  }
}

AMIS CRUD THEN:
- Reads items[] → renders table rows
- Reads count → calculates page count
- Sends page/perPage in next request
  → fetcher translates back to offset/limit
```

### Static Schemas vs Go-Driven Schemas

```markdown
CURRENT (static JSON files):
navigate('#users')
└── loads /schemas/pages/users.json (static file in web/)
    └── AMIS renders using schema in that file

FUTURE (Go-driven, when flags/tenants need it):
navigate('#users')
└── calls GET /schema/users (Go handler)
    └── Go returns schema filtered by feature flags + permissions
        └── AMIS renders using returned schema

THE TRANSITION:
Only the SchemaLoader URL changes:
  var loader = new SchemaLoader('/schemas/');        // static
  var loader = new SchemaLoader('/schema/');         // Go-driven
No other code changes needed.
```

---
