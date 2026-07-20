> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 23 — Validation Framework

| | |
|---|---|
| **Volume** | 05 — Runtime |
| **Chapter** | 23 |
| **Status** | Implemented |
| **Source** | `internal/web/dsl/ast/` (Validate), `internal/web/dsl/pipeline/` (ValidateStage, NormalizeStage) |

---

## Table of Contents

1. [23.1 Two Validation Points](#231-two-validation-points)
2. [23.2 AST Validation](#232-ast-validation)
3. [23.3 ValidateStage Rules](#233-validatestage-rules)
4. [23.4 NormalizeStage vs ValidateStage Boundary](#234-normalizestage-vs-validatestage-boundary)
5. [23.5 Error Code Prefixes](#235-error-code-prefixes)
6. [23.6 Form-Level Validation](#236-form-level-validation)

---

## 23.1 Two Validation Points

Schema validation happens at two distinct points in the pipeline, with different
responsibilities and timing:

```
ASTFn called
  ↓
CompileTree
  ├─ node.Validate()  ← AST validation (structural, before compile)
  └─ node.Compile()   ← produces JSON schema fragment
  ↓
Pipeline stages (in priority order)
  ├─ NormalizeStage  (priority 50) — canonicalise, never errors
  └─ ValidateStage   (priority 70) — enforce rules, returns BusinessError
  ↓
CacheStage          (priority 90) — store compiled schema
  ↓
Schema served to browser
```

| Stage | When | Can error? | On cache hit? |
|---|---|---|---|
| `node.Validate()` | Before compile, always | Yes | Yes (compile skipped on cache, but Validate is part of compile) |
| `ValidateStage` | After compile, in pipeline | Yes | Security rules always run; structural rules skipped if `DataKeyASTCompiled=true` |

---

## 23.2 AST Validation

Every node type implements a `Validate() error` method. `CompileTree` calls
`Validate()` before `Compile()` for each node in the tree. A validation error
from any node aborts the compile and returns an error to the pipeline.

### Error codes

AST validation errors use the `AST_` prefix:

| Code | Meaning |
|---|---|
| `AST_MISSING_FIELD` | Required field is empty (e.g. `ActionNode.Label == ""`) |
| `AST_INVALID_TYPE` | Field value is not in the allowed set |
| `AST_MISSING_DIALOG` | ActionType is "dialog" but Dialog is nil |
| `AST_MISSING_DRAWER` | ActionType is "drawer" but Drawer is nil |
| `AST_MISSING_API` | ActionType is "ajax" but API is nil |
| `AST_INVALID_METHOD` | API.Method is not GET/POST/PUT/PATCH/DELETE |

### Example: ActionNode.Validate()

```go
func (a ActionNode) Validate() error {
    if a.Label == "" {
        return NewASTError("AST_MISSING_FIELD", "ActionNode.Label is required")
    }
    validTypes := map[string]bool{
        "ajax": true, "url": true, "dialog": true,
        "drawer": true, "reload": true, "link": true,
    }
    if !validTypes[a.ActionType] {
        return NewASTError("AST_INVALID_TYPE",
            fmt.Sprintf("ActionNode.ActionType %q is not valid", a.ActionType))
    }
    if a.ActionType == "dialog" && a.Dialog == nil {
        return NewASTError("AST_MISSING_DIALOG",
            "ActionNode.Dialog is required when ActionType is 'dialog'")
    }
    if a.ActionType == "ajax" && a.API == nil {
        return NewASTError("AST_MISSING_API",
            "ActionNode.API is required when ActionType is 'ajax'")
    }
    return nil
}
```

AST validation catches programmer errors at schema build time — before the
schema ever reaches a user's browser.

---

## 23.3 ValidateStage Rules

`ValidateStage` runs in the pipeline at priority 70 (after NormalizeStage at 50,
before CacheStage at 90). It has two categories of rules:

### Structural rules (skipped on cache hit)

These rules verify the structural correctness of the compiled schema. They are
skipped when `DataKeyASTCompiled=true` is set in the pipeline context, meaning
the schema was already validated during a previous compile and the result is
being served from cache.

| Rule | What it checks |
|---|---|
| CRUDNode `syncLocation` | Every CRUDNode must have `syncLocation: true`. Without it, AMIS does not synchronise filter/pagination state to the URL, breaking the back button. |
| ChartNode transparent background | Every ChartNode must have `style.background: "transparent"`. Opaque chart backgrounds look broken in dark mode. |
| API method prefix | API `method` field must be an uppercase HTTP verb (GET, POST, PUT, PATCH, DELETE). Normalise was supposed to handle this but Validate catches any that slipped through. |

#### CRUDNode syncLocation example

Non-compliant (ValidateStage rejects):

```json
{
  "type": "crud",
  "api": "/api/v1/finance/invoices"
}
```

Compliant:

```json
{
  "type": "crud",
  "api": "/api/v1/finance/invoices",
  "syncLocation": true
}
```

Error returned: `VALIDATE_CRUD_SYNC_LOCATION` with message "CRUDNode at path
`body[0]` is missing `syncLocation: true`."

#### ChartNode background example

Non-compliant:

```json
{
  "type": "chart",
  "source": { "chart": { /* ... */ } }
}
```

Compliant:

```json
{
  "type": "chart",
  "source": { "chart": { /* ... */ } },
  "style": { "background": "transparent" }
}
```

Error returned: `VALIDATE_CHART_BG` with message "ChartNode at path `body[1]`
must have `style.background: transparent`."

### Security rules (always run, even on cache hit)

Security rules are enforced on every schema request — they are **not** skipped
when serving from cache. This ensures that even a cached schema cannot bypass
security validation.

| Rule | What it checks | Error code |
|---|---|---|
| No permission strings in expressions | `visibleOn` and `disabledOn` must not contain permission strings like `"finance:invoices:approve"`. Only `can_*` booleans are permitted. | `VALIDATE_PERM_EXPRESSION` |
| No IAM endpoint calls | No `APISpec.URL` in the compiled schema may call IAM-internal endpoints (e.g. `/api/v1/iam/roles`, `/api/v1/iam/policies`). | `VALIDATE_IAM_ENDPOINT` |

#### Permission expression check

The ValidateStage walks all `visibleOn` and `disabledOn` expression strings in
the compiled schema and rejects any that match the permission string pattern
`\w+:\w+(:\w+)+` (colon-separated segments):

Non-compliant (rejected):

```json
{ "visibleOn": "${permissions.includes('finance:invoices:approve')}" }
```

Compliant:

```json
{ "visibleOn": "${can_approve}" }
```

This prevents the browser from receiving raw permission identifiers that could be
used to infer the IAM schema.

#### IAM endpoint check

The ValidateStage checks all `api.url` fields in the compiled schema. Any URL
that starts with `/api/v1/iam/` and is not on an explicit allowlist is rejected:

```
VALIDATE_IAM_ENDPOINT: API URL "/api/v1/iam/roles" is not permitted in UI schemas.
Use the pre-resolved can_* scope variables instead.
```

---

## 23.4 NormalizeStage vs ValidateStage Boundary

| Concern | NormalizeStage (priority 50) | ValidateStage (priority 70) |
|---|---|---|
| Purpose | Canonicalise the schema | Enforce invariants |
| Can return error? | Never — always succeeds | Yes — returns `*BusinessError` |
| Example: `type` field casing | Lowercases `"Form"` → `"form"` | n/a (already normalised) |
| Example: API URL whitespace | Trims leading/trailing whitespace | n/a (already trimmed) |
| Example: `syncLocation` | n/a | Rejects CRUDNode without it |
| Example: `"disabledOn": "${!finance:approve}"` | n/a | Rejects with `VALIDATE_PERM_EXPRESSION` |

The rule of thumb: if the transformation is purely cosmetic and always safe
(lowercasing a string, trimming whitespace), it belongs in NormalizeStage. If
it requires a policy decision that might fail (enforcing a required field, a
security invariant), it belongs in ValidateStage.

NormalizeStage runs first so that ValidateStage always sees a clean, canonical
schema. ValidateStage does not need to handle mixed-case `type` values.

---

## 23.5 Error Code Prefixes

Pipeline and AST errors use consistent prefixes:

| Prefix | Origin | Example |
|---|---|---|
| `AST_*` | `node.Validate()` in the AST layer | `AST_MISSING_FIELD` |
| `VALIDATE_*` | `ValidateStage` | `VALIDATE_CRUD_SYNC_LOCATION` |
| `REGISTRY_*` | Registry lookup (route not found, duplicate registration) | `REGISTRY_NOT_FOUND` |
| `DSL_*` | DSL compilation errors (ASTFn panicked, node type unknown) | `DSL_COMPILE_ERROR` |
| `CACHE_*` | Cache layer errors (serialisation failure, store unavailable) | `CACHE_WRITE_ERROR` |

All pipeline errors are wrapped in `*BusinessError` from
`internal/shared/errors/` and map to HTTP 422 (Unprocessable Entity) when the
schema is served. The response body contains:

```json
{
  "success": false,
  "code": "VALIDATE_CRUD_SYNC_LOCATION",
  "message": "CRUDNode at path body[0] is missing syncLocation: true"
}
```

> **Schema versioning and ETags** are planned but not implemented. When
> implemented, cached schemas would carry an ETag header so browsers can send
> `If-None-Match` and receive a `304 Not Modified` response without re-running
> the pipeline.

---

## 23.6 Form-Level Validation

Client-side form validation is performed by AMIS using the `required` and
`validations` fields on form input nodes. This is **not** part of the pipeline
ValidateStage — it is AMIS runtime behaviour in the browser.

### Required fields

```json
{
  "type": "input-text",
  "name": "invoice_number",
  "label": "Invoice Number",
  "required": true
}
```

AMIS prevents form submission if a required field is empty, showing an inline
error message.

### Pattern validation

```json
{
  "type": "input-text",
  "name": "tax_id",
  "label": "Tax ID",
  "validations": {
    "matchRegexp": "/^[A-Z0-9]{8,12}$/"
  },
  "validationErrors": {
    "matchRegexp": "Tax ID must be 8–12 uppercase alphanumeric characters"
  }
}
```

### Min/max for numbers

```json
{
  "type": "input-number",
  "name": "amount",
  "label": "Amount",
  "min": 0.01,
  "validations": {
    "minimum": 0.01
  }
}
```

### Server-side validation

Client-side validation is a user experience aid only. The server always
re-validates submitted data. If the server returns `{ "status": 1, "msg": "..." }`,
AMIS shows the message as a form-level error banner above the submit button.

Field-level server validation errors can be returned in the `errors` field:

```json
{
  "status": 1,
  "msg": "Validation failed",
  "errors": {
    "invoice_number": "This invoice number already exists",
    "due_date": "Due date cannot be before issue date"
  }
}
```

AMIS maps each key in `errors` to the corresponding form field and shows the
message inline beneath the field.
