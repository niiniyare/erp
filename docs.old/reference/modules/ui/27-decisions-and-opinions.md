[<-- Back to Index](README.md)

> **📌 DECISION STATUS** (last reviewed 2026-05-18):
> | Decision | Original State | Current State |
> |----------|---------------|---------------|
> | 1: Custom Shell vs AMIS `app` | Custom HTML shell | Still custom HTML — trigger not yet hit |
> | 2: Static JSON vs Go-driven schemas | Static JSON | **Go-driven — trigger was hit** |
> | 3: One Envelope Format | Two formats (fetcher bridges) | **Data APIs standardised on AMIS envelope** |
> | Others | See below | Verify against current code |

## Decisions & Opinions

This document is the architecture decision record for the UI layer. These are not suggestions. They are the decisions made, with the reasoning. Future decisions that contradict these need an explicit reason for the reversal.

---

### Decision 1: Custom Shell Today, AMIS `app` When Nav Needs Go-Driving

**Current state:** Custom HTML shell in `web/index.html`. Menu is a JS array.

**Trigger to migrate:** The moment you need to show/hide nav items based on feature flags or permissions, migrate to AMIS `app` component driven by `/schema/app` from Go.

**Why not now:** The project is early. Static nav is fine. The custom shell gives faster iteration and easier debugging. The migration path is clear and low-risk.

**Why eventually:** Duplicating flag/permission logic between Go and the JS `menuConfig` is a maintenance tax that compounds. Go already knows what the user can see. It should tell the frontend. Full stop.

---

### Decision 2: Static JSON Schemas During Development, Go-Driven for Production

**Current state:** `web/schemas/pages/*.json` — flat JSON files, served as static assets.

**When to switch:** When the first schema needs to vary by feature flag, tenant, or permission.

**How to switch:** Change `SchemaLoader` base URL from `/schemas/` to `/schema/`. Go handlers return the schema. No other changes.

**What NOT to do:** Add if/else branches inside JSON schema files. Schemas are declarative, not imperative. Logic belongs in Go, not in JSON.

---

### Decision 3: One Envelope Format

**Current state:** The backend returns `{success, data, meta}`. The fetcher bridges this to `{status, data}`. Both formats are handled.

**Problem with this:** Two formats means every developer writing a new Go handler must remember which format they are using. It's also a test surface where mismatches cause silent failures.

**Recommendation:** When schemas move to Go-driven (Decision 2), standardise all handlers on the AMIS envelope: `{status: 0, data: ...}`. Remove the bridge translation from the fetcher. One format, one decision, no confusion.

**The transition:** Go handlers for `/schema/*` already return raw JSON. Go handlers for `/api/v1/*` can move to AMIS format incrementally — the fetcher handles both until migration is complete, then the bridge code is deleted.

---

### Decision 4: Dark Mode Is CSS Variables Only — No JavaScript Involvement

The dark mode toggle adds/removes `html.dark`. That is the ONLY thing it does. All visual changes are CSS.

**Why:** JavaScript-driven dark mode (setting inline styles, toggling AMIS props) creates race conditions and flicker. CSS variables propagate to every element including AMIS portals without any JS involvement.

**Rule:** If you are writing JavaScript to change a colour for dark mode, stop. Add a CSS variable.

---

### Decision 5: Permissions Are Go's Problem. AMIS Reads the Result.

Go decides what the user can do. AMIS renders accordingly. Never the other way around.

```markdown
The Go schema handler checks permissions.
It injects { can_create: true, can_approve: false } into page data.
The AMIS schema uses visibleOn: "${can_create}" on the create button.
The button appears or does not appear.

The Go API handler also checks the same permissions.
An attempt to call the create endpoint without permission returns 403.
```

Two layers of enforcement. Frontend shows/hides for UX. Backend enforces for security.

**Never:** Check permissions in the frontend fetcher and refuse to call the API. That is not security — it is theatre. The API must enforce permissions regardless of what the frontend does.

---

### Decision 6: The Data Chain Is Your State Manager — Use It

AMIS has no Redux, no Zustand, no MobX. The data chain IS the state system.

The rule: if two components need to share data, put that data at the closest common ancestor scope.

```markdown
Sibling components need to share selection state?
→ Put it on the parent page data.

A dialog needs data from the row that opened it?
→ The button that opens the dialog is IN the row — row data is in scope. Use ${id}.

A form step needs data from the previous step?
→ wizard component handles this automatically. Previous step data is in scope.
```

When you feel the urge to use `broadcast` actions to pass data between components: stop. You have a scope problem. Restructure the schema so the data is available at the right level.

---

### Decision 7: Mobile Is Not a Bolt-On

The employee persona (Persona 3) is mobile-first. That means the schema must be designed for mobile from the start — not adapted to fit later.

Concrete rules:
- Forms: one column, always. Two-column form layouts break on mobile without exception.
- Tables: show 3–4 columns on mobile. Everything else behind "add column" — not a horizontal scroll.
- Actions: primary action accessible without scrolling. Never hidden below the fold on a phone.
- Modals: full-screen on mobile (`size: "full"` in dialog/drawer).
- Touch targets: every interactive element ≥44px tall in CSS (AMIS respects this for its components).

---

### Decision 8: Charts Always Have `backgroundColor: transparent`

Every chart schema must include:

```json
{ "config": { "backgroundColor": "transparent" } }
```

Without this, ECharts renders a white rectangle in dark mode before AMIS finishes mounting. The user sees a flash. Always transparent.

---

### Decision 9: Empty States Are Not Optional

Every `crud` component must have a designed empty state. "Optional" is not an option. An empty table is not a valid UX state — it looks broken.

There are two kinds of empty:
1. No records exist → explain why, offer the creation action
2. Filters exclude everything → say so, offer "clear filters"

A `crud` with `syncLocation: true` can detect whether filters are active by checking `${__query}` in the data domain. Use this to show the right message.

---

### Decision 10: No Hardcoded Colors in Schemas

`"className": "text-red-500"` in a schema: fine for utility classes.
`"style": "color: #ff0000"`: never. Use a CSS variable or a semantic colour class.

When a tenant changes their brand colour, it should propagate everywhere automatically. That only works if every colour is a variable reference, not a literal value.

---

### Opinion: The Biggest Mistake You Will Make

Building a custom component for something AMIS already does.

Before writing any custom code, spend 30 minutes in the AMIS playground testing whether a standard component achieves what you need. The AMIS component catalog is large. `condition-builder`, `input-kv`, `combo`, `transfer`, `tree-select`, `switch`, `rating`, `json`, `diff`, `tag` — these exist and work well.

The 30 minutes you spend discovering the right built-in component saves 3 days of building and maintaining a custom one.

---
