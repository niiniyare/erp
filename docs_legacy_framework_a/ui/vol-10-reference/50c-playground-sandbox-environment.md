> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Playground and Sandbox Environment"
volume: "X — Reference & Guidance"
chapter: "50-C"
phase: 4
status: draft
audience: [All Engineers, ERP Implementers]
---

# Chapter 50-C — Playground and Sandbox Environment

> **Volume:** X — Reference & Guidance
> **Audience:** All Engineers, ERP Implementers
> **Prerequisites:** Chapter 07-B — UI Composition Patterns, Chapter 40-B — Developer Tooling
> **Phase:** 4

---

## 50C.1 Playground Philosophy

The AwoERP playground is a browser-based environment for experimenting with SDUI surface definitions without affecting any real tenant data. It is the fastest path from "I want to try a new surface layout" to "I can see it rendered with realistic data". The playground is used by:

- **Platform engineers** testing new component patterns before formalising them in the platform
- **Backend engineers** validating that a new API response shape renders correctly in the amis renderer
- **ERP implementers** building tenant-specific surface customisations during the implementation phase
- **Sales engineers** demonstrating AwoERP capabilities to prospects using realistic sample data

The playground follows the same compilation and rendering pipeline as production. A surface definition that renders in the playground will render identically in production. The only difference is that the playground uses sandboxed tenant context, mock data, and a sandboxed permission system.

---

## 50C.2 Playground Architecture

### 50C.2.1 Sandboxed Tenant Context

Every playground session operates within an automatically provisioned **sandbox tenant**. Sandbox tenants are real tenant records in the database with real RLS enforcement — they are not special-cased. What makes them sandboxes:

1. They are provisioned from a template (e.g. "Kenya Fuel Station Demo" or "Generic SME Demo") with pre-loaded sample data
2. They are owned by a playground session, not a real business
3. They are automatically destroyed after a configurable TTL (default: 24 hours for anonymous sessions, 7 days for authenticated sessions)
4. All modules are enabled by default, so implementers can explore the full platform
5. Writes to the sandbox tenant are allowed (you can create test invoices, employees, etc.) but are scoped to the sandbox

```go
// internal/core/playground/service.go

type PlaygroundService struct {
    tenantService  TenantService
    templateRepo   PlaygroundTemplateRepository
    cache          cache.Service
}

type SandboxSession struct {
    SessionID   uuid.UUID
    TenantID    uuid.UUID    // The sandbox tenant
    Template    string       // "fuel_station" | "generic_sme" | "blank"
    CreatedAt   time.Time
    ExpiresAt   time.Time
    AuthContext *PlaygroundUser  // nil for anonymous
}

func (s *PlaygroundService) CreateSandbox(ctx context.Context, req CreateSandboxRequest) (*SandboxSession, error) {
    // Provision a new tenant from the selected template
    tenant, err := s.tenantService.ProvisionFromTemplate(ctx, req.Template)
    if err != nil {
        return nil, fmt.Errorf("provision sandbox: %w", err)
    }

    session := &SandboxSession{
        SessionID: uuid.New(),
        TenantID:  tenant.ID,
        Template:  req.Template,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(sandboxTTL(req.AuthContext)),
    }

    // Store session in Redis
    return session, s.cache.Set(ctx,
        fmt.Sprintf("sandbox:%s", session.SessionID),
        session,
        time.Until(session.ExpiresAt),
    )
}
```

### 50C.2.2 Live amis JSON Editor with Preview Pane

The playground UI is a split-pane interface: JSON editor on the left, live amis preview on the right. The editor is powered by Monaco Editor (the VS Code editor engine) with the AwoERP JSON schema for autocomplete and validation.

```javascript
// web/src/playground/editor.js

function initPlayground(sandboxSession) {
    var editor = monaco.editor.create(document.getElementById('editor'), {
        value:    JSON.stringify(defaultSurface, null, 2),
        language: 'json',
        theme:    'vs-dark',
        automaticLayout: true,
    });

    // Configure JSON schema for autocomplete
    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: true,
        schemas: [{
            uri: '/api/v1/ui/schema-registry/surface.schema.json',
            fileMatch: ['*'],
            schema: surfaceSchema
        }]
    });

    // Debounced live preview update
    var updatePreview = debounce(function() {
        var json;
        try {
            json = JSON.parse(editor.getValue());
        } catch (e) {
            showEditorError('Invalid JSON: ' + e.message);
            return;
        }
        renderPreview(json, sandboxSession);
    }, 500);

    editor.onDidChangeModelContent(updatePreview);
}

function renderPreview(schema, session) {
    // Re-render the amis preview pane with the new schema
    if (window._amisInstance) {
        window._amisInstance.updateSchema(schema);
    } else {
        window._amisInstance = amis.embed(
            document.getElementById('preview'),
            schema,
            buildEnv(session)
        );
    }
}
```

The preview pane uses the same `env.fetcher` as production, but all API requests are intercepted by the playground middleware and routed to the sandbox tenant's data:

```go
// internal/api/handlers/playground/middleware.go

func PlaygroundMiddleware(playgroundService PlaygroundService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sessionID := c.Get("X-Playground-Session")
        if sessionID == "" {
            return c.Next()
        }
        session, err := playgroundService.GetSession(c.UserContext(), uuid.MustParse(sessionID))
        if err != nil {
            return fiber.ErrUnauthorized
        }
        // Inject sandbox tenant ID into context
        ctx := shared.WithTenantID(c.UserContext(), session.TenantID)
        c.SetUserContext(ctx)
        return c.Next()
    }
}
```

### 50C.2.3 Permission Simulator

The permission simulator allows the user to select a mock actor context and see how the rendered surface changes:

```json
{
  "type": "form",
  "title": "Simulate Actor Context",
  "className": "playground-permission-sim",
  "body": [
    { "type": "select", "name": "actor_type", "label": "Actor Type",
      "options": [
        { "label": "Tenant Staff — Admin", "value": "tenant_staff:admin" },
        { "label": "Tenant Staff — Finance Manager", "value": "tenant_staff:finance_manager" },
        { "label": "Tenant Staff — Pump Attendant", "value": "tenant_staff:pump_attendant" },
        { "label": "Portal — Supplier", "value": "portal_supplier" },
        { "label": "Portal — Customer", "value": "portal_customer" },
        { "label": "Portal — Employee", "value": "portal_employee" }
      ]
    },
    { "type": "checkboxes", "name": "custom_permissions", "label": "Custom Permissions (optional)",
      "source": "GET /api/v1/ui/emulator/permissions",
      "visibleOn": "actor_type === 'custom'" }
  ],
  "onSubmit": { "actionType": "custom", "script": "updatePlaygroundContext(form.data)" }
}
```

When the actor context changes, the playground recompiles the surface definition via `POST /api/v1/ui/playground/compile` and refreshes the preview. The diff between the previous and new compilations is shown in the editor pane as highlighted changes.

### 50C.2.4 Feature Flag Toggle Panel

A collapsible panel shows all feature flags available for the sandbox tenant with toggle switches:

```json
{
  "type": "panel",
  "title": "Feature Flags",
  "collapsible": true,
  "body": {
    "type": "crud",
    "api": "GET /api/v1/ui/emulator/flags",
    "columns": [
      { "name": "key", "label": "Flag" },
      { "name": "description", "label": "Description" },
      { "name": "enabled", "label": "Enabled", "type": "switch",
        "quickEdit": {
          "type": "switch",
          "saveImmediately": true,
          "api": "PATCH /api/v1/ui/emulator/flags/${key}"
        }
      }
    ]
  }
}
```

Toggling a flag in the playground immediately invalidates the navigation cache and surface compilation cache for the sandbox tenant, and triggers a preview refresh.

---

## 50C.3 Shareable Playground Links

Playground sessions can be saved and shared as URLs. The sharing mechanism encodes the current surface definition, actor context, and flag overrides as a compressed, base64-encoded URL parameter:

```javascript
// web/src/playground/share.js

async function generateShareLink() {
    var state = {
        schema:   JSON.parse(editor.getValue()),
        context:  currentActorContext,
        flags:    currentFlagOverrides,
        template: sandboxTemplate,
        version:  '1'
    };

    // Compress with LZ-string (browser-side compression)
    var compressed = LZString.compressToEncodedURIComponent(JSON.stringify(state));

    // For long schemas, save to server and use a short ID
    if (compressed.length > 2000) {
        var response = await api.post('/api/v1/ui/playground/share', { state: compressed });
        return window.location.origin + '/playground?s=' + response.data.share_id;
    }

    return window.location.origin + '/playground?state=' + compressed;
}
```

Server-side share IDs are stored in Redis with a 30-day TTL. Anonymous share links do not require authentication to view; authenticated share links can be restricted to specific domains or users.

```go
// internal/api/handlers/playground/share.go

func (h *PlaygroundHandler) SaveShare(c *fiber.Ctx) error {
    var body struct {
        State string `json:"state"`
    }
    if err := c.BodyParser(&body); err != nil {
        return fiber.ErrBadRequest
    }

    shareID := generateShareID()
    err := h.cache.Set(c.UserContext(),
        fmt.Sprintf("playground:share:%s", shareID),
        body.State,
        30*24*time.Hour,
    )
    if err != nil {
        return err
    }

    return c.JSON(fiber.Map{"share_id": shareID})
}
```

---

## 50C.4 Playground Limitations and Security Boundaries

The playground is a powerful tool, and its security boundaries are carefully designed to prevent it from becoming an attack vector:

### Data Isolation
- Sandbox tenants are real RLS-scoped tenants. A playground session cannot read data from real tenants.
- Sandbox tenants created from templates contain only synthetic sample data — no real business data is ever used as sandbox seed data.
- The sandbox tenant is destroyed (and all its data deleted) when the session expires.

### Schema Execution Safety
- Surface definitions submitted to the playground are validated by the same `awo ui validate` pipeline as production definitions before compilation.
- Schemas cannot contain server-side script execution nodes — only the standard amis component set is available.
- Custom amis renderer plugins (`amis.registerRendererPlugin`) cannot be registered via the playground.

### API Access
- The playground `env.fetcher` is scoped to the sandbox tenant. Requests to `/api/v1/platform/...` (platform admin endpoints) are rejected with 403 even if the user somehow obtains a platform admin token.
- Rate limiting: 100 surface compilations per session per hour; 50 API requests per minute per sandbox session.

### What the Playground Cannot Do
- Access real tenant data
- Permanently modify platform templates or component library
- Execute arbitrary Go code
- Access external networks (all external URL references in schemas are neutralised in the sandbox)
- Simulate platform admin capabilities

> **✅ Convention:** Do not use the playground for testing changes to sealed templates or platform-level components. Those changes must go through the standard PR → review → CI → staging → production pipeline. The playground is for surface-level composition and amis component usage experimentation only.

> See Chapter 40-B — Developer Tooling for the `awo ui sandbox` CLI command that integrates the playground into the local development workflow.
