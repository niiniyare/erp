[<-- Back to Index](README.md)

## Known Issues & Workarounds

### Issue 1 — `condition-builder` Output Format Mismatch

The `condition-builder` component outputs its own JSON structure that does not map 1:1 to `pkg/condition.ConditionGroup`. The structures are close but field name casing and nesting depth differ.

**Do not change either side.** Write a converter at the boundary:

```go
// awo/web/converters/conditions.go

// AmisConditionToPkg converts amis condition-builder JSON output
// to pkg/condition ConditionGroup for storage.
func AmisConditionToPkg(raw map[string]any) (*condition.ConditionGroup, error) {
    data, err := json.Marshal(raw)
    if err != nil {
        return nil, err
    }
    var group condition.ConditionGroup
    return &group, json.Unmarshal(data, &group)
}
```

If field names diverge over time, handle remapping in this converter — not in the schema or the domain package.

### Issue 2 — No Native WebSocket / SSE

AMIS polls APIs on a timer. For real-time features (payroll run progress, import status), use polling with `interval`:

```json
{
  "type":          "service",
  "api":           "get:/api/v1/payroll/runs/${id}/status",
  "interval":      3000,
  "silentPolling": true,
  "stopAutoRefreshWhen": "${status === 'completed' || status === 'failed'}",
  "body": {
    "type":   "progress",
    "source": "${progress_pct}",
    "label":  "${status_message}"
  }
}
```

`stopAutoRefreshWhen` stops polling when the terminal state is reached — prevents unnecessary requests after completion.

### Issue 3 — Large Forms (50+ Fields) Performance

AMIS renders all form fields on mount. A form with 50+ fields becomes sluggish.

**Solution:** Split into tabbed forms, one form per tab with its own `initApi`:

```json
{
  "type": "tabs",
  "tabs": [
    {
      "title": "Basic Info",
      "body": {
        "type":    "form",
        "initApi": "get:/api/v1/employees/${id}/basic",
        "body":    ["...10 fields..."]
      }
    },
    {
      "title": "Contract",
      "body": {
        "type":    "form",
        "initApi": "get:/api/v1/employees/${id}/contract",
        "body":    ["...10 fields..."]
      }
    }
  ]
}
```

Each tab loads its own data only when the user clicks on it. Saves initial render time and avoids one giant API call.

### Issue 4 — AMIS Documentation Is Primarily Chinese

The authoritative docs at `aisuda.bce.baidu.com/amis/zh-CN` are in Chinese. The English version lags by months and is incomplete for advanced features.

**Practical strategies:**
- Browser auto-translate is sufficient for reading Chinese docs — JSON examples are universal
- The GitHub TypeScript source is the most reliable reference for prop types: `github.com/baidu/amis`
- The interactive playground at `aisuda.bce.baidu.com/amis/zh-CN/examples/index` lets you test JSON schemas live without any setup
- The visual editor at `aisuda.bce.baidu.com/amis-editor-demo` generates JSON you can copy directly

### Issue 5 — AMIS Portals Escape Dark Mode

AMIS appends dropdowns, date pickers, modals, and tooltips directly to `<body>` — outside your `#content` div. CSS variables cascade correctly from `html.dark`, but some portal containers need explicit backup rules.

See [§05 Dark Mode](./05-dark-mode.md) for the full list of portal selectors that need explicit dark rules.

**Rule of thumb:** Any time you add a new component that renders a popup or overlay, check it in dark mode. If it appears light, add an explicit rule:

```css
html.dark .cxd-YourNewComponent-popover {
  background: var(--bg-surface);
  border-color: var(--borderColor);
  color: var(--text-color);
}
```

---
