> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Error Catalog
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, frontend-engineer]
related:
  - "[Error Handling Overview](01-error-handling-overview.md)"
  - "[API Design Overview](../14-api-design/01-api-design-overview.md)"
  - "[API Reference Overview](../../../08-api-reference/01-api-reference-overview.md)"
---

# Error Catalog

All error codes returned by AwoERP API endpoints. Frontend uses `error.code` to display localized messages.

## Authentication Errors (401, 403)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `UNAUTHORIZED` | 401 | Missing or expired session token | Redirect to login |
| `FORBIDDEN` | 403 | Valid session but insufficient permission | Show "access denied" message |
| `TENANT_INACTIVE` | 403 | Tenant account suspended or archived | Show tenant status page |

## Validation Errors (400, 422)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `BAD_REQUEST` | 400 | Malformed JSON or invalid UUID format | Check request format |
| `VALIDATION_ERROR` | 422 | One or more fields failed validation | Show field-level errors from `details` |
| `INVALID_DATE_RANGE` | 422 | End date before start date | Correct dates |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | Content-Type not application/json | Fix client headers |

## Conflict Errors (409)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `CONFLICT` | 409 | Version mismatch (optimistic lock) | Re-fetch record and retry |
| `DUPLICATE_CONTRACT_NUMBER` | 409 | Contract number already exists in tenant | Change contract number |
| `DUPLICATE_ACCOUNT_CODE` | 409 | Finance account code already exists | Change account code |
| `DUPLICATE_EMAIL` | 409 | User email already exists in tenant | Use different email |

## Business Rule Errors (422)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `CONTRACT_NOT_EDITABLE` | 422 | Contract not in draft status | Check contract status |
| `INVALID_STATUS_TRANSITION` | 422 | Cannot transition from current to target status | Show allowed transitions |
| `CONTRACT_VALUE_EXCEEDS_LIMIT` | 422 | Total value above tenant limit | Adjust value or contact admin |
| `UNBALANCED_JOURNAL_ENTRY` | 422 | Debits ≠ credits in transaction | Fix entry amounts |
| `FEATURE_NOT_ENABLED` | 403 | Feature flag disabled for this tenant | Contact tenant admin |

## Not Found Errors (404)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `NOT_FOUND` | 404 | Resource doesn't exist or belongs to another tenant | Show "not found" |

Note: 404 is returned for both "truly not found" and "exists but different tenant" — this is intentional (no tenant enumeration).

## Rate Limit Errors (429)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded | Back off and retry after `Retry-After` header |

## Server Errors (500)

| Code | HTTP | Meaning | Action |
|------|------|---------|--------|
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error | Log request ID, contact support |

Internal errors never expose details (no SQL, no stack traces) to clients.

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "contract_number", "message": "contract_number is required" },
      { "field": "total_value", "message": "must be greater than 0" }
    ]
  }
}
```

`details` is present only for `VALIDATION_ERROR`. All other error codes have an empty or omitted `details` array.

## Frontend Error Handling Template

```javascript
async function apiCall(url, options) {
    const resp = await fetch(url, options);
    if (!resp.ok) {
        const body = await resp.json();
        const code = body.error?.code;

        switch (code) {
            case 'UNAUTHORIZED':
                window.location = '/login';
                break;
            case 'CONFLICT':
                showToast('warning', 'Record was updated by another user. Please refresh.');
                break;
            case 'VALIDATION_ERROR':
                showFieldErrors(body.error.details);
                break;
            case 'FEATURE_NOT_ENABLED':
                showToast('info', 'This feature is not available for your account.');
                break;
            default:
                showToast('error', body.error?.message || 'An error occurred.');
        }
        throw new Error(code);
    }
    return resp.json();
}
```
