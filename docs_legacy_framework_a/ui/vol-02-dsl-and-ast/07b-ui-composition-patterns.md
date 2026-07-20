> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "UI Composition Patterns"
volume: "II — DSL & AST"
chapter: "07-B"
phase: 1
status: draft
audience: [Platform Engineers, Backend Engineers]
---

# Chapter 07-B — UI Composition Patterns

> **Volume:** II — DSL & AST
> **Audience:** Platform Engineers, Backend Engineers
> **Prerequisites:** Chapter 06 — UI DSL Architecture, Chapter 07 — AST Design
> **Phase:** 1

---

## 07B.1 Why Composition Matters in SDUI

In a traditional client-rendered application, UI composition happens in framework code — React components, Flutter widgets, Angular templates. The server sends data; the client assembles the view. In AwoERP's server-driven UI model, composition must happen on the server, inside the Go backend, before the compiled AST is sent to any renderer. This inversion places a new burden on the platform: the server must support the same compositional expressiveness that client frameworks provide — templates, slots, inheritance, mixins, fragments — without the luxury of a component tree evaluated at render time.

The stakes are high. AwoERP serves dozens of ERP modules (Finance, Procurement, HR, Fuel Retail, Inventory) across hundreds of tenants, each with custom branding, enabled-module sets, and per-role surface variants. Without a principled composition system, every surface definition becomes a copy-paste monolith. A change to the shared invoice header must be propagated to 47 places. A new tenant requires writing 300 surface definitions from scratch. The composition system exists to prevent these failure modes.

The composition system in AwoERP operates at the **AST level**, not at the JSON text level. Templates, slots, and mixins are resolved during the compilation phase (see Chapter 08 — JSON Compilation Pipeline), producing a flattened, renderer-ready AST. No renderer (amis, Flutter) ever receives an unresolved template reference. This keeps renderers simple and ensures that composition logic is tested once, centrally, in Go.

> See Chapter 08 — JSON Compilation Pipeline for how the compilation phase resolves these constructs.

The design philosophy follows three principles:
1. **No runtime composition in the browser.** All resolution happens server-side.
2. **Tenant overrides are additive, not destructive.** A tenant can extend a template but cannot break the platform's base definition.
3. **Composition is auditable.** Every resolved node in the AST carries provenance metadata so a developer can trace which template contributed each part of a surface.

---

## 07B.2 Template System

### 07B.2.1 Named Templates and Template Registry

A **named template** is a reusable AST sub-tree registered under a unique string key in the platform's template registry. Templates are defined in Go as `UITemplate` structs and registered at Wire wiring time via the `TemplateRegistry`. No template can be defined at runtime — all templates are compile-time artefacts, versioned alongside the Go codebase.

```go
// internal/ui/templates/registry.go

// UITemplate defines a reusable AST sub-tree.
type UITemplate struct {
    Key         string            // e.g. "finance.invoice-header"
    Version     int               // Monotonically increasing
    Description string
    Parameters  []TemplateParam   // Typed parameters with defaults
    Sealed      bool              // If true, tenants cannot extend
    Body        ast.Node          // The AST sub-tree
}

// TemplateRegistry holds all registered templates.
type TemplateRegistry struct {
    mu        sync.RWMutex
    templates map[string]*UITemplate
}

func (r *TemplateRegistry) Register(t *UITemplate) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, exists := r.templates[t.Key]; exists {
        return fmt.Errorf("template %q already registered", t.Key)
    }
    r.templates[t.Key] = t
    return nil
}

func (r *TemplateRegistry) Lookup(key string) (*UITemplate, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    t, ok := r.templates[key]
    return t, ok
}
```

Templates are registered in module-specific Wire providers. For example, the Finance module registers its invoice header template during application startup:

```go
// internal/core/finance/ui/templates.go

func ProvideFinanceTemplates(reg *ui.TemplateRegistry) error {
    return reg.Register(&ui.UITemplate{
        Key:         "finance.invoice-header",
        Version:     2,
        Description: "Standard invoice header with tenant logo and KRA eTIMS fields",
        Sealed:      false,
        Parameters: []ui.TemplateParam{
            {Name: "show_etims_qr", Type: "bool", Default: true},
            {Name: "logo_position", Type: "enum", Options: []string{"left", "center"}, Default: "left"},
        },
        Body: ast.Node{
            Type: "container",
            Props: map[string]any{
                "className": "invoice-header",
            },
            Children: []ast.Node{
                {Type: "image", Props: map[string]any{"src": "{{tenant.logo_url}}", "className": "invoice-logo"}},
                {Type: "tpl", Props: map[string]any{"tpl": "{{tenant.legal_name}}"}},
                // ... more nodes
            },
        },
    })
}
```

The template registry is a singleton provided via Wire to the UI compilation service. At startup, all modules register their templates before any request is served. This means template resolution is O(1) hash lookup — no database query, no network round-trip.

> **✅ Convention:** Template keys use the format `<module>.<template-name>`. Cross-module templates owned by the platform use `platform.<template-name>`. Never use generic keys like `"header"` — namespace collisions are a CI failure.

### 07B.2.2 Template Parameters and Defaults

Templates accept typed parameters that control their rendered output. Parameters are evaluated during compilation with the following type system:

| Type | Go equivalent | Validation |
|------|--------------|------------|
| `string` | `string` | Max 512 chars |
| `bool` | `bool` | — |
| `int` | `int64` | Range optional |
| `enum` | `string` | Must be in `Options` list |
| `node` | `ast.Node` | Valid AST node |
| `nodes` | `[]ast.Node` | Slice of AST nodes |

Parameters with defaults are optional at the call site. A template call without a required parameter fails compilation with a descriptive error, not a runtime panic.

```json
{
  "type": "template_ref",
  "template": "finance.invoice-header",
  "params": {
    "show_etims_qr": true,
    "logo_position": "left"
  }
}
```

During compilation, the `template_ref` node is replaced with the template body, with parameter tokens resolved:

```go
func (c *Compiler) resolveTemplateRef(node ast.Node, ctx *CompileContext) (ast.Node, error) {
    key, _ := node.Props["template"].(string)
    tmpl, ok := c.registry.Lookup(key)
    if !ok {
        return ast.Node{}, fmt.Errorf("unknown template %q referenced in surface %q", key, ctx.SurfaceID)
    }

    params, _ := node.Props["params"].(map[string]any)
    resolved, err := tmpl.Bind(params, ctx)
    if err != nil {
        return ast.Node{}, fmt.Errorf("template %q bind error: %w", key, err)
    }

    // Stamp provenance metadata onto resolved root
    resolved.Meta["composed_from"] = key
    resolved.Meta["template_version"] = tmpl.Version
    return resolved, nil
}
```

### 07B.2.3 Template Versioning

Templates carry an integer `Version` field. This enables:

1. **Backwards compatibility detection.** If a tenant has saved a surface definition that calls `finance.invoice-header` at version 1, and the platform ships version 2 with a breaking parameter change, the compilation service can detect the mismatch and either auto-migrate or surface a warning.
2. **Audit trail.** The compiled AST metadata records which template version was used, enabling incident investigation ("the invoice layout changed because the template was upgraded from v1 to v2 on 2026-05-12").
3. **Rollback.** The registry can maintain the last N versions, allowing a tenant admin to pin to a previous version while migration work is completed.

> **⚠ Warning:** Never remove a template version that is still referenced by any stored surface definition. Run `awo ui validate --all-tenants` before removing old template versions to check for active references.

---

## 07B.3 Slot-Based Composition

Slots are named insertion points within a template where the caller can inject custom content. Slots make templates extensible without requiring template parameters for every possible customisation.

### 07B.3.1 Named Slots

A template declares a slot using the `slot` node type with a required `name` attribute:

```json
{
  "type": "container",
  "className": "invoice-body",
  "children": [
    { "type": "template_ref", "template": "finance.invoice-line-items" },
    {
      "type": "slot",
      "name": "footer_extras",
      "props": {
        "description": "Optional content below line items (e.g. terms, bank details)"
      }
    },
    { "type": "template_ref", "template": "finance.invoice-totals" }
  ]
}
```

The caller fills the slot by providing a `slots` map in the `template_ref` call:

```json
{
  "type": "template_ref",
  "template": "finance.invoice-body",
  "slots": {
    "footer_extras": {
      "type": "container",
      "children": [
        { "type": "tpl", "tpl": "Payment via M-Pesa Paybill: <strong>{{tenant.mpesa_paybill}}</strong>" },
        { "type": "tpl", "tpl": "Account Number: <strong>{{invoice.number}}</strong>" }
      ]
    }
  }
}
```

This pattern is used extensively in AwoERP's Kenya-specific print templates, where the M-Pesa payment details block is injected into the standard invoice template as a slot fill — keeping the platform invoice template clean of Kenya-specific content while still rendering correctly for Kenyan tenants.

### 07B.3.2 Default Slot Content (Fallbacks)

A slot can declare default content that is used when the caller provides no fill for that slot. This is critical for the platform's progressive enhancement model — base templates work correctly for all tenants even without customisation.

```json
{
  "type": "slot",
  "name": "footer_extras",
  "default": {
    "type": "tpl",
    "tpl": "Thank you for your business."
  }
}
```

During compilation, if no slot fill is provided, the `default` node is inlined at the slot position. If neither a fill nor a default is present, the slot resolves to an empty node (zero-height, no DOM element in the amis renderer).

### 07B.3.3 Slot Override Rules (Tenant vs. User)

Slots have a priority hierarchy for overrides. When multiple sources provide content for the same slot, the following precedence applies (highest wins):

1. **User-specific override** (saved by a user with `ui.surfaces.customize` permission)
2. **Tenant-specific override** (configured in tenant settings)
3. **Module-provided default** (in the surface definition)
4. **Template default** (in the template declaration)

This hierarchy is resolved at compile time by the `SlotResolver`:

```go
// internal/ui/compiler/slot_resolver.go

type SlotResolver struct {
    tenantOverrides map[string]ast.Node  // loaded from PostgreSQL
    userOverrides   map[string]ast.Node  // loaded from PostgreSQL (if user has permission)
}

func (r *SlotResolver) Resolve(slotName string, callSiteDefault *ast.Node, templateDefault *ast.Node) ast.Node {
    if node, ok := r.userOverrides[slotName]; ok {
        return node
    }
    if node, ok := r.tenantOverrides[slotName]; ok {
        return node
    }
    if callSiteDefault != nil {
        return *callSiteDefault
    }
    if templateDefault != nil {
        return *templateDefault
    }
    return ast.EmptyNode()
}
```

> **⚠ Warning:** User overrides are only resolved when the session has the `ui.surfaces.customize` permission. Never load user override data for sessions without this permission, even if the data exists in the database.

### 07B.3.4 Recursive Slot Nesting

Slots can be nested — a slot fill can itself contain `template_ref` nodes with their own slots. The compiler resolves these recursively. Maximum recursion depth is 8 levels, enforced at compile time to prevent circular composition and stack overflows.

```go
const maxCompositionDepth = 8

func (c *Compiler) resolveNode(node ast.Node, ctx *CompileContext, depth int) (ast.Node, error) {
    if depth > maxCompositionDepth {
        return ast.Node{}, fmt.Errorf("composition depth exceeded at node type %q (surface: %q)", node.Type, ctx.SurfaceID)
    }
    // ... resolution logic
}
```

---

## 07B.4 Inheritance and Extension

Inheritance allows a surface or template to declare that it **extends** a base definition, inheriting its structure and overriding specific parts.

### 07B.4.1 Base Definitions and Derived Definitions

A base surface definition is a normal surface that can be used standalone. A derived surface declares `extends` pointing to a base:

```json
{
  "surface_id": "finance.invoice.kenya",
  "extends": "finance.invoice.base",
  "overrides": {
    "header.etims_section": {
      "type": "container",
      "children": [
        { "type": "image", "src": "{{invoice.etims_qr_url}}", "alt": "KRA eTIMS QR Code" },
        { "type": "tpl", "tpl": "CU Invoice No: {{invoice.etims_cu_invoice_number}}" }
      ]
    }
  }
}
```

The compilation pipeline resolves inheritance before slot resolution, producing a merged AST that contains all base nodes plus the derived overrides.

### 07B.4.2 Override Semantics (Replace vs. Merge vs. Extend)

Each override entry carries a `strategy` field controlling how the base node is modified:

| Strategy | Behaviour |
|----------|-----------|
| `replace` | Base node discarded; override node used verbatim. Default. |
| `merge` | Props of base node and override are merged (override wins on conflict). Children from base are kept unless override provides children. |
| `extend` | Base node kept; override's children are **appended** after base children. |
| `prepend` | Override's children prepended before base children. Base props unchanged. |

```json
{
  "overrides": {
    "shared.page-actions": {
      "strategy": "extend",
      "children": [
        {
          "type": "action",
          "label": "Submit to KRA eTIMS",
          "actionType": "ajax",
          "api": "POST /api/v1/finance/invoices/${id}/etims-submit"
        }
      ]
    }
  }
}
```

This `extend` strategy adds the "Submit to KRA eTIMS" button to the existing page actions bar without replacing the standard Save/Cancel buttons that the base surface already defines.

### 07B.4.3 Sealed vs. Open Templates

Templates marked `Sealed: true` in their `UITemplate` definition cannot be extended or have their named slots overridden by tenants. Sealed templates are appropriate for:

- Security-sensitive surfaces (permission grant screens, platform billing)
- Legally required layouts (KRA eTIMS invoice format mandates from KRA)
- Platform-internal admin surfaces

```go
reg.Register(&ui.UITemplate{
    Key:    "platform.tenant-admin.billing",
    Sealed: true,  // Tenants cannot extend or override slots in this template
    // ...
})
```

When a tenant surface definition attempts to extend a sealed template, compilation fails with:
```
compilation error: surface "acme.billing-override" attempts to extend sealed template "platform.tenant-admin.billing"
```

---

## 07B.5 Mixin System

Mixins are reusable bundles of **props and behaviours** (not structural AST nodes) that can be applied to any node in a surface definition. Where templates add structure, mixins add cross-cutting concerns.

### 07B.5.1 Declaring Mixins

Mixins are declared as `UIMixin` structs:

```go
// internal/ui/mixins/registry.go

type UIMixin struct {
    Key         string
    Description string
    // Props to merge onto the target node
    Props       map[string]any
    // Actions to append to the target node's actions
    Actions     []ast.Action
    // CSS class names to add to the target node
    ClassNames  []string
}
```

Example — a mixin that adds standard audit fields visibility to any form:

```go
&ui.UIMixin{
    Key:         "platform.audit-fields",
    Description: "Appends created_by, created_at, updated_by, updated_at display fields to a form",
    Props:       map[string]any{"className": "has-audit-fields"},
    Actions:     nil,
    // These children get appended to the target form's field list
    // (via a special 'children_append' mechanism)
}
```

### 07B.5.2 Applying Mixins to Nodes

Mixins are applied via the `mixins` prop on any AST node:

```json
{
  "type": "form",
  "api": "POST /api/v1/finance/invoices",
  "mixins": ["platform.audit-fields", "finance.kenya-tax-fields"],
  "body": [
    { "type": "input-text", "name": "invoice_number", "label": "Invoice Number" }
  ]
}
```

During compilation, each listed mixin is resolved and its contributions merged onto the node:

```go
func (c *Compiler) applyMixins(node ast.Node, ctx *CompileContext) (ast.Node, error) {
    mixinKeys, _ := node.Props["mixins"].([]string)
    for _, key := range mixinKeys {
        mixin, ok := c.mixinRegistry.Lookup(key)
        if !ok {
            return node, fmt.Errorf("unknown mixin %q in surface %q node type %q", key, ctx.SurfaceID, node.Type)
        }
        node = mergeMixin(node, mixin)
    }
    // Remove the 'mixins' prop from the final AST — renderers never see it
    delete(node.Props, "mixins")
    return node, nil
}
```

### 07B.5.3 Mixin Conflict Resolution

When two mixins both contribute values for the same prop, the conflict is resolved using the following rules:

1. **String props:** Last mixin wins (append order matters; document your mixin order).
2. **Array props (e.g. `classNames`, `actions`):** Arrays are concatenated. Duplicates are removed.
3. **Boolean props:** Last mixin wins.
4. **Object props:** Deep merge; last mixin wins on leaf conflicts.

> **✅ Convention:** If two mixins both set the same string prop, treat it as a bug. File an issue against the mixin that should defer. Use the `mixin_priority` override field to explicitly state which mixin wins when conflict is expected and acceptable.

---

## 07B.6 Fragment Reuse (Partial ASTs)

Fragments are **saved partial AST definitions** stored in the database per tenant. Unlike templates (which are code-defined, versioned Go structs), fragments are user-created and can be composed at the surface level by tenant admins through the AwoERP UI builder.

A fragment is stored in the `ui_fragments` table:

```sql
CREATE TABLE ui_fragments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    key         TEXT NOT NULL,
    description TEXT,
    body        JSONB NOT NULL,          -- The partial AST
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, key)
);
```

Fragments are referenced in surface definitions using the `fragment_ref` node type:

```json
{
  "type": "fragment_ref",
  "fragment": "acme.mpesa-payment-block",
  "params": {
    "amount_field": "total_due",
    "currency": "KES"
  }
}
```

During compilation, the `UICompilationService` loads all fragments for the current tenant at the start of compilation (single SQL query, not per-reference), then resolves `fragment_ref` nodes from the in-memory map. This keeps compilation latency low even when a surface uses dozens of fragments.

```go
func (s *UICompilationService) Compile(ctx context.Context, surfaceID string, viewCtx ViewContext) (*ast.CompiledSurface, error) {
    // Pre-load all tenant fragments once
    fragments, err := s.fragmentRepo.LoadAllForTenant(ctx, viewCtx.TenantID)
    if err != nil {
        return nil, fmt.Errorf("load fragments: %w", err)
    }

    compileCtx := &CompileContext{
        ViewContext: viewCtx,
        Fragments:   fragments,
    }

    return s.compiler.Compile(surface, compileCtx)
}
```

> **ℹ Note:** Fragments are tenant-scoped — a fragment created by tenant A is never accessible to tenant B. The `tenant_id` column is enforced by PostgreSQL RLS policy on the `ui_fragments` table.

---

## 07B.7 Composition Anti-Patterns

The following patterns are explicitly prohibited and caught by the `awo ui validate` CLI tool:

### Anti-Pattern 1: Deep Template Chains (> 4 levels)

A template that extends a template that references a template is difficult to reason about and dramatically slows compilation. Keep composition depth to 3 levels or fewer for templates. The hard limit of 8 exists as a safety net, not a target.

### Anti-Pattern 2: Circular References

Surface A extends Surface B, which references a fragment that includes Surface A. The compiler detects cycles using DFS with a visited set and fails compilation with the full cycle path:

```
composition cycle detected: finance.invoice.base → finance.shared-header → fragment:acme.header-v2 → finance.invoice.base
```

### Anti-Pattern 3: Template Parameters for Structural Variation

If a template accepts a boolean parameter that completely changes its child structure (e.g. `show_full_layout: true` renders 20 nodes, `false` renders 3 entirely different nodes), you have two templates that should be separate declarations. Parameters should control **values** (colors, labels, counts), not **structure** (which nodes exist).

### Anti-Pattern 4: Slot Fill with Unvalidated User Input

Slot fills can contain arbitrary AST nodes. Never allow portal users (supplier, customer, employee) to provide raw AST as slot fills. Portal users can only fill slots with pre-approved **widget types** from an allow-list:

```go
var portalAllowedSlotFillTypes = map[string]bool{
    "tpl":        true,
    "image":      true,
    "plain-text": true,
}
```

Tenant admins can fill slots with a broader set. Only platform engineers can fill slots with arbitrary nodes.

### Anti-Pattern 5: `map[string]any` in Go Surface Definitions

Surface definition code in Go **must not** use `map[string]any`. All node props must be typed structs. This is enforced by CI:

```yaml
# .github/workflows/ui-lint.yml
- name: No map[string]any in UI surface definitions
  run: grep -r 'map\[string\]any' internal/ui/surfaces/ && exit 1 || exit 0
```

Using `map[string]any` bypasses the type system and makes refactoring impossible. Use `ast.Props` or concrete prop structs.

### Anti-Pattern 6: Composition at the JSON Layer (String Interpolation)

Never construct surface JSON by string concatenation or `text/template` rendering of JSON blobs. This produces invalid JSON on edge-case inputs and is a XSS vector if user data is interpolated into JSON string. All composition happens at the typed AST level.

---

## 07B.8 Composition and Multi-Tenancy

Multi-tenancy adds a third axis to composition: in addition to module-level templates and surface-level definitions, tenants can provide their own overrides. The composition pipeline processes these in the correct order:

```
Platform base templates
  → Module surface definitions (use templates + slots)
    → Tenant surface overrides (extend/slot-fill from tenant DB)
      → User surface personalizations (if permitted)
        → Compiled AST (sent to renderer)
```

Each layer is isolated by PostgreSQL RLS. The compilation service runs under a tenant-scoped database connection established via `store.WithTenant(ctx, viewCtx.TenantID, ...)`. Tenant override data is never visible to other tenants.

### Tenant Override Storage

Tenant surface overrides are stored in the `tenant_surface_overrides` table:

```sql
CREATE TABLE tenant_surface_overrides (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    surface_id  TEXT NOT NULL,              -- e.g. "finance.invoice.view"
    override    JSONB NOT NULL,             -- Partial override descriptor
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, surface_id)
);
```

### Module Enablement and Composition Reduction

When a tenant does not have a module enabled (e.g. the Fuel Retail module is disabled for a tenant that only uses Finance), all templates, slots, and fragments from that module are **unavailable at compile time**. The compiler's `ModuleFilter` removes references to disabled-module templates before resolution, producing a compilation error if a surface in an enabled module depends on a template from a disabled module:

```
compilation error: surface "finance.invoice.view" references template "fuel.etims-station-details" but module "fuel" is not enabled for tenant "ACME Ltd" (tenant_id: 3f2e...)
```

This is a fast-fail at compile time — not a runtime rendering error — ensuring the compilation cache is only populated with valid, fully-resolvable surface definitions.

### Tenant-Specific Fragment Isolation

Fragments are the primary mechanism for tenant customisation of shared templates. Tenants cannot modify platform templates (those are sealed or require a code PR), but they can:

1. Create fragments that fill defined slots in platform templates
2. Create derived surface definitions that `extend` platform surfaces
3. Override specific AST paths using the `overrides` map

The fragment system is deliberately tenant-constrained. Even a tenant admin with the `ui.surfaces.admin` permission cannot create fragments that reference nodes from other tenants, cannot fill slots with server-side script execution nodes, and cannot produce fragments larger than 50KB (enforced at API ingestion time).

> See Chapter 30 — Customization Framework for the full tenant customisation capability matrix and permission model.

> See Chapter 44 — Migration Strategy for how tenant compositions are migrated when platform templates are upgraded.
